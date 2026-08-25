package repository

import (
	"context"

	"bookstore-api/internal/model"
)

type OrderRepository interface {
	CreateDraftOrder(ctx context.Context, order *model.Order) error
	ConfirmOrder(ctx context.Context, orderID string) (*model.Order, error)
	CancelOrder(ctx context.Context, orderID string) (*model.Order, error)
	GetByID(ctx context.Context, id string) (*model.Order, error)
	GetByUserID(ctx context.Context, userID string) ([]*model.Order, error)
}
