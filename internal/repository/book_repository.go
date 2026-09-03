package repository

import (
	"bookstore-api/internal/model"
	"context"
)

type BookRepository interface {
	Create(ctx context.Context, book *model.Book) error
	GetByID(ctx context.Context, id string) (*model.Book, error)
	GetAll(ctx context.Context, offset, limit int64) ([]*model.Book, error)
	Update(ctx context.Context, book *model.Book) error
	Delete(ctx context.Context, id string) error
}
