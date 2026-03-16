package me

import (
	"context"

	"github.com/google/uuid"

	"github.com/jonathanCaamano/inventory-back/internal/domain/group"
	"github.com/jonathanCaamano/inventory-back/internal/domain/user"
)

type Service struct {
	users  user.Repository
	groups group.Repository
}

func New(users user.Repository, groups group.Repository) *Service {
	return &Service{users: users, groups: groups}
}

type GroupRole struct {
	Group group.Group
	Role  group.Role
}

type Profile struct {
	User   user.User
	Groups []GroupRole
}

func (s *Service) WhoAmI(ctx context.Context, userID uuid.UUID, isAdmin bool) (Profile, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return Profile{}, err
	}
	gs, err := s.groups.ListForUser(ctx, userID, isAdmin)
	if err != nil {
		return Profile{}, err
	}
	out := make([]GroupRole, 0, len(gs))
	for _, g := range gs {
		r, err := s.groups.RoleForUser(ctx, userID, g.ID)
		if err != nil {
			if isAdmin {
				out = append(out, GroupRole{Group: g, Role: group.RoleWriter})
				continue
			}
			return Profile{}, err
		}
		out = append(out, GroupRole{Group: g, Role: r})
	}
	return Profile{User: u, Groups: out}, nil
}
