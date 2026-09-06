package supabase

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/ports"
)

var _ ports.TaskRepository = (*TaskRepository)(nil)

// TaskRepository is the Supabase-backed implementation of ports.TaskRepository.
type TaskRepository struct {
	pool *pgxpool.Pool
}

// NewTaskRepository builds a TaskRepository bound to the given pool.
func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

const taskColumns = `id, company_id, title, description, state, risk, branch,
	pr_number, retries, max_retries, assignee_id, created_at, updated_at`

func scanTask(row pgx.Row) (domain.Task, error) {
	var t domain.Task
	err := row.Scan(
		&t.ID, &t.CompanyID, &t.Title, &t.Description, &t.State, &t.Risk, &t.Branch,
		&t.PRNumber, &t.Retries, &t.MaxRetries, &t.AssigneeID, &t.CreatedAt, &t.UpdatedAt,
	)
	return t, err
}

// Create inserts a task. State and Risk must already be set; the database assigns
// the id and timestamps and applies defaults for any zero lifecycle field.
func (r *TaskRepository) Create(ctx context.Context, t domain.Task) (domain.Task, error) {
	const q = `
		INSERT INTO tasks (company_id, title, description, state, risk, branch,
			pr_number, retries, max_retries, assignee_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING ` + taskColumns
	created, err := scanTask(r.pool.QueryRow(ctx, q,
		t.CompanyID, t.Title, t.Description, t.State, t.Risk, t.Branch,
		t.PRNumber, t.Retries, t.MaxRetries, nullableUUID(t.AssigneeID),
	))
	if err != nil {
		return domain.Task{}, fmt.Errorf("create task: %w", err)
	}
	return created, nil
}

// Get returns the task with the given id, or ports.ErrTaskNotFound.
func (r *TaskRepository) Get(ctx context.Context, id uuid.UUID) (domain.Task, error) {
	t, err := scanTask(r.pool.QueryRow(ctx, `SELECT `+taskColumns+` FROM tasks WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Task{}, ports.ErrTaskNotFound
	}
	if err != nil {
		return domain.Task{}, fmt.Errorf("get task: %w", err)
	}
	return t, nil
}

// Update persists the full state of a task so a restarted process can resume it.
func (r *TaskRepository) Update(ctx context.Context, t domain.Task) (domain.Task, error) {
	const q = `
		UPDATE tasks SET company_id = $2, title = $3, description = $4, state = $5, risk = $6,
			branch = $7, pr_number = $8, retries = $9, max_retries = $10, assignee_id = $11
		WHERE id = $1
		RETURNING ` + taskColumns
	updated, err := scanTask(r.pool.QueryRow(ctx, q,
		t.ID, t.CompanyID, t.Title, t.Description, t.State, t.Risk,
		t.Branch, t.PRNumber, t.Retries, t.MaxRetries, nullableUUID(t.AssigneeID),
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Task{}, ports.ErrTaskNotFound
	}
	if err != nil {
		return domain.Task{}, fmt.Errorf("update task: %w", err)
	}
	return updated, nil
}

// ListByCompany returns every task of a company, oldest first.
func (r *TaskRepository) ListByCompany(ctx context.Context, companyID uuid.UUID) ([]domain.Task, error) {
	return r.query(ctx, `SELECT `+taskColumns+` FROM tasks WHERE company_id = $1 ORDER BY created_at`, companyID)
}

// ListByState returns every task in a given lifecycle state, oldest first.
func (r *TaskRepository) ListByState(ctx context.Context, state domain.TaskState) ([]domain.Task, error) {
	return r.query(ctx, `SELECT `+taskColumns+` FROM tasks WHERE state = $1 ORDER BY created_at`, state)
}

func (r *TaskRepository) query(ctx context.Context, sql string, args ...any) ([]domain.Task, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}
