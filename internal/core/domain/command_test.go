package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

// TestCommandEnumValues pins the wire values of the command enums, which must
// match the command_kind and command_status Postgres enums in
// migrations/0001_init.up.sql.
func TestCommandEnumValues(t *testing.T) {
	kinds := map[domain.CommandKind]string{
		domain.CommandKill:   "kill",
		domain.CommandPanic:  "panic",
		domain.CommandPause:  "pause",
		domain.CommandResume: "resume",
	}
	for kind, want := range kinds {
		require.Equal(t, want, string(kind))
	}

	statuses := map[domain.CommandStatus]string{
		domain.CommandPending: "pending",
		domain.CommandApplied: "applied",
		domain.CommandFailed:  "failed",
	}
	for status, want := range statuses {
		require.Equal(t, want, string(status))
	}
}

func TestCommandAppliedAtIsOptional(t *testing.T) {
	pending := domain.Command{
		ID:     testutil.RandomUUID(),
		Kind:   testutil.RandomCommandKind(),
		Status: domain.CommandPending,
	}
	require.Nil(t, pending.AppliedAt)

	now := time.Now()
	applied := pending
	applied.Status = domain.CommandApplied
	applied.AppliedAt = &now
	require.NotNil(t, applied.AppliedAt)
	require.Equal(t, now, *applied.AppliedAt)
}
