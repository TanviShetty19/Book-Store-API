package handler
import (
	"encoding/json"
	"net/http"
	"bookstore-api/internal/model"
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

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var input model.Book
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid JSON body"}`))
		return
	}
	createdBook, err := h.service.Create(input)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdBook)
}
func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var input model.Book
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid JSON body"}`))
		return
	}

	updatedBook, err := h.service.Update(id, input)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if err.Error() == "book not found" {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedBook)
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.service.Delete(id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Book not found"}`))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}