package http

import (
	"net/http"
	"uuid"

	"github.com/gin-gonic/gin"

	"github.com/ViitoJooj/foreman/internal/core/service"
)

// ChannelHandler serves the channel endpoints.
type ChannelHandler struct {
	channels *service.Channel
}

// NewChannelHandler wires the handler to its use cases.
func NewChannelHandler(channels *service.Channel) *ChannelHandler {
	return &ChannelHandler{channels: channels}
}

// Create handles POST /api/v1/channels.
func (h *ChannelHandler) Create(c *gin.Context) {
	var req createChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
		return
	}

	channel, err := h.channels.Create(c.Request.Context(), service.CreateChannelInput{
		CompanyID: companyID,
		Name:      req.Name,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, newChannelResponse(channel))
}

// Get handles GET /api/v1/channels/:id.
func (h *ChannelHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	channel, err := h.channels.Get(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newChannelResponse(channel))
}

// List handles GET /api/v1/channels?company=<id>.
func (h *ChannelHandler) List(c *gin.Context) {
	companyID, err := uuid.Parse(c.Query("company"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "a valid company query parameter is required"})
		return
	}

	channels, err := h.channels.List(c.Request.Context(), companyID)
	if err != nil {
		respondError(c, err)
		return
	}

	resp := make([]channelResponse, 0, len(channels))
	for _, channel := range channels {
		resp = append(resp, newChannelResponse(channel))
	}
	c.JSON(http.StatusOK, resp)
}
