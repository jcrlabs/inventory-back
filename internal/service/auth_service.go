package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserAuthLookup func(ctx context.Context, username string) (id uuid.UUID, passwordHash string, isAdmin bool, isActive bool, err error)

type AuthService struct {
	lookup    UserAuthLookup
	jwtSecret string
	jwtTTL    time.Duration
}

func NewAuthService(lookup UserAuthLookup, jwtSecret string, jwtTTL time.Duration) *AuthService {
	return &AuthService{lookup: lookup, jwtSecret: jwtSecret, jwtTTL: jwtTTL}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	if username == "" || password == "" {
		return "", errors.New("invalid_credentials")
	}

	uid, hash, isAdmin, isActive, err := s.lookup(ctx, username)
	if err != nil {
		return "", errors.New("invalid_credentials")
	}
	if !isActive {
		return "", errors.New("user_inactive")
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return "", errors.New("invalid_credentials")
	}

	now := time.Now()
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
