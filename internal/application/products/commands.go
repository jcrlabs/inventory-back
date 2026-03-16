package products

import (
	"time"

	"github.com/google/uuid"
)

type CreateCommand struct {
	GroupID      uuid.UUID
	Name         string
	Description  string
	EntryDate    time.Time
	ExitDate     *time.Time
	Status       string
	Paid         bool
	Price        float64
	Observations string
	Contact      *ContactCommand
	Images       []ImageCommand
}

type ContactCommand struct {
	FirstName   string
	LastName    string
	PhoneNumber string
}

type ImageCommand struct {
	URL      string
	Position int
}

type UpdateCommand struct {
	Name         string
	Description  string
	EntryDate    time.Time
	ExitDate     *time.Time
	Status       string
	Paid         bool
	Price        float64
	Observations string
}
