package repository

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"
	"bookstore-api/internal/model"
)

type JSONBookRepository struct { //implements the BookRepository interface
	filePath string //Path to data/books.json
	mu       sync.RWMutex //Protects file from concurrent read/write
}

func NewJSONBookRepository(filePath string) BookRepository { //need more clarification on this
	return &JSONBookRepository{
		filePath: filePath,
	}
}

func (r *JSONBookRepository) loadBooks() ([]model.Book, error) {
	file, err := os.ReadFile(r.filePath)
	if err != nil {
		if err != nil {
			if errors.Is(err, os.ErrNotExist){
				return [] model.Book{}, nil
			}
		}
	}
	if len(file) == 0{
		return []model.Book{},nil
	}
	var books []model.Book
	if err := json.Unmarshal(file, &books); err !=nil {
		return nil, err
	}
	return books, nil
}

func (r *JSONBookRepository) saveBooks(books []model.Book) error{
	data, err:= json.MarshalIndent(books,"","  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.filePath, data, 0644) //Owner-Group-Other permissions
}

func (r * JSONBookRepository) GetAll() ([]model.Book, error){
	r.mu.RLock()  //Lock for reading ( doesn't block other readers)
	defer r.mu.RUnlock()
	return r.loadBooks() //read the data and return the slice

}

func (r *JSONBookRepository) GetByID(id string) (*model.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock() // Handles unlocking automatically upon function return

	books, err := r.loadBooks()
	if err != nil {
		return nil, err
	}

	for _, book := range books {
		if book.ID == id {
			return &book, nil // Just return; defer handles the unlock!
		}
	}

	return nil, errors.New("book not found")
}

func (r *JSONBookRepository) Create(book model.Book) (*model.Book, error){
	//First acquire lock
	r.mu.Lock()
	defer r.mu.Unlock()

	//Second load books
	books,err := r.loadBooks()
	if err != nil {
		return nil, err
	}
	//Keep track of creation and update times
	book.CreatedAt = time.Now()
	book.UpdatedAt = time.Now()

	books= append(books,book)
	if err := r.saveBooks(books); err != nil {
		return nil, err
	}
	return &book, nil

}

func (r *JSONBookRepository) Update(id string,updatedBook model.Book)(*model.Book, error){
	r.mu.Lock()
	defer r.mu.Unlock()

	//load books
	books, err := r.loadBooks()
	if err != nil {
		return nil, err
	}
	for i , book := range books {
		if book.ID == id {
			updatedBook.ID = id
			updatedBook.CreatedAt = book.CreatedAt
			updatedBook.UpdatedAt = time.Now()
			books[i] = updatedBook

			if err := r.saveBooks(books); err!=nil{
				return nil, err
			}
			return &books[i],nil
		} 
	}
	return nil, errors.New("book not found")
}

func (r *JSONBookRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	books, err := r.loadBooks()
	if err != nil {
		return err
	}

	for i, book := range books {
		if book.ID == id {
			//delete element at index 'i' by merging elements before and after it
			books = append(books[:i], books[i+1:]...)
			return r.saveBooks(books)
		}
	}

	return errors.New("book not found")
}