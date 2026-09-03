package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"bookstore-api/internal/apperrors"
	"bookstore-api/internal/model"
)

type JsonBookRepository struct {
	mu       sync.RWMutex
	filePath string
}

func NewJsonBookRepository(filePath string) (*JsonBookRepository, error) {
	repo := &JsonBookRepository{filePath: filePath}
	if err := repo.initStorage(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *JsonBookRepository) initStorage() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, err := os.Stat(r.filePath); os.IsNotExist(err) {
		data, _ := json.MarshalIndent([]model.Book{}, "", "  ")
		return os.WriteFile(r.filePath, data, 0644)
	}
	return nil
}

func (r *JsonBookRepository) loadBooks() ([]model.Book, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, err
	}
	var books []model.Book
	if err := json.Unmarshal(data, &books); err != nil {
		return nil, err
	}
	return books, nil
}

func (r *JsonBookRepository) saveBooks(books []model.Book) error {
	data, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.filePath, data, 0644)
}

func (r *JsonBookRepository) Create(ctx context.Context, book *model.Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	books, err := r.loadBooks()
	if err != nil {
		return err
	}

	books = append(books, *book)
	return r.saveBooks(books)
}

func (r *JsonBookRepository) GetByID(ctx context.Context, id string) (*model.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	books, err := r.loadBooks()
	if err != nil {
		return nil, err
	}

	for _, book := range books {
		if book.ID == id && book.DeletedAt == nil {
			return &book, nil
		}
	}
	return nil, errors.New("book not found")
}

func (r *JsonBookRepository) GetAll(ctx context.Context, offset, limit int64) ([]*model.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	books, err := r.loadBooks()
	if err != nil {
		return nil, err
	}

	var activeBooks []*model.Book
	for i := range books {
		if books[i].DeletedAt == nil {
			activeBooks = append(activeBooks, &books[i])
		}
	}
	// Apply pagination
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = int64(len(activeBooks))
	}
	end := offset + limit
	if end > int64(len(activeBooks)) {
		end = int64(len(activeBooks))
	}
	if offset >= int64(len(activeBooks)) {
		return []*model.Book{}, nil
	}
	return activeBooks[offset:end], nil

}

// Update performs an atomic compare-and-swap on the stored book record.
// Contract: updatedBook.Version must equal the version the caller last
// read (NOT pre-incremented). On match, the repository applies the
// mutated fields, increments the stored version internally, and
// persists — all within the same lock, closing the GetByID-then-Update
// TOCTOU gap. On mismatch, an apperrors.ErrConflict is returned.
func (r *JsonBookRepository) Update(ctx context.Context, updatedBook *model.Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	books, err := r.loadBooks()
	if err != nil {
		return err
	}

	idx := -1
	for i, book := range books {
		if book.ID == updatedBook.ID && book.DeletedAt == nil {
			idx = i
			break
		}
	}

	if idx == -1 {
		return errors.New("book not found or deleted")
	}

	if books[idx].Version != updatedBook.Version {
		return fmt.Errorf("%w: book version mismatch (current: %d, provided: %d)", apperrors.ErrConflict, books[idx].Version, updatedBook.Version)
	}

	books[idx].Title = updatedBook.Title
	books[idx].Author = updatedBook.Author
	books[idx].Price = updatedBook.Price
	books[idx].Stock = updatedBook.Stock
	books[idx].DeletedAt = updatedBook.DeletedAt
	books[idx].Version++
	books[idx].UpdatedAt = time.Now()

	return r.saveBooks(books)
}

func (r *JsonBookRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	books, err := r.loadBooks()
	if err != nil {
		return err
	}

	found := false
	now := time.Now()
	for i, book := range books {
		if book.ID == id && book.DeletedAt == nil {
			books[i].DeletedAt = &now
			found = true
			break
		}
	}

	if !found {
		return errors.New("book not found")
	}

	return r.saveBooks(books)
}
