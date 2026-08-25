package dto

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// CreateOrderItemRequestDTO represents a single line item in the incoming checkout payload
type CreateOrderItemRequestDTO struct {
	BookID   string `json:"book_id"`
	Quantity int    `json:"quantity"`
}

// CreateOrderRequestDTO represents the top-level incoming payload for POST /orders
type CreateOrderRequestDTO struct {
	Items []CreateOrderItemRequestDTO `json:"items"`
}

func (r *CreateOrderRequestDTO) Validate() error {
	// 1. Ensure order contains at least one item
	if len(r.Items) == 0 {
		return errors.New("order must contain at least one item")
	}

	seenBooks := make(map[string]bool)

	// 2. Loop through each line item for sanitization and validation
	for i := range r.Items {
		// Trim surrounding whitespace from BookID
		r.Items[i].BookID = strings.TrimSpace(r.Items[i].BookID)

		if r.Items[i].BookID == "" {
			return fmt.Errorf("item %d: book_id is required", i+1)
		}

		if r.Items[i].Quantity <= 0 {
			return fmt.Errorf("item %d: quantity must be at least 1", i+1)
		}

		// Prevent duplicate line items for the same book in one order payload
		if seenBooks[r.Items[i].BookID] {
			return fmt.Errorf("duplicate book_id '%s' found in order items", r.Items[i].BookID)
		}
		seenBooks[r.Items[i].BookID] = true
	}

	return nil
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