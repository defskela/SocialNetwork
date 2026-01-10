package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/defskela/SocialNetwork/internal/entity"
	"github.com/defskela/SocialNetwork/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthClaim struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
}

type authService struct {
	userRepo   repository.UserRepository
	tokenTTL   time.Duration
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewAuthService(userRepo repository.UserRepository, tokenTTL time.Duration, privKeyPath, pubKeyPath string) (AuthService, error) {
	privBytes, err := os.ReadFile(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("could not read private key file: %w", err)
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privBytes)
	if err != nil {
		return nil, fmt.Errorf("could not parse private key: %w", err)
	}

	pubBytes, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return nil, fmt.Errorf("could not read public key file: %w", err)
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		return nil, fmt.Errorf("could not parse public key: %w", err)
	}

	return &authService{
		userRepo:   userRepo,
		tokenTTL:   tokenTTL,
		privateKey: privKey,
		publicKey:  pubKey,
	}, nil
}

func (s *authService) SignUp(ctx context.Context, input SignUpInput) (uuid.UUID, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, err
	}

	user := &entity.User{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: string(passwordHash),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return uuid.Nil, err
	}

	return user.ID, nil
}

func (s *authService) SignIn(ctx context.Context, input SignInInput) (Tokens, error) {
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return Tokens{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return Tokens{}, fmt.Errorf("invalid password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &AuthClaim{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: user.ID,
	})

	accessToken, err := token.SignedString(s.privateKey)
	if err != nil {
		return Tokens{}, err
	}

	refreshToken := make([]byte, 32)
	if _, err := rand.Read(refreshToken); err != nil {
		return Tokens{}, err
	}

	return Tokens{
		AccessToken:  accessToken,
		RefreshToken: hex.EncodeToString(refreshToken),
	}, nil
}
