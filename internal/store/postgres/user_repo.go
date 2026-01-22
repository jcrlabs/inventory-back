package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) GetUserForLogin(ctx context.Context, username string) (id uuid.UUID, passwordHash string, isAdmin bool, isActive bool, err error) {
	err = r.pool.QueryRow(ctx, `SELECT id, password_hash, is_admin, is_active FROM users WHERE username=$1`, username).Scan(&id, &passwordHash, &isAdmin, &isActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, "", false, false, err
		}
		return uuid.Nil, "", false, false, err
	}
	return id, passwordHash, isAdmin, isActive, nil
}

func (r *UserRepo) EnsureBootstrapAdmin(ctx context.Context, username, password string) error {
	if username == "" || password == "" {
		return nil
	}
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE is_admin=true)`).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `INSERT INTO users (username, password_hash, is_admin, is_active) VALUES ($1,$2,true,true)`, username, string(hash))
	return err
}

func (r *UserRepo) CreateUser(ctx context.Context, username, password string, isAdmin bool) (uuid.UUID, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, err
	}
	var id uuid.UUID
	err = r.pool.QueryRow(ctx, `INSERT INTO users (username, password_hash, is_admin, is_active) VALUES ($1,$2,$3,true) RETURNING id`, username, string(hash), isAdmin).Scan(&id)
	return id, err
}
