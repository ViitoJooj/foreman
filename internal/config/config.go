// Package config loads and validates the harness runtime configuration from
// environment variables.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultPort                 = "8080"
	defaultOrchestratorInterval = 30 * time.Second
)

// Config holds the settings the API server needs to start.
type Config struct {
	DatabaseURL string // Postgres/Supabase connection string
	APIKey      string // shared secret expected in the X-Api-Key header
	Port        string // TCP port the HTTP server listens on

	OrchestratorEnabled  bool          // run the orchestrator loop in-process
	OrchestratorInterval time.Duration // cycle interval; defaults to 30s
	BudgetHardCapUSD     float64       // orchestrator hard spend cap; 0 disables it
	LLMRunnersEnabled    bool          // use LLM-backed runners instead of stubs
}

// Load reads the configuration from the environment. DATABASE_URL and API_KEY are
// required; every other value has a default.
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

	enabled, err := boolEnv("ORCHESTRATOR_ENABLED", false)
	if err != nil {
		return Config{}, err
	}
	cfg.OrchestratorEnabled = enabled

	interval, err := durationEnv("ORCHESTRATOR_INTERVAL", defaultOrchestratorInterval)
	if err != nil {
		return Config{}, err
	}
	cfg.OrchestratorInterval = interval

	hardCap, err := floatEnv("BUDGET_HARD_CAP_USD", 0)
	if err != nil {
		return Config{}, err
	}
	cfg.BudgetHardCapUSD = hardCap

	llmRunners, err := boolEnv("LLM_RUNNERS_ENABLED", false)
	if err != nil {
		return Config{}, err
	}
	cfg.LLMRunnersEnabled = llmRunners

	return cfg, nil
}

func boolEnv(key string, fallback bool) (bool, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("config: %s must be a boolean: %w", key, err)
	}
	return v, nil
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	v, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("config: %s must be a duration: %w", key, err)
	}
	if v <= 0 {
		return 0, fmt.Errorf("config: %s must be positive", key)
	}
	return v, nil
}

func floatEnv(key string, fallback float64) (float64, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("config: %s must be a number: %w", key, err)
	}
	return v, nil
}
