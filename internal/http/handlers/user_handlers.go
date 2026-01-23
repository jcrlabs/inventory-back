package handlers

import (
	"net/http"

	"github.com/jonathanCaamano/inventory-back/internal/application/users"
	"github.com/jonathanCaamano/inventory-back/internal/http/response"
)

type Users struct {
	svc *users.Service
}

func NewUsers(svc *users.Service) *Users {
	return &Users{svc: svc}
}

type adminCreateUserReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}

func (h *Users) AdminCreate(w http.ResponseWriter, r *http.Request) {
	var in adminCreateUserReq
	if err := decodeJSON(w, r, &in); err != nil {
		response.Error(w, 400, "invalid_json")
		return
	}
	id, err := h.svc.Create(r.Context(), in.Username, in.Password, in.IsAdmin)
	if err != nil {
		response.Error(w, 400, err.Error())
		return
	}
	response.JSON(w, 201, map[string]any{"id": id.String()})
}

func (h *Users) AdminList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	search := q.Get("search")
	limit := parseInt(q.Get("limit"), 20)
	offset := parseInt(q.Get("offset"), 0)
	out, err := h.svc.List(r.Context(), search, limit, offset)
	if err != nil {
		internalError(w, r, err)
		return
	}
	response.JSON(w, 200, map[string]any{"items": out.Items, "total": out.Total, "limit": out.Limit, "offset": out.Offset})
}
