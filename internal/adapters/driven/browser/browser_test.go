package browser_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/adapters/driven/browser"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func TestNotImplementedResearch(t *testing.T) {
	out, err := browser.NotImplemented{}.Research(context.Background(), testutil.RandomText())
	require.ErrorIs(t, err, browser.ErrNotImplemented)
	require.Empty(t, out)
}
