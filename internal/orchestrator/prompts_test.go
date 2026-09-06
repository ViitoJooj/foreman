package orchestrator_test

import (
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/orchestrator"
)

func TestSystemPromptForKnownRoles(t *testing.T) {
	for _, role := range []domain.AgentRole{
		domain.RoleTaskCreator, domain.RoleCoder, domain.RoleTester, domain.RolePRReviewer,
	} {
		p, ok := orchestrator.SystemPromptFor(role)
		require.True(t, ok)
		require.NotEmpty(t, p)
	}

	_, ok := orchestrator.SystemPromptFor(domain.AgentRole("nope"))
	require.False(t, ok)
}

func TestLLMRunnersCoverDispatchedRoles(t *testing.T) {
	runners := orchestrator.LLMRunners(nil, slog.New(slog.NewTextHandler(io.Discard, nil)))

	for _, state := range orchestrator.ActionableStates() {
		role, ok := orchestrator.RoleForState(state)
		require.True(t, ok)
		runner, present := runners[role]
		require.True(t, present, "no LLM runner for role %q", role)
		require.Equal(t, role, runner.Role())
	}
}
