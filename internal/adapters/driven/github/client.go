// Package github implements ports.GitHubClient as a thin wrapper over the `gh`
// CLI, invoked via os/exec. The command runner is injectable so the wrapper is
// testable without a real `gh`.
package github

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/ViitoJooj/foreman/internal/core/ports"
)

var _ ports.GitHubClient = (*Client)(nil)

const (
	defaultBinary        = "gh"
	defaultMergeStrategy = "squash"
)

// CommandRunner runs an external command and returns its stdout.
type CommandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

// Client wraps the `gh` CLI.
type Client struct {
	run           CommandRunner
	bin           string
	mergeStrategy string
}

// Option customises a Client.
type Option func(*Client)

// WithRunner overrides the command runner (used in tests).
func WithRunner(r CommandRunner) Option { return func(c *Client) { c.run = r } }

// WithBinary overrides the `gh` executable path/name.
func WithBinary(path string) Option { return func(c *Client) { c.bin = path } }

// WithMergeStrategy sets the merge strategy: "squash", "merge" or "rebase".
func WithMergeStrategy(s string) Option { return func(c *Client) { c.mergeStrategy = s } }

// New builds a Client. By default it shells out to `gh` and merges with --squash.
func New(opts ...Option) *Client {
	c := &Client{run: execRunner, bin: defaultBinary, mergeStrategy: defaultMergeStrategy}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func execRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	// The whole point of this adapter is to run the `gh` binary with structured
	// arguments. exec.Command does not invoke a shell, so the args are not
	// interpreted; nothing here is attacker-influenced beyond the caller's repo.
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // G204: gh CLI wrapper, no shell
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

// CreateBranch creates branch off base on the remote, via the git refs API.
func (c *Client) CreateBranch(ctx context.Context, repo, base, branch string) error {
	sha, err := c.run(ctx, c.bin, "api", "repos/"+repo+"/git/ref/heads/"+base, "--jq", ".object.sha")
	if err != nil {
		return fmt.Errorf("github: resolve base %q: %w", base, err)
	}
	_, err = c.run(ctx, c.bin, "api", "-X", "POST", "repos/"+repo+"/git/refs",
		"-f", "ref=refs/heads/"+branch,
		"-f", "sha="+strings.TrimSpace(string(sha)))
	if err != nil {
		return fmt.Errorf("github: create branch %q: %w", branch, err)
	}
	return nil
}

// OpenPR opens a pull request and returns its number.
func (c *Client) OpenPR(ctx context.Context, repo string, in ports.PullRequestInput) (int, error) {
	out, err := c.run(ctx, c.bin, "pr", "create",
		"--repo", repo, "--base", in.Base, "--head", in.Head,
		"--title", in.Title, "--body", in.Body)
	if err != nil {
		return 0, fmt.Errorf("github: open pr: %w", err)
	}

	if n, perr := prNumberFromURL(string(out)); perr == nil {
		return n, nil
	}

	view, err := c.run(ctx, c.bin, "pr", "view", in.Head, "--repo", repo, "--json", "number", "--jq", ".number")
	if err != nil {
		return 0, fmt.Errorf("github: resolve pr number: %w", err)
	}
	return atoiTrim(view)
}

// UpdatePR edits the title and body of an existing pull request.
func (c *Client) UpdatePR(ctx context.Context, repo string, number int, in ports.PullRequestInput) error {
	_, err := c.run(ctx, c.bin, "pr", "edit", strconv.Itoa(number),
		"--repo", repo, "--title", in.Title, "--body", in.Body)
	if err != nil {
		return fmt.Errorf("github: update pr #%d: %w", number, err)
	}
	return nil
}

// MergePR merges a pull request using the configured strategy and deletes the branch.
func (c *Client) MergePR(ctx context.Context, repo string, number int) error {
	_, err := c.run(ctx, c.bin, "pr", "merge", strconv.Itoa(number),
		"--repo", repo, "--"+c.mergeStrategy, "--delete-branch")
	if err != nil {
		return fmt.Errorf("github: merge pr #%d: %w", number, err)
	}
	return nil
}

// ClosePR closes a pull request without merging.
func (c *Client) ClosePR(ctx context.Context, repo string, number int) error {
	_, err := c.run(ctx, c.bin, "pr", "close", strconv.Itoa(number), "--repo", repo)
	if err != nil {
		return fmt.Errorf("github: close pr #%d: %w", number, err)
	}
	return nil
}

// DeleteBranch removes a branch from the remote.
func (c *Client) DeleteBranch(ctx context.Context, repo, branch string) error {
	_, err := c.run(ctx, c.bin, "api", "-X", "DELETE", "repos/"+repo+"/git/refs/heads/"+branch)
	if err != nil {
		return fmt.Errorf("github: delete branch %q: %w", branch, err)
	}
	return nil
}

func prNumberFromURL(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty output")
	}
	fields := strings.Fields(s)
	last := fields[len(fields)-1]
	i := strings.LastIndex(last, "/")
	if i < 0 {
		return 0, fmt.Errorf("no pr number in %q", last)
	}
	return strconv.Atoi(last[i+1:])
}

func atoiTrim(b []byte) (int, error) {
	return strconv.Atoi(strings.TrimSpace(string(b)))
}
