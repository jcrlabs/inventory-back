package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/jonathanCaamano/inventory-back/internal/domain/user"
)

type Service struct {
	repo      user.Repository
	jwtSecret string
	jwtTTL    time.Duration
}

func New(repo user.Repository, jwtSecret string, jwtTTL time.Duration) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret, jwtTTL: jwtTTL}
}

func (s *Service) Login(ctx context.Context, username, password string) (string, error) {
	if username == "" || password == "" {
		return "", errors.New("invalid_credentials")
	}

	uid, hash, isAdmin, isActive, err := s.repo.GetForLogin(ctx, username)
	if err != nil {
		return "", errors.New("invalid_credentials")
	}
	if !isActive {
		return "", errors.New("user_inactive")
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return "", errors.New("invalid_credentials")
	}

	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"sub":   uid.String(),
		"admin": isAdmin,
		"iat":   now.Unix(),
		"exp":   now.Add(s.jwtTTL).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", errors.New("token_error")
	}
	return signed, nil
}
