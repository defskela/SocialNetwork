package v1

import (
	"bytes"
	"context"
	"fmt"
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

type ProfileHandlerSuite struct {
	suite.Suite
	pool        *pgxpool.Pool
	handler     *Handler
	authService service.AuthService
	userService service.UserService
	privKeyPath string
	pubKeyPath  string
	router      *chi.Mux
}

func (s *ProfileHandlerSuite) SetupSuite() {
	_ = godotenv.Load("../../../../.env")

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

func (s *ProfileHandlerSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *ProfileHandlerSuite) SetupTest() {
	s.privKeyPath = testPrivKeyPath
	s.pubKeyPath = testPubKeyPath

	repo := postgres.NewUserRepository(s.pool)

	var err error
	s.authService, err = service.NewAuthService(repo, time.Hour, s.privKeyPath, s.pubKeyPath)
	s.Require().NoError(err)

	s.userService = service.NewUserService(repo)

	services := &service.Service{Auth: s.authService, User: s.userService}
	s.handler = NewHandler(services)

	s.router = chi.NewRouter()
	s.handler.Init(s.router)
}

func (s *ProfileHandlerSuite) createAndLoginUser() (token string) {
	uniqueIdx := strconv.FormatInt(time.Now().UnixNano(), 10)
	username := "user_" + uniqueIdx
	email := "user_" + uniqueIdx + "@example.com"
	password := "password123"

	id, err := s.authService.SignUp(context.Background(), service.SignUpInput{
		Username: username,
		Email:    email,
		Password: password,
	})
	s.Require().NoError(err)
	s.Require().NotNil(id)

	tokens, err := s.authService.SignIn(context.Background(), service.SignInInput{
		Email:    email,
		Password: password,
	})
	s.Require().NoError(err)

	return tokens.AccessToken
}

func (s *ProfileHandlerSuite) TestGetProfile() {
	token := s.createAndLoginUser()

	req := httptest.NewRequest("GET", "/users/me", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *ProfileHandlerSuite) TestUpdateProfile() {
	token := s.createAndLoginUser()

	newBio := "My new bio"
	body := fmt.Sprintf(`{"bio": %q}`, newBio)
	req := httptest.NewRequest("PATCH", "/users/me", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	reqGet := httptest.NewRequest("GET", "/users/me", http.NoBody)
	reqGet.Header.Set("Authorization", "Bearer "+token)
	wGet := httptest.NewRecorder()

	s.router.ServeHTTP(wGet, reqGet)
	s.Equal(http.StatusOK, wGet.Code)
	s.Contains(wGet.Body.String(), newBio)
}

func TestProfileHandlerSuite(t *testing.T) {
	suite.Run(t, new(ProfileHandlerSuite))
}
