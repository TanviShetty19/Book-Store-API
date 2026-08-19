package handler

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"

	"bookstore-api/internal/dto"
	"bookstore-api/internal/model"
	"bookstore-api/internal/service"
)

type BookHandler struct {
	service service.BookService
}

func NewBookHandler(service service.BookService) *BookHandler {
	return &BookHandler{service: service}
}

// respondJSON writes pretty-printed JSON responses with appropriate HTTP headers
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(data)
	}
}

func (h *BookHandler) GetAllBooks(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	books, totalItems, err := h.service.GetAll(page, limit)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve books"})
		return
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

	response := dto.PaginatedBookResponse{
		Data: dto.NewBookResponseList(books),
		Meta: dto.PaginationMeta{
			CurrentPage: page,
			PageSize:    limit,
			TotalItems:  totalItems,
			TotalPages:  totalPages,
		},
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *BookHandler) GetBookByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	book, err := h.service.GetByID(id)
	if err != nil || book == nil {
		status := http.StatusNotFound
		if err != nil && err.Error() == "invalid UUID format" {
			status = http.StatusBadRequest
		}
		errMsg := "book not found"
		if err != nil {
			errMsg = err.Error()
		}
		respondJSON(w, status, map[string]string{"error": errMsg})
		return
	}

	respondJSON(w, http.StatusOK, dto.NewBookResponse(book))
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateBookRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON payload"})
		return
	}

	if err := req.Validate(); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	createdBook, err := h.service.Create(*req.ToDomain())
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "a book with the same title and author already exists" {
			status = http.StatusConflict
		}
		respondJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, dto.NewBookResponse(createdBook))
}

func (h *BookHandler) CreateBatchBooks(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateBatchBookRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON payload"})
		return
	}

	if err := req.Validate(); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	domainBooks := make([]model.Book, len(req.Books))
	for i, bReq := range req.Books {
		domainBooks[i] = *bReq.ToDomain()
	}

	createdBooks, err := h.service.CreateBatch(domainBooks)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "a book with the same title and author already exists" {
			status = http.StatusConflict
		}
		respondJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, dto.NewBookResponseList(createdBooks))
}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req dto.UpdateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON payload"})
		return
	}

	if err := req.Validate(); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	updatedBook, err := h.service.Update(id, *req.ToDomain(id))
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "precondition failed: resource has been modified by another process" {
			status = http.StatusPreconditionFailed
		} else if err.Error() == "book not found" {
			status = http.StatusNotFound
		} else if err.Error() == "invalid UUID format" {
			status = http.StatusBadRequest
		} else if err.Error() == "another book with the same title and author already exists" {
			status = http.StatusConflict
		}
		respondJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, dto.NewBookResponse(updatedBook))
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.service.Delete(id); err != nil {
		status := http.StatusNotFound
		if err.Error() == "invalid UUID format" {
			status = http.StatusBadRequest
		}
		respondJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func (h *BookHandler) DeleteBatchBooks(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteBatchBookRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON payload"})
		return
	}

	if err := req.Validate(); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := h.service.DeleteBatch(req.IDs); err != nil {
		status := http.StatusBadRequest
		if err.Error() == "book not found" {
			status = http.StatusNotFound
		}
		respondJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}