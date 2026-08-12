package handler
import (
	"encoding/json"
	"net/http"
	"bookstore-api/internal/service"
)

type BookHandler struct{
	service service.BookService
}

func NewBookHandler(service service.BookService) *BookHandler{
	return &BookHandler{
		service: service,
	}
}
func (h * BookHandler) GetAllBooks(w http.ResponseWriter, r *http.Request){
	books, err := h.service.GetAll()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve books"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(books)
}

func (h *BookHandler) GetBookByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	//Ask Bob (service) for the book
	book, err := h.service.GetByID(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		// If book was not found in books.json, return HTTP 404
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Book not found"}`))
		return
	}
w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(book)
}