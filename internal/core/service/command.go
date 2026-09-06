package service

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/ports"
)

// ErrInvalidCommandInput is returned when Command.Issue is given an inconsistent
// kind/target combination.
var ErrInvalidCommandInput = errors.New("invalid command input")

// Command groups the operator control-action use cases. Issuing a command only
// records it as pending; the orchestrator applies pending commands on its next
// cycle (idempotent replay).
type Command struct {
	repo ports.CommandRepository
}

// NewCommand wires the use cases to a command repository.
func NewCommand(repo ports.CommandRepository) *Command {
	return &Command{repo: repo}
}

// IssueCommandInput carries the fields required to record a control action.
// CompanyID is optional (zero means harness-global). TargetAgentID is required
// for pause/resume and forbidden for kill/panic.
type IssueCommandInput struct {
	Kind          domain.CommandKind
	CompanyID     uuid.UUID
	TargetAgentID uuid.UUID
	Reason        string
}

// Issue validates the input and records a pending command.
func (uc *Command) Issue(ctx context.Context, in IssueCommandInput) (domain.Command, error) {
	hasTarget := in.TargetAgentID != (uuid.UUID{})

	switch in.Kind {
	case domain.CommandKill, domain.CommandPanic:
		if hasTarget {
			return domain.Command{}, fmt.Errorf("%w: %s does not target an agent", ErrInvalidCommandInput, in.Kind)
		}
	case domain.CommandPause, domain.CommandResume:
		if !hasTarget {
			return domain.Command{}, fmt.Errorf("%w: %s requires a target agent", ErrInvalidCommandInput, in.Kind)
		}
	default:
		return domain.Command{}, fmt.Errorf("%w: unknown kind %q", ErrInvalidCommandInput, in.Kind)
	}

	return uc.repo.Create(ctx, domain.Command{
		Kind:          in.Kind,
		CompanyID:     in.CompanyID,
		TargetAgentID: in.TargetAgentID,
		Status:        domain.CommandPending,
		Reason:        in.Reason,
	})
}

// Get returns the command with the given id.
func (uc *Command) Get(ctx context.Context, id uuid.UUID) (domain.Command, error) {
	return uc.repo.Get(ctx, id)
}

// List returns every command in the given status, oldest first.
func (uc *Command) List(ctx context.Context, status domain.CommandStatus) ([]domain.Command, error) {
	return uc.repo.ListByStatus(ctx, status)
}
