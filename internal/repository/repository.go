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

type FollowerRepository interface {
	Follow(ctx context.Context, followerID, followeeID uuid.UUID) error
	Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error
	GetFollowers(ctx context.Context, userID uuid.UUID) ([]entity.User, error)
	GetFollowing(ctx context.Context, userID uuid.UUID) ([]entity.User, error)
}

type Repository struct {
	User     UserRepository
	Post     PostRepository
	Follower FollowerRepository
}

func NewRepository(user UserRepository, post PostRepository, follower FollowerRepository) *Repository {
	return &Repository{
		User:     user,
		Post:     post,
		Follower: follower,
	}
}
