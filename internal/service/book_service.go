package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"bookstore-api/internal/model"
	"bookstore-api/internal/repository"
)

type BookService interface {
	GetAll() ([]model.Book, error)
	GetByID(id string) (*model.Book, error)
	Create(book model.Book) (*model.Book, error)
	Update(id string, book model.Book) (*model.Book, error)
	Delete(id string) error
}

type bookService struct {
	repo repository.BookRepository
}

func NewBookService(repo repository.BookRepository) BookService {
	return &bookService{repo: repo}
}

func (s *bookService) GetAll() ([]model.Book, error) {
	return s.repo.GetAll()
}

func (s *bookService) GetByID(id string) (*model.Book, error) {
	return s.repo.GetByID(id)
}

func (s *bookService) Create(b model.Book) (*model.Book, error) {
	b.ID = uuid.New().String()
	now := time.Now()
	b.CreatedAt = now
	b.UpdatedAt = now

	return s.repo.Create(b)
}

func (s *bookService) Update(id string, b model.Book) (*model.Book, error) {
	existing, err := s.repo.GetByID(id)
	if err != nil || existing == nil {
		return nil, errors.New("book not found")
	}

	b.ID = id
	b.CreatedAt = existing.CreatedAt
	b.UpdatedAt = time.Now()

	return s.repo.Update(id, b)
}

func (s *bookService) Delete(id string) error {
	return s.repo.Delete(id)
}