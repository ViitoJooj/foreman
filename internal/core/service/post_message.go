package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"uuid"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/ports"
)

// ErrInvalidMessageInput is returned when PostMessage is called with missing or
// malformed fields.
var ErrInvalidMessageInput = errors.New("invalid message input")

// PostMessage persists a channel message and broadcasts it to live subscribers.
type PostMessage struct {
	messages ports.MessageRepository
	bus      ports.MessageBus
}

// NewPostMessage wires the use case to a message repository and a message bus.
func NewPostMessage(messages ports.MessageRepository, bus ports.MessageBus) *PostMessage {
	return &PostMessage{messages: messages, bus: bus}
}

// PostMessageInput carries the fields a caller must provide to post a message.
type PostMessageInput struct {
	ChannelID        uuid.UUID
	Type             domain.MessageType
	FromAgentID      uuid.UUID
	ToAgentID        uuid.UUID
	InReplyTo        uuid.UUID
	RequiresResponse bool
	Body             string
	Payload          json.RawMessage
}

// Execute validates the input, stores the message and then broadcasts it. The
// broadcast is best-effort: a publish failure is reported but the message is
// already durably persisted.
func (uc *PostMessage) Execute(ctx context.Context, in PostMessageInput) (domain.Message, error) {
	if in.ChannelID == (uuid.UUID{}) {
		return domain.Message{}, fmt.Errorf("%w: channel id is required", ErrInvalidMessageInput)
	}
	if in.FromAgentID == (uuid.UUID{}) {
		return domain.Message{}, fmt.Errorf("%w: from agent id is required", ErrInvalidMessageInput)
	}
	if !validMessageType(in.Type) {
		return domain.Message{}, fmt.Errorf("%w: unknown type %q", ErrInvalidMessageInput, in.Type)
	}
	if len(in.Payload) > 0 && !json.Valid(in.Payload) {
		return domain.Message{}, fmt.Errorf("%w: payload is not valid JSON", ErrInvalidMessageInput)
	}

	created, err := uc.messages.Create(ctx, domain.Message{
		ChannelID:        in.ChannelID,
		Type:             in.Type,
		FromAgentID:      in.FromAgentID,
		ToAgentID:        in.ToAgentID,
		InReplyTo:        in.InReplyTo,
		RequiresResponse: in.RequiresResponse,
		Body:             in.Body,
		Payload:          in.Payload,
	})
	if err != nil {
		return domain.Message{}, err
	}

	if err := uc.bus.Publish(ctx, created); err != nil {
		return created, fmt.Errorf("publish message: %w", err)
	}
	return created, nil
}

func validMessageType(t domain.MessageType) bool {
	switch t {
	case domain.MessageRequest, domain.MessageResponse, domain.MessageStatus:
		return true
	default:
		return false
	}
}
