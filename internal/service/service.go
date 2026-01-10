package service

import (
	"context"
	"time"

	"github.com/defskela/SocialNetwork/internal/entity"
	"github.com/defskela/SocialNetwork/internal/repository"

	"github.com/google/uuid"
)

type SignUpInput struct {
	Username string `json:"username" validate:"required,min=3,max=32" example:"johndoe"`
	Email    string `json:"email" validate:"required,email" example:"john@example.com"`
	Password string `json:"password" validate:"required,min=8" example:"secret123"`
}

type SignInInput struct {
	Email    string `json:"email" validate:"required,email" example:"john@example.com"`
	Password string `json:"password" validate:"required" example:"secret123"`
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthService interface {
	SignUp(ctx context.Context, input SignUpInput) (uuid.UUID, error)
	SignIn(ctx context.Context, input SignInInput) (Tokens, error)
	ParseToken(accessToken string) (uuid.UUID, error)
}

type UserService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*entity.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, input UpdateUserInput) (*entity.User, error)
}

type UpdateUserInput struct {
	Username *string `json:"username" validate:"omitempty,min=3,max=32" example:"johndoe"`
	Email    *string `json:"email" validate:"omitempty,email" example:"john@example.com"`
	Bio      *string `json:"bio" validate:"omitempty,max=500" example:"Software Engineer"`
	Birthday *string `json:"birthday" validate:"omitempty,datetime=2006-01-02" example:"2006-01-02"`
}

type Service struct {
	Auth AuthService
	User UserService
}

func NewService(repos *repository.Repository, privKeyPath, pubKeyPath string) (*Service, error) {
	authService, err := NewAuthService(repos.User, 12*time.Hour, privKeyPath, pubKeyPath)
	if err != nil {
		return nil, err
	}

	userService := NewUserService(repos.User)

	return &Service{
		Auth: authService,
		User: userService,
	}, nil
}
