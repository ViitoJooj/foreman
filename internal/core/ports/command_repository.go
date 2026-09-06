package ports

import (
	"context"
	"errors"
	"uuid"

	"github.com/ViitoJooj/foreman/internal/core/domain"
)

// ErrCommandNotFound is returned when no command matches the given identifier.
var ErrCommandNotFound = errors.New("command not found")

// CommandRepository persists operator control actions. ListByStatus lets the
// harness pick up pending commands after a restart and apply them idempotently.
type CommandRepository interface {
	Create(ctx context.Context, c domain.Command) (domain.Command, error)
	Get(ctx context.Context, id uuid.UUID) (domain.Command, error)
	Update(ctx context.Context, c domain.Command) (domain.Command, error)
	ListByStatus(ctx context.Context, status domain.CommandStatus) ([]domain.Command, error)
}
