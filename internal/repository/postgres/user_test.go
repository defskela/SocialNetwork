package postgres

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"

	"github.com/defskela/SocialNetwork/internal/config"
	"github.com/defskela/SocialNetwork/internal/entity"
	"github.com/defskela/SocialNetwork/pkg/client/postgresql"
)

type UserRepoSuite struct {
	suite.Suite
	pool *pgxpool.Pool
	repo *userRepository
}

func (s *UserRepoSuite) SetupSuite() {
	cfg := config.MustLoadPath("../../../configs/local.yaml")
	cfg.Postgres.Host = "localhost"

	var err error
	s.pool, err = postgresql.NewClient(context.Background(), 3, &cfg.Postgres)
	s.Require().NoError(err)
	s.repo = &userRepository{client: s.pool}
}

func (s *UserRepoSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *UserRepoSuite) TestCreate() {
	user := &entity.User{
		Username:     "test_user_create_" + strconv.FormatInt(time.Now().UnixNano(), 10),
		Email:        "test_create_" + strconv.FormatInt(time.Now().UnixNano(), 10) + "@example.com",
		PasswordHash: "hash",
	}

	err := s.repo.Create(context.Background(), user)
	s.NoError(err)
	s.NotEqual(uuid.Nil, user.ID)

	err = s.repo.Create(context.Background(), user)
	s.Error(err)
	s.Equal("user already exists", err.Error())
}

func (s *UserRepoSuite) TestGetByID() {
	user := &entity.User{
		Username:     "test_user_get_" + strconv.FormatInt(time.Now().UnixNano(), 10),
		Email:        "test_get_" + strconv.FormatInt(time.Now().UnixNano(), 10) + "@example.com",
		PasswordHash: "hash",
	}
	err := s.repo.Create(context.Background(), user)
	s.Require().NoError(err)

	foundUser, err := s.repo.GetByID(context.Background(), user.ID)
	s.NoError(err)
	s.Equal(user.Username, foundUser.Username)

	_, err = s.repo.GetByID(context.Background(), uuid.New())
	s.Error(err)
	s.Equal("user not found", err.Error())
}

func (s *UserRepoSuite) TestGetByEmail() {
	user := &entity.User{
		Username:     "test_user_email_" + strconv.FormatInt(time.Now().UnixNano(), 10),
		Email:        "test_email_" + strconv.FormatInt(time.Now().UnixNano(), 10) + "@example.com",
		PasswordHash: "hash",
	}
	err := s.repo.Create(context.Background(), user)
	s.Require().NoError(err)

	foundUser, err := s.repo.GetByEmail(context.Background(), user.Email)
	s.NoError(err)
	s.Equal(user.ID, foundUser.ID)

	_, err = s.repo.GetByEmail(context.Background(), "nonexistent@example.com")
	s.Error(err)
	s.Equal("user not found", err.Error())
}

func (s *UserRepoSuite) TestUpdate() {
	user := &entity.User{
		Username:     "test_user_update_" + strconv.FormatInt(time.Now().UnixNano(), 10),
		Email:        "test_update_" + strconv.FormatInt(time.Now().UnixNano(), 10) + "@example.com",
		PasswordHash: "hash",
	}
	err := s.repo.Create(context.Background(), user)
	s.Require().NoError(err)

	newBio := "Updated Bio"
	user.Bio = &newBio
	err = s.repo.Update(context.Background(), user)
	s.NoError(err)

	updatedUser, err := s.repo.GetByID(context.Background(), user.ID)
	s.NoError(err)
	s.NotNil(updatedUser.Bio)
	s.Equal(newBio, *updatedUser.Bio)
}

func TestUserRepoSuite(t *testing.T) {
	suite.Run(t, new(UserRepoSuite))
}
