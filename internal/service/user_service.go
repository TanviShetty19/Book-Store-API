package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"bookstore-api/internal/apperrors"
	"bookstore-api/internal/dto"
	"bookstore-api/internal/model"
	"bookstore-api/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, req dto.RegisterUserRequestDTO) (*dto.UserResponseDTO, error)
	GetUserByID(ctx context.Context, id string) (*dto.UserResponseDTO, error)
}

type userServiceImpl struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userServiceImpl{userRepo: userRepo}
}

func (s *userServiceImpl) Register(ctx context.Context, req dto.RegisterUserRequestDTO) (*dto.UserResponseDTO, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", apperrors.ErrValidation, err)
	}

	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("%w: user with this email already exists", apperrors.ErrConflict)
	}
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return nil, fmt.Errorf("storage error: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	role := model.RoleCustomer
	if strings.EqualFold(req.Role, string(model.RoleAdmin)) {
		role = model.RoleAdmin
	}

	user := &model.User{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      role,
		CreatedAt: time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &dto.UserResponseDTO{
		ID:        user.ID,
		Email:     user.Email,
		Role:      string(user.Role),
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *userServiceImpl) GetUserByID(ctx context.Context, id string) (*dto.UserResponseDTO, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: user not found", apperrors.ErrNotFound)
	}
	return &dto.UserResponseDTO{
		ID:        user.ID,
		Email:     user.Email,
		Role:      string(user.Role),
		CreatedAt: user.CreatedAt,
	}, nil
}