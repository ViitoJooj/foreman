// Package service holds the application use cases. Each use case receives its
// dependencies as ports through its constructor (manual injection, no framework).
package service

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/ports"
)

const defaultMaxRetries = 3

// ErrInvalidTaskInput is returned when CreateTask is called with missing or
// malformed fields.
var ErrInvalidTaskInput = errors.New("invalid task input")

// CreateTask adds a new task to a company's queue in the "created" state.
type CreateTask struct {
	tasks ports.TaskRepository
}

// NewCreateTask wires the use case to a task repository.
func NewCreateTask(tasks ports.TaskRepository) *CreateTask {
	return &CreateTask{tasks: tasks}
}

// CreateTaskInput carries the fields a caller must provide to open a task.
type CreateTaskInput struct {
	CompanyID   uuid.UUID
	Title       string
	Description string
	Risk        domain.RiskLevel
	MaxRetries  int // <= 0 falls back to the default
}

// Execute validates the input and persists the new task.
func (uc *CreateTask) Execute(ctx context.Context, in CreateTaskInput) (domain.Task, error) {
	if in.CompanyID == (uuid.UUID{}) {
		return domain.Task{}, fmt.Errorf("%w: company id is required", ErrInvalidTaskInput)
	}
	if in.Title == "" {
		return domain.Task{}, fmt.Errorf("%w: title is required", ErrInvalidTaskInput)
	}
	if !validRisk(in.Risk) {
		return domain.Task{}, fmt.Errorf("%w: unknown risk %q", ErrInvalidTaskInput, in.Risk)
	}

	maxRetries := in.MaxRetries
	if maxRetries <= 0 {
		maxRetries = defaultMaxRetries
	}

	// A task added through the API/CLI is a human-authored work item that is
	// already specified, so it enters the pipeline queue directly. The
	// created -> researching -> queued states are reserved for the autonomous
	// Task Creator, which is not wired yet.
	return uc.tasks.Create(ctx, domain.Task{
		CompanyID:   in.CompanyID,
		Title:       in.Title,
		Description: in.Description,
		State:       domain.TaskQueued,
		Risk:        in.Risk,
		MaxRetries:  maxRetries,
	})
}

func validRisk(r domain.RiskLevel) bool {
	switch r {
	case domain.RiskLow, domain.RiskMedium, domain.RiskHigh:
		return true
	default:
		return false
	}
}
