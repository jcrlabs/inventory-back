package products

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jonathanCaamano/inventory-back/internal/domain/group"
	"github.com/jonathanCaamano/inventory-back/internal/domain/product"
)

type Actor struct {
	UserID  uuid.UUID
	IsAdmin bool
}

type Service struct {
	products product.Repository
	groups   group.Repository
}

func New(products product.Repository, groups group.Repository) *Service {
	return &Service{products: products, groups: groups}
}

type SearchRequest struct {
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

type SearchResult struct {
	Items  []*product.Product
	Total  int
	Limit  int
	Offset int
}

func (s *Service) Create(ctx context.Context, a Actor, in CreateCommand) (*product.Product, error) {
	st, err := product.ParseStatus(in.Status)
	if err != nil {
		return nil, err
	}
	if !a.IsAdmin {
		if err := s.requireGroupRole(ctx, a.UserID, in.GroupID, group.RoleWriter); err != nil {
			return nil, err
		}
	}
	p, err := product.New(in.GroupID, in.Name, in.Description, in.EntryDate, in.ExitDate, st, in.Paid, in.Price, in.Observations)
	if err != nil {
		return nil, err
	}
	if in.Contact != nil {
		if err := p.UpsertContact(in.Contact.FirstName, in.Contact.LastName, in.Contact.PhoneNumber); err != nil {
			return nil, err
		}
	}
	for _, img := range in.Images {
		if err := p.AddImage(img.URL, img.Position); err != nil {
			return nil, err
		}
	}
	created, err := s.products.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *Service) GetByID(ctx context.Context, a Actor, id uuid.UUID) (*product.Product, error) {
	if !a.IsAdmin {
		gid, err := s.products.GetGroupID(ctx, id)
		if err != nil {
			return nil, err
		}
		if err := s.requireGroupRole(ctx, a.UserID, gid, group.RoleReader); err != nil {
			return nil, err
		}
	}
	return s.products.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, a Actor, id uuid.UUID, in UpdateCommand) (*product.Product, error) {
	if !a.IsAdmin {
		gid, err := s.products.GetGroupID(ctx, id)
		if err != nil {
			return nil, err
		}
		if err := s.requireGroupRole(ctx, a.UserID, gid, group.RoleWriter); err != nil {
			return nil, err
		}
	}
	existing, err := s.products.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	st, err := product.ParseStatus(in.Status)
	if err != nil {
		return nil, err
	}
	if err := existing.Update(in.Name, in.Description, in.EntryDate, in.ExitDate, st, in.Paid, in.Price, in.Observations); err != nil {
		return nil, err
	}
	return s.products.Update(ctx, existing)
}

func (s *Service) Delete(ctx context.Context, a Actor, id uuid.UUID) error {
	if !a.IsAdmin {
		gid, err := s.products.GetGroupID(ctx, id)
		if err != nil {
			return err
		}
		if err := s.requireGroupRole(ctx, a.UserID, gid, group.RoleWriter); err != nil {
			return err
		}
	}
	return s.products.Delete(ctx, id)
}

func (s *Service) Search(ctx context.Context, a Actor, req SearchRequest) (SearchResult, error) {
	if !a.IsAdmin {
		if req.GroupID == uuid.Nil {
			return SearchResult{}, errors.New("group_id_required")
		}
		if err := s.requireGroupRole(ctx, a.UserID, req.GroupID, group.RoleReader); err != nil {
			return SearchResult{}, err
		}
	}
	if req.Limit <= 0 || req.Limit > 200 {
		req.Limit = 20
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	q := product.SearchQuery{
		GroupID:   req.GroupID,
		Search:    req.Search,
		Status:    req.Status,
		Paid:      req.Paid,
		MinPrice:  req.MinPrice,
		MaxPrice:  req.MaxPrice,
		FromEntry: req.FromEntry,
		ToEntry:   req.ToEntry,
		Sort:      req.Sort,
		Limit:     req.Limit,
		Offset:    req.Offset,
	}
	items, total, err := s.products.Search(ctx, q, a.IsAdmin)
	if err != nil {
		return SearchResult{}, err
	}
	return SearchResult{Items: items, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}

func (s *Service) AddImage(ctx context.Context, a Actor, id uuid.UUID, url string, position int) (*product.Product, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return nil, errors.New("image_url_required")
	}
	if !a.IsAdmin {
		gid, err := s.products.GetGroupID(ctx, id)
		if err != nil {
			return nil, err
		}
		if err := s.requireGroupRole(ctx, a.UserID, gid, group.RoleWriter); err != nil {
			return nil, err
		}
	}
	img := product.Image{ID: uuid.New(), URL: url, Position: position, CreatedAt: time.Now().UTC()}
	return s.products.AddImage(ctx, id, img)
}

func (s *Service) UpsertContact(ctx context.Context, a Actor, id uuid.UUID, first, last, phone string) (*product.Product, error) {
	if !a.IsAdmin {
		gid, err := s.products.GetGroupID(ctx, id)
		if err != nil {
			return nil, err
		}
		if err := s.requireGroupRole(ctx, a.UserID, gid, group.RoleWriter); err != nil {
			return nil, err
		}
	}
	c := product.Contact{FirstName: strings.TrimSpace(first), LastName: strings.TrimSpace(last), PhoneNumber: strings.TrimSpace(phone)}
	if c.FirstName == "" || c.LastName == "" || c.PhoneNumber == "" {
		return nil, errors.New("contact_required")
	}
	return s.products.UpsertContact(ctx, id, c)
}

func (s *Service) requireGroupRole(ctx context.Context, userID, groupID uuid.UUID, want group.Role) error {
	role, err := s.groups.RoleForUser(ctx, userID, groupID)
	if err != nil {
		return errors.New("forbidden")
	}
	if want == group.RoleReader {
		if role.AllowsRead() {
			return nil
		}
		return errors.New("forbidden")
	}
	if want == group.RoleWriter {
		if role.AllowsWrite() {
			return nil
		}
		return errors.New("forbidden")
	}
	return errors.New("forbidden")
}

func (s *Service) RequireWriterForProduct(ctx context.Context, a Actor, productID uuid.UUID) error {
	if a.IsAdmin {
		return nil
	}
	gid, err := s.products.GetGroupID(ctx, productID)
	if err != nil {
		return err
	}
	return s.requireGroupRole(ctx, a.UserID, gid, group.RoleWriter)
}
