package service

import (
	"context"
	"uuid"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/ports"
)

// TailMessages streams a channel's messages: a bounded backlog followed by a
// live feed until the caller's context is cancelled.
type TailMessages struct {
	messages ports.MessageRepository
	bus      ports.MessageBus
}

// NewTailMessages wires the use case to a message repository and a message bus.
func NewTailMessages(messages ports.MessageRepository, bus ports.MessageBus) *TailMessages {
	return &TailMessages{messages: messages, bus: bus}
}

// Execute returns the last historyLimit messages of the channel plus a channel
// that yields every message posted afterwards. A non-positive historyLimit skips
// the backlog.
func (uc *TailMessages) Execute(ctx context.Context, channelID uuid.UUID, historyLimit int) ([]domain.Message, <-chan domain.Message, error) {
	live, err := uc.bus.Subscribe(ctx, channelID)
	if err != nil {
		return nil, nil, err
	}

	var backlog []domain.Message
	if historyLimit > 0 {
		backlog, err = uc.messages.ListByChannel(ctx, channelID, historyLimit)
		if err != nil {
			return nil, nil, err
		}
	}
	return backlog, live, nil
}
