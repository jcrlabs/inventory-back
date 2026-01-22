package config

import "os"

type Config struct {
  HTTPAddr  string
  DBURL     string
  JWTSecret string
  Env       string

  S3Endpoint        string
  S3Region          string
  S3Bucket          string
  S3AccessKey       string
  S3SecretKey       string
  S3ForcePathStyle  bool
  S3PublicBaseURL   string
}


func Load() Config {
	return Config{
		HTTPAddr:  get("HTTP_ADDR", ":8080"),
		DBURL:     must("DATABASE_URL"),
		JWTSecret: must("JWT_SECRET"),
		Env:       get("ENV", "dev"),

		S3Endpoint:       must("S3_ENDPOINT"),
		S3Region:         get("S3_REGION", "us-east-1"),
		S3Bucket:         must("S3_BUCKET"),
		S3AccessKey:      must("S3_ACCESS_KEY"),
		S3SecretKey:      must("S3_SECRET_KEY"),
		S3ForcePathStyle: get("S3_FORCE_PATH_STYLE", "true") == "true",
		S3PublicBaseURL:  get("S3_PUBLIC_BASE_URL", ""),

	}
}

func get(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func must(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic("missing env: " + k)
	}
	return v
}
