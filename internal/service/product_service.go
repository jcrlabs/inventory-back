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
	GetProductGroupID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
}

type GroupAuthzRepo interface {
	GetUserRoleInGroup(ctx context.Context, userID, groupID uuid.UUID) (string, error)
}

type ProductService struct {
	repo      ProductRepo
	groupRepo GroupAuthzRepo
}

func NewProductService(repo ProductRepo, groupRepo GroupAuthzRepo) *ProductService {
	return &ProductService{repo: repo, groupRepo: groupRepo}
}

type SearchRequest struct {
	GroupID   *uuid.UUID
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

func (s *ProductService) Create(ctx context.Context, actor uuid.UUID, isAdmin bool, in dto.ProductCreate) (any, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, errors.New("name_required")
	}
	if in.GroupID == uuid.Nil {
		return nil, errors.New("group_id_required")
	}
	if strings.TrimSpace(in.Status) == "" {
		return nil, errors.New("status_required")
	}
	if in.EntryDate.IsZero() {
		return nil, errors.New("entry_date_required")
	}
	if !isAdmin {
		if err := s.requireGroupRole(ctx, actor, in.GroupID, GroupRoleWriter); err != nil {
			return nil, err
		}
	}
	return s.repo.Create(ctx, in)
}

func (s *ProductService) GetByID(ctx context.Context, actor uuid.UUID, isAdmin bool, id uuid.UUID) (any, error) {
	if !isAdmin {
		gid, err := s.repo.GetProductGroupID(ctx, id)
		if err != nil {
			return nil, err
		}
		if err := s.requireGroupRole(ctx, actor, gid, GroupRoleReader); err != nil {
			return nil, err
		}
	}
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) Update(ctx context.Context, actor uuid.UUID, isAdmin bool, id uuid.UUID, in dto.ProductUpdate) (any, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, errors.New("name_required")
	}
	if strings.TrimSpace(in.Status) == "" {
		return nil, errors.New("status_required")
	}
	if in.EntryDate.IsZero() {
		return nil, errors.New("entry_date_required")
	}
	if !isAdmin {
		gid, err := s.repo.GetProductGroupID(ctx, id)
		if err != nil {
			return nil, err
		}
		if err := s.requireGroupRole(ctx, actor, gid, GroupRoleWriter); err != nil {
			return nil, err
		}
	}
	return s.repo.Update(ctx, id, in)
}

func (s *ProductService) Delete(ctx context.Context, actor uuid.UUID, isAdmin bool, id uuid.UUID) error {
	if !isAdmin {
		gid, err := s.repo.GetProductGroupID(ctx, id)
		if err != nil {
			return err
		}
		if err := s.requireGroupRole(ctx, actor, gid, GroupRoleWriter); err != nil {
			return err
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *ProductService) Search(ctx context.Context, actor uuid.UUID, isAdmin bool, req SearchRequest) (any, error) {
	if !isAdmin {
		if req.GroupID == nil || *req.GroupID == uuid.Nil {
			return nil, errors.New("group_id_required")
		}
		if err := s.requireGroupRole(ctx, actor, *req.GroupID, GroupRoleReader); err != nil {
			return nil, err
		}
	}
	if req.Limit <= 0 || req.Limit > 200 {
		req.Limit = 20
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	return s.repo.Search(ctx, req)
}

func (s *ProductService) AddImage(ctx context.Context, actor uuid.UUID, isAdmin bool, id uuid.UUID, img dto.Image) (any, error) {
	if strings.TrimSpace(img.ImageURL) == "" {
		return nil, errors.New("image_url_required")
	}
	if !isAdmin {
		gid, err := s.repo.GetProductGroupID(ctx, id)
		if err != nil {
			return nil, err
		}
		if err := s.requireGroupRole(ctx, actor, gid, GroupRoleWriter); err != nil {
			return nil, err
		}
	}
	return s.repo.AddImage(ctx, id, img)
}

func (s *ProductService) UpsertContact(ctx context.Context, actor uuid.UUID, isAdmin bool, id uuid.UUID, c dto.Contact) (any, error) {
	if strings.TrimSpace(c.FirstName) == "" || strings.TrimSpace(c.LastName) == "" || strings.TrimSpace(c.PhoneNumber) == "" {
		return nil, errors.New("contact_required")
	}
	if !isAdmin {
		gid, err := s.repo.GetProductGroupID(ctx, id)
		if err != nil {
			return nil, err
		}
		if err := s.requireGroupRole(ctx, actor, gid, GroupRoleWriter); err != nil {
			return nil, err
		}
	}
	return s.repo.UpsertContact(ctx, id, c)
}

func (s *ProductService) requireGroupRole(ctx context.Context, actor uuid.UUID, groupID uuid.UUID, want string) error {
	role, err := s.groupRepo.GetUserRoleInGroup(ctx, actor, groupID)
	if err != nil {
		return errors.New("forbidden")
	}
	if want == GroupRoleReader {
		if role == GroupRoleReader || role == GroupRoleWriter {
			return nil
		}
		return errors.New("forbidden")
	}
	if want == GroupRoleWriter {
		if role == GroupRoleWriter {
			return nil
		}
		return errors.New("forbidden")
	}
	return errors.New("forbidden")
}

func (s *ProductService) RequireWriterForProduct(ctx context.Context, actor uuid.UUID, isAdmin bool, productID uuid.UUID) error {
	if isAdmin {
		return nil
	}
	gid, err := s.repo.GetProductGroupID(ctx, productID)
	if err != nil {
		return err
	}
	return s.requireGroupRole(ctx, actor, gid, GroupRoleWriter)
}
