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

var _ ports.AgentRepository = (*AgentRepository)(nil)

// AgentRepository is the Supabase-backed implementation of ports.AgentRepository.
type AgentRepository struct {
	pool *pgxpool.Pool
}

// NewAgentRepository builds an AgentRepository bound to the given pool.
func NewAgentRepository(pool *pgxpool.Pool) *AgentRepository {
	return &AgentRepository{pool: pool}
}

const agentColumns = `id, company_id, name, role, status, created_at, updated_at`

func scanAgent(row pgx.Row) (domain.Agent, error) {
	var a domain.Agent
	err := row.Scan(&a.ID, &a.CompanyID, &a.Name, &a.Role, &a.Status, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}

// Create inserts an agent and returns it with the database-assigned id and timestamps.
func (r *AgentRepository) Create(ctx context.Context, a domain.Agent) (domain.Agent, error) {
	const q = `
		INSERT INTO agents (company_id, name, role, status)
		VALUES ($1, $2, $3, $4)
		RETURNING ` + agentColumns
	created, err := scanAgent(r.pool.QueryRow(ctx, q, a.CompanyID, a.Name, a.Role, a.Status))
	if err != nil {
		return domain.Agent{}, fmt.Errorf("create agent: %w", err)
	}
	return created, nil
}

// Get returns the agent with the given id, or ports.ErrAgentNotFound.
func (r *AgentRepository) Get(ctx context.Context, id uuid.UUID) (domain.Agent, error) {
	a, err := scanAgent(r.pool.QueryRow(ctx, `SELECT `+agentColumns+` FROM agents WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Agent{}, ports.ErrAgentNotFound
	}
	if err != nil {
		return domain.Agent{}, fmt.Errorf("get agent: %w", err)
	}
	return a, nil
}

// Update overwrites the mutable fields of an existing agent, including its status.
func (r *AgentRepository) Update(ctx context.Context, a domain.Agent) (domain.Agent, error) {
	const q = `
		UPDATE agents SET company_id = $2, name = $3, role = $4, status = $5
		WHERE id = $1
		RETURNING ` + agentColumns
	updated, err := scanAgent(r.pool.QueryRow(ctx, q, a.ID, a.CompanyID, a.Name, a.Role, a.Status))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Agent{}, ports.ErrAgentNotFound
	}
	if err != nil {
		return domain.Agent{}, fmt.Errorf("update agent: %w", err)
	}
	return updated, nil
}

// ListByCompany returns every agent of a company, oldest first.
func (r *AgentRepository) ListByCompany(ctx context.Context, companyID uuid.UUID) ([]domain.Agent, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+agentColumns+` FROM agents WHERE company_id = $1 ORDER BY created_at`, companyID)
	if err != nil {
		return nil, fmt.Errorf("list agents by company: %w", err)
	}
	defer rows.Close()

	var agents []domain.Agent
	for rows.Next() {
		a, err := scanAgent(rows)
		if err != nil {
			return nil, fmt.Errorf("scan agent: %w", err)
		}
		agents = append(agents, a)
	}
	return agents, rows.Err()
}
