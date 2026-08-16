package dto
import (
	"errors"
	"fmt"
	"strings"
	"time"
	"bookstore-api/internal/model"

)

//defines the allowed fields
type CreateBookRequest struct {
	Title string `json:"title"`
	Author string `json:"author"`
	Price float64 `json:"price"`
}

//validate checks structural input rules at the HTTP boundary
func (r *CreateBookRequest) Validate() error{
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required and cannot be empty")
	}
	if strings.TrimSpace(r.Author) =="" {
		return errors.New("author is required and cannot be empty")
	}
	if r.Price <=0 {
		return errors.New("price must be greater than 0")
	}
	return nil
}

//ToDomain converts the validated DTO into internal domain entity
func (r *CreateBookRequest) ToDomain() *model.Book {
	return &model.Book {
		Title : strings.TrimSpace(r.Title),
		Author: srings.TrimSpace(r.Author),
		Price: r.Price,
	}
}

func (r *UpdateBookRequest) Validate() error {
	if strings.TrimSpace(r.Title) == ""{
		return error.New("title cannot be empty")
	}
	if strings.TrimSpace(r.Author) == ""{
		return error.New("author cannot be empty")
	}
	if r.Price <= 0 {
		return error.New("price must be greater than 0")
	}
	return nil
}
//BookResponse controls to outgoing payload structure sent back to the clients
type BookResponse struct {
	ID string `json:"id"`
	Title string `json:"title"`
	Author string `json:"author"`
	Price float64 `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

//Map a single domain book to outgoing DTO
func NewResponse(b *model.Book) *BookRepsonse{
	return &BookResponse {
		ID: b.ID,
		Title: b.Title,
		Author: b.Author,
		Price: b.Price,
		Formatted: fmt.Sprintf("$%.2f", b.Price),
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt
	}
}
//Map slice of domain Books into outgoin DTOs
func NewResponseSlice(books []*model.Book) []*BookResponse {
	reposnses := make([]*BookRepsonse, len(books))
	for i, b := range books {
		responses[i] = NewBookResponse(&b)
	}
	return responses
}