package model

import (
	"errors"
	"time"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusCompleted OrderStatus = "COMPLETED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

type OrderItem struct {
	BookID    string  `json:"book_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type Order struct {
	ID         string      `json:"id"`
	UserID     string      `json:"user_id"`
	Items      []OrderItem `json:"items"`
	TotalPrice float64     `json:"total_price"`
	Status     OrderStatus `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

func (o *Order) Validate() error {
	if o.UserID == "" {
		return errors.New("user ID is required")
	}
	if len(o.Items) == 0 {
		return errors.New("order must contain at least one item")
	}
	for _, item := range o.Items {
		if item.BookID == "" {
			return errors.New("item book ID cannot be empty")
		}
		if item.Quantity <= 0 {
			return errors.New("item quantity must be greater than zero")
		}
		if item.UnitPrice <= 0 {
			return errors.New("item unit price must be greater than zero")
		}
	}
	return nil
}