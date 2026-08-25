package service

import (
	"context"

	"bookstore-api/internal/dto"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID string, req dto.CreateOrderRequestDTO) (*dto.OrderResponseDTO, error)
	ConfirmOrder(ctx context.Context, userID string, userRole string, orderID string) (*dto.OrderResponseDTO, error)
	CancelOrder(ctx context.Context, userID string, userRole string, orderID string) (*dto.OrderResponseDTO, error)
	GetOrderByID(ctx context.Context, userID string, userRole string, orderID string) (*dto.OrderResponseDTO, error)
	GetUserOrders(ctx context.Context, userID string) ([]*dto.OrderResponseDTO, error)
}
