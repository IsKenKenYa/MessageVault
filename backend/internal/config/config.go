package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	Driver      string
	DSN         string
	DatabaseURL string
	ListenAddr  string
	SchemaRoot  string
}

func Load() (Config, error) {
	return Config{
		Driver:      env("COMMORY_DB_DRIVER", "sqlite"),
		DSN:         env("COMMORY_DB_DSN", filepath.Join(".", "data", "commory-store.json")),
		DatabaseURL: env("COMMORY_DATABASE_URL", ""),
		ListenAddr:  env("COMMORY_LISTEN_ADDR", ":3000"),
		SchemaRoot:  env("COMMORY_SCHEMA_ROOT", filepath.Join("..", "msglayer", "schema", "v0.1", "root.schema.json")),
	}, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
