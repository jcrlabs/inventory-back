package user

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	GetForLogin(ctx context.Context, username string) (id uuid.UUID, passwordHash string, isAdmin bool, isActive bool, err error)
	Create(ctx context.Context, username, password string, isAdmin bool) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	List(ctx context.Context, search string, limit, offset int) ([]User, int, error)
}
