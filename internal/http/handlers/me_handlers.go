package handlers

import (
	"net/http"

	"github.com/jonathanCaamano/inventory-back/internal/application/me"
	"github.com/jonathanCaamano/inventory-back/internal/http/middleware"
	"github.com/jonathanCaamano/inventory-back/internal/http/response"
)

type Me struct {
	svc *me.Service
}

func NewMe(svc *me.Service) *Me {
	return &Me{svc: svc}
}

func (h *Me) WhoAmI(w http.ResponseWriter, r *http.Request) {
	uid, err := middleware.UserID(r.Context())
	if err != nil {
		response.Error(w, 401, "missing_token")
		return
	}
	isAdmin := middleware.IsAdmin(r.Context())
	out, err := h.svc.WhoAmI(r.Context(), uid, isAdmin)
	if err != nil {
		response.Error(w, 500, "internal_error")
		return
	}
	response.JSON(w, 200, out)
}
