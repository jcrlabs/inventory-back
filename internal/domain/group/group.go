package group

import "github.com/google/uuid"

type Group struct {
	ID   uuid.UUID
	Slug string
	Name string
}

type Role string

const (
	RoleReader Role = "reader"
	RoleWriter Role = "writer"
)

func (r Role) AllowsWrite() bool { return r == RoleWriter }
func (r Role) AllowsRead() bool  { return r == RoleReader || r == RoleWriter }
