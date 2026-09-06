// Package cli implements the `foreman` command tree. Every command is a thin
// HTTP client of the harness API, the way `gh` talks to the GitHub API; no
// domain logic lives here.
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const requestTimeout = 30 * time.Second

// apiClient issues authenticated JSON requests against the harness API.
type apiClient struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func newAPIClient(baseURL, apiKey string) *apiClient {
	return &apiClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		http:    &http.Client{Timeout: requestTimeout},
	}
}

// do sends method/path with an optional JSON body and decodes a 2xx response
// into out. Non-2xx responses are turned into a readable error carrying the
// API's own message.
func (c *apiClient) do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request: %w", err)
		}
		reader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Api-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("calling %s: %w", c.baseURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New(apiErrorMessage(resp.Status, data))
	}
	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}
	return nil
}

func apiErrorMessage(status string, body []byte) string {
	var parsed struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &parsed) == nil && parsed.Error != "" {
		return fmt.Sprintf("api error (%s): %s", status, parsed.Error)
	}
	if msg := strings.TrimSpace(string(body)); msg != "" {
		return fmt.Sprintf("api error (%s): %s", status, msg)
	}
	return fmt.Sprintf("api error (%s)", status)
}

// createCompanyBody is the JSON payload for POST /api/v1/companies.
type createCompanyBody struct {
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	RepoOwner string `json:"repo_owner"`
	RepoName  string `json:"repo_name"`
}

// createAgentBody is the JSON payload for POST /api/v1/agents.
type createAgentBody struct {
	CompanyID string `json:"company_id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
}

// createChannelBody is the JSON payload for POST /api/v1/channels.
type createChannelBody struct {
	CompanyID string `json:"company_id"`
	Name      string `json:"name"`
}

// companyView mirrors the company JSON returned by the API.
type companyView struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	RepoOwner string    `json:"repo_owner"`
	RepoName  string    `json:"repo_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// agentView mirrors the agent JSON returned by the API.
type agentView struct {
	ID        string    `json:"id"`
	CompanyID string    `json:"company_id"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// channelView mirrors the channel JSON returned by the API.
type channelView struct {
	ID        string    `json:"id"`
	CompanyID string    `json:"company_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// issueCommandBody is the JSON payload for POST /api/v1/commands.
type issueCommandBody struct {
	Kind          string `json:"kind"`
	CompanyID     string `json:"company_id,omitempty"`
	TargetAgentID string `json:"target_agent_id,omitempty"`
	Reason        string `json:"reason,omitempty"`
}

// commandView mirrors the command JSON returned by the API.
type commandView struct {
	ID            string     `json:"id"`
	Kind          string     `json:"kind"`
	Status        string     `json:"status"`
	CompanyID     string     `json:"company_id"`
	TargetAgentID string     `json:"target_agent_id"`
	Reason        string     `json:"reason"`
	CreatedAt     time.Time  `json:"created_at"`
	AppliedAt     *time.Time `json:"applied_at"`
}

// createTaskBody is the JSON payload for POST /api/v1/tasks.
type createTaskBody struct {
	CompanyID   string `json:"company_id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Risk        string `json:"risk"`
	MaxRetries  int    `json:"max_retries,omitempty"`
}

// postMessageBody is the JSON payload for POST /api/v1/channels/:id/messages.
type postMessageBody struct {
	Type             string          `json:"type"`
	FromAgentID      string          `json:"from_agent_id"`
	ToAgentID        string          `json:"to_agent_id,omitempty"`
	InReplyTo        string          `json:"in_reply_to,omitempty"`
	RequiresResponse bool            `json:"requires_response,omitempty"`
	Body             string          `json:"body,omitempty"`
	Payload          json.RawMessage `json:"payload,omitempty"`
}

// taskView mirrors the task JSON returned by the API.
type taskView struct {
	ID          string    `json:"id"`
	CompanyID   string    `json:"company_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	State       string    `json:"state"`
	Risk        string    `json:"risk"`
	Branch      string    `json:"branch"`
	PRNumber    int       `json:"pr_number"`
	Retries     int       `json:"retries"`
	MaxRetries  int       `json:"max_retries"`
	AssigneeID  string    `json:"assignee_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// messageView mirrors the message JSON returned by the API.
type messageView struct {
	ID               string    `json:"id"`
	ChannelID        string    `json:"channel_id"`
	Type             string    `json:"type"`
	FromAgentID      string    `json:"from_agent_id"`
	ToAgentID        string    `json:"to_agent_id"`
	InReplyTo        string    `json:"in_reply_to"`
	RequiresResponse bool      `json:"requires_response"`
	Body             string    `json:"body"`
	CreatedAt        time.Time `json:"created_at"`
}
