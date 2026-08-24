package repository

import (
	"context"

	"bookstore-api/internal/model"
)

type OrderRepository interface {
	CreateOrderWithStockDeduction(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id string) (*model.Order, error)
	GetByUserID(ctx context.Context, userID string) ([]*model.Order, error)
}