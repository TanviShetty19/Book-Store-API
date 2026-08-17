package repository

import (
	"errors"
	"sync"

	"bookstore-api/internal/model"
)

type UserRepository interface {
	Create(user *model.User) error
	GetByEmail(email string) (*model.User, error)
}

type MemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]model.User // email -> user
}

func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{users: make(map[string]model.User)}
}

func (r *MemoryUserRepository) Create(user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.Email]; exists {
		return errors.New("user with this email already exists")
	}
	r.users[user.Email] = *user
	return nil
}

func (r *MemoryUserRepository) GetByEmail(email string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return &user, nil
}