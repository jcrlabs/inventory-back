package handlers

import (
	"log"
	"net/http"

	"github.com/jonathanCaamano/inventory-back/internal/http/middleware"
	"github.com/jonathanCaamano/inventory-back/internal/http/response"
)

func internalError(w http.ResponseWriter, r *http.Request, err error) {
	rid := middleware.GetRequestID(r.Context())
	log.Printf("rid=%s method=%s path=%s err=%v", rid, r.Method, r.URL.Path, err)
	response.Error(w, http.StatusInternalServerError, "internal_error")
}

func serverError(w http.ResponseWriter, r *http.Request, code int, msg string, err error) {
	rid := middleware.GetRequestID(r.Context())
	log.Printf("rid=%s method=%s path=%s err=%v", rid, r.Method, r.URL.Path, err)
	response.Error(w, code, msg)
}
