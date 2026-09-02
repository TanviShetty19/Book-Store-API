package model

import (
	"errors"
	"strings"
	"time"
)

type Book struct {
	ID        string     `json:"id" bson:"_id"` // BSON _id maps Go ID to Mongo primary key
	Title     string     `json:"title" bson:"title"`
	Author    string     `json:"author" bson:"author"`
	Price     float64    `json:"price" bson:"price"`
	Stock     int        `json:"stock" bson:"stock"`
	Version   int        `json:"version" bson:"version"` // Used for Optimistic Locking
	CreatedAt time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" bson:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"` // Soft delete
}

func (b *Book) CheckStock(qty int) bool {
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
