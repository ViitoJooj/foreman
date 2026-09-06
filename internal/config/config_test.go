package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/config"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func setEnv(t *testing.T, dsn, key, port string) {
	t.Helper()
	t.Setenv("DATABASE_URL", dsn)
	t.Setenv("API_KEY", key)
	t.Setenv("PORT", port)
}

func TestLoadReadsAllValues(t *testing.T) {
	dsn := testutil.RandomDatabaseURL()
	key := testutil.RandomAPIKey()
	port := testutil.RandomPort()
	setEnv(t, dsn, key, port)

	cfg, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, dsn, cfg.DatabaseURL)
	require.Equal(t, key, cfg.APIKey)
	require.Equal(t, port, cfg.Port)
}

func TestLoadDefaultsPort(t *testing.T) {
	setEnv(t, testutil.RandomDatabaseURL(), testutil.RandomAPIKey(), "")

	cfg, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, "8080", cfg.Port)
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
