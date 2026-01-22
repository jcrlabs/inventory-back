package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jonathanCaamano/inventory-back/internal/http/dto"
)

type fakeProductRepo struct {
	groupID uuid.UUID
}

func (f *fakeProductRepo) Create(ctx context.Context, p dto.ProductCreate) (any, error) {
	return map[string]any{"id": uuid.NewString()}, nil
}
func (f *fakeProductRepo) GetByID(ctx context.Context, id uuid.UUID) (any, error) {
	return map[string]any{"id": id.String()}, nil
}
func (f *fakeProductRepo) Update(ctx context.Context, id uuid.UUID, p dto.ProductUpdate) (any, error) {
	return map[string]any{"id": id.String()}, nil
}
func (f *fakeProductRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (f *fakeProductRepo) Search(ctx context.Context, req SearchRequest) (any, error) {
	return map[string]any{"items": []any{}}, nil
}
func (f *fakeProductRepo) AddImage(ctx context.Context, id uuid.UUID, img dto.Image) (any, error) {
	return nil, nil
}
func (f *fakeProductRepo) UpsertContact(ctx context.Context, id uuid.UUID, c dto.Contact) (any, error) {
	return nil, nil
}
func (f *fakeProductRepo) GetProductGroupID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	return f.groupID, nil
}

type fakeGroupRepo struct {
	role string
	err  error
}

func (f fakeGroupRepo) GetUserRoleInGroup(ctx context.Context, userID, groupID uuid.UUID) (string, error) {
	return f.role, f.err
}

func TestProductService_Create_RequiresWriterOnGroup(t *testing.T) {
	ctx := context.Background()
	groupID := uuid.New()
	ps := NewProductService(&fakeProductRepo{groupID: groupID}, fakeGroupRepo{role: GroupRoleReader})

	_, err := ps.Create(ctx, uuid.New(), false, dto.ProductCreate{
		GroupID:   groupID,
		Name:      "x",
		Status:    "in_stock",
		EntryDate: time.Now(),
	})
	if err == nil || err.Error() != "forbidden" {
		t.Fatalf("expected forbidden, got %v", err)
	}

	ps = NewProductService(&fakeProductRepo{groupID: groupID}, fakeGroupRepo{role: GroupRoleWriter})
	_, err = ps.Create(ctx, uuid.New(), false, dto.ProductCreate{
		GroupID:   groupID,
		Name:      "x",
		Status:    "in_stock",
		EntryDate: time.Now(),
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestProductService_Search_RequiresGroupForNonAdmin(t *testing.T) {
	ctx := context.Background()
	ps := NewProductService(&fakeProductRepo{}, fakeGroupRepo{role: GroupRoleReader})
	_, err := ps.Search(ctx, uuid.New(), false, SearchRequest{})
	if err == nil || err.Error() != "group_id_required" {
		t.Fatalf("expected group_id_required, got %v", err)
	}
}

func TestProductService_RequireWriterForProduct(t *testing.T) {
	ctx := context.Background()
	groupID := uuid.New()
	ps := NewProductService(&fakeProductRepo{groupID: groupID}, fakeGroupRepo{role: GroupRoleReader})
	if err := ps.RequireWriterForProduct(ctx, uuid.New(), false, uuid.New()); err == nil {
		t.Fatalf("expected error")
	}

	ps = NewProductService(&fakeProductRepo{groupID: groupID}, fakeGroupRepo{role: GroupRoleWriter})
	if err := ps.RequireWriterForProduct(ctx, uuid.New(), false, uuid.New()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}
