package v1

import (
	"bytes"
	"context"
	"fmt"
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
	"github.com/google/uuid"
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
	if err := godotenv.Load("../../../../.env"); err != nil {
		s.T().Log("Error loading .env file")
	}
	cfg := config.MustLoad()
	cfg.Postgres.Host = testDBHost

	var err error
	s.pool, err = postgresql.NewClient(context.Background(), 3, &cfg.Postgres)
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

func (s *ProfileHandlerSuite) createAndLoginUser() (string, uuid.UUID) {
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
	s.Require().NotEqual(uuid.Nil, id)

	tokens, err := s.authService.SignIn(context.Background(), service.SignInInput{
		Email:    email,
		Password: password,
	})
	s.Require().NoError(err)

	return tokens.AccessToken, id
}

func (s *ProfileHandlerSuite) TestGetProfile() {
	token, _ := s.createAndLoginUser()

	req := httptest.NewRequest("GET", "/users/me", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *ProfileHandlerSuite) TestGetProfile_UserNotFound() {
	token, id := s.createAndLoginUser()

	_, err := s.pool.Exec(context.Background(), "DELETE FROM social.users WHERE id = $1", id)
	s.Require().NoError(err)

	req := httptest.NewRequest("GET", "/users/me", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), "user not found")
}

func (s *ProfileHandlerSuite) TestUpdateProfile() {
	token, _ := s.createAndLoginUser()

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

func (s *ProfileHandlerSuite) TestUpdateProfile_InvalidInput() {
	token, _ := s.createAndLoginUser()

	body := `{"birthday": "invalid-date"}`
	req := httptest.NewRequest("PATCH", "/users/me", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ProfileHandlerSuite) TestUpdateProfile_InvalidJSON() {
	token, _ := s.createAndLoginUser()

	body := `{"bio": "unclosed_json`
	req := httptest.NewRequest("PATCH", "/users/me", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ProfileHandlerSuite) TestUpdateProfile_UserNotFound() {
	token, id := s.createAndLoginUser()

	_, err := s.pool.Exec(context.Background(), "DELETE FROM social.users WHERE id = $1", id)
	s.Require().NoError(err)

	body := `{"bio": "new bio"}`
	req := httptest.NewRequest("PATCH", "/users/me", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), "user not found")
}

func (s *ProfileHandlerSuite) TestGetProfile_ContextMissingUserID() {
	req := httptest.NewRequest("GET", "/users/me", http.NoBody)
	w := httptest.NewRecorder()

	s.handler.getProfile(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), "user id not found")
}

func (s *ProfileHandlerSuite) TestUpdateProfile_ContextMissingUserID() {
	req := httptest.NewRequest("PATCH", "/users/me", http.NoBody)
	w := httptest.NewRecorder()

	s.handler.updateProfile(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), "user id not found")
}

func (s *ProfileHandlerSuite) TestAuthMiddleware_Errors() {
	req := httptest.NewRequest("GET", "/users/me", http.NoBody)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusUnauthorized, w.Code)
	s.Contains(w.Body.String(), "empty auth header")

	req = httptest.NewRequest("GET", "/users/me", http.NoBody)
	req.Header.Set("Authorization", "Basic user:pass")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusUnauthorized, w.Code)
	s.Contains(w.Body.String(), "invalid auth header")

	req = httptest.NewRequest("GET", "/users/me", http.NoBody)
	req.Header.Set("Authorization", "Bearer")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusUnauthorized, w.Code)
	s.Contains(w.Body.String(), "invalid auth header")

	req = httptest.NewRequest("GET", "/users/me", http.NoBody)
	req.Header.Set("Authorization", "Bearer ")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusUnauthorized, w.Code)
	s.Contains(w.Body.String(), "token is empty")

	req = httptest.NewRequest("GET", "/users/me", http.NoBody)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusUnauthorized, w.Code)
	s.Contains(w.Body.String(), "token is malformed")
}

func TestProfileHandlerSuite(t *testing.T) {
	suite.Run(t, new(ProfileHandlerSuite))
}
