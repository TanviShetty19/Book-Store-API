package model

import (
	"errors"
	"strings"
	"time"
)

// UserRole defines typed role constants for Role-Based Access Control (RBAC).
type UserRole string

const (
	RoleAdmin    UserRole = "ADMIN"
	RoleCustomer UserRole = "CUSTOMER"
)

type User struct {
	ID        string    `json:"id" bson:"_id"`
	Email     string    `json:"email" bson:"email"`
	Password  string    `json:"-" bson:"password"` // Omitted from JSON output, preserved in BSON
	Role      UserRole  `json:"role" bson:"role"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

func (u *User) Validate() error {
	if strings.TrimSpace(u.Email) == "" {
		return errors.New("email is required")
	}
	if strings.TrimSpace(u.Password) == "" {
		return errors.New("password is required")
	}
	if u.Role != RoleAdmin && u.Role != RoleCustomer {
		return errors.New("invalid user role")
	}
	return nil
}