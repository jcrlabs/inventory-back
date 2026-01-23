package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jonathanCaamano/inventory-back/internal/storage/s3client"

	"github.com/jonathanCaamano/inventory-back/internal/application/products"
	"github.com/jonathanCaamano/inventory-back/internal/config"
	"github.com/jonathanCaamano/inventory-back/internal/domain/product"
	"github.com/jonathanCaamano/inventory-back/internal/http/dto"
	"github.com/jonathanCaamano/inventory-back/internal/http/middleware"
	"github.com/jonathanCaamano/inventory-back/internal/http/response"
)

type Products struct {
	cfg config.Config
	svc *products.Service
	s3  *s3client.Client
}

func NewProducts(cfg config.Config, svc *products.Service, s3 *s3client.Client) *Products {
	return &Products{cfg: cfg, svc: svc, s3: s3}
}

func (h *Products) Create(w http.ResponseWriter, r *http.Request) {
	actor, err := middleware.UserID(r.Context())
	if err != nil {
		response.Error(w, 401, "missing_token")
		return
	}
	isAdmin := middleware.IsAdmin(r.Context())
	var in dto.ProductCreate
	if err := decodeJSON(w, r, &in); err != nil {
		response.Error(w, 400, "invalid_json")
		return
	}
	a := products.Actor{UserID: actor, IsAdmin: isAdmin}
	cmd := products.CreateCommand{
		GroupID:      in.GroupID,
		Name:         in.Name,
		Description:  in.Description,
		EntryDate:    in.EntryDate,
		ExitDate:     in.ExitDate,
		Status:       in.Status,
		Paid:         in.Paid,
		Price:        in.Price,
		Observations: in.Observations,
	}
	if in.Contact != nil {
		cmd.Contact = &products.ContactCommand{FirstName: in.Contact.FirstName, LastName: in.Contact.LastName, PhoneNumber: in.Contact.PhoneNumber}
	}
	for _, img := range in.Images {
		cmd.Images = append(cmd.Images, products.ImageCommand{URL: img.ImageURL, Position: img.Position})
	}
	p, err := h.svc.Create(r.Context(), a, cmd)
	if err != nil {
		response.Error(w, 400, err.Error())
		return
	}
	response.JSON(w, 201, toDTOProduct(p))
}

func (h *Products) GetByID(w http.ResponseWriter, r *http.Request) {
	actor, err := middleware.UserID(r.Context())
	if err != nil {
		response.Error(w, 401, "missing_token")
		return
	}
	isAdmin := middleware.IsAdmin(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, 400, "invalid_id")
		return
	}
	a := products.Actor{UserID: actor, IsAdmin: isAdmin}
	out, err := h.svc.GetByID(r.Context(), a, id)
	if err != nil {
		if err.Error() == "forbidden" {
			response.Error(w, 403, "forbidden")
			return
		}
		response.Error(w, 404, "not_found")
		return
	}
	response.JSON(w, 200, toDTOProduct(out))
}

func (h *Products) Update(w http.ResponseWriter, r *http.Request) {
	actor, err := middleware.UserID(r.Context())
	if err != nil {
		response.Error(w, 401, "missing_token")
		return
	}
	isAdmin := middleware.IsAdmin(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, 400, "invalid_id")
		return
	}
	var in dto.ProductUpdate
	if err := decodeJSON(w, r, &in); err != nil {
		response.Error(w, 400, "invalid_json")
		return
	}
	a := products.Actor{UserID: actor, IsAdmin: isAdmin}
	cmd := products.UpdateCommand{
		Name:         in.Name,
		Description:  in.Description,
		EntryDate:    in.EntryDate,
		ExitDate:     in.ExitDate,
		Status:       in.Status,
		Paid:         in.Paid,
		Price:        in.Price,
		Observations: in.Observations,
	}
	out, err := h.svc.Update(r.Context(), a, id, cmd)
	if err != nil {
		if err.Error() == "forbidden" {
			response.Error(w, 403, "forbidden")
			return
		}
		response.Error(w, 400, err.Error())
		return
	}
	response.JSON(w, 200, toDTOProduct(out))
}

func (h *Products) Delete(w http.ResponseWriter, r *http.Request) {
	actor, err := middleware.UserID(r.Context())
	if err != nil {
		response.Error(w, 401, "missing_token")
		return
	}
	isAdmin := middleware.IsAdmin(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, 400, "invalid_id")
		return
	}
	a := products.Actor{UserID: actor, IsAdmin: isAdmin}
	if err := h.svc.Delete(r.Context(), a, id); err != nil {
		if err.Error() == "forbidden" {
			response.Error(w, 403, "forbidden")
			return
		}
		response.Error(w, 404, "not_found")
		return
	}
	w.WriteHeader(204)
}

func (h *Products) Search(w http.ResponseWriter, r *http.Request) {
	actor, err := middleware.UserID(r.Context())
	if err != nil {
		response.Error(w, 401, "missing_token")
		return
	}
	isAdmin := middleware.IsAdmin(r.Context())
	q := r.URL.Query()
	limit := parseInt(q.Get("limit"), 20)
	offset := parseInt(q.Get("offset"), 0)

	groupID := uuid.Nil
	if v := q.Get("group_id"); v != "" {
		gid, err := uuid.Parse(v)
		if err != nil {
			response.Error(w, 400, "invalid_group_id")
			return
		}
		groupID = gid
	}

	var fromEntry, toEntry *time.Time
	if v := q.Get("from_entry"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			fromEntry = &t
		}
	}
	if v := q.Get("to_entry"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			toEntry = &t
		}
	}

	a := products.Actor{UserID: actor, IsAdmin: isAdmin}
	req := products.SearchRequest{
		GroupID:   groupID,
		Search:    q.Get("search"),
		Status:    q.Get("status"),
		Paid:      q.Get("paid"),
		MinPrice:  parseFloatPtr(q.Get("min_price")),
		MaxPrice:  parseFloatPtr(q.Get("max_price")),
		FromEntry: fromEntry,
		ToEntry:   toEntry,
		Sort:      q.Get("sort"),
		Limit:     limit,
		Offset:    offset,
	}

	out, err := h.svc.Search(r.Context(), a, req)
	if err != nil {
		if err.Error() == "forbidden" {
			response.Error(w, 403, "forbidden")
			return
		}
		response.Error(w, 400, err.Error())
		return
	}
	items := make([]dtoProduct, 0, len(out.Items))
	for _, p := range out.Items {
		items = append(items, toDTOProduct(p))
	}
	response.JSON(w, 200, map[string]any{"items": items, "total": out.Total, "limit": out.Limit, "offset": out.Offset})
}

func (h *Products) AddImage(w http.ResponseWriter, r *http.Request) {
	actor, err := middleware.UserID(r.Context())
	if err != nil {
		response.Error(w, 401, "missing_token")
		return
	}
	isAdmin := middleware.IsAdmin(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, 400, "invalid_id")
		return
	}
	var in dto.Image
	if err := decodeJSON(w, r, &in); err != nil {
		response.Error(w, 400, "invalid_json")
		return
	}
	a := products.Actor{UserID: actor, IsAdmin: isAdmin}
	out, err := h.svc.AddImage(r.Context(), a, id, in.ImageURL, in.Position)
	if err != nil {
		if err.Error() == "forbidden" {
			response.Error(w, 403, "forbidden")
			return
		}
		response.Error(w, 400, err.Error())
		return
	}
	response.JSON(w, 200, toDTOProduct(out))
}

func (h *Products) UpsertContact(w http.ResponseWriter, r *http.Request) {
	actor, err := middleware.UserID(r.Context())
	if err != nil {
		response.Error(w, 401, "missing_token")
		return
	}
	isAdmin := middleware.IsAdmin(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, 400, "invalid_id")
		return
	}
	var in dto.Contact
	if err := decodeJSON(w, r, &in); err != nil {
		response.Error(w, 400, "invalid_json")
		return
	}
	a := products.Actor{UserID: actor, IsAdmin: isAdmin}
	out, err := h.svc.UpsertContact(r.Context(), a, id, in.FirstName, in.LastName, in.PhoneNumber)
	if err != nil {
		if err.Error() == "forbidden" {
			response.Error(w, 403, "forbidden")
			return
		}
		response.Error(w, 400, err.Error())
		return
	}
	response.JSON(w, 200, toDTOProduct(out))
}

func parseInt(v string, def int) int {
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func parseFloatPtr(v string) *float64 {
	if v == "" {
		return nil
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil
	}
	return &f
}

func (h *Products) PresignImage(w http.ResponseWriter, r *http.Request) {
	if h.s3 == nil {
		response.Error(w, http.StatusServiceUnavailable, "storage_not_configured")
		return
	}
	actor, err := middleware.UserID(r.Context())
	if err != nil {
		response.Error(w, 401, "missing_token")
		return
	}
	isAdmin := middleware.IsAdmin(r.Context())

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, 400, "invalid_id")
		return
	}
	a := products.Actor{UserID: actor, IsAdmin: isAdmin}
	gid, err := h.svc.GroupIDForProduct(r.Context(), a, id)
	if err != nil {
		response.Error(w, 403, "forbidden")
		return
	}
	var in dto.PresignImageRequest
	if err := decodeJSON(w, r, &in); err != nil {
		response.Error(w, 400, "invalid_json")
		return
	}
	if in.FileName == "" || in.ContentType == "" {
		response.Error(w, 400, "file_name_and_content_type_required")
		return
	}
	if !isAllowedImageContentType(in.ContentType) {
		response.Error(w, 400, "content_type_not_allowed")
		return
	}

	key := "groups/" + gid.String() + "/products/" + id.String() + "/" + uuid.NewString() + "-" + sanitizeFileName(in.FileName)

	uploadURL, objectURL, err := h.s3.PresignPutObject(r.Context(), key, in.ContentType, 10*time.Minute)
	if err != nil {
		serverError(w, r, 500, "presign_failed", err)
		return
	}

	out := dto.PresignImageResponse{
		UploadURL: uploadURL,
		ObjectURL: objectURL,
		ObjectKey: key,
	}
	response.JSON(w, 200, out)
}

func sanitizeFileName(v string) string {
	v = strings.TrimSpace(v)
	v = strings.ReplaceAll(v, "..", "")
	v = strings.ReplaceAll(v, "/", "-")
	v = strings.ReplaceAll(v, "\\", "-")
	return v
}

func isAllowedImageContentType(ct string) bool {
	switch strings.ToLower(strings.TrimSpace(ct)) {
	case "image/jpeg", "image/jpg", "image/png", "image/webp":
		return true
	default:
		return false
	}
}

type dtoProduct struct {
	ID           uuid.UUID    `json:"id"`
	GroupID      uuid.UUID    `json:"group_id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Price        float64      `json:"price"`
	Paid         bool         `json:"paid"`
	Status       string       `json:"status"`
	EntryDate    time.Time    `json:"entry_date"`
	ExitDate     *time.Time   `json:"exit_date"`
	Observations string       `json:"observations"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	Contact      *dto.Contact `json:"contact,omitempty"`
	Images       []dto.Image  `json:"images,omitempty"`
}

func toDTOProduct(p *product.Product) dtoProduct {
	out := dtoProduct{
		ID:           p.ID,
		GroupID:      p.GroupID,
		Name:         p.Name,
		Description:  p.Description,
		Price:        p.Price,
		Paid:         p.Paid,
		Status:       string(p.Status),
		EntryDate:    p.EntryDate,
		ExitDate:     p.ExitDate,
		Observations: p.Observations,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
	if p.Contact != nil {
		out.Contact = &dto.Contact{FirstName: p.Contact.FirstName, LastName: p.Contact.LastName, PhoneNumber: p.Contact.PhoneNumber}
	}
	if len(p.Images) > 0 {
		out.Images = make([]dto.Image, 0, len(p.Images))
		for _, img := range p.Images {
			out.Images = append(out.Images, dto.Image{ImageURL: img.URL, Position: img.Position})
		}
	}
	return out
}
