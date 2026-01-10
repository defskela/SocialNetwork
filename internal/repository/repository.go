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
}

type Repository struct {
	User UserRepository
}

func NewRepository(user UserRepository) *Repository {
	return &Repository{
		User: user,
	}
}
