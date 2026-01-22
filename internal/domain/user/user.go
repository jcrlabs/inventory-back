package user

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID
	Username string
	IsAdmin  bool
	IsActive bool
}
