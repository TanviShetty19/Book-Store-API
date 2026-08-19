package service

import (
	"errors"
	"time"

	"bookstore-api/internal/auth"
	"bookstore-api/internal/dto"
	"bookstore-api/internal/model"
	"bookstore-api/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(req dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(req dto.LoginRequest) (*dto.AuthResponse, error)
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to process password")
	}

	role := req.Role
	if role == "" {
		role = "customer"
	}

	user := &model.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         role,
		CreatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		if err.Error() == "user with this email already exists" {
			return nil, err
		}
		return nil, errors.New("internal server error")
	}

	token, err := auth.GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, errors.New("internal server error")
	}

	return &dto.AuthResponse{Token: token, Type: "Bearer"}, nil
}

func (s *authService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := auth.GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{Token: token, Type: "Bearer"}, nil
}