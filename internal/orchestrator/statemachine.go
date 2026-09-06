// Package orchestrator drives tasks through the pipeline state machine, applies
// operator commands and guards the token budget. It depends only on the core
// ports and domain plus the standard library; no framework lives here.
package orchestrator

import (
	"fmt"

	"github.com/ViitoJooj/foreman/internal/core/domain"
)

// Outcome is the result a runner reports after acting on a task.
type Outcome string

const (
	// OutcomeAdvance means the step succeeded and the task moves forward.
	OutcomeAdvance Outcome = "advance"
	// OutcomeBuildFailed is reported from the coding stage.
	OutcomeBuildFailed Outcome = "build_failed"
	// OutcomeTestFailed is reported from the testing stage.
	OutcomeTestFailed Outcome = "test_failed"
	// OutcomeRejected is reported from the reviewing stage.
	OutcomeRejected Outcome = "rejected"
	// OutcomeNeedsHuman escalates the task from any stage.
	OutcomeNeedsHuman Outcome = "needs_human"
)

// actionableStates are the states a runner can be dispatched for.
var actionableStates = []domain.TaskState{
	domain.TaskQueued,
	domain.TaskCoding,
	domain.TaskTesting,
	domain.TaskReviewing,
	domain.TaskFailedBuild,
	domain.TaskFailedTest,
}

// ActionableStates returns the task states the scheduler picks up each cycle.
func ActionableStates() []domain.TaskState {
	return append([]domain.TaskState(nil), actionableStates...)
}

// AllowedOutcomes lists the outcomes a runner may legally report for a task in
// the given state. It is used to constrain an LLM runner's response.
func AllowedOutcomes(state domain.TaskState) []Outcome {
	switch state {
	case domain.TaskQueued:
		return []Outcome{OutcomeAdvance, OutcomeNeedsHuman}
	case domain.TaskCoding:
		return []Outcome{OutcomeAdvance, OutcomeBuildFailed, OutcomeNeedsHuman}
	case domain.TaskTesting:
		return []Outcome{OutcomeAdvance, OutcomeTestFailed, OutcomeNeedsHuman}
	case domain.TaskReviewing:
		return []Outcome{OutcomeAdvance, OutcomeRejected, OutcomeNeedsHuman}
	case domain.TaskFailedBuild, domain.TaskFailedTest:
		return []Outcome{OutcomeAdvance, OutcomeNeedsHuman}
	default:
		return nil
	}
}

// Next returns the task after a runner step: the new state, and an incremented
// retry count when a failed build/test is retried. It is a pure function.
func Next(task domain.Task, outcome Outcome) (domain.Task, error) {
	next := task

	if outcome == OutcomeNeedsHuman {
		next.State = domain.TaskNeedsHuman
		return next, nil
	}

	switch task.State {
	case domain.TaskQueued:
		if outcome != OutcomeAdvance {
			return task, unexpectedOutcome(task.State, outcome)
		}
		next.State = domain.TaskCoding

	case domain.TaskCoding:
		switch outcome {
		case OutcomeAdvance:
			next.State = domain.TaskTesting
		case OutcomeBuildFailed:
			next.State = domain.TaskFailedBuild
		default:
			return task, unexpectedOutcome(task.State, outcome)
		}

	case domain.TaskTesting:
		switch outcome {
		case OutcomeAdvance:
			next.State = domain.TaskReviewing
		case OutcomeTestFailed:
			next.State = domain.TaskFailedTest
		default:
			return task, unexpectedOutcome(task.State, outcome)
		}

	case domain.TaskReviewing:
		switch outcome {
		case OutcomeAdvance:
			if task.CanAutoMerge() {
				next.State = domain.TaskMerged
			} else {
				next.State = domain.TaskNeedsHuman
			}
		case OutcomeRejected:
			next.State = domain.TaskRejected
		default:
			return task, unexpectedOutcome(task.State, outcome)
		}

	case domain.TaskFailedBuild, domain.TaskFailedTest:
		if outcome != OutcomeAdvance {
			return task, unexpectedOutcome(task.State, outcome)
		}
		next.Retries = task.Retries + 1
		if next.RetriesExhausted() {
			next.State = domain.TaskNeedsHuman
		} else {
			next.State = domain.TaskCoding
		}

	default:
		return task, fmt.Errorf("orchestrator: state %q is not runner-actionable", task.State)
	}

	return next, nil
}

func unexpectedOutcome(state domain.TaskState, outcome Outcome) error {
	return fmt.Errorf("orchestrator: outcome %q is not valid in state %q", outcome, state)
}
