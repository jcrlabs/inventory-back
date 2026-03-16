package middleware

import (
	"net/http"

	"github.com/jonathanCaamano/inventory-back/internal/http/response"
)

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsAdmin(r.Context()) {
			response.Error(w, 403, "forbidden")
			return
		}
		next.ServeHTTP(w, r)
	})
}
