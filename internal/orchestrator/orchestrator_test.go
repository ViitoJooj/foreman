package orchestrator_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/orchestrator"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

type funcRunner struct {
	role domain.AgentRole
	fn   func(domain.Task) (orchestrator.Outcome, error)
}

func (r funcRunner) Role() domain.AgentRole { return r.role }

func (r funcRunner) Step(_ context.Context, t domain.Task) (orchestrator.Outcome, error) {
	return r.fn(t)
}

type orchFixture struct {
	orch     *orchestrator.Orchestrator
	tasks    *testutil.MockTaskRepository
	agents   *testutil.MockAgentRepository
	channels *testutil.MockChannelRepository
	messages *testutil.MockMessageRepository
	bus      *testutil.MockMessageBus
	commands *testutil.MockCommandRepository
}

func newFixture(t *testing.T, runners map[domain.AgentRole]orchestrator.AgentRunner, budget *orchestrator.Budget) *orchFixture {
	t.Helper()
	ctrl := gomock.NewController(t)

	f := &orchFixture{
		tasks:    testutil.NewMockTaskRepository(ctrl),
		agents:   testutil.NewMockAgentRepository(ctrl),
		channels: testutil.NewMockChannelRepository(ctrl),
		messages: testutil.NewMockMessageRepository(ctrl),
		bus:      testutil.NewMockMessageBus(ctrl),
		commands: testutil.NewMockCommandRepository(ctrl),
	}
	f.orch = orchestrator.New(orchestrator.Deps{
		Tasks:    f.tasks,
		Agents:   f.agents,
		Channels: f.channels,
		Messages: f.messages,
		Bus:      f.bus,
		Commands: f.commands,
		Runners:  runners,
		Budget:   budget,
		Interval: time.Millisecond,
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	return f
}

func advancingRunners() map[domain.AgentRole]orchestrator.AgentRunner {
	adv := func(domain.Task) (orchestrator.Outcome, error) { return orchestrator.OutcomeAdvance, nil }
	return map[domain.AgentRole]orchestrator.AgentRunner{
		domain.RoleCoder:      funcRunner{domain.RoleCoder, adv},
		domain.RoleTester:     funcRunner{domain.RoleTester, adv},
		domain.RolePRReviewer: funcRunner{domain.RolePRReviewer, adv},
	}
}

// expectNoPendingCommands wires the command lookup to return nothing.
func (f *orchFixture) expectNoPendingCommands() {
	f.commands.EXPECT().ListByStatus(gomock.Any(), domain.CommandPending).Return(nil, nil)
}

// expectTasksByState returns byState[s] for each actionable state.
func (f *orchFixture) expectTasksByState(byState map[domain.TaskState][]domain.Task) {
	f.tasks.EXPECT().ListByState(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, s domain.TaskState) ([]domain.Task, error) {
			return byState[s], nil
		}).AnyTimes()
}

func TestCycleAdvancesQueuedTask(t *testing.T) {
	f := newFixture(t, advancingRunners(), nil)

	companyID := testutil.RandomUUID()
	coder := domain.Agent{ID: testutil.RandomUUID(), CompanyID: companyID, Name: "Ada", Role: domain.RoleCoder}
	channel := domain.Channel{ID: testutil.RandomUUID(), CompanyID: companyID, Name: "general"}
	queued := domain.Task{ID: testutil.RandomUUID(), CompanyID: companyID, State: domain.TaskQueued, Risk: domain.RiskLow, MaxRetries: 3}

	f.expectNoPendingCommands()
	f.expectTasksByState(map[domain.TaskState][]domain.Task{domain.TaskQueued: {queued}})
	f.agents.EXPECT().ListByCompany(gomock.Any(), companyID).Return([]domain.Agent{coder}, nil)

	var saved domain.Task
	f.tasks.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, tk domain.Task) (domain.Task, error) {
			saved = tk
			return tk, nil
		})
	f.channels.EXPECT().ListByCompany(gomock.Any(), companyID).Return([]domain.Channel{channel}, nil)
	f.messages.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, m domain.Message) (domain.Message, error) {
			require.Equal(t, channel.ID, m.ChannelID)
			require.Equal(t, coder.ID, m.FromAgentID)
			require.Equal(t, domain.MessageStatus, m.Type)
			m.ID = testutil.RandomUUID()
			return m, nil
		})
	f.bus.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil)

	require.NoError(t, f.orch.Cycle(context.Background()))
	require.Equal(t, domain.TaskCoding, saved.State)
	require.Equal(t, coder.ID, saved.AssigneeID)
}

func TestCycleKillSwitchStopsAdvancement(t *testing.T) {
	f := newFixture(t, advancingRunners(), nil)

	kill := domain.Command{ID: testutil.RandomUUID(), Kind: domain.CommandKill, Status: domain.CommandPending}
	f.commands.EXPECT().ListByStatus(gomock.Any(), domain.CommandPending).Return([]domain.Command{kill}, nil)
	f.commands.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, c domain.Command) (domain.Command, error) {
			require.Equal(t, domain.CommandApplied, c.Status)
			require.NotNil(t, c.AppliedAt)
			return c, nil
		})
	// No ListByState / Update calls: the kill switch is engaged.

	require.NoError(t, f.orch.Cycle(context.Background()))
}

func TestCyclePauseSkipsThatAgentsTask(t *testing.T) {
	f := newFixture(t, advancingRunners(), nil)

	companyID := testutil.RandomUUID()
	coder := domain.Agent{ID: testutil.RandomUUID(), CompanyID: companyID, Role: domain.RoleCoder}
	queued := domain.Task{ID: testutil.RandomUUID(), CompanyID: companyID, State: domain.TaskQueued, Risk: domain.RiskLow, MaxRetries: 3}
	pause := domain.Command{ID: testutil.RandomUUID(), Kind: domain.CommandPause, TargetAgentID: coder.ID, Status: domain.CommandPending}

	f.commands.EXPECT().ListByStatus(gomock.Any(), domain.CommandPending).Return([]domain.Command{pause}, nil)
	f.commands.EXPECT().Update(gomock.Any(), gomock.Any()).Return(pause, nil)
	f.expectTasksByState(map[domain.TaskState][]domain.Task{domain.TaskQueued: {queued}})
	f.agents.EXPECT().ListByCompany(gomock.Any(), companyID).Return([]domain.Agent{coder}, nil)
	// No task Update: the coder is paused.

	require.NoError(t, f.orch.Cycle(context.Background()))
}

func TestCycleBudgetExhaustedStopsAdvancement(t *testing.T) {
	budget := orchestrator.NewBudget(1)
	budget.Record(2)
	f := newFixture(t, advancingRunners(), budget)

	f.expectNoPendingCommands()
	// No ListByState: budget blocks the cycle.

	require.NoError(t, f.orch.Cycle(context.Background()))
}

func TestCycleMissingRoleAgentDoesNotUpdate(t *testing.T) {
	f := newFixture(t, advancingRunners(), nil)

	companyID := testutil.RandomUUID()
	queued := domain.Task{ID: testutil.RandomUUID(), CompanyID: companyID, State: domain.TaskQueued, Risk: domain.RiskLow, MaxRetries: 3}

	f.expectNoPendingCommands()
	f.expectTasksByState(map[domain.TaskState][]domain.Task{domain.TaskQueued: {queued}})
	f.agents.EXPECT().ListByCompany(gomock.Any(), companyID).Return(nil, nil)
	// No Update: there is no coder agent.

	require.NoError(t, f.orch.Cycle(context.Background()))
}

func TestCycleBuildFailureMovesTaskToFailedBuild(t *testing.T) {
	buildFails := map[domain.AgentRole]orchestrator.AgentRunner{
		domain.RoleCoder: funcRunner{domain.RoleCoder, func(domain.Task) (orchestrator.Outcome, error) {
			return orchestrator.OutcomeBuildFailed, nil
		}},
	}
	f := newFixture(t, buildFails, nil)

	companyID := testutil.RandomUUID()
	coder := domain.Agent{ID: testutil.RandomUUID(), CompanyID: companyID, Role: domain.RoleCoder}
	coding := domain.Task{ID: testutil.RandomUUID(), CompanyID: companyID, State: domain.TaskCoding, Risk: domain.RiskLow, MaxRetries: 3}

	f.expectNoPendingCommands()
	f.expectTasksByState(map[domain.TaskState][]domain.Task{domain.TaskCoding: {coding}})
	f.agents.EXPECT().ListByCompany(gomock.Any(), companyID).Return([]domain.Agent{coder}, nil)

	var saved domain.Task
	f.tasks.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, tk domain.Task) (domain.Task, error) { saved = tk; return tk, nil })
	f.channels.EXPECT().ListByCompany(gomock.Any(), companyID).Return([]domain.Channel{{ID: testutil.RandomUUID()}}, nil)
	f.messages.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, m domain.Message) (domain.Message, error) {
			m.ID = testutil.RandomUUID()
			return m, nil
		})
	f.bus.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil)

	require.NoError(t, f.orch.Cycle(context.Background()))
	require.Equal(t, domain.TaskFailedBuild, saved.State)
}

func TestNewAppliesDefaults(t *testing.T) {
	o := orchestrator.New(orchestrator.Deps{})
	require.NotNil(t, o)
}

func TestRunStopsOnContextCancel(t *testing.T) {
	f := newFixture(t, advancingRunners(), nil)
	f.commands.EXPECT().ListByStatus(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	f.expectTasksByState(nil)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := f.orch.Run(ctx)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
