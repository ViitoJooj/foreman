package http

import (
	"net/http"
	"uuid"

	"github.com/gin-gonic/gin"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/service"
)

// MessageHandler serves the channel message endpoints.
type MessageHandler struct {
	post *service.PostMessage
}

// NewMessageHandler wires the handler to its use case.
func NewMessageHandler(post *service.PostMessage) *MessageHandler {
	return &MessageHandler{post: post}
}

// Create handles POST /api/v1/channels/:channelID/messages.
func (h *MessageHandler) Create(c *gin.Context) {
	channelID, err := uuid.Parse(c.Param("channelID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channelID"})
		return
	}

	var req postMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fromAgentID, err := uuid.Parse(req.FromAgentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from_agent_id"})
		return
	}
	toAgentID, err := parseOptionalUUID(req.ToAgentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to_agent_id"})
		return
	}
	inReplyTo, err := parseOptionalUUID(req.InReplyTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid in_reply_to"})
		return
	}

	msg, err := h.post.Execute(c.Request.Context(), service.PostMessageInput{
		ChannelID:        channelID,
		Type:             domain.MessageType(req.Type),
		FromAgentID:      fromAgentID,
		ToAgentID:        toAgentID,
		InReplyTo:        inReplyTo,
		RequiresResponse: req.RequiresResponse,
		Body:             req.Body,
		Payload:          req.Payload,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, newMessageResponse(msg))
}
