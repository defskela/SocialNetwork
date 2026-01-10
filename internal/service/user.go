package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/defskela/SocialNetwork/internal/entity"
	"github.com/defskela/SocialNetwork/internal/repository"
)

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) GetProfile(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	return s.repo.GetByID(ctx, userID)
}

func (s *userService) UpdateProfile(
	ctx context.Context,
	userID uuid.UUID,
	input UpdateUserInput,
) (*entity.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if input.Username != nil {
		user.Username = *input.Username
	}
	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Bio != nil {
		user.Bio = input.Bio
	}
	if input.Birthday != nil {
		t, err := time.Parse("2006-01-02", *input.Birthday)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
		user.Birthday = &t
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
