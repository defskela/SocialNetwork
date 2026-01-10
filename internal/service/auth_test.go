package service

import (
	"context"
	"os"
	"strconv"
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

func (s *AuthServiceVerifySuite) SetupSuite() {
	_ = godotenv.Load("../../.env")

	requiredEnv := []string{"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB"}
	for _, env := range requiredEnv {
		if os.Getenv(env) == "" {
			s.T().Skipf("Skipping test: %s is not set", env)
		}
	}

	port, _ := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
	cfg := config.Postgres{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     port,
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB"),
		SSLMode:  "disable",
	}

	var err error
	s.pool, err = postgresql.NewClient(context.Background(), 3, &cfg)
	s.Require().NoError(err)
}

func (s *AuthServiceVerifySuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *AuthServiceVerifySuite) SetupTest() {
	s.privKeyPath = "../../certs/local/private.pem"
	s.pubKeyPath = "../../certs/local/public.pem"

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
