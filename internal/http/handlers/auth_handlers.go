package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/jonathanCaamano/inventory-back/internal/config"
	"github.com/jonathanCaamano/inventory-back/internal/http/response"
)

type Auth struct {
	cfg config.Config
}

func NewAuth(cfg config.Config) *Auth {
	return &Auth{cfg: cfg}
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

	role := "user"
	if in.Username == "admin" && in.Password == "admin" {
		role = "admin"
	} else if in.Username == "" || in.Password == "" {
		response.Error(w, 401, "invalid_credentials")
		return
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  in.Username,
		"role": role,
		"iat":  now.Unix(),
		"exp":  now.Add(30 * time.Minute).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString([]byte(a.cfg.JWTSecret))
	if err != nil {
		response.Error(w, 500, "token_error")
		return
	}

	response.JSON(w, 200, map[string]any{"access_token": s, "token_type": "Bearer"})
}
