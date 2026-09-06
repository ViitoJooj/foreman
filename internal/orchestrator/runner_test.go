package orchestrator_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/orchestrator"
)

func TestRoleForState(t *testing.T) {
	tests := []struct {
		state domain.TaskState
		want  domain.AgentRole
		ok    bool
	}{
		{domain.TaskQueued, domain.RoleCoder, true},
		{domain.TaskCoding, domain.RoleCoder, true},
		{domain.TaskFailedBuild, domain.RoleCoder, true},
		{domain.TaskFailedTest, domain.RoleCoder, true},
		{domain.TaskTesting, domain.RoleTester, true},
		{domain.TaskReviewing, domain.RolePRReviewer, true},
		{domain.TaskCreated, "", false},
		{domain.TaskMerged, "", false},
	}
	for _, tc := range tests {
		t.Run(string(tc.state), func(t *testing.T) {
			got, ok := orchestrator.RoleForState(tc.state)
			require.Equal(t, tc.ok, ok)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestStubRunnerAdvances(t *testing.T) {
	r := orchestrator.NewStubRunner(domain.RoleCoder)
	require.Equal(t, domain.RoleCoder, r.Role())

	res, err := r.Step(context.Background(), domain.Task{})
	require.NoError(t, err)
	require.Equal(t, orchestrator.OutcomeAdvance, res.Outcome)
	require.Empty(t, res.Note)
	require.Zero(t, res.Usage.InputTokens)
}

func TestStubRunnersCoverEveryActionableRole(t *testing.T) {
	runners := orchestrator.StubRunners()
	for _, state := range orchestrator.ActionableStates() {
		role, ok := orchestrator.RoleForState(state)
		require.True(t, ok)
		_, present := runners[role]
		require.True(t, present, "no stub runner for role %q (state %q)", role, state)
	}
}
