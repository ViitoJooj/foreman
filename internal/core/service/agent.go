package service

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/ports"
)

// ErrInvalidAgentInput is returned when Agent.Create is given missing or
// malformed fields.
var ErrInvalidAgentInput = errors.New("invalid agent input")

// Agent groups the agent use cases.
type Agent struct {
	repo ports.AgentRepository
}

// NewAgent wires the use cases to an agent repository.
func NewAgent(repo ports.AgentRepository) *Agent {
	return &Agent{repo: repo}
}

// CreateAgentInput carries the fields required to add an agent to a company.
type CreateAgentInput struct {
	CompanyID uuid.UUID
	Name      string
	Role      domain.AgentRole
}

// Create validates the input and persists a new agent in the idle state.
func (uc *Agent) Create(ctx context.Context, in CreateAgentInput) (domain.Agent, error) {
	if in.CompanyID == (uuid.UUID{}) {
		return domain.Agent{}, fmt.Errorf("%w: company id is required", ErrInvalidAgentInput)
	}
	if in.Name == "" {
		return domain.Agent{}, fmt.Errorf("%w: name is required", ErrInvalidAgentInput)
	}
	if !validAgentRole(in.Role) {
		return domain.Agent{}, fmt.Errorf("%w: unknown role %q", ErrInvalidAgentInput, in.Role)
	}
	return uc.repo.Create(ctx, domain.Agent{
		CompanyID: in.CompanyID,
		Name:      in.Name,
		Role:      in.Role,
		Status:    domain.AgentIdle,
	})
}

// Get returns the agent with the given id.
func (uc *Agent) Get(ctx context.Context, id uuid.UUID) (domain.Agent, error) {
	return uc.repo.Get(ctx, id)
}

// List returns every agent of a company, oldest first.
func (uc *Agent) List(ctx context.Context, companyID uuid.UUID) ([]domain.Agent, error) {
	return uc.repo.ListByCompany(ctx, companyID)
}

func validAgentRole(r domain.AgentRole) bool {
	switch r {
	case domain.RoleTaskCreator, domain.RoleCoder, domain.RoleTester, domain.RolePRReviewer:
		return true
	default:
		return false
	}
}
