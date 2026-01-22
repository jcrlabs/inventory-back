package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jonathanCaamano/inventory-back/internal/storage/s3client"

	"github.com/jonathanCaamano/inventory-back/internal/config"
	ihttp "github.com/jonathanCaamano/inventory-back/internal/http"
	"github.com/jonathanCaamano/inventory-back/internal/service"
	"github.com/jonathanCaamano/inventory-back/internal/store/postgres"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	repo := postgres.NewProductRepo(pool)
	svc := service.NewProductService(repo)
	s3c, err := s3client.New(ctx, s3client.Options{
		Endpoint:       cfg.S3Endpoint,
		Region:         cfg.S3Region,
		Bucket:         cfg.S3Bucket,
		AccessKey:      cfg.S3AccessKey,
		SecretKey:      cfg.S3SecretKey,
		ForcePathStyle: cfg.S3ForcePathStyle,
		PublicBaseURL:  cfg.S3PublicBaseURL,
	})
	if err != nil {
		log.Fatal(err)
	}

	h := ihttp.NewRouter(cfg, svc, s3c)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
