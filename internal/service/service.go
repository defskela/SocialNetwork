package service

import (
	"context"
	"time"

	"github.com/defskela/SocialNetwork/internal/repository"

	"github.com/google/uuid"
)

type SignUpInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignInInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
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
