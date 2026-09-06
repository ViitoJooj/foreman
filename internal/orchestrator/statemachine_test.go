package orchestrator_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/orchestrator"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func task(state domain.TaskState, risk domain.RiskLevel) domain.Task {
	return domain.Task{
		ID:         testutil.RandomUUID(),
		CompanyID:  testutil.RandomUUID(),
		Title:      testutil.RandomTaskTitle(),
		State:      state,
		Risk:       risk,
		MaxRetries: 3,
	}
}

func TestNextHappyPath(t *testing.T) {
	tests := []struct {
		from domain.TaskState
		want domain.TaskState
	}{
		{domain.TaskQueued, domain.TaskCoding},
		{domain.TaskCoding, domain.TaskTesting},
		{domain.TaskTesting, domain.TaskReviewing},
	}
	for _, tc := range tests {
		t.Run(string(tc.from), func(t *testing.T) {
			got, err := orchestrator.Next(task(tc.from, domain.RiskLow), orchestrator.OutcomeAdvance)
			require.NoError(t, err)
			require.Equal(t, tc.want, got.State)
		})
	}
}

func TestNextReviewingAutoMergesLowRisk(t *testing.T) {
	got, err := orchestrator.Next(task(domain.TaskReviewing, domain.RiskLow), orchestrator.OutcomeAdvance)
	require.NoError(t, err)
	require.Equal(t, domain.TaskMerged, got.State)
}

func TestNextReviewingEscalatesRiskyTask(t *testing.T) {
	for _, risk := range []domain.RiskLevel{domain.RiskMedium, domain.RiskHigh} {
		got, err := orchestrator.Next(task(domain.TaskReviewing, risk), orchestrator.OutcomeAdvance)
		require.NoError(t, err)
		require.Equal(t, domain.TaskNeedsHuman, got.State)
	}
}

func TestNextReviewingRejected(t *testing.T) {
	got, err := orchestrator.Next(task(domain.TaskReviewing, domain.RiskLow), orchestrator.OutcomeRejected)
	require.NoError(t, err)
	require.Equal(t, domain.TaskRejected, got.State)
}

func TestNextBuildAndTestFailures(t *testing.T) {
	got, err := orchestrator.Next(task(domain.TaskCoding, domain.RiskLow), orchestrator.OutcomeBuildFailed)
	require.NoError(t, err)
	require.Equal(t, domain.TaskFailedBuild, got.State)

	got, err = orchestrator.Next(task(domain.TaskTesting, domain.RiskLow), orchestrator.OutcomeTestFailed)
	require.NoError(t, err)
	require.Equal(t, domain.TaskFailedTest, got.State)
}

func TestNextRetryIncrementsUntilExhausted(t *testing.T) {
	for _, from := range []domain.TaskState{domain.TaskFailedBuild, domain.TaskFailedTest} {
		t.Run(string(from), func(t *testing.T) {
			tk := task(from, domain.RiskLow)
			tk.Retries = 1

			got, err := orchestrator.Next(tk, orchestrator.OutcomeAdvance)
			require.NoError(t, err)
			require.Equal(t, domain.TaskCoding, got.State)
			require.Equal(t, 2, got.Retries)

			tk.Retries = tk.MaxRetries
			got, err = orchestrator.Next(tk, orchestrator.OutcomeAdvance)
			require.NoError(t, err)
			require.Equal(t, domain.TaskNeedsHuman, got.State)
		})
	}
}

func TestNextNeedsHumanFromAnyStage(t *testing.T) {
	for _, from := range orchestrator.ActionableStates() {
		got, err := orchestrator.Next(task(from, domain.RiskLow), orchestrator.OutcomeNeedsHuman)
		require.NoError(t, err)
		require.Equal(t, domain.TaskNeedsHuman, got.State)
	}
}

func TestNextRejectsInvalidTransitions(t *testing.T) {
	tests := []struct {
		state   domain.TaskState
		outcome orchestrator.Outcome
	}{
		{domain.TaskQueued, orchestrator.OutcomeBuildFailed},
		{domain.TaskCoding, orchestrator.OutcomeTestFailed},
		{domain.TaskReviewing, orchestrator.OutcomeBuildFailed},
		{domain.TaskMerged, orchestrator.OutcomeAdvance},
		{domain.TaskCreated, orchestrator.OutcomeAdvance},
		{domain.TaskNeedsHuman, orchestrator.OutcomeAdvance},
	}
	for _, tc := range tests {
		t.Run(string(tc.state)+"_"+string(tc.outcome), func(t *testing.T) {
			_, err := orchestrator.Next(task(tc.state, domain.RiskLow), tc.outcome)
			require.Error(t, err)
		})
	}
}

func TestActionableStatesIsACopy(t *testing.T) {
	a := orchestrator.ActionableStates()
	a[0] = domain.TaskMerged
	require.NotEqual(t, domain.TaskMerged, orchestrator.ActionableStates()[0])
}
