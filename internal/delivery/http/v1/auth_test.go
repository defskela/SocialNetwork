package v1

import (
	"bytes"
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

const (
	testPrivKeyPath = "../../../../certs/local/private.pem"
	testPubKeyPath  = "../../../../certs/local/public.pem"
)

type AuthHandlerSuite struct {
	suite.Suite
	pool        *pgxpool.Pool
	handler     *Handler
	authService service.AuthService
	privKeyPath string
	pubKeyPath  string
	router      *chi.Mux
}

func (s *AuthHandlerSuite) SetupSuite() {
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

func (s *AuthHandlerSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *AuthHandlerSuite) SetupTest() {
	s.privKeyPath = testPrivKeyPath
	s.pubKeyPath = testPubKeyPath

	repo := postgres.NewUserRepository(s.pool)

	var err error
	s.authService, err = service.NewAuthService(repo, time.Hour, s.privKeyPath, s.pubKeyPath)
	s.Require().NoError(err)

	services := &service.Service{Auth: s.authService}
	s.handler = NewHandler(services)

	s.router = chi.NewRouter()
	s.handler.Init(s.router)
}

func (s *AuthHandlerSuite) TestSignUp() {
	tests := []struct {
		name                 string
		inputBody            string
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "OK",
			inputBody: `{"username": "testhttp_` + strconv.FormatInt(time.Now().UnixNano(), 10) +
				`", "email": "testhttp_` + strconv.FormatInt(time.Now().UnixNano(), 10) +
				`@test.com", "password": "password"}`,
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `{"id":`,
		},
		{
			name:                 "Invalid Input",
			inputBody:            `{"username": "test"`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "unexpected EOF",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/auth/register", bytes.NewBufferString(tt.inputBody))

			s.router.ServeHTTP(w, req)

			s.Equal(tt.expectedStatusCode, w.Code)
			s.Contains(w.Body.String(), tt.expectedResponseBody)
		})
	}
}

func (s *AuthHandlerSuite) TestSignIn() {
	signUpInput := service.SignUpInput{
		Username: "testlogin_" + strconv.FormatInt(time.Now().UnixNano(), 10),
		Email:    "testlogin_" + strconv.FormatInt(time.Now().UnixNano(), 10) + "@test.com",
		Password: "password",
	}

	_, err := s.authService.SignUp(context.Background(), signUpInput)
	s.Require().NoError(err)

	tests := []struct {
		name                 string
		inputBody            string
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:                 "OK",
			inputBody:            `{"email": "` + signUpInput.Email + `", "password": "password"}`,
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `"access_token":`,
		},
		{
			name:                 "Invalid Credentials",
			inputBody:            `{"email": "` + signUpInput.Email + `", "password": "wrongpassword"}`,
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "invalid password",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString(tt.inputBody))

			s.router.ServeHTTP(w, req)

			s.Equal(tt.expectedStatusCode, w.Code)
			s.Contains(w.Body.String(), tt.expectedResponseBody)
		})
	}
}

func TestAuthHandlerSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerSuite))
}
