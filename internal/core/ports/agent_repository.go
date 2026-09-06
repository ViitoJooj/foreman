package ports

import (
	"context"
	"errors"
	"uuid"

	"github.com/ViitoJooj/foreman/internal/core/domain"
)

// ErrAgentNotFound is returned when no agent matches the given identifier.
var ErrAgentNotFound = errors.New("agent not found")

// AgentRepository persists the team members of each company. Update is used to
// move an agent between idle, working and paused.
type AgentRepository interface {
	Create(ctx context.Context, a domain.Agent) (domain.Agent, error)
	Get(ctx context.Context, id uuid.UUID) (domain.Agent, error)
	Update(ctx context.Context, a domain.Agent) (domain.Agent, error)
	ListByCompany(ctx context.Context, companyID uuid.UUID) ([]domain.Agent, error)
}
