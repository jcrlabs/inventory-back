package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jonathanCaamano/inventory-back/internal/storage/s3client"

	"github.com/jonathanCaamano/inventory-back/internal/config"
	"github.com/jonathanCaamano/inventory-back/internal/http/handlers"
	"github.com/jonathanCaamano/inventory-back/internal/http/middleware"
	"github.com/jonathanCaamano/inventory-back/internal/service"
)

func NewRouter(cfg config.Config, svc *service.ProductService, s3c *s3client.Client) http.Handler {

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-Id"},
		ExposedHeaders:   []string{"X-Request-Id"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(middleware.RequestID)
	r.Use(middleware.Recover)

	hh := handlers.NewHealth()
	auth := handlers.NewAuth(cfg)
	ph := handlers.NewProducts(cfg, svc, s3c)


	r.Get("/health", hh.Health)

	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/auth/login", auth.Login)

		api.Group(func(pr chi.Router) {
			pr.Use(middleware.JWT(cfg.JWTSecret))
				pr.Route("/products", func(rp chi.Router) {
					rp.Get("/", ph.Search)
					rp.Post("/", ph.Create)
					rp.Get("/{id}", ph.GetByID)
					rp.Put("/{id}", ph.Update)
					rp.Delete("/{id}", ph.Delete)
					rp.Post("/{id}/images/presign", ph.PresignImage)
					rp.Post("/{id}/images", ph.AddImage)
					rp.Put("/{id}/contact", ph.UpsertContact)
				})
		})
	})

	return r
}
