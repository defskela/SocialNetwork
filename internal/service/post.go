package service

import (
	"context"
	"errors"

	"github.com/defskela/SocialNetwork/internal/entity"
	"github.com/defskela/SocialNetwork/internal/repository"

	"github.com/google/uuid"
)

type postService struct {
	repo repository.PostRepository
}

func NewPostService(repo repository.PostRepository) PostService {
	return &postService{
		repo: repo,
	}
}

func (s *postService) Create(ctx context.Context, userID uuid.UUID, input CreatePostInput) (uuid.UUID, error) {
	post := &entity.Post{
		UserID:  userID,
		Content: input.Content,
	}

	if err := s.repo.Create(ctx, post); err != nil {
		return uuid.Nil, err
	}

	return post.ID, nil
}

func (s *postService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Post, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *postService) Update(
	ctx context.Context,
	userID, postID uuid.UUID,
	input UpdatePostInput,
) (*entity.Post, error) {
	post, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	if post.UserID != userID {
		return nil, errors.New("forbidden")
	}

	post.Content = input.Content

	if err := s.repo.Update(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *postService) Delete(ctx context.Context, userID, postID uuid.UUID) error {
	post, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	if post.UserID != userID {
		return errors.New("forbidden")
	}

	return s.repo.Delete(ctx, postID)
}
