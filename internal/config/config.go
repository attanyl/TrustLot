package config

import "os"

// Config holds application configuration loaded from environment variables.
type Config struct {
	DatabaseURL string
	RedisURL    string
	HTTPAddr    string
	Environment string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	return Config{
		DatabaseURL: envOr("DATABASE_URL", "postgres://trustlot:trustlot@localhost:5432/trustlot?sslmode=disable"),
		RedisURL:    envOr("REDIS_URL", "redis://localhost:6379/0"),
		HTTPAddr:    envOr("HTTP_ADDR", ":8080"),
		Environment: envOr("ENVIRONMENT", "development"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
