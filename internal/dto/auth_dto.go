package dto

import (
	"errors"
	"strings"
)

// RegisterRequest defines the payload for POST /auth/register
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // Optional: defaults to "customer"
}

func (r *RegisterRequest) Validate() error {
	if !strings.Contains(r.Email, "@") || strings.TrimSpace(r.Email) == "" {
		return errors.New("valid email is required")
	}
	if len(r.Password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}
	return nil
}

// LoginRequest defines the payload for POST /auth/login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (l *LoginRequest) Validate() error {
	if strings.TrimSpace(l.Email) == "" || strings.TrimSpace(l.Password) == "" {
		return errors.New("email and password are required")
	}
	return nil
}

// AuthResponse returns the JWT bearer token payload to the client
type AuthResponse struct {
	Token string `json:"token"`
	Type  string `json:"token_type"`
}