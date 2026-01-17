package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/defskela/SocialNetwork/internal/config"
	"github.com/defskela/SocialNetwork/internal/entity"
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
}

func (s *E2ESuite) SetupTest() {
	if err := godotenv.Load("../../../../.env"); err != nil {
		s.T().Log("Error loading .env file")
	}
	cfg := config.MustLoad()
	cfg.Postgres.Host = testDBHost

	privKeyPath := "../../../../" + cfg.JWT.PrivateKeyPath
	pubKeyPath := "../../../../" + cfg.JWT.PublicKeyPath

	ctx := context.Background()
	pool, err := postgresql.NewClient(ctx, 3, &cfg.Postgres)
	if err != nil {
		s.T().Skipf("Could not connect to database: %v", err)
	}
	s.pool = pool

	repo := postgres.NewUserRepository(s.pool)
	s.authService, err = service.NewAuthService(repo, time.Hour, privKeyPath, pubKeyPath)
	s.Require().NoError(err)
	s.userService = service.NewUserService(repo)

	postRepo := postgres.NewPostRepository(s.pool)
	postService := service.NewPostService(postRepo)

	services := &service.Service{Auth: s.authService, User: s.userService, Post: postService}
	s.handler = NewHandler(services)
	s.router = chi.NewRouter()
	s.handler.Init(s.router)
}

func (s *E2ESuite) TearDownTest() {
	if s.pool != nil {
		s.pool.Close()
	}
}

const testPassword = "securePass123"

func (s *E2ESuite) TestFullFlow_RegisterLoginProfile() {
	username := "e2e_user_" + time.Now().Format("20060102150405")
	email := username + "@example.com"
	password := testPassword

	id, err := s.authService.SignUp(context.Background(), service.SignUpInput{
		Username: username,
		Email:    email,
		Password: password,
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(id)

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

func (s *E2ESuite) TestPostCRUD() {
	username := "e2e_post_" + time.Now().Format("20060102150405")
	email := username + "@example.com"
	password := testPassword

	_, err := s.authService.SignUp(context.Background(), service.SignUpInput{
		Username: username,
		Email:    email,
		Password: password,
	})
	s.Require().NoError(err)

	tokens, err := s.authService.SignIn(context.Background(), service.SignInInput{
		Email:    email,
		Password: password,
	})
	s.Require().NoError(err)
	accessToken := tokens.AccessToken

	body := `{"content": "Hello via E2E"}`
	req := httptest.NewRequest("POST", "/posts", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	postID := resp["id"].(string)
	s.NotEmpty(postID)

	req = httptest.NewRequest("GET", "/posts/"+postID, http.NoBody)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)
	var post entity.Post
	json.NewDecoder(w.Body).Decode(&post)
	s.Equal("Hello via E2E", post.Content)

	body = `{"content": "Updated via E2E"}`
	req = httptest.NewRequest("PATCH", "/posts/"+postID, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)
	json.NewDecoder(w.Body).Decode(&post)
	s.Equal("Updated via E2E", post.Content)

	req = httptest.NewRequest("DELETE", "/posts/"+postID, http.NoBody)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)

	req = httptest.NewRequest("GET", "/posts/"+postID, http.NoBody)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusNotFound, w.Code)
}

func TestE2ESuite(t *testing.T) {
	suite.Run(t, new(E2ESuite))
}
