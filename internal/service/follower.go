package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/defskela/SocialNetwork/internal/entity"
	"github.com/defskela/SocialNetwork/internal/repository"
)

type followerService struct {
	repo repository.FollowerRepository
}

func NewFollowerService(repo repository.FollowerRepository) FollowerService {
	return &followerService{
		repo: repo,
	}
}

func (s *followerService) Follow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	return s.repo.Follow(ctx, followerID, followeeID)
}

func (s *followerService) Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	return s.repo.Unfollow(ctx, followerID, followeeID)
}

func (s *followerService) GetFollowers(ctx context.Context, userID uuid.UUID) ([]entity.User, error) {
	return s.repo.GetFollowers(ctx, userID)
}

func (s *followerService) GetFollowing(ctx context.Context, userID uuid.UUID) ([]entity.User, error) {
	return s.repo.GetFollowing(ctx, userID)
}
