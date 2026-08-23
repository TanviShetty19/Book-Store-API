package dto

import "time"

// CreateBookRequestDTO represents the incoming payload to add a new book
type CreateBookRequestDTO struct {
	Title  string  `json:"title" validate:"required,min=1,max=255"`
	Author string  `json:"author" validate:"required,min=1,max=255"`
	Price  float64 `json:"price" validate:"required,gt=0"`
	Stock  int     `json:"stock" validate:"gte=0"` // gte=0 allows starting with 0 stock
}

// UpdateBookRequestDTO represents the incoming payload for PATCH/PUT /books/{id}
type UpdateBookRequestDTO struct {
	Title   *string  `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Author  *string  `json:"author,omitempty" validate:"omitempty,min=1,max=255"`
	Price   *float64 `json:"price,omitempty" validate:"omitempty,gt=0"`
	Stock   *int     `json:"stock,omitempty" validate:"omitempty,gte=0"`
	Version *int     `json:"version,omitempty" validate:"omitempty,gte=1"` // Required for optimistic concurrency control
}

// BookResponseDTO represents the outgoing JSON format for a book
type BookResponseDTO struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Price     float64   `json:"price"`
	Stock     int       `json:"stock"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}