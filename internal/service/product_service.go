package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jonathanCaamano/inventory-back/internal/http/dto"
)

type ProductRepo interface {
	Create(ctx context.Context, p dto.ProductCreate) (any, error)
	GetByID(ctx context.Context, id uuid.UUID) (any, error)
	Update(ctx context.Context, id uuid.UUID, p dto.ProductUpdate) (any, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Search(ctx context.Context, req SearchRequest) (any, error)
	AddImage(ctx context.Context, id uuid.UUID, img dto.Image) (any, error)
	UpsertContact(ctx context.Context, id uuid.UUID, c dto.Contact) (any, error)
}

type ProductService struct {
	repo ProductRepo
}

func NewProductService(repo ProductRepo) *ProductService {
	return &ProductService{repo: repo}
}

type SearchRequest struct {
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

func (s *ProductService) Create(ctx context.Context, in dto.ProductCreate) (any, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, errors.New("name_required")
	}
	if strings.TrimSpace(in.Status) == "" {
		return nil, errors.New("status_required")
	}
	if in.EntryDate.IsZero() {
		return nil, errors.New("entry_date_required")
	}
	return s.repo.Create(ctx, in)
}

func (s *ProductService) GetByID(ctx context.Context, id uuid.UUID) (any, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) Update(ctx context.Context, id uuid.UUID, in dto.ProductUpdate) (any, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, errors.New("name_required")
	}
	if strings.TrimSpace(in.Status) == "" {
		return nil, errors.New("status_required")
	}
	if in.EntryDate.IsZero() {
		return nil, errors.New("entry_date_required")
	}
	return s.repo.Update(ctx, id, in)
}

func (s *ProductService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *ProductService) Search(ctx context.Context, req SearchRequest) (any, error) {
	if req.Limit <= 0 || req.Limit > 200 {
		req.Limit = 20
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	return s.repo.Search(ctx, req)
}

func (s *ProductService) AddImage(ctx context.Context, id uuid.UUID, img dto.Image) (any, error) {
	if strings.TrimSpace(img.ImageURL) == "" {
		return nil, errors.New("image_url_required")
	}
	return s.repo.AddImage(ctx, id, img)
}

func (s *ProductService) UpsertContact(ctx context.Context, id uuid.UUID, c dto.Contact) (any, error) {
	if strings.TrimSpace(c.FirstName) == "" || strings.TrimSpace(c.LastName) == "" || strings.TrimSpace(c.PhoneNumber) == "" {
		return nil, errors.New("contact_required")
	}
	return s.repo.UpsertContact(ctx, id, c)
}
