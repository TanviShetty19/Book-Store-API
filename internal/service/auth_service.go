package service

import (
	"context"
	"fmt"
	"time"

	"bookstore-api/internal/apperrors"
	"bookstore-api/internal/dto"
	"bookstore-api/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequestDTO) (string, error)
}

type authServiceImpl struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string) AuthService {
	return &authServiceImpl{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *authServiceImpl) Login(ctx context.Context, req dto.LoginRequestDTO) (string, error) {
	if err := req.Validate(); err != nil {
		return "", fmt.Errorf("%w: %v", apperrors.ErrValidation, err)
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return "", fmt.Errorf("%w: invalid email or password", apperrors.ErrUnauthorized)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", fmt.Errorf("%w: invalid email or password", apperrors.ErrUnauthorized)
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    string(user.Role),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}