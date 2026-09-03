package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"bookstore-api/internal/apperrors"
	"bookstore-api/internal/dto"
	"bookstore-api/internal/model"
	"bookstore-api/internal/repository"

	"github.com/google/uuid"
)

type BookService interface {
	CreateBook(ctx context.Context, req dto.CreateBookRequestDTO) (*dto.BookResponseDTO, error)
	GetBookByID(ctx context.Context, id string) (*dto.BookResponseDTO, error)
	GetAllBooks(ctx context.Context, offset, limit int64) ([]*dto.BookResponseDTO, error)
	UpdateBook(ctx context.Context, id string, req dto.UpdateBookRequestDTO) (*dto.BookResponseDTO, error)
	DeleteBook(ctx context.Context, id string) error
}

type bookServiceImpl struct {
	bookRepo repository.BookRepository
}

func NewBookService(bookRepo repository.BookRepository) BookService {
	return &bookServiceImpl{bookRepo: bookRepo}
}

func (s *bookServiceImpl) CreateBook(ctx context.Context, req dto.CreateBookRequestDTO) (*dto.BookResponseDTO, error) {
	if req.Stock < 0 {
		return nil, fmt.Errorf("%w: stock cannot be negative", apperrors.ErrValidation)
	}
	if req.Price <= 0 {
		return nil, fmt.Errorf("%w: price must be greater than zero", apperrors.ErrValidation)
	}

	// Fetch up to the repository max limit (100) to check for duplicate entries
	existingBooks, err := s.bookRepo.GetAll(ctx, 0, 100)
	if err == nil {
		for _, b := range existingBooks {
			if strings.EqualFold(strings.TrimSpace(b.Title), strings.TrimSpace(req.Title)) &&
				strings.EqualFold(strings.TrimSpace(b.Author), strings.TrimSpace(req.Author)) {
				return nil, fmt.Errorf("%w: a book with title '%s' by author '%s' already exists", apperrors.ErrConflict, req.Title, req.Author)
			}
		}
	}

	book := &model.Book{
		ID:        uuid.New().String(),
		Title:     strings.TrimSpace(req.Title),
		Author:    strings.TrimSpace(req.Author),
		Price:     req.Price,
		Stock:     req.Stock,
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.bookRepo.Create(ctx, book); err != nil {
		return nil, err
	}

	return mapBookToDTO(book), nil
}

func (s *bookServiceImpl) GetBookByID(ctx context.Context, id string) (*dto.BookResponseDTO, error) {
	book, err := s.bookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: book not found", apperrors.ErrNotFound)
	}
	return mapBookToDTO(book), nil
}

func (s *bookServiceImpl) GetAllBooks(ctx context.Context, offset, limit int64) ([]*dto.BookResponseDTO, error) {
	books, err := s.bookRepo.GetAll(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	var dtos []*dto.BookResponseDTO
	for _, book := range books {
		dtos = append(dtos, mapBookToDTO(book))
	}
	return dtos, nil
}

func (s *bookServiceImpl) UpdateBook(ctx context.Context, id string, req dto.UpdateBookRequestDTO) (*dto.BookResponseDTO, error) {
	book, err := s.bookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: book not found", apperrors.ErrNotFound)
	}

	if req.Title != nil {
		book.Title = strings.TrimSpace(*req.Title)
	}
	if req.Author != nil {
		book.Author = strings.TrimSpace(*req.Author)
	}
	if req.Price != nil {
		if *req.Price <= 0 {
			return nil, fmt.Errorf("%w: price must be greater than zero", apperrors.ErrValidation)
		}
		book.Price = *req.Price
	}
	if req.Stock != nil {
		if *req.Stock < 0 {
			return nil, fmt.Errorf("%w: stock cannot be negative", apperrors.ErrValidation)
		}
		book.Stock = *req.Stock
	}

	book.UpdatedAt = time.Now()

	if err := s.bookRepo.Update(ctx, book); err != nil {
		return nil, err
	}

	return mapBookToDTO(book), nil
}

func (s *bookServiceImpl) DeleteBook(ctx context.Context, id string) error {
	if _, err := s.bookRepo.GetByID(ctx, id); err != nil {
		return fmt.Errorf("%w: book not found", apperrors.ErrNotFound)
	}
	return s.bookRepo.Delete(ctx, id)
}

func mapBookToDTO(book *model.Book) *dto.BookResponseDTO {
	return &dto.BookResponseDTO{
		ID:        book.ID,
		Title:     book.Title,
		Author:    book.Author,
		Price:     book.Price,
		Stock:     book.Stock,
		Version:   book.Version,
		CreatedAt: book.CreatedAt,
		UpdatedAt: book.UpdatedAt,
	}
}
