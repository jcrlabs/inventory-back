package dto

import (
	"time"

	"github.com/google/uuid"
)

type Contact struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
}

type Image struct {
	ImageURL string `json:"image_url"`
	Position int    `json:"position"`
}

type ProductCreate struct {
	GroupID      uuid.UUID  `json:"group_id"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	Price        float64    `json:"price"`
	Paid         bool       `json:"paid"`
	Status       string     `json:"status"`
	EntryDate    time.Time  `json:"entry_date"`
	ExitDate     *time.Time `json:"exit_date"`
	Observations string     `json:"observations"`
	Contact      *Contact   `json:"contact"`
	Images       []Image    `json:"images"`
}

type ProductUpdate struct {
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	Price        float64    `json:"price"`
	Paid         bool       `json:"paid"`
	Status       string     `json:"status"`
	EntryDate    time.Time  `json:"entry_date"`
	ExitDate     *time.Time `json:"exit_date"`
	Observations string     `json:"observations"`
	Contact      *Contact   `json:"contact"`
	Images       []Image    `json:"images"`
}
