package domain

import (
	"time"
	"uuid"
)

// CommandKind is a control action issued by the human operator against the harness.
type CommandKind string

// The control actions the operator can issue.
const (
	CommandKill   CommandKind = "kill"   // graceful stop: agents stop taking new tasks
	CommandPanic  CommandKind = "panic"  // kill, then close agent PRs and delete their branches
	CommandPause  CommandKind = "pause"  // suspend a single agent
	CommandResume CommandKind = "resume" // resume a paused agent
)

// CommandStatus tracks whether a control command has been applied yet.
type CommandStatus string

// The lifecycle of a control command.
const (
	CommandPending CommandStatus = "pending"
	CommandApplied CommandStatus = "applied"
	CommandFailed  CommandStatus = "failed"
)

// Command is an audit record of an operator control action. It is persisted so
// the harness can replay it idempotently after a restart.
type Command struct {
	ID            uuid.UUID
	CompanyID     uuid.UUID // zero for a harness-global command
	TargetAgentID uuid.UUID // set only for pause / resume; zero otherwise
	Kind          CommandKind
	Status        CommandStatus
	Reason        string
	CreatedAt     time.Time
	AppliedAt     *time.Time // nil until applied
}
