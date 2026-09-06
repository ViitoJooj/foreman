package orchestrator_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/ports"
	"github.com/ViitoJooj/foreman/internal/orchestrator"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func newLLMRunner(t *testing.T) (*orchestrator.LLMRunner, *testutil.MockLLMClient) {
	t.Helper()
	client := testutil.NewMockLLMClient(gomock.NewController(t))
	return orchestrator.NewLLMRunner(domain.RoleCoder, "system", client), client
}

func codingTask() domain.Task {
	return domain.Task{
		ID:         testutil.RandomUUID(),
		CompanyID:  testutil.RandomUUID(),
		Title:      testutil.RandomTaskTitle(),
		State:      domain.TaskCoding,
		Risk:       domain.RiskLow,
		MaxRetries: 3,
	}
}

func TestLLMRunnerStepAdvances(t *testing.T) {
	runner, client := newLLMRunner(t)

	client.EXPECT().Complete(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, req ports.LLMRequest) (ports.LLMResponse, error) {
			require.Equal(t, "system", req.System)
			require.Contains(t, req.Messages[0].Content, "estado: coding")
			require.Contains(t, req.Messages[0].Content, "build_failed")
			return ports.LLMResponse{
				Text:  `{"outcome": "advance", "note": "build verde, handoff pro tester"}`,
				Usage: ports.LLMUsage{InputTokens: 120, OutputTokens: 30},
			}, nil
		})

	res, err := runner.Step(context.Background(), codingTask())
	require.NoError(t, err)
	require.Equal(t, orchestrator.OutcomeAdvance, res.Outcome)
	require.Equal(t, "build verde, handoff pro tester", res.Note)
	require.Equal(t, 120, res.Usage.InputTokens)
}

func TestLLMRunnerStepToleratesProseAroundJSON(t *testing.T) {
	runner, client := newLLMRunner(t)
	client.EXPECT().Complete(gomock.Any(), gomock.Any()).Return(ports.LLMResponse{
		Text: "Claro! Aqui vai:\n```json\n{\"outcome\":\"build_failed\",\"note\":\"lint reprovou\"}\n```\n",
	}, nil)

	res, err := runner.Step(context.Background(), codingTask())
	require.NoError(t, err)
	require.Equal(t, orchestrator.OutcomeBuildFailed, res.Outcome)
	require.Equal(t, "lint reprovou", res.Note)
}

func TestLLMRunnerStepRejectsOutcomeNotAllowedForState(t *testing.T) {
	runner, client := newLLMRunner(t)
	client.EXPECT().Complete(gomock.Any(), gomock.Any()).Return(ports.LLMResponse{
		Text: `{"outcome": "test_failed", "note": "x"}`,
	}, nil)

	_, err := runner.Step(context.Background(), codingTask())
	require.ErrorContains(t, err, "not allowed in state")
}

func TestLLMRunnerStepRejectsNonJSON(t *testing.T) {
	runner, client := newLLMRunner(t)
	client.EXPECT().Complete(gomock.Any(), gomock.Any()).Return(ports.LLMResponse{Text: "desculpa, nao entendi"}, nil)

	_, err := runner.Step(context.Background(), codingTask())
	require.ErrorContains(t, err, "no json object")
}

func TestLLMRunnerStepPropagatesClientError(t *testing.T) {
	runner, client := newLLMRunner(t)
	client.EXPECT().Complete(gomock.Any(), gomock.Any()).Return(ports.LLMResponse{}, errors.New("429 slow down"))

	_, err := runner.Step(context.Background(), codingTask())
	require.ErrorContains(t, err, "slow down")
}

func TestLLMRunnerStepRejectsTerminalState(t *testing.T) {
	runner, _ := newLLMRunner(t)
	task := codingTask()
	task.State = domain.TaskMerged

	_, err := runner.Step(context.Background(), task)
	require.ErrorContains(t, err, "no action for state")
}
