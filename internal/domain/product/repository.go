package product

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SearchQuery struct {
	GroupID   uuid.UUID
	Search    string
	Status    string
	Paid      string
	MinPrice  *float64
	MaxPrice  *float64
	FromEntry *time.Time
	ToEntry   *time.Time
	Sort      string
	Limit     int
	Offset    int
}

type Repository interface {
	Create(ctx context.Context, p *Product) (*Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	Update(ctx context.Context, p *Product) (*Product, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Search(ctx context.Context, q SearchQuery, isAdmin bool) (items []*Product, total int, err error)
	GetGroupID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	AddImage(ctx context.Context, productID uuid.UUID, img Image) (*Product, error)
	UpsertContact(ctx context.Context, productID uuid.UUID, c Contact) (*Product, error)
}
