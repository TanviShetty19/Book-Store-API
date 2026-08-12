package service

import (
	"bookstore-api/internal/model"
	"bookstore-api/internal/repository"
)

// BookService defines the business operations contract
type BookService interface {
	GetAll() ([]model.Book, error)
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