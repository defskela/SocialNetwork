package service

import (
	"context"
	"time"

	"social-network/internal/repository"

	"github.com/google/uuid"
)

type SignUpInput struct {
	Username string
	Email    string
	Password string
}

type SignInInput struct {
	Email    string
	Password string
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type AuthService interface {
	SignUp(ctx context.Context, input SignUpInput) (uuid.UUID, error)
	SignIn(ctx context.Context, input SignInInput) (Tokens, error)
}

type Service struct {
	Auth AuthService
}

func NewService(repos *repository.Repository, privKeyPath, pubKeyPath string) (*Service, error) {
	authService, err := NewAuthService(repos.User, 12*time.Hour, privKeyPath, pubKeyPath)
	if err != nil {
		return nil, err
	}
	return &Service{
		Auth: authService,
	}, nil
}
