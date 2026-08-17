package model

import "time"

type User struct {
	ID string `json:"id"`
	Email string `json:"email"`
	PasswordHash string `json:"-"` //never expose in json
	Role string `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
