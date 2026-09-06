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

var _ ports.CommandRepository = (*CommandRepository)(nil)

// CommandRepository is the Supabase-backed implementation of ports.CommandRepository.
type CommandRepository struct {
	pool *pgxpool.Pool
}

// NewCommandRepository builds a CommandRepository bound to the given pool.
func NewCommandRepository(pool *pgxpool.Pool) *CommandRepository {
	return &CommandRepository{pool: pool}
}

const commandColumns = `id, company_id, target_agent_id, kind, status, reason, created_at, applied_at`

func scanCommand(row pgx.Row) (domain.Command, error) {
	var c domain.Command
	err := row.Scan(
		&c.ID, &c.CompanyID, &c.TargetAgentID, &c.Kind, &c.Status, &c.Reason,
		&c.CreatedAt, &c.AppliedAt,
	)
	return c, err
}

// Create inserts a control command and returns it with the database-assigned id
// and timestamp.
func (r *CommandRepository) Create(ctx context.Context, c domain.Command) (domain.Command, error) {
	const q = `
		INSERT INTO commands (company_id, target_agent_id, kind, status, reason)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING ` + commandColumns
	created, err := scanCommand(r.pool.QueryRow(ctx, q,
		nullableUUID(c.CompanyID), nullableUUID(c.TargetAgentID), c.Kind, c.Status, c.Reason,
	))
	if err != nil {
		return domain.Command{}, fmt.Errorf("create command: %w", err)
	}
	return created, nil
}

// Get returns the command with the given id, or ports.ErrCommandNotFound.
func (r *CommandRepository) Get(ctx context.Context, id uuid.UUID) (domain.Command, error) {
	c, err := scanCommand(r.pool.QueryRow(ctx, `SELECT `+commandColumns+` FROM commands WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Command{}, ports.ErrCommandNotFound
	}
	if err != nil {
		return domain.Command{}, fmt.Errorf("get command: %w", err)
	}
	return c, nil
}

// Update overwrites the status, reason and applied timestamp of a command.
func (r *CommandRepository) Update(ctx context.Context, c domain.Command) (domain.Command, error) {
	const q = `
		UPDATE commands SET company_id = $2, target_agent_id = $3, kind = $4, status = $5,
			reason = $6, applied_at = $7
		WHERE id = $1
		RETURNING ` + commandColumns
	updated, err := scanCommand(r.pool.QueryRow(ctx, q,
		c.ID, nullableUUID(c.CompanyID), nullableUUID(c.TargetAgentID), c.Kind, c.Status,
		c.Reason, c.AppliedAt,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Command{}, ports.ErrCommandNotFound
	}
	if err != nil {
		return domain.Command{}, fmt.Errorf("update command: %w", err)
	}
	return updated, nil
}

// ListByStatus returns every command in a given status, oldest first, so the
// harness can replay pending commands after a restart.
func (r *CommandRepository) ListByStatus(ctx context.Context, status domain.CommandStatus) ([]domain.Command, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+commandColumns+` FROM commands WHERE status = $1 ORDER BY created_at`, status)
	if err != nil {
		return nil, fmt.Errorf("list commands by status: %w", err)
	}
	defer rows.Close()

	var commands []domain.Command
	for rows.Next() {
		c, err := scanCommand(rows)
		if err != nil {
			return nil, fmt.Errorf("scan command: %w", err)
		}
		commands = append(commands, c)
	}
	return commands, rows.Err()
}
