package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"uuid"

	"github.com/gin-gonic/gin"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/service"
)

const defaultTailHistory = 20

// MessageHandler serves the channel message endpoints.
type MessageHandler struct {
	post *service.PostMessage
	tail *service.TailMessages
}

// NewMessageHandler wires the handler to its use cases.
func NewMessageHandler(post *service.PostMessage, tail *service.TailMessages) *MessageHandler {
	return &MessageHandler{post: post, tail: tail}
}

// Create handles POST /api/v1/channels/:id/messages.
func (h *MessageHandler) Create(c *gin.Context) {
	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel id"})
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

// Stream handles GET /api/v1/channels/:id/messages/stream as Server-Sent Events:
// a bounded backlog, then every message posted while the connection is open.
func (h *MessageHandler) Stream(c *gin.Context) {
	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel id"})
		return
	}

	history := defaultTailHistory
	if raw := c.Query("history"); raw != "" {
		if n, convErr := strconv.Atoi(raw); convErr == nil {
			history = n
		}
	}

	backlog, live, err := h.tail.Execute(c.Request.Context(), channelID, history)
	if err != nil {
		respondError(c, err)
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	for _, m := range backlog {
		if !writeSSE(c, m) {
			return
		}
	}

	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case m, ok := <-live:
			if !ok || !writeSSE(c, m) {
				return
			}
		}
	}
}

func writeSSE(c *gin.Context, m domain.Message) bool {
	body, err := json.Marshal(newMessageResponse(m))
	if err != nil {
		return true
	}
	if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", body); err != nil {
		return false
	}
	c.Writer.Flush()
	return true
}
