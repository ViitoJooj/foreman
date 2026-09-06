package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/config"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func setEnv(t *testing.T, dsn, key, port string) {
	t.Helper()
	t.Setenv("DATABASE_URL", dsn)
	t.Setenv("API_KEY", key)
	t.Setenv("PORT", port)
	// Clear optional orchestrator settings so tests start from a known state.
	t.Setenv("ORCHESTRATOR_ENABLED", "")
	t.Setenv("ORCHESTRATOR_INTERVAL", "")
	t.Setenv("BUDGET_HARD_CAP_USD", "")
	t.Setenv("LLM_RUNNERS_ENABLED", "")
}

func TestLoadReadsAllValues(t *testing.T) {
	dsn := testutil.RandomDatabaseURL()
	key := testutil.RandomAPIKey()
	port := testutil.RandomPort()
	setEnv(t, dsn, key, port)
	t.Setenv("ORCHESTRATOR_ENABLED", "true")
	t.Setenv("ORCHESTRATOR_INTERVAL", "45s")
	t.Setenv("BUDGET_HARD_CAP_USD", "12.5")
	t.Setenv("LLM_RUNNERS_ENABLED", "1")

	cfg, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, dsn, cfg.DatabaseURL)
	require.Equal(t, key, cfg.APIKey)
	require.Equal(t, port, cfg.Port)
	require.True(t, cfg.OrchestratorEnabled)
	require.Equal(t, 45*time.Second, cfg.OrchestratorInterval)
	require.InDelta(t, 12.5, cfg.BudgetHardCapUSD, 1e-9)
	require.True(t, cfg.LLMRunnersEnabled)
}

func TestLoadDefaults(t *testing.T) {
	setEnv(t, testutil.RandomDatabaseURL(), testutil.RandomAPIKey(), "")

	cfg, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, "8080", cfg.Port)
	require.False(t, cfg.OrchestratorEnabled)
	require.Equal(t, 30*time.Second, cfg.OrchestratorInterval)
	require.Zero(t, cfg.BudgetHardCapUSD)
}

func TestLoadRejectsMalformedOrchestratorSettings(t *testing.T) {
	tests := []struct{ key, value string }{
		{"ORCHESTRATOR_ENABLED", "maybe"},
		{"ORCHESTRATOR_INTERVAL", "soon"},
		{"ORCHESTRATOR_INTERVAL", "-5s"},
		{"BUDGET_HARD_CAP_USD", "lots"},
	}
	for _, tc := range tests {
		t.Run(tc.key+"="+tc.value, func(t *testing.T) {
			setEnv(t, testutil.RandomDatabaseURL(), testutil.RandomAPIKey(), testutil.RandomPort())
			t.Setenv(tc.key, tc.value)

			_, err := config.Load()
			require.Error(t, err)
		})
	}
}

func TestLoadRequiresMandatoryValues(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
		key  string
	}{
		{"missing database url", "", testutil.RandomAPIKey()},
		{"missing api key", testutil.RandomDatabaseURL(), ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setEnv(t, tc.dsn, tc.key, testutil.RandomPort())

			_, err := config.Load()
			require.Error(t, err)
		})
	}
}
