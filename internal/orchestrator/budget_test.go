package orchestrator_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/orchestrator"
)

func TestBudgetDisabledWithNonPositiveCap(t *testing.T) {
	b := orchestrator.NewBudget(0)
	b.Record(1_000_000)
	require.True(t, b.Allow())
}

func TestBudgetBlocksAtCap(t *testing.T) {
	b := orchestrator.NewBudget(10)
	require.True(t, b.Allow())

	b.Record(7.5)
	require.True(t, b.Allow())
	require.InDelta(t, 7.5, b.SpentUSD(), 1e-9)

	b.Record(2.5)
	require.False(t, b.Allow(), "spend reached the cap")
}

func TestBudgetRecordTokens(t *testing.T) {
	b := orchestrator.NewBudget(0)

	b.RecordTokens(0, 0)
	require.Zero(t, b.SpentUSD())

	b.RecordTokens(500_000, 500_000) // 1M tokens
	require.InDelta(t, 6.0, b.SpentUSD(), 1e-9, "blended rate is $6 / million tokens")
}
