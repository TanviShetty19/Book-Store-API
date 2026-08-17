package dto

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"github.com/google/uuid" // Add this line
	"bookstore-api/internal/model"
)

// CreateBookRequest defines the strictly allowed payload for POST /books
type CreateBookRequest struct {
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Price  float64 `json:"price"`
}
// PaginationMeta holds metadata for paginated query responses
type PaginationMeta struct {
	CurrentPage int `json:"current_page"`
	PageSize    int `json:"page_size"`
	TotalItems  int `json:"total_items"`
	TotalPages  int `json:"total_pages"`
}

// PaginatedBookResponse defines the API payload structure for GET /books
type PaginatedBookResponse struct {
	Data []*BookResponse `json:"data"`
	Meta PaginationMeta  `json:"meta"`
}
// Validate checks structural input rules at the HTTP boundary
func (r *CreateBookRequest) Validate() error {
	trimmedTitle := strings.TrimSpace(r.Title) // [NEW EDGE CASE] Clean leading/trailing spaces
	if trimmedTitle == "" {
		return errors.New("title is required and cannot be empty")
	}
	if len(trimmedTitle) > 255 { // [NEW EDGE CASE] Max string length boundary guard
		return errors.New("title cannot exceed 255 characters")
	}

	trimmedAuthor := strings.TrimSpace(r.Author) // [NEW EDGE CASE] Clean leading/trailing spaces
	if trimmedAuthor == "" {
		return errors.New("author is required and cannot be empty")
	}
	if len(trimmedAuthor) > 255 { // [NEW EDGE CASE] Max string length boundary guard
		return errors.New("author cannot exceed 255 characters")
	}

	if r.Price <= 0 || r.Price > 10000.00 { // [NEW EDGE CASE] Upper & lower price boundary checks ($0.01 to $10,000.00)
		return errors.New("price must be between 0.01 and 10000.00")
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

// UpdateBookRequest defines the strictly allowed payload for PUT /books/{id}
type UpdateBookRequest struct {
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Price  float64 `json:"price"`
}

func (r *UpdateBookRequest) Validate() error {
	trimmedTitle := strings.TrimSpace(r.Title) // [NEW EDGE CASE] Clean leading/trailing spaces
	if trimmedTitle == "" {
		return errors.New("title cannot be empty")
	}
	if len(trimmedTitle) > 255 { // [NEW EDGE CASE] Max string length boundary guard
		return errors.New("title cannot exceed 255 characters")
	}

	trimmedAuthor := strings.TrimSpace(r.Author) // [NEW EDGE CASE] Clean leading/trailing spaces
	if trimmedAuthor == "" {
		return errors.New("author cannot be empty")
	}
	if len(trimmedAuthor) > 255 { // [NEW EDGE CASE] Max string length boundary guard
		return errors.New("author cannot exceed 255 characters")
	}

	if r.Price <= 0 || r.Price > 10000.00 { // [NEW EDGE CASE] Upper & lower price boundary checks ($0.01 to $10,000.00)
		return errors.New("price must be between 0.01 and 10000.00")
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

// BookResponse controls the outgoing JSON payload sent back to clients
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

// CreateBatchBookRequest defines payload for POST /books/batch
type CreateBatchBookRequest struct {
	Books []CreateBookRequest `json:"books"`
}

func (r *CreateBatchBookRequest) Validate() error {
	if len(r.Books) == 0 {
		return errors.New("books slice cannot be empty")
	}
	for i, b := range r.Books {
		if err := b.Validate(); err != nil {
			return fmt.Errorf("item [%d]: %w", i, err)
		}
	}
	return nil
}

// DeleteBatchBookRequest defines payload for DELETE /books/batch
type DeleteBatchBookRequest struct {
	IDs []string `json:"ids"`
}

func (r *DeleteBatchBookRequest) Validate() error {
	if len(r.IDs) == 0 {
		return errors.New("ids slice cannot be empty")
	}
	for _, id := range r.IDs {
		if _, err := uuid.Parse(id); err != nil {
			return fmt.Errorf("invalid UUID: %s", id)
		}
	}
	return nil
}