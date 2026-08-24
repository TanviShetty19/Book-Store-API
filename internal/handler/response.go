package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"bookstore-api/internal/apperrors"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// RespondWithError unwraps domain errors and sets REST-compliant status codes
func RespondWithError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError

	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		statusCode = http.StatusNotFound // 404 Not Found
	case errors.Is(err, apperrors.ErrConflict):
		statusCode = http.StatusConflict // 409 Conflict
	case errors.Is(err, apperrors.ErrForbidden):
		statusCode = http.StatusForbidden // 403 Forbidden
	case errors.Is(err, apperrors.ErrUnauthorized):
		statusCode = http.StatusUnauthorized // 401 Unauthorized
	case errors.Is(err, apperrors.ErrValidation):
		statusCode = http.StatusBadRequest // 400 Bad Request
	}

	RespondWithJSON(w, statusCode, ErrorResponse{Error: err.Error()})
}

// RespondWithJSON handles standardizing JSON responses
func RespondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}