package service

import (
	"context"
	"errors"
	"uuid"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/ports"
)

// ErrTaskFilterRequired is returned when ListTasks is called without any filter.
var ErrTaskFilterRequired = errors.New("at least one filter is required")

// ListTasks reads tasks, optionally narrowed by company and/or lifecycle state.
type ListTasks struct {
	tasks ports.TaskRepository
}

// NewListTasks wires the use case to a task repository.
func NewListTasks(tasks ports.TaskRepository) *ListTasks {
	return &ListTasks{tasks: tasks}
}

// ListTasksFilter narrows the result. A zero CompanyID or empty State means
// "any"; at least one of them must be set.
type ListTasksFilter struct {
	CompanyID uuid.UUID
	State     domain.TaskState
}

// Execute returns the tasks matching the filter, oldest first.
func (uc *ListTasks) Execute(ctx context.Context, f ListTasksFilter) ([]domain.Task, error) {
	hasCompany := f.CompanyID != (uuid.UUID{})
	hasState := f.State != ""

	switch {
	case hasCompany && hasState:
		byCompany, err := uc.tasks.ListByCompany(ctx, f.CompanyID)
		if err != nil {
			return nil, err
		}
		var filtered []domain.Task
		for _, t := range byCompany {
			if t.State == f.State {
				filtered = append(filtered, t)
			}
		}
		return filtered, nil
	case hasCompany:
		return uc.tasks.ListByCompany(ctx, f.CompanyID)
	case hasState:
		return uc.tasks.ListByState(ctx, f.State)
	default:
		return nil, ErrTaskFilterRequired
	}
}
