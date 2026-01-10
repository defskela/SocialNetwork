package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"social-network/internal/config"
	httpHandler "social-network/internal/delivery/http"
	"social-network/internal/repository"
	"social-network/internal/repository/postgres"
	"social-network/internal/service"
	"social-network/pkg/client/postgresql"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
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
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	err := godotenv.Load(filepath.Join(basepath, "../../.env"))
	if err != nil {
		s.T().Logf(".env file not found, using system environment")
	}

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

	s.pool, err = postgresql.NewClient(context.Background(), 3, cfg)
	s.Require().NoError(err)
}

func (s *AuthIntegrationSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *AuthIntegrationSuite) SetupTest() {
	s.privKeyPath = "../../certs/private.pem"
	s.pubKeyPath = "../../certs/public.pem"

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

func (s *AuthIntegrationSuite) POST(path string, body interface{}) (*http.Response, string) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		s.Require().NoError(err)
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest("POST", s.server.URL+path, bodyReader)
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	s.Require().NoError(err)

	respBody, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	resp.Body.Close()

	return resp, string(respBody)
}

func (s *AuthIntegrationSuite) TestRegisterAndLogin() {
	// 1. REGISTER
	username := "integration_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	email := username + "@test.com"
	password := "password123"

	registerInput := map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}

	resp, body := s.POST("/api/v1/auth/register", registerInput)

	s.Equal(http.StatusCreated, resp.StatusCode)
	s.Contains(body, `"id":`)

	// 2. LOGIN
	loginInput := map[string]string{
		"email":    email,
		"password": password,
	}

	resp, body = s.POST("/api/v1/auth/login", loginInput)

	s.Equal(http.StatusOK, resp.StatusCode)
	s.Contains(body, `"access_token":`)
}

func TestAuthIntegrationSuite(t *testing.T) {
	suite.Run(t, new(AuthIntegrationSuite))
}
