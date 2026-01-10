package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/defskela/SocialNetwork/internal/config"
	"github.com/defskela/SocialNetwork/internal/repository/postgres"
	"github.com/defskela/SocialNetwork/internal/service"
	"github.com/defskela/SocialNetwork/pkg/client/postgresql"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type E2ESuite struct {
	suite.Suite
	pool        *pgxpool.Pool
	handler     *Handler
	authService service.AuthService
	userService service.UserService
	router      *chi.Mux
}

func (s *E2ESuite) SetupSuite() {
	_ = godotenv.Load("../../../../.env")

	port, _ := os.LookupEnv("POSTGRES_PORT")
	if port == "" {
		s.T().Log("POSTGRES_PORT not set, assuming CI or local env already setup")
	}
}

func (s *E2ESuite) SetupTest() {
	privKeyPath := "../../../../certs/local/private.pem"
	pubKeyPath := "../../../../certs/local/public.pem"

	cfg := config.Postgres{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     5432,
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB"),
		SSLMode:  "disable",
	}
	if p := os.Getenv("POSTGRES_PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			cfg.Port = val
		}
	}

	ctx := context.Background()
	pool, err := postgresql.NewClient(ctx, 3, &cfg)
	if err != nil {
		s.T().Skipf("Could not connect to database: %v", err)
	}
	s.pool = pool

	repo := postgres.NewUserRepository(s.pool)
	s.authService, err = service.NewAuthService(repo, time.Hour, privKeyPath, pubKeyPath)
	s.Require().NoError(err)
	s.userService = service.NewUserService(repo)

	services := &service.Service{Auth: s.authService, User: s.userService}
	s.handler = NewHandler(services)
	s.router = chi.NewRouter()
	s.handler.Init(s.router)
}

func (s *E2ESuite) TearDownTest() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *E2ESuite) TestFullFlow_RegisterLoginProfile() {
	username := "e2e_user_" + time.Now().Format("20060102150405")
	email := username + "@example.com"
	password := "securePass123"

	id, err := s.authService.SignUp(context.Background(), service.SignUpInput{
		Username: username,
		Email:    email,
		Password: password,
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(id)

	// 2. Login
	tokens, err := s.authService.SignIn(context.Background(), service.SignInInput{
		Email:    email,
		Password: password,
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(tokens.AccessToken)

	accessToken := tokens.AccessToken

	req := httptest.NewRequest("GET", "/users/me", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)

	reqInvalid := httptest.NewRequest("GET", "/users/me", http.NoBody)
	reqInvalid.Header.Set("Authorization", "Bearer "+accessToken+"\"")
	wInvalid := httptest.NewRecorder()
	s.router.ServeHTTP(wInvalid, reqInvalid)

	s.Equal(http.StatusUnauthorized, wInvalid.Code)
	s.Contains(wInvalid.Body.String(), "token is malformed")
}

func TestE2ESuite(t *testing.T) {
	suite.Run(t, new(E2ESuite))
}
