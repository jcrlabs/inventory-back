package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupRepo struct {
	pool *pgxpool.Pool
}

type GroupRow struct {
	ID        uuid.UUID `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type GroupMemberRow struct {
	UserID    uuid.UUID `json:"user_id"`
	GroupID   uuid.UUID `json:"group_id"`
	Role      string    `json:"role"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

func NewGroupRepo(pool *pgxpool.Pool) *GroupRepo {
	return &GroupRepo{pool: pool}
}

func (r *GroupRepo) Create(ctx context.Context, slug, name string) (GroupRow, error) {
	var g GroupRow
	err := r.pool.QueryRow(ctx, `INSERT INTO groups (slug, name) VALUES ($1,$2) RETURNING id, slug, name, created_at`, slug, name).Scan(&g.ID, &g.Slug, &g.Name, &g.CreatedAt)
	return g, err
}

func (r *GroupRepo) ListForUser(ctx context.Context, userID uuid.UUID, isAdmin bool) ([]GroupRow, error) {
	rows, err := func() (pgx.Rows, error) {
		if isAdmin {
			return r.pool.Query(ctx, `SELECT id, slug, name, created_at FROM groups ORDER BY name ASC`)
		}
		return r.pool.Query(ctx, `
			SELECT g.id, g.slug, g.name, g.created_at
			FROM groups g
			JOIN group_memberships gm ON gm.group_id=g.id
			WHERE gm.user_id=$1
			ORDER BY g.name ASC`, userID)
	}()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []GroupRow{}
	for rows.Next() {
		var g GroupRow
		if err := rows.Scan(&g.ID, &g.Slug, &g.Name, &g.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, nil
}

func (r *GroupRepo) AddMember(ctx context.Context, groupID, userID uuid.UUID, role string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO group_memberships (user_id, group_id, role) VALUES ($1,$2,$3)
		ON CONFLICT (user_id, group_id) DO UPDATE SET role=EXCLUDED.role`, userID, groupID, role)
	return err
}

func (r *GroupRepo) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM group_memberships WHERE user_id=$1 AND group_id=$2`, userID, groupID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("not_found")
	}
	return nil
}

func (r *GroupRepo) GetUserRoleInGroup(ctx context.Context, userID, groupID uuid.UUID) (string, error) {
	var role string
	err := r.pool.QueryRow(ctx, `SELECT role FROM group_memberships WHERE user_id=$1 AND group_id=$2`, userID, groupID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("not_member")
		}
		return "", err
	}
	return role, nil
}

func (r *GroupRepo) FindUserIDByUsername(ctx context.Context, username string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT id FROM users WHERE username=$1`, username).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, errors.New("not_found")
		}
		return uuid.Nil, err
	}
	return id, nil
}

func (r *GroupRepo) ListMembers(ctx context.Context, groupID uuid.UUID) ([]GroupMemberRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT gm.user_id, gm.group_id, gm.role, u.username, gm.created_at
		FROM group_memberships gm
		JOIN users u ON u.id=gm.user_id
		WHERE gm.group_id=$1
		ORDER BY u.username ASC`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []GroupMemberRow{}
	for rows.Next() {
		var m GroupMemberRow
		if err := rows.Scan(&m.UserID, &m.GroupID, &m.Role, &m.Username, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}
