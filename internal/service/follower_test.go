package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"

	"github.com/defskela/SocialNetwork/internal/config"
	"github.com/defskela/SocialNetwork/internal/repository/postgres"
	"github.com/defskela/SocialNetwork/pkg/client/postgresql"
)

type FollowerServiceSuite struct {
	suite.Suite
	pool            *pgxpool.Pool
	followerService FollowerService
	authService     AuthService
}

func (s *FollowerServiceSuite) SetupSuite() {
	if err := godotenv.Load("../../.env"); err != nil {
		s.T().Log("Error loading .env file")
	}
	cfg := config.MustLoad()
	cfg.Postgres.Host = testDBHost

	var err error
	s.pool, err = postgresql.NewClient(context.Background(), 3, &cfg.Postgres)
	s.Require().NoError(err)
}

func (s *FollowerServiceSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *FollowerServiceSuite) SetupTest() {
	followerRepo := postgres.NewFollowerRepository(s.pool)
	s.followerService = NewFollowerService(followerRepo)

	userRepo := postgres.NewUserRepository(s.pool)
	privKeyPath := "../../certs/local/private.pem"
	pubKeyPath := "../../certs/local/public.pem"
	var err error
	s.authService, err = NewAuthService(userRepo, time.Hour, privKeyPath, pubKeyPath)
	s.Require().NoError(err)
}

func (s *FollowerServiceSuite) TestFollowUnfollow() {
	ctx := context.Background()

	user1 := SignUpInput{
		Username: "user1_" + uuid.New().String(),
		Email:    "user1_" + uuid.New().String() + "@example.com",
		Password: "password123",
	}
	id1, err := s.authService.SignUp(ctx, user1)
	s.Require().NoError(err)

	user2 := SignUpInput{
		Username: "user2_" + uuid.New().String(),
		Email:    "user2_" + uuid.New().String() + "@example.com",
		Password: "password123",
	}
	id2, err := s.authService.SignUp(ctx, user2)
	s.Require().NoError(err)

	err = s.followerService.Follow(ctx, id1, id2)
	s.NoError(err)

	following, err := s.followerService.GetFollowing(ctx, id1)
	s.NoError(err)
	s.Len(following, 1)
	s.Equal(id2, following[0].ID)

	followers, err := s.followerService.GetFollowers(ctx, id2)
	s.NoError(err)
	s.Len(followers, 1)
	s.Equal(id1, followers[0].ID)

	err = s.followerService.Unfollow(ctx, id1, id2)
	s.NoError(err)

	following, err = s.followerService.GetFollowing(ctx, id1)
	s.NoError(err)
	s.Len(following, 0)
}

func TestFollowerService(t *testing.T) {
	suite.Run(t, new(FollowerServiceSuite))
}
