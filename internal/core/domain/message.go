package domain

import (
	"encoding/json"
	"time"
	"uuid"
)

// MessageType is the protocol envelope kind exchanged between agents.
type MessageType string

// The protocol envelope kinds.
const (
	MessageRequest  MessageType = "request"
	MessageResponse MessageType = "response"
	MessageStatus   MessageType = "status"
)

// Message is a single entry in a channel, following the inter-agent protocol.
type Message struct {
	ID               uuid.UUID
	ChannelID        uuid.UUID
	Type             MessageType
	FromAgentID      uuid.UUID
	ToAgentID        uuid.UUID // zero for broadcast / status messages
	InReplyTo        uuid.UUID // zero when the message is not a reply
	RequiresResponse bool
	Body             string
	Payload          json.RawMessage // free-form protocol payload, stored verbatim
	CreatedAt        time.Time
}
