package apperrors

import "errors"

var (
	ErrNotFound     = errors.New("resource not found")
	ErrValidation   = errors.New("validation failed")
	ErrConflict     = errors.New("resource conflict")
	ErrUnauthorized = errors.New("unauthorized access")
	ErrForbidden    = errors.New("forbidden action")
)