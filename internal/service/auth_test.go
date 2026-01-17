package service

import (
	"context"
	"testing"
	"time"

	"github.com/defskela/SocialNetwork/internal/config"
	"github.com/defskela/SocialNetwork/internal/repository/postgres"
	"github.com/defskela/SocialNetwork/pkg/client/postgresql"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type AuthServiceVerifySuite struct {
	suite.Suite
	pool        *pgxpool.Pool
	authService AuthService
	privKeyPath string
	pubKeyPath  string
}

const (
	testDBHost  = "localhost"
	privKeyPath = "../../certs/local/private.pem"
	pubKeyPath  = "../../certs/local/public.pem"
)

func (s *AuthServiceVerifySuite) SetupSuite() {
	if err := godotenv.Load("../../.env"); err != nil {
		s.T().Log("Error loading .env file")
	}
	cfg := config.MustLoad()
	cfg.Postgres.Host = testDBHost

	var err error
	s.pool, err = postgresql.NewClient(context.Background(), 3, &cfg.Postgres)
	s.Require().NoError(err)
}

func (s *AuthServiceVerifySuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *AuthServiceVerifySuite) SetupTest() {
	s.privKeyPath = privKeyPath
	s.pubKeyPath = pubKeyPath

	authRepo := postgres.NewUserRepository(s.pool)
	var err error
	s.authService, err = NewAuthService(authRepo, time.Hour, s.privKeyPath, s.pubKeyPath)
	s.Require().NoError(err)
}

func (s *AuthServiceVerifySuite) TestSignUp() {
	input := SignUpInput{
		Username: "testuser_" + uuid.New().String(),
		Email:    "test_" + uuid.New().String() + "@example.com",
		Password: "password123",
	}

	id, err := s.authService.SignUp(context.Background(), input)

	s.NoError(err)
	s.NotEqual(uuid.Nil, id)

	repo := postgres.NewUserRepository(s.pool)
	user, err := repo.GetByID(context.Background(), id)
	s.NoError(err)
	s.Equal(input.Email, user.Email)
}

func (s *AuthServiceVerifySuite) TestSignIn() {
	input := SignUpInput{
		Username: "testuser_" + uuid.New().String(),
		Email:    "test_" + uuid.New().String() + "@example.com",
		Password: "password123",
	}

	_, err := s.authService.SignUp(context.Background(), input)
	s.Require().NoError(err)

	signInInput := SignInInput{
		Email:    input.Email,
		Password: input.Password,
	}

	tokens, err := s.authService.SignIn(context.Background(), signInInput)

	s.NoError(err)
	s.NotEmpty(tokens.AccessToken)
}

func TestAuthServiceVerifySuite(t *testing.T) {
	suite.Run(t, new(AuthServiceVerifySuite))
}
