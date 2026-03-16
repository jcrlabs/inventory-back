package config

import (
	"os"
	"testing"
)

func TestLoad_MissingJWTSecret(t *testing.T) {
	os.Unsetenv("JWT_SECRET")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when JWT_SECRET is missing")
	}
}

func TestLoad_ShortJWTSecret(t *testing.T) {
	os.Setenv("JWT_SECRET", "tooshort")
	defer os.Unsetenv("JWT_SECRET")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when JWT_SECRET is too short")
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	os.Setenv("JWT_SECRET", "a-very-long-secret-key-that-is-at-least-32-chars")
	defer os.Unsetenv("JWT_SECRET")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.JWTSecret == "" {
		t.Error("expected non-empty JWT secret")
	}
	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.JWTAccessTTLHours != 24 {
		t.Errorf("expected default TTL 24, got %d", cfg.JWTAccessTTLHours)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	os.Setenv("JWT_SECRET", "a-very-long-secret-key-that-is-at-least-32-chars")
	os.Setenv("PORT", "9090")
	os.Setenv("ENV", "production")
	os.Setenv("JWT_ACCESS_TTL_HOURS", "48")
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com,https://api.example.com")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("PORT")
		os.Unsetenv("ENV")
		os.Unsetenv("JWT_ACCESS_TTL_HOURS")
		os.Unsetenv("CORS_ALLOWED_ORIGINS")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.JWTAccessTTLHours != 48 {
		t.Errorf("expected TTL 48, got %d", cfg.JWTAccessTTLHours)
	}
	if len(cfg.AllowedOrigins) != 2 {
		t.Errorf("expected 2 origins, got %d", len(cfg.AllowedOrigins))
	}
}

func TestLoad_InvalidTTLDefaultsTo24(t *testing.T) {
	os.Setenv("JWT_SECRET", "a-very-long-secret-key-that-is-at-least-32-chars")
	os.Setenv("JWT_ACCESS_TTL_HOURS", "-5")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_ACCESS_TTL_HOURS")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.JWTAccessTTLHours != 24 {
		t.Errorf("expected TTL to default to 24, got %d", cfg.JWTAccessTTLHours)
	}
}

func TestIsProduction(t *testing.T) {
	os.Setenv("JWT_SECRET", "a-very-long-secret-key-that-is-at-least-32-chars")
	defer os.Unsetenv("JWT_SECRET")

	tests := []struct {
		env      string
		expected bool
	}{
		{"production", true},
		{"development", false},
		{"staging", false},
	}

	for _, tt := range tests {
		os.Setenv("ENV", tt.env)
		cfg, _ := Load()
		if cfg.IsProduction() != tt.expected {
			t.Errorf("env=%s: expected IsProduction()=%v", tt.env, tt.expected)
		}
		os.Unsetenv("ENV")
	}
}

func TestDSN(t *testing.T) {
	os.Setenv("JWT_SECRET", "a-very-long-secret-key-that-is-at-least-32-chars")
	defer os.Unsetenv("JWT_SECRET")

	cfg, _ := Load()
	dsn := cfg.DSN()
	if dsn == "" {
		t.Error("expected non-empty DSN")
	}
	// Should contain key components
	for _, part := range []string{"host=", "user=", "dbname=", "port="} {
		found := false
		for i := 0; i+len(part) <= len(dsn); i++ {
			if dsn[i:i+len(part)] == part {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("DSN missing %q: %s", part, dsn)
		}
	}
}
