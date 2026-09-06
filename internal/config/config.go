// Package config loads and validates the harness runtime configuration from
// environment variables.
package config

import (
	"errors"
	"os"
)

const defaultPort = "8080"

// Config holds the settings the API server needs to start.
type Config struct {
	DatabaseURL string // Postgres/Supabase connection string
	APIKey      string // shared secret expected in the X-Api-Key header
	Port        string // TCP port the HTTP server listens on
}

// Load reads the configuration from the environment. DATABASE_URL and API_KEY are
// required; PORT defaults to 8080 when unset.
func Load() (Config, error) {
	cfg := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		APIKey:      os.Getenv("API_KEY"),
		Port:        os.Getenv("PORT"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("config: DATABASE_URL is required")
	}
	if cfg.APIKey == "" {
		return Config{}, errors.New("config: API_KEY is required")
	}
	if cfg.Port == "" {
		cfg.Port = defaultPort
	}
	return cfg, nil
}
