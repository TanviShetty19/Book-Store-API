package model

import "time"

type Book struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Author    string     `json:"author"`
	Price     float64    `json:"price"`
	Version   int        `json:"version"`            // [NEW] Optimistic Locking counter
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"` // [NEW] Soft delete timestamp (nil if active)
}