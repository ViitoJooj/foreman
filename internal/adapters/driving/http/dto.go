package http

import (
	"encoding/json"
	"time"
	"uuid"

	"github.com/ViitoJooj/foreman/internal/core/domain"
)

type createCompanyRequest struct {
	Name      string `json:"name" binding:"required"`
	Slug      string `json:"slug" binding:"required"`
	RepoOwner string `json:"repo_owner" binding:"required"`
	RepoName  string `json:"repo_name" binding:"required"`
}

type companyResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	RepoOwner string    `json:"repo_owner"`
	RepoName  string    `json:"repo_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func newCompanyResponse(c domain.Company) companyResponse {
	return companyResponse{
		ID:        c.ID.String(),
		Name:      c.Name,
		Slug:      c.Slug,
		RepoOwner: c.RepoOwner,
		RepoName:  c.RepoName,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

type createAgentRequest struct {
	CompanyID string `json:"company_id" binding:"required,uuid"`
	Name      string `json:"name" binding:"required"`
	Role      string `json:"role" binding:"required,oneof=task_creator coder tester pr_reviewer"`
}

type agentResponse struct {
	ID        string    `json:"id"`
	CompanyID string    `json:"company_id"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func newAgentResponse(a domain.Agent) agentResponse {
	return agentResponse{
		ID:        a.ID.String(),
		CompanyID: a.CompanyID.String(),
		Name:      a.Name,
		Role:      string(a.Role),
		Status:    string(a.Status),
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

type createChannelRequest struct {
	CompanyID string `json:"company_id" binding:"required,uuid"`
	Name      string `json:"name" binding:"required"`
}

type channelResponse struct {
	ID        string    `json:"id"`
	CompanyID string    `json:"company_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func newChannelResponse(ch domain.Channel) channelResponse {
	return channelResponse{
		ID:        ch.ID.String(),
		CompanyID: ch.CompanyID.String(),
		Name:      ch.Name,
		CreatedAt: ch.CreatedAt,
	}
}

type issueCommandRequest struct {
	Kind          string `json:"kind" binding:"required,oneof=kill panic pause resume"`
	CompanyID     string `json:"company_id" binding:"omitempty,uuid"`
	TargetAgentID string `json:"target_agent_id" binding:"omitempty,uuid"`
	Reason        string `json:"reason"`
}

type commandResponse struct {
	ID            string     `json:"id"`
	Kind          string     `json:"kind"`
	Status        string     `json:"status"`
	CompanyID     string     `json:"company_id,omitempty"`
	TargetAgentID string     `json:"target_agent_id,omitempty"`
	Reason        string     `json:"reason,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	AppliedAt     *time.Time `json:"applied_at,omitempty"`
}

func newCommandResponse(c domain.Command) commandResponse {
	return commandResponse{
		ID:            c.ID.String(),
		Kind:          string(c.Kind),
		Status:        string(c.Status),
		CompanyID:     optionalUUID(c.CompanyID),
		TargetAgentID: optionalUUID(c.TargetAgentID),
		Reason:        c.Reason,
		CreatedAt:     c.CreatedAt,
		AppliedAt:     c.AppliedAt,
	}
}

type createTaskRequest struct {
	CompanyID   string `json:"company_id" binding:"required,uuid"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Risk        string `json:"risk" binding:"required,oneof=low medium high"`
	MaxRetries  int    `json:"max_retries"`
}

type postMessageRequest struct {
	Type             string          `json:"type" binding:"required,oneof=request response status"`
	FromAgentID      string          `json:"from_agent_id" binding:"required,uuid"`
	ToAgentID        string          `json:"to_agent_id" binding:"omitempty,uuid"`
	InReplyTo        string          `json:"in_reply_to" binding:"omitempty,uuid"`
	RequiresResponse bool            `json:"requires_response"`
	Body             string          `json:"body"`
	Payload          json.RawMessage `json:"payload"`
}

type taskResponse struct {
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
	AssigneeID  string    `json:"assignee_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func newTaskResponse(t domain.Task) taskResponse {
	return taskResponse{
		ID:          t.ID.String(),
		CompanyID:   t.CompanyID.String(),
		Title:       t.Title,
		Description: t.Description,
		State:       string(t.State),
		Risk:        string(t.Risk),
		Branch:      t.Branch,
		PRNumber:    t.PRNumber,
		Retries:     t.Retries,
		MaxRetries:  t.MaxRetries,
		AssigneeID:  optionalUUID(t.AssigneeID),
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

type messageResponse struct {
	ID               string          `json:"id"`
	ChannelID        string          `json:"channel_id"`
	Type             string          `json:"type"`
	FromAgentID      string          `json:"from_agent_id"`
	ToAgentID        string          `json:"to_agent_id,omitempty"`
	InReplyTo        string          `json:"in_reply_to,omitempty"`
	RequiresResponse bool            `json:"requires_response"`
	Body             string          `json:"body"`
	Payload          json.RawMessage `json:"payload,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}

func newMessageResponse(m domain.Message) messageResponse {
	return messageResponse{
		ID:               m.ID.String(),
		ChannelID:        m.ChannelID.String(),
		Type:             string(m.Type),
		FromAgentID:      m.FromAgentID.String(),
		ToAgentID:        optionalUUID(m.ToAgentID),
		InReplyTo:        optionalUUID(m.InReplyTo),
		RequiresResponse: m.RequiresResponse,
		Body:             m.Body,
		Payload:          m.Payload,
		CreatedAt:        m.CreatedAt,
	}
}

// optionalUUID renders the zero UUID as an empty string so absent foreign keys
// are omitted from responses.
func optionalUUID(id uuid.UUID) string {
	if id == (uuid.UUID{}) {
		return ""
	}
	return id.String()
}

// parseOptionalUUID parses s, treating the empty string as the zero UUID.
func parseOptionalUUID(s string) (uuid.UUID, error) {
	if s == "" {
		return uuid.UUID{}, nil
	}
	return uuid.Parse(s)
}
