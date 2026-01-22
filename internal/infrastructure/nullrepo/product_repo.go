package nullrepo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jonathanCaamano/inventory-back/internal/domain/product"
)

type ProductRepo struct{}

func NewProductRepo() *ProductRepo { return &ProductRepo{} }

func (r *ProductRepo) Create(ctx context.Context, p *product.Product) (*product.Product, error) {
	return nil, errors.New("storage_not_configured")
}
func (r *ProductRepo) GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	return nil, errors.New("storage_not_configured")
}
func (r *ProductRepo) UpdatePassword(ctx context.Context, id uuid.UUID, newPasswordHash string) error {
	return errors.New("storage_not_configured")
}
func (r *ProductRepo) EnsureBootstrapAdmin(ctx context.Context, username, passwordHash string) error {
	return errors.New("storage_not_configured")
}

func (r *ProductRepo) Update(ctx context.Context, p *product.Product) (*product.Product, error) {
	return nil, errors.New("storage_not_configured")
}

func (r *ProductRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return errors.New("storage_not_configured")
}

func (r *ProductRepo) GetGroupID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, errors.New("storage_not_configured")
}

func (r *ProductRepo) AddImage(ctx context.Context, productID uuid.UUID, img product.Image) (*product.Product, error) {
	return nil, errors.New("storage_not_configured")
}

func (r *ProductRepo) UpsertContact(ctx context.Context, productID uuid.UUID, c product.Contact) (*product.Product, error) {
	return nil, errors.New("storage_not_configured")
}

func (r *ProductRepo) Search(ctx context.Context, q product.SearchQuery, isAdmin bool) (items []*product.Product, total int, err error) {
	return nil, 0, errors.New("storage_not_configured")
}
