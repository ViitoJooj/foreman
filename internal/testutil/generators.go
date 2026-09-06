// Package testutil provides random-but-valid input generators and generated
// port mocks shared across the test suite. It is importable by any package's
// tests, so it is deliberately not a _test.go file.
package testutil

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"uuid"

	"github.com/ViitoJooj/foreman/internal/core/domain"
)

func token() string {
	return fmt.Sprintf("%08x", rand.Uint32())
}

func pick[T any](options ...T) T {
	return options[rand.IntN(len(options))]
}

// RandomUUID returns a fresh random UUID.
func RandomUUID() uuid.UUID {
	return uuid.NewV4()
}

// RandomEmail returns a syntactically valid, unique email address.
func RandomEmail() string {
	return fmt.Sprintf("user-%s@example.com", token())
}

// RandomCompanyName returns a plausible company display name.
func RandomCompanyName() string {
	return "Company " + token()
}

// RandomSlug returns a lowercase, URL-safe slug.
func RandomSlug() string {
	return "slug-" + token()
}

// RandomRepoOwner returns a GitHub-style owner login.
func RandomRepoOwner() string {
	return "owner-" + token()
}

// RandomRepoName returns a GitHub-style repository name.
func RandomRepoName() string {
	return "repo-" + token()
}

// RandomAgentName returns a unique agent persona name.
func RandomAgentName() string {
	return "Agent " + token()
}

// RandomChannelName returns a unique channel name.
func RandomChannelName() string {
	return "channel-" + token()
}

// RandomTaskTitle returns a non-empty task title.
func RandomTaskTitle() string {
	return "Task " + token()
}

// RandomText returns a short free-text blob.
func RandomText() string {
	return strings.Join([]string{token(), token(), token()}, " ")
}

// RandomBranchSlug returns an autodev-style branch name.
func RandomBranchSlug() string {
	return fmt.Sprintf("autodev/%s/%s", RandomSlug(), RandomSlug())
}

// RandomJSONPayload returns a valid, non-trivial JSON object.
func RandomJSONPayload() json.RawMessage {
	b, _ := json.Marshal(map[string]string{"key-" + token(): "value-" + token()})
	return b
}

// RandomRiskLevel returns one of the valid risk levels.
func RandomRiskLevel() domain.RiskLevel {
	return pick(domain.RiskLow, domain.RiskMedium, domain.RiskHigh)
}

// RandomAgentRole returns one of the valid agent roles.
func RandomAgentRole() domain.AgentRole {
	return pick(domain.RoleTaskCreator, domain.RoleCoder, domain.RoleTester, domain.RolePRReviewer)
}

// RandomAgentStatus returns one of the valid agent statuses.
func RandomAgentStatus() domain.AgentStatus {
	return pick(domain.AgentIdle, domain.AgentWorking, domain.AgentPaused)
}

// RandomMessageType returns one of the valid message types.
func RandomMessageType() domain.MessageType {
	return pick(domain.MessageRequest, domain.MessageResponse, domain.MessageStatus)
}

// RandomTaskState returns one of the valid task states.
func RandomTaskState() domain.TaskState {
	return pick(
		domain.TaskCreated, domain.TaskResearching, domain.TaskQueued, domain.TaskCoding,
		domain.TaskTesting, domain.TaskReviewing, domain.TaskMerged, domain.TaskNeedsHuman,
		domain.TaskRejected, domain.TaskFailedBuild, domain.TaskFailedTest,
	)
}

// RandomCommandKind returns one of the valid command kinds.
func RandomCommandKind() domain.CommandKind {
	return pick(domain.CommandKill, domain.CommandPanic, domain.CommandPause, domain.CommandResume)
}

// RandomPort returns a valid TCP port string in the ephemeral range.
func RandomPort() string {
	return fmt.Sprintf("%d", 20000+rand.IntN(20000))
}

// RandomAPIKey returns a unique API key.
func RandomAPIKey() string {
	return "key-" + token() + token()
}

// RandomDatabaseURL returns a syntactically valid Postgres DSN.
func RandomDatabaseURL() string {
	return fmt.Sprintf(
		"postgres://user-%s:pass-%s@db.example.com:5432/app-%s?sslmode=require",
		token(), token(), token(),
	)
}
