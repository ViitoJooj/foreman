package orchestrator

import (
	"context"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/ports"
)

// StepResult is what a runner reports after acting on a task.
type StepResult struct {
	Outcome Outcome        // where the state machine should move
	Note    string         // channel message body; empty lets the orchestrator use a default
	Usage   ports.LLMUsage // tokens consumed; zero for non-LLM runners
}

// AgentRunner performs one pipeline step for a task and reports the result.
// Real runners drive an LLM, GitHub and the sandbox; the stub just advances.
type AgentRunner interface {
	Role() domain.AgentRole
	Step(ctx context.Context, task domain.Task) (StepResult, error)
}

// StubRunner advances every task it is handed. It is the placeholder until the
// real role runners land.
type StubRunner struct {
	role domain.AgentRole
}

// NewStubRunner returns a StubRunner for the given role.
func NewStubRunner(role domain.AgentRole) *StubRunner {
	return &StubRunner{role: role}
}

// Role reports which pipeline role this runner fills.
func (r *StubRunner) Role() domain.AgentRole { return r.role }

// Step always reports OutcomeAdvance with no note or token usage.
func (r *StubRunner) Step(context.Context, domain.Task) (StepResult, error) {
	return StepResult{Outcome: OutcomeAdvance}, nil
}

// RoleForState maps an actionable task state to the role responsible for it.
func RoleForState(s domain.TaskState) (domain.AgentRole, bool) {
	switch s {
	case domain.TaskQueued, domain.TaskCoding, domain.TaskFailedBuild, domain.TaskFailedTest:
		return domain.RoleCoder, true
	case domain.TaskTesting:
		return domain.RoleTester, true
	case domain.TaskReviewing:
		return domain.RolePRReviewer, true
	default:
		return "", false
	}
}

// StubRunners returns a stub runner for every pipeline role.
func StubRunners() map[domain.AgentRole]AgentRunner {
	return map[domain.AgentRole]AgentRunner{
		domain.RoleCoder:      NewStubRunner(domain.RoleCoder),
		domain.RoleTester:     NewStubRunner(domain.RoleTester),
		domain.RolePRReviewer: NewStubRunner(domain.RolePRReviewer),
	}
}
