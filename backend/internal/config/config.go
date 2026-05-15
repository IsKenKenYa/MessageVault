package config

import (
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Driver           string
	DSN              string
	DatabaseURL      string
	ListenAddr       string
	SchemaRoot       string
	AuthSecret       string
	TLS              bool
	Env              string
	AllowedImportDir []string
}

func Load() (Config, error) {
	schemaRoot := env("COMMORY_SCHEMA_ROOT", filepath.Join("..", "msglayer", "schema", "v0.1", "root.schema.json"))
	if !filepath.IsAbs(schemaRoot) {
		if abs, err := filepath.Abs(schemaRoot); err == nil {
			schemaRoot = abs
		}
	}
	return Config{
		Driver:           env("COMMORY_DB_DRIVER", "sqlite"),
		DSN:              env("COMMORY_DB_DSN", filepath.Join(".", "data", "commory-store.json")),
		DatabaseURL:      env("COMMORY_DATABASE_URL", ""),
		ListenAddr:       env("COMMORY_LISTEN_ADDR", ":3000"),
		SchemaRoot:       schemaRoot,
		AuthSecret:       env("COMMORY_AUTH_SECRET", "commory-dev-secret"),
		TLS:              envBool("COMMORY_TLS", false),
		Env:              env("COMMORY_ENV", "development"),
		AllowedImportDir: splitAndClean(env("COMMORY_ALLOWED_IMPORT_DIRS", filepath.Join("..", "msglayer", "examples"))),
	}, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return strings.EqualFold(value, "true") || value == "1"
}

func (c Config) IsDefaultAuthSecret() bool {
	return c.AuthSecret == "commory-dev-secret"
}

func splitAndClean(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if !filepath.IsAbs(part) {
			if abs, err := filepath.Abs(part); err == nil {
				part = abs
			}
		}
		result = append(result, part)
	}
	return result
}
