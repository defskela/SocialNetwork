package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/defskela/SocialNetwork/internal/config"
	"github.com/defskela/SocialNetwork/internal/entity"
	"github.com/defskela/SocialNetwork/internal/repository/postgres"
	"github.com/defskela/SocialNetwork/internal/service"
	"github.com/defskela/SocialNetwork/pkg/client/postgresql"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type FollowerHandlerSuite struct {
	suite.Suite
	pool            *pgxpool.Pool
	handler         *Handler
	authService     service.AuthService
	followerService service.FollowerService
	privKeyPath     string
	pubKeyPath      string
	router          *chi.Mux
}

func (s *FollowerHandlerSuite) SetupSuite() {
	if err := godotenv.Load("../../../../.env"); err != nil {
		s.T().Log("Error loading .env file")
	}
	cfg := config.MustLoad()
	cfg.Postgres.Host = testDBHost

	var err error
	s.pool, err = postgresql.NewClient(context.Background(), 3, &cfg.Postgres)
	s.Require().NoError(err)
}

func (s *FollowerHandlerSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *FollowerHandlerSuite) SetupTest() {
	s.privKeyPath = testPrivKeyPath
	s.pubKeyPath = testPubKeyPath

	userRepo := postgres.NewUserRepository(s.pool)
	followerRepo := postgres.NewFollowerRepository(s.pool)

	var err error
	s.authService, err = service.NewAuthService(userRepo, time.Hour, s.privKeyPath, s.pubKeyPath)
	s.Require().NoError(err)
	s.followerService = service.NewFollowerService(followerRepo)

	services := &service.Service{
		Auth:     s.authService,
		Follower: s.followerService,
	}
	s.handler = NewHandler(services)

	s.router = chi.NewRouter()
	s.handler.Init(s.router)
}

const (
	followerPassword = "password123"
)

func (s *FollowerHandlerSuite) createAndLoginUser() (string, uuid.UUID) {
	uniqueIdx := strconv.FormatInt(time.Now().UnixNano(), 10)
	username := "user_" + uniqueIdx
	email := "user_" + uniqueIdx + "@example.com"
	password := followerPassword

	id, err := s.authService.SignUp(context.Background(), service.SignUpInput{
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

	return tokens.AccessToken, id
}

func (s *FollowerHandlerSuite) createUser() uuid.UUID {
	uniqueIdx := strconv.FormatInt(time.Now().UnixNano(), 10)
	username := "target_" + uniqueIdx
	email := "target_" + uniqueIdx + "@example.com"
	password := followerPassword

	id, err := s.authService.SignUp(context.Background(), service.SignUpInput{
		Username: username,
		Email:    email,
		Password: password,
	})
	s.Require().NoError(err)
	return id
}

func (s *FollowerHandlerSuite) TestFollow() {
	token, _ := s.createAndLoginUser()
	targetID := s.createUser()

	req := httptest.NewRequest(http.MethodPost, "/users/"+targetID.String()+"/follow", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	s.router.ServeHTTP(rr, req)

	s.Equal(http.StatusOK, rr.Code)

	// Test Follow Yourself
	token2, id2 := s.createAndLoginUser()
	reqSelf := httptest.NewRequest(http.MethodPost, "/users/"+id2.String()+"/follow", http.NoBody)
	reqSelf.Header.Set("Authorization", "Bearer "+token2)
	rrSelf := httptest.NewRecorder()
	s.router.ServeHTTP(rrSelf, reqSelf)
	s.Equal(http.StatusBadRequest, rrSelf.Code)
}

func (s *FollowerHandlerSuite) TestUnfollow() {
	token, followerID := s.createAndLoginUser()
	targetID := s.createUser()

	// Follow first
	err := s.followerService.Follow(context.Background(), followerID, targetID)
	s.Require().NoError(err)

	// Unfollow
	req := httptest.NewRequest(http.MethodDelete, "/users/"+targetID.String()+"/follow", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	s.router.ServeHTTP(rr, req)

	s.Equal(http.StatusOK, rr.Code)

	// Unfollow again (Not Found)
	req2 := httptest.NewRequest(http.MethodDelete, "/users/"+targetID.String()+"/follow", http.NoBody)
	req2.Header.Set("Authorization", "Bearer "+token)
	rr2 := httptest.NewRecorder()
	s.router.ServeHTTP(rr2, req2)
	s.Equal(http.StatusNotFound, rr2.Code)
}

func (s *FollowerHandlerSuite) TestGetFollowers() {
	token, followerID := s.createAndLoginUser()
	targetID := s.createUser()

	// Follow
	err := s.followerService.Follow(context.Background(), followerID, targetID)
	s.Require().NoError(err)

	// Get Followers of target
	req := httptest.NewRequest(http.MethodGet, "/users/"+targetID.String()+"/followers", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token) // Auth required? Yes
	rr := httptest.NewRecorder()

	s.router.ServeHTTP(rr, req)

	s.Equal(http.StatusOK, rr.Code)
	var users []entity.User
	err = json.Unmarshal(rr.Body.Bytes(), &users)
	s.NoError(err)
	s.Len(users, 1)
	s.Equal(followerID, users[0].ID)
}

func (s *FollowerHandlerSuite) TestGetFollowing() {
	token, followerID := s.createAndLoginUser()
	targetID := s.createUser()

	// Follow
	err := s.followerService.Follow(context.Background(), followerID, targetID)
	s.Require().NoError(err)

	// Get Following of follower
	req := httptest.NewRequest(http.MethodGet, "/users/"+followerID.String()+"/following", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	s.router.ServeHTTP(rr, req)

	s.Equal(http.StatusOK, rr.Code)
	var users []entity.User
	err = json.Unmarshal(rr.Body.Bytes(), &users)
	s.NoError(err)
	s.Len(users, 1)
	s.Equal(targetID, users[0].ID)
}

func TestFollowerHandlerSuite(t *testing.T) {
	suite.Run(t, new(FollowerHandlerSuite))
}
