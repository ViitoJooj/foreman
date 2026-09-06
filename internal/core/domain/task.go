package domain

import (
	"time"
	"uuid"
)

// TaskState is a node in the task lifecycle state machine:
//
//	created → researching → queued → coding → testing → reviewing → { merged | needs_human | rejected }
//	                                    │         │
//	                                    └── failed_build / failed_test (retry up to MaxRetries) ──┘
type TaskState string

// The nodes of the task lifecycle.
const (
	TaskCreated     TaskState = "created"
	TaskResearching TaskState = "researching"
	TaskQueued      TaskState = "queued"
	TaskCoding      TaskState = "coding"
	TaskTesting     TaskState = "testing"
	TaskReviewing   TaskState = "reviewing"
	TaskMerged      TaskState = "merged"
	TaskNeedsHuman  TaskState = "needs_human"
	TaskRejected    TaskState = "rejected"
	TaskFailedBuild TaskState = "failed_build"
	TaskFailedTest  TaskState = "failed_test"
)

// RiskLevel drives the auto-merge decision: only low-risk tasks may auto-merge.
type RiskLevel string

// The task risk levels.
const (
	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"
)

// Task is a single unit of work moving through the pipeline.
type Task struct {
	ID          uuid.UUID
	CompanyID   uuid.UUID
	Title       string
	Description string
	State       TaskState
	Risk        RiskLevel
	Branch      string    // autodev/<company-slug>/<task-slug>; empty until coding starts
	PRNumber    int       // 0 until a PR is opened
	Retries     int       // failed build/test attempts so far
	MaxRetries  int       // cap after which the task goes to needs_human
	AssigneeID  uuid.UUID // agent currently owning the task; zero when unassigned
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CanAutoMerge reports whether the task's risk allows merging without a human.
func (t Task) CanAutoMerge() bool {
	return t.Risk == RiskLow
}

// RetriesExhausted reports whether the retry budget for this task is spent.
func (t Task) RetriesExhausted() bool {
	return t.Retries >= t.MaxRetries
}
