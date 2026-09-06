// Package domain holds the pure business entities of the harness.
// Types here never perform I/O and never import adapters or external SDKs;
// only the Go standard library is allowed.
package domain

import (
	"time"
	"uuid"
)

// Company is a GitHub repository the harness maintains autonomously, together
// with the team of agents that work on it.
type Company struct {
	ID        uuid.UUID
	Name      string // human-facing display name
	Slug      string // used in automation branch prefixes: autodev/<slug>/<task>
	RepoOwner string
	RepoName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// RepoFullName returns the "owner/name" identifier used by the GitHub API.
func (c Company) RepoFullName() string {
	return c.RepoOwner + "/" + c.RepoName
}
