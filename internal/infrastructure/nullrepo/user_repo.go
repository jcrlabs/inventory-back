package nullrepo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jonathanCaamano/inventory-back/internal/domain/user"
)

type UserRepo struct{}

func NewUserRepo() *UserRepo { return &UserRepo{} }

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (user.User, error) {
	return user.User{}, errors.New("storage_not_configured")
}
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	return nil, errors.New("storage_not_configured")
}
func (r *UserRepo) Create(ctx context.Context, username, password string, isAdmin bool) (uuid.UUID, error) {
	return uuid.Nil, errors.New("storage_not_configured")
}

func (r *UserRepo) EnsureBootstrapAdmin(ctx context.Context, username, passwordHash string) error {
	return errors.New("storage_not_configured")
}

func (r *UserRepo) GetForLogin(ctx context.Context, username string) (id uuid.UUID, passwordHash string, isAdmin bool, isActive bool, err error) {
	return uuid.Nil, "", false, false, errors.New("storage_not_configured")
}
func (r *UserRepo) List(ctx context.Context, search string, limit, offset int) ([]user.User, int, error) {
	return nil, 0, errors.New("storage_not_configured")
}
