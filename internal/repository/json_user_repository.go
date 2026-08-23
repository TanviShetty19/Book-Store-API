package repository

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"bookstore-api/internal/model"
)

type JsonUserRepository struct {
	mu       sync.RWMutex
	filePath string
}

func NewJsonUserRepository(filePath string) (*JsonUserRepository, error) {
	repo := &JsonUserRepository{filePath: filePath}
	if err := repo.initStorage(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *JsonUserRepository) initStorage() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, err := os.Stat(r.filePath); os.IsNotExist(err) {
		data, _ := json.MarshalIndent([]model.User{}, "", "  ")
		return os.WriteFile(r.filePath, data, 0644)
	}
	return nil
}

func (r *JsonUserRepository) loadUsers() ([]model.User, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, err
	}
	var users []model.User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *JsonUserRepository) saveUsers(users []model.User) error {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.filePath, data, 0644)
}

func (r *JsonUserRepository) Create(ctx context.Context, user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	users, err := r.loadUsers()
	if err != nil {
		return err
	}

	for _, existing := range users {
		if existing.Email == user.Email {
			return errors.New("user with this email already exists")
		}
	}

	users = append(users, *user)
	return r.saveUsers(users)
}

func (r *JsonUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users, err := r.loadUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.ID == id {
			return &user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *JsonUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users, err := r.loadUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Email == email {
			return &user, nil
		}
	}
	return nil, errors.New("user not found")
}