package product

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusDraft    Status = "draft"
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
)

func ParseStatus(v string) (Status, error) {
	s := Status(strings.ToLower(strings.TrimSpace(v)))
	switch s {
	case StatusDraft, StatusActive, StatusArchived:
		return s, nil
	default:
		return "", errors.New("invalid_status")
	}
}

type Image struct {
	ID        uuid.UUID
	URL       string
	Position  int
	CreatedAt time.Time
}

type Contact struct {
	FirstName   string
	LastName    string
	PhoneNumber string
}

type Product struct {
	ID           uuid.UUID
	GroupID      uuid.UUID
	Name         string
	Description  string
	EntryDate    time.Time
	ExitDate     *time.Time
	Status       Status
	Paid         bool
	Price        float64
	Observations string

	Contact *Contact
	Images  []Image

	CreatedAt time.Time
	UpdatedAt time.Time
}

func New(groupID uuid.UUID, name, desc string, entry time.Time, exit *time.Time, status Status, paid bool, price float64, obs string) (*Product, error) {
	name = strings.TrimSpace(name)
	if groupID == uuid.Nil {
		return nil, errors.New("group_id_required")
	}
	if name == "" {
		return nil, errors.New("name_required")
	}
	if entry.IsZero() {
		return nil, errors.New("entry_date_required")
	}
	if exit != nil && exit.Before(entry) {
		return nil, errors.New("exit_before_entry")
	}
	if price < 0 {
		return nil, errors.New("invalid_price")
	}
	p := &Product{
		ID:           uuid.New(),
		GroupID:      groupID,
		Name:         name,
		Description:  strings.TrimSpace(desc),
		EntryDate:    entry,
		ExitDate:     exit,
		Status:       status,
		Paid:         paid,
		Price:        price,
		Observations: strings.TrimSpace(obs),
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	return p, nil
}

func (p *Product) Update(name, desc string, entry time.Time, exit *time.Time, status Status, paid bool, price float64, obs string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("name_required")
	}
	if entry.IsZero() {
		return errors.New("entry_date_required")
	}
	if exit != nil && exit.Before(entry) {
		return errors.New("exit_before_entry")
	}
	if price < 0 {
		return errors.New("invalid_price")
	}
	p.Name = name
	p.Description = strings.TrimSpace(desc)
	p.EntryDate = entry
	p.ExitDate = exit
	p.Status = status
	p.Paid = paid
	p.Price = price
	p.Observations = strings.TrimSpace(obs)
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (p *Product) AddImage(url string, position int) error {
	url = strings.TrimSpace(url)
	if url == "" {
		return errors.New("image_url_required")
	}
	p.Images = append(p.Images, Image{ID: uuid.New(), URL: url, Position: position, CreatedAt: time.Now().UTC()})
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (p *Product) UpsertContact(first, last, phone string) error {
	first = strings.TrimSpace(first)
	last = strings.TrimSpace(last)
	phone = strings.TrimSpace(phone)
	if first == "" || last == "" || phone == "" {
		return errors.New("contact_required")
	}
	p.Contact = &Contact{FirstName: first, LastName: last, PhoneNumber: phone}
	p.UpdatedAt = time.Now().UTC()
	return nil
}
