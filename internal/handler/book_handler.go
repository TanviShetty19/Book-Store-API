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

