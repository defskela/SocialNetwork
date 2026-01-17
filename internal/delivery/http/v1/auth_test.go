package v1

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
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
	testDBHost      = "localhost"
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
	if err := godotenv.Load("../../../../.env"); err != nil {
		s.T().Log("Error loading .env file")
	}
	cfg := config.MustLoad()
	cfg.Postgres.Host = testDBHost

	var err error
	s.pool, err = postgresql.NewClient(context.Background(), 3, &cfg.Postgres)
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
	uniqueID := strconv.FormatInt(time.Now().UnixNano(), 10)
	dupEmail := "dup_" + uniqueID + "@test.com"
	dupUsername := "dupUser_" + uniqueID
	_, err := s.authService.SignUp(context.Background(), service.SignUpInput{
		Username: dupUsername,
		Email:    dupEmail,
		Password: "password",
	})
	s.Require().NoError(err)

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
		{
			name:                 "Validation Error",
			inputBody:            `{"username": "us", "email": "bad-email", "password": "short"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Field validation for",
		},
		{
			name: "Duplicate Email",
			inputBody: `{"username": "dupUser2_` + uniqueID +
				`", "email": "` + dupEmail +
				`", "password": "password"}`,
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "user already exists",
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
		{
			name:                 "Validation Error",
			inputBody:            `{"email": "not-email", "password": ""}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Field validation for",
		},
		{
			name:                 "Invalid JSON",
			inputBody:            `{"email": "test"`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "unexpected EOF",
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
