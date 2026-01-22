package handlers

import (
	"net/http"

	"github.com/jonathanCaamano/inventory-back/internal/http/response"
)

type Health struct{}

func NewHealth() *Health { return &Health{} }

func (h *Health) Health(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, 200, map[string]any{"status": "ok"})
}
