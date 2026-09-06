package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/testutil"
)

func TestRunFailsWhenConfigIsIncomplete(t *testing.T) {
	// run() must reject a bad environment before it opens any connection.
	t.Setenv("DATABASE_URL", "")
	t.Setenv("API_KEY", testutil.RandomAPIKey())
	t.Setenv("PORT", testutil.RandomPort())

	require.Error(t, run())
}
