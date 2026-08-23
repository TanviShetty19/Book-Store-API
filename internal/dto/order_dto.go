package dto

import "time"

// CreateOrderItemRequestDTO represents a single line item in the incoming checkout payload
type CreateOrderItemRequestDTO struct {
	BookID   string `json:"book_id" validate:"required,uuid"`
	Quantity int    `json:"quantity" validate:"required,gt=0"`
}

// CreateOrderRequestDTO represents the top-level incoming payload for POST /orders
type CreateOrderRequestDTO struct {
	Items []CreateOrderItemRequestDTO `json:"items" validate:"required,min=1,dive"`
}

// OrderItemResponseDTO represents a single line item returned in the order receipt
type OrderItemResponseDTO struct {
	BookID    string  `json:"book_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

// OrderResponseDTO represents the full receipt returned after an order is processed
type OrderResponseDTO struct {
	ID         string                 `json:"id"`
	UserID     string                 `json:"user_id"`
	Items      []OrderItemResponseDTO `json:"items"`
	TotalPrice float64                `json:"total_price"`
	Status     string                 `json:"status"`
	CreatedAt  time.Time              `json:"created_at"`
}