package repository

import (
	"context"

	"github.com/defskela/SocialNetwork/internal/entity"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
}

type PostRepository interface {
	Create(ctx context.Context, post *entity.Post) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Post, error)
	Update(ctx context.Context, post *entity.Post) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type Repository struct {
	User UserRepository
	Post PostRepository
}

func NewRepository(user UserRepository, post PostRepository) *Repository {
	return &Repository{
		User: user,
		Post: post,
	}
}
