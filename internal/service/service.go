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

type CreatePostInput struct {
	Content string `json:"content" validate:"required,min=1,max=2000" example:"Hello, world!"`
}

type UpdatePostInput struct {
	Content string `json:"content" validate:"required,min=1,max=2000" example:"Updated content"`
}

type PostService interface {
	Create(ctx context.Context, userID uuid.UUID, input CreatePostInput) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Post, error)
	Update(ctx context.Context, userID uuid.UUID, postID uuid.UUID, input UpdatePostInput) (*entity.Post, error)
	Delete(ctx context.Context, userID uuid.UUID, postID uuid.UUID) error
}

type FollowerService interface {
	Follow(ctx context.Context, followerID, followeeID uuid.UUID) error
	Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error
	GetFollowers(ctx context.Context, userID uuid.UUID) ([]entity.User, error)
	GetFollowing(ctx context.Context, userID uuid.UUID) ([]entity.User, error)
}

type Service struct {
	Auth     AuthService
	User     UserService
	Post     PostService
	Follower FollowerService
}

func NewService(repos *repository.Repository, privKeyPath, pubKeyPath string) (*Service, error) {
	authService, err := NewAuthService(repos.User, 12*time.Hour, privKeyPath, pubKeyPath)
	if err != nil {
		return nil, err
	}

	userService := NewUserService(repos.User)
	postService := NewPostService(repos.Post)
	followerService := NewFollowerService(repos.Follower)

	return &Service{
		Auth:     authService,
		User:     userService,
		Post:     postService,
		Follower: followerService,
	}, nil
}
