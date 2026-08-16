package dto

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"bookstore-api/internal/model"
)

// CreateBookRequest defines the strictly allowed fields for POST /books
type CreateBookRequest struct {
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Price  float64 `json:"price"`
}

// Validate checks structural input rules at the HTTP boundary
func (r *CreateBookRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required and cannot be empty")
	}
	if strings.TrimSpace(r.Author) == "" {
		return errors.New("author is required and cannot be empty")
	}
	if r.Price <= 0 {
		return errors.New("price must be greater than 0")
	}
	return nil
}

// ToDomain converts the validated DTO into an internal domain entity
func (r *CreateBookRequest) ToDomain() *model.Book {
	return &model.Book{
		Title:  strings.TrimSpace(r.Title),
		Author: strings.TrimSpace(r.Author),
		Price:  r.Price,
	}
}

// UpdateBookRequest defines the strictly allowed fields for PUT /books/{id}
type UpdateBookRequest struct {
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Price  float64 `json:"price"`
}

func (r *UpdateBookRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title cannot be empty")
	}
	if strings.TrimSpace(r.Author) == "" {
		return errors.New("author cannot be empty")
	}
	if r.Price <= 0 {
		return errors.New("price must be greater than 0")
	}
	return nil
}

func (r *UpdateBookRequest) ToDomain(id string) *model.Book {
	return &model.Book{
		ID:     id,
		Title:  strings.TrimSpace(r.Title),
		Author: strings.TrimSpace(r.Author),
		Price:  r.Price,
	}
}

// BookResponse controls the outgoing payload structure sent back to clients
type BookResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Price     float64   `json:"price"`
	Formatted string    `json:"formatted_price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewBookResponse maps a single domain Book to an outgoing DTO
func NewBookResponse(b *model.Book) *BookResponse {
	return &BookResponse{
		ID:        b.ID,
		Title:     b.Title,
		Author:    b.Author,
		Price:     b.Price,
		Formatted: fmt.Sprintf("$%.2f", b.Price),
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
}

// NewBookResponseList maps a slice of domain Books into outgoing DTOs
func NewBookResponseList(books []model.Book) []*BookResponse {
	responses := make([]*BookResponse, len(books))
	for i, b := range books {
		responses[i] = NewBookResponse(&b)
	}
	return responses
}