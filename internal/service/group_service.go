package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	GroupRoleReader = "reader"
	GroupRoleWriter = "writer"
)

type GroupService struct {
	repo GroupRepo
}

type GroupRepo interface {
	ListForUser(ctx context.Context, userID uuid.UUID, isAdmin bool) ([]GroupRow, error)
	Create(ctx context.Context, slug, name string) (GroupRow, error)
	AddMember(ctx context.Context, groupID, userID uuid.UUID, role string) error
	ListMembers(ctx context.Context, groupID uuid.UUID) ([]GroupMemberRow, error)
	RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error
	GetUserRoleInGroup(ctx context.Context, userID, groupID uuid.UUID) (string, error)
	FindUserIDByUsername(ctx context.Context, username string) (uuid.UUID, error)
}

type GroupRow struct {
	ID        uuid.UUID
	Slug      string
	Name      string
	CreatedAt time.Time
}

type GroupMemberRow struct {
	UserID    uuid.UUID
	GroupID   uuid.UUID
	Role      string
	Username  string
	CreatedAt time.Time
}

func NewGroupService(repo GroupRepo) *GroupService {
	return &GroupService{repo: repo}
}

func (s *GroupService) ListForUser(ctx context.Context, userID uuid.UUID, isAdmin bool) ([]GroupRow, error) {
	return s.repo.ListForUser(ctx, userID, isAdmin)
}

func (s *GroupService) Create(ctx context.Context, slug, name string) (GroupRow, error) {
	slug = strings.TrimSpace(slug)
	name = strings.TrimSpace(name)
	if slug == "" || name == "" {
		return GroupRow{}, errors.New("slug_and_name_required")
	}
	return s.repo.Create(ctx, slug, name)
}

func (s *GroupService) AddMemberByUsername(ctx context.Context, groupID uuid.UUID, username, role string) error {
	username = strings.TrimSpace(username)
	role = strings.TrimSpace(role)
	if username == "" {
		return errors.New("username_required")
	}
	if role != GroupRoleReader && role != GroupRoleWriter {
		return errors.New("invalid_role")
	}
	uid, err := s.repo.FindUserIDByUsername(ctx, username)
	if err != nil {
		return err
	}
	return s.repo.AddMember(ctx, groupID, uid, role)
}

func (s *GroupService) ListMembers(ctx context.Context, groupID uuid.UUID) ([]GroupMemberRow, error) {
	return s.repo.ListMembers(ctx, groupID)
}

func (s *GroupService) RemoveMemberByUsername(ctx context.Context, groupID uuid.UUID, username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return errors.New("username_required")
	}
	uid, err := s.repo.FindUserIDByUsername(ctx, username)
	if err != nil {
		return err
	}
	return s.repo.RemoveMember(ctx, groupID, uid)
}

func (s *GroupService) UserRole(ctx context.Context, userID, groupID uuid.UUID) (string, error) {
	return s.repo.GetUserRoleInGroup(ctx, userID, groupID)
}
