package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr  string
	DBURL     string
	JWTSecret string
	JWTTTL    time.Duration
	Env       string

	BootstrapAdminUsername string
	BootstrapAdminPassword string

	S3Endpoint       string
	S3Region         string
	S3Bucket         string
	S3AccessKey      string
	S3SecretKey      string
	S3ForcePathStyle bool
	S3PublicBaseURL  string

	DevNoDeps bool
}

func Load() Config {
	env := get("ENV", "dev")

	cfg := Config{
		Env:       env,
		HTTPAddr:  get("HTTP_ADDR", ":8080"),
		JWTSecret: get("JWT_SECRET", "dev-secret"),
		JWTTTL:    time.Duration(getInt("JWT_TTL_MINUTES", 30)) * time.Minute,

		BootstrapAdminUsername: get("BOOTSTRAP_ADMIN_USERNAME", "admin"),
		BootstrapAdminPassword: get("BOOTSTRAP_ADMIN_PASSWORD", "admin"),
		DBURL:                  get("DATABASE_URL", "postgres://user:pass@test.example.com:5432/inventory?sslmode=disable"),

		S3Endpoint:       get("S3_ENDPOINT", "http://test.example.com"),
		S3Region:         get("S3_REGION", "us-east-1"),
		S3Bucket:         get("S3_BUCKET", "inventory"),
		S3AccessKey:      get("S3_ACCESS_KEY", "test"),
		S3SecretKey:      get("S3_SECRET_KEY", "test"),
		S3ForcePathStyle: get("S3_FORCE_PATH_STYLE", "true") == "true",
		S3PublicBaseURL:  get("S3_PUBLIC_BASE_URL", "http://test.example.com/inventory"),

		DevNoDeps: get("DEV_NO_DEPS", "false") == "true",
	}

	// En producción, exigimos variables críticas y no permitimos bootstrap por defecto
	if isProd(env) {
		if cfg.JWTSecret == "dev-secret" {
			panic("JWT_SECRET must be set in production")
		}
		if cfg.BootstrapAdminUsername == "admin" && cfg.BootstrapAdminPassword == "admin" {
			panic("BOOTSTRAP_ADMIN_USERNAME/BOOTSTRAP_ADMIN_PASSWORD must be set in production")
		}
		if cfg.DBURL == "" {
			panic("DATABASE_URL must be set in production")
		}
		if cfg.S3Endpoint == "" || cfg.S3AccessKey == "" || cfg.S3SecretKey == "" {
			panic("S3 configuration must be set in production")
		}
	}

	return cfg
}

func get(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func getInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func isProd(env string) bool {
	return env == "prod" || env == "production"
}
