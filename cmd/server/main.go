package main

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jonathanCaamano/inventory-back/internal/config"
	"github.com/jonathanCaamano/inventory-back/internal/database"
	"github.com/jonathanCaamano/inventory-back/internal/handlers"
	"github.com/jonathanCaamano/inventory-back/internal/middleware"
	"github.com/jonathanCaamano/inventory-back/internal/models"
	"github.com/jonathanCaamano/inventory-back/internal/repository"
	"github.com/jonathanCaamano/inventory-back/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// MinIO (optional - graceful degradation)
	var minioSvc *services.MinIOService
	minioSvc, err = services.NewMinIOService(cfg)
	if err != nil {
		log.Printf("Warning: MinIO not available: %v", err)
	}

	// Repositories
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	// Services
	authSvc := services.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpirationHours)

	// Handlers
	authHandler := handlers.NewAuthHandler(authSvc, userRepo)
	userHandler := handlers.NewUserHandler(userRepo, authSvc)
	productHandler := handlers.NewProductHandler(productRepo, categoryRepo, minioSvc)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)

	// Seed default admin user if no users exist
	seedAdmin(userRepo, authSvc)

	// Router
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")

	// Public routes
	api.POST("/auth/login", authHandler.Login)

	// Authenticated routes
	auth := api.Group("")
	auth.Use(middleware.AuthRequired(authSvc))
	{
		auth.GET("/auth/me", authHandler.Me)

		// Products (viewer+)
		auth.GET("/products", productHandler.List)
		auth.GET("/products/:id", productHandler.Get)

		// Products write (manager+)
		manage := auth.Group("")
		manage.Use(middleware.RequireRole(models.RoleAdmin, models.RoleManager))
		{
			manage.POST("/products", productHandler.Create)
			manage.PUT("/products/:id", productHandler.Update)
			manage.POST("/products/:id/image", productHandler.UploadImage)
		}

		// Products delete (admin only)
		adminOnly := auth.Group("")
		adminOnly.Use(middleware.RequireRole(models.RoleAdmin))
		{
			adminOnly.DELETE("/products/:id", productHandler.Delete)
		}

		// Categories (viewer+)
		auth.GET("/categories", categoryHandler.List)
		auth.GET("/categories/:id", categoryHandler.Get)

		// Categories write (manager+)
		manage.POST("/categories", categoryHandler.Create)
		manage.PUT("/categories/:id", categoryHandler.Update)
		adminOnly.DELETE("/categories/:id", categoryHandler.Delete)

		// Users (admin only)
		adminOnly.GET("/users", userHandler.List)
		adminOnly.GET("/users/:id", userHandler.Get)
		adminOnly.POST("/users", userHandler.Create)
		adminOnly.PUT("/users/:id", userHandler.Update)
		adminOnly.DELETE("/users/:id", userHandler.Delete)
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func seedAdmin(userRepo *repository.UserRepository, authSvc *services.AuthService) {
	users, err := userRepo.FindAll()
	if err != nil || len(users) > 0 {
		return
	}

	hash, err := authSvc.HashPassword("admin123!")
	if err != nil {
		log.Printf("Failed to seed admin: %v", err)
		return
	}

	admin := &models.User{
		Username:     "admin",
		Email:        "admin@inventory.local",
		PasswordHash: hash,
		Role:         models.RoleAdmin,
		Active:       true,
	}

	if err := userRepo.Create(admin); err != nil {
		log.Printf("Failed to seed admin user: %v", err)
		return
	}

	log.Println("Admin user seeded: admin / admin123!")
}
