package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Clear any env vars that might be set
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("REDIS_URL")
	os.Unsetenv("HTTP_ADDR")
	os.Unsetenv("ENVIRONMENT")

	cfg := Load()

	if cfg.DatabaseURL != "postgres://trustlot:trustlot@localhost:5432/trustlot?sslmode=disable" {
		t.Errorf("unexpected DatabaseURL: %s", cfg.DatabaseURL)
	}
	if cfg.RedisURL != "redis://localhost:6379/0" {
		t.Errorf("unexpected RedisURL: %s", cfg.RedisURL)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Errorf("unexpected HTTPAddr: %s", cfg.HTTPAddr)
	}
	if cfg.Environment != "development" {
		t.Errorf("unexpected Environment: %s", cfg.Environment)
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://test:test@db:5432/test")
	os.Setenv("HTTP_ADDR", ":9090")
	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("HTTP_ADDR")

	cfg := Load()

	if cfg.DatabaseURL != "postgres://test:test@db:5432/test" {
		t.Errorf("expected DATABASE_URL override, got %s", cfg.DatabaseURL)
	}
	if cfg.HTTPAddr != ":9090" {
		t.Errorf("expected HTTP_ADDR override, got %s", cfg.HTTPAddr)
	}
}
