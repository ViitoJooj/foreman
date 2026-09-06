package http

import (
	"net/http"
	"uuid"

	"github.com/gin-gonic/gin"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/service"
)

// AgentHandler serves the agent endpoints.
type AgentHandler struct {
	agents *service.Agent
}

// NewAgentHandler wires the handler to its use cases.
func NewAgentHandler(agents *service.Agent) *AgentHandler {
	return &AgentHandler{agents: agents}
}

// Create handles POST /api/v1/agents.
func (h *AgentHandler) Create(c *gin.Context) {
	var req createAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
		return
	}

	agent, err := h.agents.Create(c.Request.Context(), service.CreateAgentInput{
		CompanyID: companyID,
		Name:      req.Name,
		Role:      domain.AgentRole(req.Role),
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, newAgentResponse(agent))
}

// Get handles GET /api/v1/agents/:id.
func (h *AgentHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	agent, err := h.agents.Get(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newAgentResponse(agent))
}

// List handles GET /api/v1/agents?company=<id>.
func (h *AgentHandler) List(c *gin.Context) {
	companyID, err := uuid.Parse(c.Query("company"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "a valid company query parameter is required"})
		return
	}

	agents, err := h.agents.List(c.Request.Context(), companyID)
	if err != nil {
		respondError(c, err)
		return
	}

	resp := make([]agentResponse, 0, len(agents))
	for _, agent := range agents {
		resp = append(resp, newAgentResponse(agent))
	}
	c.JSON(http.StatusOK, resp)
}
