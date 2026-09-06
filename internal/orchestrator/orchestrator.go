package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"uuid"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/ports"
)

const defaultInterval = 30 * time.Second

// Deps are the collaborators the orchestrator needs. All persistence is reached
// through ports so the loop is testable without a database.
type Deps struct {
	Tasks    ports.TaskRepository
	Agents   ports.AgentRepository
	Channels ports.ChannelRepository
	Messages ports.MessageRepository
	Bus      ports.MessageBus
	Commands ports.CommandRepository
	Runners  map[domain.AgentRole]AgentRunner
	Budget   *Budget
	Interval time.Duration
	Logger   *slog.Logger
}

// Orchestrator advances tasks through the pipeline, applies operator commands
// and respects the budget guard.
type Orchestrator struct {
	deps Deps

	mu           sync.Mutex
	stopped      bool
	pausedAgents map[uuid.UUID]bool
}

// New builds an Orchestrator, filling in sensible defaults.
func New(d Deps) *Orchestrator {
	if d.Logger == nil {
		d.Logger = slog.Default()
	}
	if d.Interval <= 0 {
		d.Interval = defaultInterval
	}
	if d.Budget == nil {
		d.Budget = NewBudget(0)
	}
	if d.Runners == nil {
		d.Runners = map[domain.AgentRole]AgentRunner{}
	}
	return &Orchestrator{deps: d, pausedAgents: make(map[uuid.UUID]bool)}
}

// Run drives one cycle per interval tick until ctx is cancelled, at which point
// it returns ctx.Err().
func (o *Orchestrator) Run(ctx context.Context) error {
	ticker := time.NewTicker(o.deps.Interval)
	defer ticker.Stop()

	o.deps.Logger.Info("orchestrator started", "interval", o.deps.Interval)
	for {
		select {
		case <-ctx.Done():
			o.deps.Logger.Info("orchestrator stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := o.Cycle(ctx); err != nil {
				o.deps.Logger.Error("cycle failed", "err", err)
			}
		}
	}
}

// Cycle runs a single pass: apply pending commands, then advance every
// actionable task one step unless the kill switch or budget stops it.
func (o *Orchestrator) Cycle(ctx context.Context) error {
	if err := o.applyPendingCommands(ctx); err != nil {
		return fmt.Errorf("apply commands: %w", err)
	}

	if o.isStopped() {
		o.deps.Logger.Info("kill switch engaged; not advancing tasks")
		return nil
	}
	if !o.deps.Budget.Allow() {
		o.deps.Logger.Warn("budget exhausted; not advancing tasks", "spent_usd", o.deps.Budget.SpentUSD())
		return nil
	}

	return o.advanceTasks(ctx)
}

func (o *Orchestrator) applyPendingCommands(ctx context.Context) error {
	pending, err := o.deps.Commands.ListByStatus(ctx, domain.CommandPending)
	if err != nil {
		return err
	}
	for _, cmd := range pending {
		o.apply(cmd)

		cmd.Status = domain.CommandApplied
		applied := time.Now().UTC()
		cmd.AppliedAt = &applied
		if _, err := o.deps.Commands.Update(ctx, cmd); err != nil {
			o.deps.Logger.Error("marking command applied", "command_id", cmd.ID, "err", err)
		}
	}
	return nil
}

func (o *Orchestrator) apply(cmd domain.Command) {
	o.mu.Lock()
	defer o.mu.Unlock()
	switch cmd.Kind {
	case domain.CommandKill, domain.CommandPanic:
		o.stopped = true
	case domain.CommandPause:
		o.pausedAgents[cmd.TargetAgentID] = true
	case domain.CommandResume:
		delete(o.pausedAgents, cmd.TargetAgentID)
	}
}

func (o *Orchestrator) advanceTasks(ctx context.Context) error {
	for _, state := range ActionableStates() {
		tasks, err := o.deps.Tasks.ListByState(ctx, state)
		if err != nil {
			return fmt.Errorf("list %s tasks: %w", state, err)
		}
		for _, task := range tasks {
			if err := o.advance(ctx, task); err != nil {
				o.deps.Logger.Error("advancing task", "task_id", task.ID, "state", task.State, "err", err)
			}
		}
	}
	return nil
}

func (o *Orchestrator) advance(ctx context.Context, task domain.Task) error {
	role, ok := RoleForState(task.State)
	if !ok {
		return nil
	}
	runner, ok := o.deps.Runners[role]
	if !ok {
		return fmt.Errorf("no runner for role %q", role)
	}

	agent, err := o.agentFor(ctx, task.CompanyID, role)
	if err != nil {
		return err
	}
	if o.isPaused(agent.ID) {
		o.deps.Logger.Info("agent paused; skipping task", "agent_id", agent.ID, "task_id", task.ID)
		return nil
	}

	result, err := runner.Step(ctx, task)
	if err != nil {
		return fmt.Errorf("runner step: %w", err)
	}
	o.deps.Budget.RecordTokens(result.Usage.InputTokens, result.Usage.OutputTokens)

	next, err := Next(task, result.Outcome)
	if err != nil {
		return err
	}
	next.AssigneeID = agent.ID

	saved, err := o.deps.Tasks.Update(ctx, next)
	if err != nil {
		return fmt.Errorf("persist task: %w", err)
	}

	o.postTransition(ctx, saved, agent, task.State, result)
	return nil
}

func (o *Orchestrator) agentFor(ctx context.Context, companyID uuid.UUID, role domain.AgentRole) (domain.Agent, error) {
	agents, err := o.deps.Agents.ListByCompany(ctx, companyID)
	if err != nil {
		return domain.Agent{}, err
	}
	for _, a := range agents {
		if a.Role == role {
			return a, nil
		}
	}
	return domain.Agent{}, fmt.Errorf("company %s has no %s agent", companyID, role)
}

func (o *Orchestrator) postTransition(ctx context.Context, task domain.Task, agent domain.Agent, from domain.TaskState, result StepResult) {
	channels, err := o.deps.Channels.ListByCompany(ctx, task.CompanyID)
	if err != nil || len(channels) == 0 {
		o.deps.Logger.Warn("no channel for status message", "company_id", task.CompanyID, "err", err)
		return
	}

	payload, _ := json.Marshal(map[string]any{
		"task_id":       task.ID.String(),
		"from":          from,
		"to":            task.State,
		"outcome":       result.Outcome,
		"input_tokens":  result.Usage.InputTokens,
		"output_tokens": result.Usage.OutputTokens,
	})

	body := result.Note
	if body == "" {
		body = fmt.Sprintf("%s: %s -> %s", agent.Name, from, task.State)
	}

	msg, err := o.deps.Messages.Create(ctx, domain.Message{
		ChannelID:   channels[0].ID,
		Type:        domain.MessageStatus,
		FromAgentID: agent.ID,
		Body:        body,
		Payload:     payload,
	})
	if err != nil {
		o.deps.Logger.Error("posting status message", "task_id", task.ID, "err", err)
		return
	}
	if err := o.deps.Bus.Publish(ctx, msg); err != nil {
		o.deps.Logger.Error("broadcasting status message", "message_id", msg.ID, "err", err)
	}
}

func (o *Orchestrator) isStopped() bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.stopped
}

func (o *Orchestrator) isPaused(agentID uuid.UUID) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.pausedAgents[agentID]
}
