package dto

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// RegisterUserRequestDTO defines the payload for POST /users/register
type RegisterUserRequestDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role,omitempty"` // Optional: "ADMIN" or "CUSTOMER"
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func (r *RegisterUserRequestDTO) Validate() error {
	// 1. Trim surrounding whitespace and lowercase email
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))

	// 2. Trim whitespace from role
	r.Role = strings.TrimSpace(r.Role)

	// 3. Validate email presence and regex structure
	if r.Email == "" || !emailRegex.MatchString(r.Email) {
		return errors.New("a valid email address is required (e.g., user@example.com)")
	}

	// 4. Validate password criteria
	if len(strings.TrimSpace(r.Password)) < 6 {
		return errors.New("password must be at least 6 characters long and cannot be only spaces")
	}

	return nil
}

// LoginRequestDTO defines the payload for POST /auth/login
type LoginRequestDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (l *LoginRequestDTO) Validate() error {
	l.Email = strings.ToLower(strings.TrimSpace(l.Email))
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