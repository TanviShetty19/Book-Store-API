package service

import (
	"errors"
	"strings"
	"time"
	"bookstore-api/internal/model"
	"bookstore-api/internal/repository"
)

// BookService defines the business operations contract
type BookService interface {
	GetAll() ([]model.Book, error)
	GetByID(id string) (*model.Book, error)
	Create(book model.Book) (*model.Book, error) 
	Update(id string, book model.Book) (*model.Book, error)
	Delete(id string) error
}

// bookService is the concrete implementation holding the repository dependency
type bookService struct {
	repo repository.BookRepository
}

// NewBookService constructs a BookService with the injected repository
func NewBookService(repo repository.BookRepository) BookService {
	return &bookService{
		repo: repo,
	}
}

// GetAll delegates data fetching to the repository layer
func (s *bookService) GetAll() ([]model.Book, error) {
	return s.repo.GetAll()
}

func (s *bookService) GetByID(id string) (*model.Book, error) {
	return s.repo.GetByID(id)
}
func (s *bookService) Create(book model.Book) (*model.Book, error) {
	//Business Validation Rules
	if strings.TrimSpace(book.Title) == "" {
		return nil, errors.New("title is required")
	}
	if strings.TrimSpace(book.Author) == "" {
		return nil, errors.New("author is required")
	}
	if book.Price <= 0 {
		return nil, errors.New("price must be greater than 0")
	}
	// Assign timestamp
	now := time.Now()
	book.CreatedAt = now
	book.UpdatedAt = now

	// Delegate to repository
	return s.repo.Create(book)
}
func (s *bookService) Update(id string, book model.Book) (*model.Book, error) {
	if strings.TrimSpace(book.Title) == "" {
		return nil, errors.New("title is required")
	}
	if strings.TrimSpace(book.Author) == "" {
		return nil, errors.New("author is required")
	}
	if book.Price <= 0 {
		return nil, errors.New("price must be greater than 0")
	}

	book.UpdatedAt = time.Now()
	return s.repo.Update(id, book)
}

func (s *bookService) Delete(id string) error {
	return s.repo.Delete(id)
}