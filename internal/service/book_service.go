package service

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"bookstore-api/internal/model"
	"bookstore-api/internal/repository"
)

type BookService interface {
	GetAll(page, limit int) ([]model.Book, int, error) // [NEW] Pagination signature update
	GetByID(id string) (*model.Book, error)
	Create(book model.Book) (*model.Book, error)
	Update(id string, book model.Book) (*model.Book, error)
	Delete(id string) error
	// Batch Operations
	CreateBatch(books []model.Book) ([]model.Book, error)
	DeleteBatch(ids []string) error
}

type bookService struct {
	repo repository.BookRepository
}

func NewBookService(repo repository.BookRepository) BookService {
	return &bookService{repo: repo}
}

func (s *bookService) GetAll(page, limit int) ([]model.Book, int, error) {
	// [NEW] Defaults and safety bounds for unbounded queries
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	allBooks, err := s.repo.GetAll()
	if err != nil {
		return nil, 0, err
	}

	// [NEW] Soft Delete Filter: Exclude records where DeletedAt != nil
	var activeBooks []model.Book
	for _, b := range allBooks {
		if b.DeletedAt == nil {
			activeBooks = append(activeBooks, b)
		}
	}

	totalItems := len(activeBooks)
	startIndex := (page - 1) * limit
	if startIndex >= totalItems {
		return []model.Book{}, totalItems, nil
	}

	endIndex := startIndex + limit
	if endIndex > totalItems {
		endIndex = totalItems
	}

	return activeBooks[startIndex:endIndex], totalItems, nil
}

func (s *bookService) GetByID(id string) (*model.Book, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid UUID format")
	}

	book, err := s.repo.GetByID(id)
	if err != nil || book == nil {
		return nil, errors.New("book not found")
	}

	// [NEW] Soft Delete Filter: Rejection check for soft-deleted items
	if book.DeletedAt != nil {
		return nil, errors.New("book not found")
	}
	if book.Version == 0 {
		book.Version = 1
	}

	return book, nil
}

func (s *bookService) Create(b model.Book) (*model.Book, error) {
	// [NEW EDGE CASE] Sanitize string inputs prior to business processing
	b.Title = strings.TrimSpace(b.Title)
	b.Author = strings.TrimSpace(b.Author)

	// Edge Case 1: Check for duplicate Title + Author
	existingBooks, err := s.repo.GetAll()
	if err == nil {
		for _, existing := range existingBooks {
			// [NEW] Exclude soft-deleted books when checking duplicates
			if existing.DeletedAt == nil && strings.EqualFold(existing.Title, b.Title) && strings.EqualFold(existing.Author, b.Author) {
				return nil, errors.New("a book with the same title and author already exists")
			}
		}
	}

	b.ID = uuid.New().String()
	b.Version = 1 // [NEW] Set initial optimistic locking version
	now := time.Now()
	b.CreatedAt = now
	b.UpdatedAt = now

	return s.repo.Create(b)
}

func (s *bookService) Update(id string, b model.Book) (*model.Book, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid UUID format")
	}

	b.Title = strings.TrimSpace(b.Title)
	b.Author = strings.TrimSpace(b.Author)

	existing, err := s.GetByID(id) // GetByID filters out soft-deleted records automatically
	if err != nil || existing == nil {
		return nil, errors.New("book not found")
	}

	// [NEW] Optimistic Locking Check: Compare incoming version against store version
	if existing.Version != b.Version {
		return nil, errors.New("precondition failed: resource has been modified by another process")
	}

	// Edge Case 2: No-Op Update check (skip write if data identical)
	if existing.Title == b.Title && existing.Author == b.Author && existing.Price == b.Price {
		return existing, nil
	}

	existingBooks, err := s.repo.GetAll()
	if err == nil {
		for _, other := range existingBooks {
			if other.DeletedAt == nil && other.ID != id && strings.EqualFold(other.Title, b.Title) && strings.EqualFold(other.Author, b.Author) {
				return nil, errors.New("another book with the same title and author already exists")
			}
		}
	}

	b.ID = id
	b.Version = existing.Version + 1 // [NEW] Increment version on valid mutation
	b.CreatedAt = existing.CreatedAt
	b.UpdatedAt = time.Now()

	return s.repo.Update(id, b)
}

func (s *bookService) Delete(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid UUID format")
	}

	existing, err := s.GetByID(id)
	if err != nil || existing == nil {
		return errors.New("book not found")
	}

	// [NEW] Soft Delete Action: Set DeletedAt timestamp rather than removing from repository
	now := time.Now()
	existing.DeletedAt = &now
	existing.UpdatedAt = now

	_, err = s.repo.Update(id, *existing)
	return err
}

func (s *bookService) CreateBatch(books []model.Book) ([]model.Book, error) {
	var createdBooks []model.Book
	for _, b := range books {
		created, err := s.Create(b)
		if err != nil {
			return nil, err // Abort batch on first validation or duplicate error
		}
		createdBooks = append(createdBooks, *created)
	}
	return createdBooks, nil
}

func (s *bookService) DeleteBatch(ids []string) error {
	for _, id := range ids {
		if err := s.Delete(id); err != nil {
			return err
		}
	}
	return nil
}