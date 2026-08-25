package dto

import (
	"errors"
	"strings"
	"time" // Re-enabled for BookResponseDTO timestamps
)

// CreateBookRequestDTO represents the incoming payload to add a new book
type CreateBookRequestDTO struct {
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Price  float64 `json:"price"`
	Stock  int     `json:"stock"`
}

func (c *CreateBookRequestDTO) Validate() error {
	// 1. Trim leading and trailing whitespace from string fields
	c.Title = strings.TrimSpace(c.Title)
	c.Author = strings.TrimSpace(c.Author)

	// 2. Validate non-empty strings
	if c.Title == "" {
		return errors.New("book title is required")
	}
	if c.Author == "" {
		return errors.New("book author is required")
	}

	// 3. Validate numeric boundaries
	if c.Price <= 0 {
		return errors.New("price must be greater than zero")
	}
	if c.Stock < 0 {
		return errors.New("stock cannot be negative")
	}

	return nil
}

// UpdateBookRequestDTO represents the incoming payload for PATCH/PUT /books/{id}
type UpdateBookRequestDTO struct {
	Title   *string  `json:"title,omitempty"`
	Author  *string  `json:"author,omitempty"`
	Price   *float64 `json:"price,omitempty"`
	Stock   *int     `json:"stock,omitempty"`
	Version *int     `json:"version,omitempty"`
}

func (u *UpdateBookRequestDTO) Validate() error {
	// 1. Ensure at least one field is provided for update
	if u.Title == nil && u.Author == nil && u.Price == nil && u.Stock == nil {
		return errors.New("at least one field must be provided for update")
	}

	// 2. Sanitize and validate Title pointer if provided
	if u.Title != nil {
		trimmed := strings.TrimSpace(*u.Title)
		if trimmed == "" {
			return errors.New("book title cannot be empty")
		}
		u.Title = &trimmed // Store trimmed string back to pointer
	}

	// 3. Sanitize and validate Author pointer if provided
	if u.Author != nil {
		trimmed := strings.TrimSpace(*u.Author)
		if trimmed == "" {
			return errors.New("book author cannot be empty")
		}
		u.Author = &trimmed
	}

	// 4. Validate numeric boundaries on pointers
	if u.Price != nil && *u.Price <= 0 {
		return errors.New("price must be greater than zero")
	}
	if u.Stock != nil && *u.Stock < 0 {
		return errors.New("stock cannot be negative")
	}

	return nil
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