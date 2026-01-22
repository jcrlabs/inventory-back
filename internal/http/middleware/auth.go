package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/jonathanCaamano/inventory-back/internal/http/response"
)

type authKey string

const (
	userIDKey authKey = "user_id"
	roleKey   authKey = "role"
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

			parsed, err := jwt.Parse(tok, func(t *jwt.Token) (any, error) {
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

			uid, _ := claims["sub"].(string)
			role, _ := claims["role"].(string)
			ctx := context.WithValue(r.Context(), userIDKey, uid)
			ctx = context.WithValue(ctx, roleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			v := r.Context().Value(roleKey)
			rt, _ := v.(string)
			if rt != role {
				response.Error(w, http.StatusForbidden, "forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
