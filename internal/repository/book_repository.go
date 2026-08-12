package repository

import "bookstore-api/internal/model"

type BookRepository interface {
	GetAll() ([]model.Book,error)
	GetByID(id string) (*model.Book, error)
	Create(book model.Book) (*model.Book, error)
	Update(id string, book model.Book) (*model.Book, error)
	Delete(id string) error
}