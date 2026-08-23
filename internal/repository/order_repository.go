package repository

import (
	"context"
	"bookstore-api/internal/model"
)

type OrderRepository interface {
	CreateOrderWithDeduction(ctx context.Context, order *model.Order) error
	GetOrderByID(ctx context.Context, id string) (*model.Order, error)
	GetOrdersByUserID(ctx context.Context, userID string) ([]*model.Order, error)
}
