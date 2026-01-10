package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/defskela/SocialNetwork/internal/config"
	httpHandler "github.com/defskela/SocialNetwork/internal/delivery/http"
	"github.com/defskela/SocialNetwork/internal/repository"
	"github.com/defskela/SocialNetwork/internal/repository/postgres"
	"github.com/defskela/SocialNetwork/internal/service"
	"github.com/defskela/SocialNetwork/pkg/client/postgresql"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
)

type AuthIntegrationSuite struct {
	suite.Suite
	pool        *pgxpool.Pool
	server      *httptest.Server
	privKeyPath string
	pubKeyPath  string
	client      *http.Client
}

func (s *AuthIntegrationSuite) SetupSuite() {
	cfg := config.MustLoadPath("../../configs/test.yaml")

	var err error
	s.pool, err = postgresql.NewClient(context.Background(), 3, &cfg.Postgres)
	s.Require().NoError(err)
}

func (s *AuthIntegrationSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *AuthIntegrationSuite) SetupTest() {
	s.privKeyPath = "../../certs/local/private.pem"
	s.pubKeyPath = "../../certs/local/public.pem"

	repo := repository.NewRepository(postgres.NewUserRepository(s.pool))

	authService, err := service.NewAuthService(repo.User, time.Hour, s.privKeyPath, s.pubKeyPath)
	s.Require().NoError(err)

	services := &service.Service{Auth: authService}

	_ = &config.Config{
		HTTPServer: config.HTTPServer{
			Address: "localhost:8081",
			Timeout: 4 * time.Second,
		},
	}

	handler := httpHandler.NewHandler(services)
	router := handler.Init()

	s.server = httptest.NewServer(router)
	s.client = s.server.Client()
}

func (s *AuthIntegrationSuite) TearDownTest() {
	if s.server != nil {
		s.server.Close()
	}
}

func (s *AuthIntegrationSuite) POST(path string, body interface{}) (statusCode int, responseBody string) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		s.Require().NoError(err)
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", s.server.URL+path, bodyReader)
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	return resp.StatusCode, string(respBody)
}

func (s *AuthIntegrationSuite) TestRegisterAndLogin() {
	username := "integration_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	email := username + "@test.com"
	password := "password123"

	registerInput := map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}

	statusCode, body := s.POST("/api/v1/auth/register", registerInput)

	s.Equal(http.StatusCreated, statusCode)
	s.Contains(body, `"id":`)

	loginInput := map[string]string{
		"email":    email,
		"password": password,
	}

	statusCode, body = s.POST("/api/v1/auth/login", loginInput)

	s.Equal(http.StatusOK, statusCode)
	s.Contains(body, `"access_token":`)
}

func TestAuthIntegrationSuite(t *testing.T) {
	suite.Run(t, new(AuthIntegrationSuite))
}
