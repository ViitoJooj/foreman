package domain

import (
	"time"
	"uuid"
)

// AgentRole identifies which of the four pipeline roles an agent fills.
type AgentRole string

const (
	RoleTaskCreator AgentRole = "task_creator"
	RoleCoder       AgentRole = "coder"
	RoleTester      AgentRole = "tester"
	RolePRReviewer  AgentRole = "pr_reviewer"
)

// AgentStatus is the current working state of an agent, as shown on the TUI.
type AgentStatus string

const (
	AgentIdle    AgentStatus = "idle"
	AgentWorking AgentStatus = "working"
	AgentPaused  AgentStatus = "paused"
)

// Agent is one member of a company's team: a named persona bound to a single role.
type Agent struct {
	ID        uuid.UUID
	CompanyID uuid.UUID
	Name      string // generated persona name, unique within the company
	Role      AgentRole
	Status    AgentStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}
