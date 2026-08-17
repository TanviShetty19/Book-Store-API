package repository

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"bookstore-api/internal/model"
)

// BookRepository defines the persistence contract for Book entities
type BookRepository interface {
	GetAll() ([]model.Book, error)
	GetByID(id string) (*model.Book, error)
	Create(book model.Book) (*model.Book, error)
	Update(id string, book model.Book) (*model.Book, error)
	Delete(id string) error
}

// JSONBookRepository implements BookRepository backed by a local JSON file
type JSONBookRepository struct {
	mu       sync.RWMutex
	filePath string
}

func NewJSONBookRepository(filePath string) *JSONBookRepository {
	repo := &JSONBookRepository{filePath: filePath}
	// Ensure file exists on startup
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		_ = os.WriteFile(filePath, []byte("[]"), 0644)
	}
	return repo
}

func (r *JSONBookRepository) load() ([]model.Book, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.Book{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return []model.Book{}, nil
	}

	var books []model.Book
	if err := json.Unmarshal(data, &books); err != nil {
		return nil, err
	}
	return books, nil
}

func (r *JSONBookRepository) save(books []model.Book) error {
	data, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.filePath, data, 0644)
}

func (r *JSONBookRepository) GetAll() ([]model.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.load()
}

func (r *JSONBookRepository) GetByID(id string) (*model.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	books, err := r.load()
	if err != nil {
		return nil, err
	}

	for _, b := range books {
		if b.ID == id {
			return &b, nil
		}
	}
	return nil, errors.New("book not found")
}

func (r *JSONBookRepository) Create(book model.Book) (*model.Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	books, err := r.load()
	if err != nil {
		return nil, err
	}

	books = append(books, book)
	if err := r.save(books); err != nil {
		return nil, err
	}

	return &book, nil
}

func (r *JSONBookRepository) Update(id string, book model.Book) (*model.Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	books, err := r.load()
	if err != nil {
		return nil, err
	}

	found := false
	for i, b := range books {
		if b.ID == id {
			books[i] = book
			found = true
			break
		}
	}

	if !found {
		return nil, errors.New("book not found")
	}

	if err := r.save(books); err != nil {
		return nil, err
	}

	return &book, nil
}

func (r *JSONBookRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	books, err := r.load()
	if err != nil {
		return err
	}

	index := -1
	for i, b := range books {
		if b.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		return errors.New("book not found")
	}

	books = append(books[:index], books[index+1:]...)
	return r.save(books)
}