package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

type UserService struct {
	repo UserRepo
}

type UserRepo interface {
	CreateUser(ctx context.Context, username, password string, isAdmin bool) (uuid.UUID, error)
}

func NewUserService(repo UserRepo) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, username, password string, isAdmin bool) (uuid.UUID, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return uuid.Nil, errors.New("username_and_password_required")
	}
	if len(password) < 8 {
		return uuid.Nil, errors.New("password_too_short")
	}
	return s.repo.CreateUser(ctx, username, password, isAdmin)
}
