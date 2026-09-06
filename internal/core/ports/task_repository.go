// Package ports declares the interfaces the core depends on. Implementations
// live in internal/adapters; nothing here imports a framework or an SDK.
package ports

import (
	"context"
	"errors"
	"uuid"

	"github.com/ViitoJooj/foreman/internal/core/domain"
)

// ErrTaskNotFound is returned when no task matches the given identifier.
var ErrTaskNotFound = errors.New("task not found")

// TaskRepository persists tasks and their lifecycle state. Update must save the
// full state so a crashed process can resume a task from its last known state.
type TaskRepository interface {
	Create(ctx context.Context, t domain.Task) (domain.Task, error)
	Get(ctx context.Context, id uuid.UUID) (domain.Task, error)
	Update(ctx context.Context, t domain.Task) (domain.Task, error)
	ListByCompany(ctx context.Context, companyID uuid.UUID) ([]domain.Task, error)
	ListByState(ctx context.Context, state domain.TaskState) ([]domain.Task, error)
}
