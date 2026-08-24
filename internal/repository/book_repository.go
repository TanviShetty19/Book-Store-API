package repository

import (
	"context"
	"bookstore-api/internal/model"
)

type BookRepository interface {
	Create(ctx context.Context, book *model.Book) error
	GetByID(ctx context.Context, id string) (*model.Book, error)
	GetAll(ctx context.Context) ([]*model.Book, error)
	Update(ctx context.Context, book *model.Book) error
	Delete(ctx context.Context, id string) error
}