package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/jonathanCaamano/inventory-back/internal/http/response"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				rid := GetRequestID(r.Context())
				log.Printf("panic rid=%s method=%s path=%s err=%v\n%s", rid, r.Method, r.URL.Path, rec, string(debug.Stack()))
				response.Error(w, http.StatusInternalServerError, "internal_error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
