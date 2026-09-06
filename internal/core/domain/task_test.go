package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func TestTaskCanAutoMerge(t *testing.T) {
	tests := []struct {
		name string
		risk domain.RiskLevel
		want bool
	}{
		{"low risk auto-merges", domain.RiskLow, true},
		{"medium risk needs a human", domain.RiskMedium, false},
		{"high risk needs a human", domain.RiskHigh, false},
		{"unknown risk never auto-merges", domain.RiskLevel(testutil.RandomSlug()), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			task := domain.Task{ID: testutil.RandomUUID(), Risk: tc.risk}
			require.Equal(t, tc.want, task.CanAutoMerge())
		})
	}
}

func TestTaskRetriesExhausted(t *testing.T) {
	tests := []struct {
		name       string
		retries    int
		maxRetries int
		want       bool
	}{
		{"below cap", 1, 3, false},
		{"one before cap", 2, 3, false},
		{"at cap", 3, 3, true},
		{"past cap", 5, 3, true},
		{"zero cap is immediately exhausted", 0, 0, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			task := domain.Task{
				ID:         testutil.RandomUUID(),
				Retries:    tc.retries,
				MaxRetries: tc.maxRetries,
			}
			require.Equal(t, tc.want, task.RetriesExhausted())
		})
	}
}

// TestTaskEnumValues pins the wire values of the task enums, which must match the
// task_state and risk_level Postgres enums in migrations/0001_init.up.sql.
func TestTaskEnumValues(t *testing.T) {
	states := map[domain.TaskState]string{
		domain.TaskCreated:     "created",
		domain.TaskResearching: "researching",
		domain.TaskQueued:      "queued",
		domain.TaskCoding:      "coding",
		domain.TaskTesting:     "testing",
		domain.TaskReviewing:   "reviewing",
		domain.TaskMerged:      "merged",
		domain.TaskNeedsHuman:  "needs_human",
		domain.TaskRejected:    "rejected",
		domain.TaskFailedBuild: "failed_build",
		domain.TaskFailedTest:  "failed_test",
	}
	for state, want := range states {
		require.Equal(t, want, string(state))
	}

	risks := map[domain.RiskLevel]string{
		domain.RiskLow:    "low",
		domain.RiskMedium: "medium",
		domain.RiskHigh:   "high",
	}
	for risk, want := range risks {
		require.Equal(t, want, string(risk))
	}
}
