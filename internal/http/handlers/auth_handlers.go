package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jonathanCaamano/inventory-back/internal/application/auth"
	"github.com/jonathanCaamano/inventory-back/internal/http/response"
)

type Auth struct {
	svc *auth.Service
}

func NewAuth(svc *auth.Service) *Auth {
	return &Auth{svc: svc}
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *Auth) Login(w http.ResponseWriter, r *http.Request) {
	var in loginReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.Error(w, 400, "invalid_json")
		return
	}

	tok, err := a.svc.Login(r.Context(), in.Username, in.Password)
	if err != nil {
		code := 401
		if err.Error() == "user_inactive" {
			code = 403
		}
		response.Error(w, code, err.Error())
		return
	}

	response.JSON(w, 200, map[string]any{"access_token": tok, "token_type": "Bearer"})
}
