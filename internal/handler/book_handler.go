package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"bookstore-api/internal/dto"
	"bookstore-api/internal/service"

	"github.com/gorilla/mux"
)

type BookHandler struct {
	bookService service.BookService
}

func NewBookHandler(bookService service.BookService) *BookHandler {
	return &BookHandler{bookService: bookService}
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateBookRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	response, err := h.bookService.CreateBook(r.Context(), req)
	if err != nil {
		writeJSONError(w, mapErrorToStatusCode(err), err.Error())
		return
	}

	writeJSONResponse(w, http.StatusCreated, response)
}

func (h *BookHandler) GetBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookID := vars["id"]

	response, err := h.bookService.GetBookByID(r.Context(), bookID)
	if err != nil {
		writeJSONError(w, mapErrorToStatusCode(err), err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func (h *BookHandler) GetAllBooks(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)

	books, err := h.bookService.GetAllBooks(r.Context(), offset, limit)
	if err != nil {
		writeJSONError(w, mapErrorToStatusCode(err), err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, books)
}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookID := vars["id"]

	var req dto.UpdateBookRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	response, err := h.bookService.UpdateBook(r.Context(), bookID, req)
	if err != nil {
		writeJSONError(w, mapErrorToStatusCode(err), err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookID := vars["id"]

	if err := h.bookService.DeleteBook(r.Context(), bookID); err != nil {
		writeJSONError(w, mapErrorToStatusCode(err), err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
