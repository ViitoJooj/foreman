package http

import (
	"net/http"
	"uuid"

	"github.com/gin-gonic/gin"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/service"
)

// CommandHandler serves the operator control-action endpoints.
type CommandHandler struct {
	commands *service.Command
}

// NewCommandHandler wires the handler to its use cases.
func NewCommandHandler(commands *service.Command) *CommandHandler {
	return &CommandHandler{commands: commands}
}

// Create handles POST /api/v1/commands. It records the control action as
// pending; the orchestrator applies it on its next cycle.
func (h *CommandHandler) Create(c *gin.Context) {
	var req issueCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID, err := parseOptionalUUID(req.CompanyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
		return
	}
	targetAgentID, err := parseOptionalUUID(req.TargetAgentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target_agent_id"})
		return
	}

	command, err := h.commands.Issue(c.Request.Context(), service.IssueCommandInput{
		Kind:          domain.CommandKind(req.Kind),
		CompanyID:     companyID,
		TargetAgentID: targetAgentID,
		Reason:        req.Reason,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, newCommandResponse(command))
}

// Get handles GET /api/v1/commands/:id.
func (h *CommandHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	command, err := h.commands.Get(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newCommandResponse(command))
}

// List handles GET /api/v1/commands?status=<status>, defaulting to pending.
func (h *CommandHandler) List(c *gin.Context) {
	status := domain.CommandStatus(c.DefaultQuery("status", string(domain.CommandPending)))

	commands, err := h.commands.List(c.Request.Context(), status)
	if err != nil {
		respondError(c, err)
		return
	}

	resp := make([]commandResponse, 0, len(commands))
	for _, command := range commands {
		resp = append(resp, newCommandResponse(command))
	}
	c.JSON(http.StatusOK, resp)
}
