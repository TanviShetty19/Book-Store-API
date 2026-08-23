package dto

import (
	"errors"
	"strings"
	"time"
)

// RegisterUserRequestDTO defines the payload for POST /users/register
type RegisterUserRequestDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role,omitempty"` // Optional: "ADMIN" or "CUSTOMER"
}

func (r *RegisterUserRequestDTO) Validate() error {
	if !strings.Contains(r.Email, "@") || strings.TrimSpace(r.Email) == "" {
		return errors.New("valid email is required")
	}
	if len(r.Password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}
	return nil
}

// LoginRequestDTO defines the payload for POST /auth/login
type LoginRequestDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (l *LoginRequestDTO) Validate() error {
	if strings.TrimSpace(l.Email) == "" || strings.TrimSpace(l.Password) == "" {
		return errors.New("email and password are required")
	}
	return nil
}

// UserResponseDTO returns user profile information without exposing credentials
type UserResponseDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthResponseDTO returns the JWT bearer token payload to the client
type AuthResponseDTO struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
}