package postgres

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"

	"github.com/defskela/SocialNetwork/internal/config"
	"github.com/defskela/SocialNetwork/internal/entity"
	"github.com/defskela/SocialNetwork/pkg/client/postgresql"
)

type FollowerRepoSuite struct {
	suite.Suite
	pool     *pgxpool.Pool
	repo     *followerRepository
	userRepo *userRepository
}

const testDBHost = "localhost"

func (s *FollowerRepoSuite) SetupSuite() {
	if err := godotenv.Load("../../../.env"); err != nil {
		s.T().Log("Error loading .env file")
	}
	cfg := config.MustLoad()
	cfg.Postgres.Host = testDBHost

	var err error
	s.pool, err = postgresql.NewClient(context.Background(), 3, &cfg.Postgres)
	s.Require().NoError(err)
	s.repo = &followerRepository{client: s.pool}
	s.userRepo = &userRepository{client: s.pool}
}

func (s *FollowerRepoSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *FollowerRepoSuite) createUser() *entity.User {
	user := &entity.User{
		Username:     "user_" + strconv.FormatInt(time.Now().UnixNano(), 10),
		Email:        "email_" + strconv.FormatInt(time.Now().UnixNano(), 10) + "@example.com",
		PasswordHash: "hash",
	}
	err := s.userRepo.Create(context.Background(), user)
	s.Require().NoError(err)
	return user
}

func (s *FollowerRepoSuite) TestFollowUnfollow() {
	follower := s.createUser()
	followee := s.createUser()

	// Initial Follow
	err := s.repo.Follow(context.Background(), follower.ID, followee.ID)
	s.NoError(err)

	// Verify Following
	following, err := s.repo.GetFollowing(context.Background(), follower.ID)
	s.NoError(err)
	s.Len(following, 1)
	s.Equal(followee.ID, following[0].ID)

	// Verify Followers
	followers, err := s.repo.GetFollowers(context.Background(), followee.ID)
	s.NoError(err)
	s.Len(followers, 1)
	s.Equal(follower.ID, followers[0].ID)

	// Duplicate Follow (should safely ignore or return no error due to ON CONFLICT DO NOTHING)
	err = s.repo.Follow(context.Background(), follower.ID, followee.ID)
	s.NoError(err)

	// Unfollow
	err = s.repo.Unfollow(context.Background(), follower.ID, followee.ID)
	s.NoError(err)

	// Verify Unfollow
	following, err = s.repo.GetFollowing(context.Background(), follower.ID)
	s.NoError(err)
	s.Len(following, 0)

	// Unfollow non-existent relationship
	err = s.repo.Unfollow(context.Background(), follower.ID, followee.ID)
	s.Error(err)
	s.Equal("relationship not found", err.Error())
}

func TestFollowerRepoSuite(t *testing.T) {
	suite.Run(t, new(FollowerRepoSuite))
}
