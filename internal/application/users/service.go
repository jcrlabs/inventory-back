package users

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/jonathanCaamano/inventory-back/internal/domain/user"
)

type Service struct {
	repo user.Repository
}

func New(repo user.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, username, password string, isAdmin bool) (uuid.UUID, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return uuid.Nil, errors.New("username_and_password_required")
	}
	if len(password) < 8 {
		return uuid.Nil, errors.New("password_too_short")
	}
	return s.repo.Create(ctx, username, password, isAdmin)
}

type ListResult struct {
	Items  []user.User
	Total  int
	Limit  int
	Offset int
}

func (s *Service) List(ctx context.Context, search string, limit, offset int) (ListResult, error) {
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	items, total, err := s.repo.List(ctx, strings.TrimSpace(search), limit, offset)
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{Items: items, Total: total, Limit: limit, Offset: offset}, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (user.User, error) {
	return s.repo.GetByID(ctx, id)
}
