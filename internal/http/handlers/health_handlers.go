package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jonathanCaamano/inventory-back/internal/http/response"
	"github.com/jonathanCaamano/inventory-back/internal/storage/s3client"
)

type Health struct {
	dbPing func(context.Context) error
	s3     *s3client.Client
}

func NewHealth(dbPing func(context.Context) error, s3 *s3client.Client) *Health {
	return &Health{dbPing: dbPing, s3: s3}
}

func (h *Health) Liveness(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, 200, map[string]any{"status": "ok"})
}

func (h *Health) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	checks := map[string]any{}

	if h.dbPing != nil {
		if err := h.dbPing(ctx); err != nil {
			checks["db"] = "down"
			response.JSON(w, http.StatusServiceUnavailable, map[string]any{"status": "not_ready", "checks": checks})
			return
		}
		checks["db"] = "up"
	} else {
		checks["db"] = "skipped"
	}

	if h.s3 != nil {
		if err := h.s3.CheckBucket(ctx); err != nil {
			checks["s3"] = "down"
			response.JSON(w, http.StatusServiceUnavailable, map[string]any{"status": "not_ready", "checks": checks})
			return
		}
		checks["s3"] = "up"
	} else {
		checks["s3"] = "skipped"
	}

	response.JSON(w, 200, map[string]any{"status": "ready", "checks": checks})
}
