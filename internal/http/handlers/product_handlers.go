package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jonathanCaamano/inventory-back/internal/storage/s3client"

	"github.com/jonathanCaamano/inventory-back/internal/config"
	"github.com/jonathanCaamano/inventory-back/internal/http/dto"
	"github.com/jonathanCaamano/inventory-back/internal/http/response"
	"github.com/jonathanCaamano/inventory-back/internal/service"
)

type Products struct {
	cfg config.Config
	svc *service.ProductService
	s3  *s3client.Client
}

func NewProducts(cfg config.Config, svc *service.ProductService, s3 *s3client.Client) *Products {
	return &Products{cfg: cfg, svc: svc, s3: s3}
}

func (h *Products) Create(w http.ResponseWriter, r *http.Request) {
	var in dto.ProductCreate
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.Error(w, 400, "invalid_json")
		return
	}
	out, err := h.svc.Create(r.Context(), in)
	if err != nil {
		response.Error(w, 400, err.Error())
		return
	}
	response.JSON(w, 201, out)
}

func (h *Products) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, 400, "invalid_id")
		return
	}
	out, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, 404, "not_found")
		return
	}
	response.JSON(w, 200, out)
}

func (h *Products) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, 400, "invalid_id")
		return
	}
	var in dto.ProductUpdate
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.Error(w, 400, "invalid_json")
		return
	}
	out, err := h.svc.Update(r.Context(), id, in)
	if err != nil {
		response.Error(w, 400, err.Error())
		return
	}
	response.JSON(w, 200, out)
}

func (h *Products) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, 400, "invalid_id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		response.Error(w, 404, "not_found")
		return
	}
	w.WriteHeader(204)
}

func (h *Products) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := parseInt(q.Get("limit"), 20)
	offset := parseInt(q.Get("offset"), 0)

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

	req := service.SearchRequest{
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

	out, err := h.svc.Search(r.Context(), req)
	if err != nil {
		response.Error(w, 400, err.Error())
		return
	}
	response.JSON(w, 200, out)
}

func (h *Products) AddImage(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, 400, "invalid_id")
		return
	}
	var in dto.Image
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.Error(w, 400, "invalid_json")
		return
	}
	out, err := h.svc.AddImage(r.Context(), id, in)
	if err != nil {
		response.Error(w, 400, err.Error())
		return
	}
	response.JSON(w, 200, out)
}

func (h *Products) UpsertContact(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, 400, "invalid_id")
		return
	}
	var in dto.Contact
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.Error(w, 400, "invalid_json")
		return
	}
	out, err := h.svc.UpsertContact(r.Context(), id, in)
	if err != nil {
		response.Error(w, 400, err.Error())
		return
	}
	response.JSON(w, 200, out)
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
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, 400, "invalid_id")
		return
	}
	_ = id // el id lo usamos para componer el objectKey (organización por producto)

	var in dto.PresignImageRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.Error(w, 400, "invalid_json")
		return
	}
	if in.FileName == "" || in.ContentType == "" {
		response.Error(w, 400, "file_name_and_content_type_required")
		return
	}

	// key: products/{productId}/{uuid}-{fileName}
	key := "products/" + id.String() + "/" + uuid.NewString() + "-" + sanitizeFileName(in.FileName)

	uploadURL, objectURL, err := h.s3.PresignPutObject(r.Context(), key, in.ContentType, 10*time.Minute)
	if err != nil {
		response.Error(w, 500, "presign_failed")
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
