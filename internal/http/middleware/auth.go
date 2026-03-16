package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/jonathanCaamano/inventory-back/internal/http/response"
)

type authKey string

const (
	userIDKey  authKey = "user_id"
	isAdminKey authKey = "is_admin"
)

func JWT(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" || !strings.HasPrefix(h, "Bearer ") {
				response.Error(w, http.StatusUnauthorized, "missing_token")
				return
			}
			tok := strings.TrimPrefix(h, "Bearer ")

			parsed, err := jwt.Parse(tok, func(_ *jwt.Token) (any, error) {
				return []byte(secret), nil
			})
			if err != nil || !parsed.Valid {
				response.Error(w, http.StatusUnauthorized, "invalid_token")
				return
			}

			claims, ok := parsed.Claims.(jwt.MapClaims)
			if !ok {
				response.Error(w, http.StatusUnauthorized, "invalid_claims")
				return
			}

			sub, _ := claims["sub"].(string)
			uid, err := uuid.Parse(sub)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "invalid_claims")
				return
			}
			isAdmin, _ := claims["admin"].(bool)

			ctx := context.WithValue(r.Context(), userIDKey, uid)
			ctx = context.WithValue(ctx, isAdminKey, isAdmin)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserID(ctx context.Context) (uuid.UUID, error) {
	v := ctx.Value(userIDKey)
	uid, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("missing_user")
	}
	return uid, nil
}

func IsAdmin(ctx context.Context) bool {
	v := ctx.Value(isAdminKey)
	b, _ := v.(bool)
	return b
}
