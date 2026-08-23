package model

import (
	"errors"
	"strings"
	"time"
)

type Book struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Author    string     `json:"author"`
	Price     float64    `json:"price"`
	Stock     int        `json:"stock"`
	Version   int        `json:"version"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

func (b *Book) CanFulfill(qty int) bool {
	return qty > 0 && b.Stock >= qty
}

func (b *Book) Validate() error {
	if strings.TrimSpace(b.Title) == "" {
		return errors.New("title cannot be empty")
	}
	if strings.TrimSpace(b.Author) == "" {
		return errors.New("author cannot be empty")
	}
	if b.Price <= 0 {
		return errors.New("price must be greater than zero")
	}
	if b.Stock < 0 {
		return errors.New("stock cannot be negative")
	}
	return nil
}