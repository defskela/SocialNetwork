package service

import (
	"context"
	"testing"

	"github.com/defskela/SocialNetwork/internal/config"
	"github.com/defskela/SocialNetwork/internal/entity"
	"github.com/defskela/SocialNetwork/internal/repository"
	"github.com/defskela/SocialNetwork/internal/repository/postgres"
	"github.com/defskela/SocialNetwork/pkg/client/postgresql"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
)

type PostServiceSuite struct {
	suite.Suite
	pool        *pgxpool.Pool
	postService PostService
	userRepo    repository.UserRepository
}

func (s *PostServiceSuite) SetupSuite() {
	cfg := config.MustLoadPath("../../configs/local.yaml")
	cfg.Postgres.Host = "localhost"

	var err error
	s.pool, err = postgresql.NewClient(context.Background(), 3, &cfg.Postgres)
	s.Require().NoError(err)
}

func (s *PostServiceSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *PostServiceSuite) SetupTest() {
	s.userRepo = postgres.NewUserRepository(s.pool)
	postRepo := postgres.NewPostRepository(s.pool)
	s.postService = NewPostService(postRepo)
}

func (s *PostServiceSuite) TestCRUD() {
	ctx := context.Background()

	uniqueName := "post_tester_" + uuid.New().String()
	user := &entity.User{
		Username:     uniqueName,
		Email:        uniqueName + "@example.com",
		PasswordHash: "hash",
	}
	s.NoError(s.userRepo.Create(ctx, user))

	input := CreatePostInput{Content: "Hello World"}
	id, err := s.postService.Create(ctx, user.ID, input)
	s.Require().NoError(err)
	s.NotEqual(uuid.Nil, id)

	post, err := s.postService.GetByID(ctx, id)
	s.Require().NoError(err)
	s.Equal(input.Content, post.Content)
	s.Equal(user.ID, post.UserID)

	updateInput := UpdatePostInput{Content: "Updated World"}
	post, err = s.postService.Update(ctx, user.ID, id, updateInput)
	s.Require().NoError(err)
	s.Equal(updateInput.Content, post.Content)

	otherUserID := uuid.New()
	_, err = s.postService.Update(ctx, otherUserID, id, updateInput)
	s.Error(err)
	s.Equal("forbidden", err.Error())

	err = s.postService.Delete(ctx, otherUserID, id)
	s.Error(err)
	s.Equal("forbidden", err.Error())

	err = s.postService.Delete(ctx, user.ID, id)
	s.Require().NoError(err)

	_, err = s.postService.GetByID(ctx, id)
	s.Error(err)
}

func TestPostService(t *testing.T) {
	suite.Run(t, new(PostServiceSuite))
}
