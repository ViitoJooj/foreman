package ports

import (
	"context"
	"uuid"

	"github.com/ViitoJooj/foreman/internal/core/domain"
)

// MessageBus delivers channel messages in real time. It only broadcasts and
// streams; persisting messages is MessageRepository's responsibility.
type MessageBus interface {
	// Publish broadcasts an already-persisted message to live subscribers.
	Publish(ctx context.Context, m domain.Message) error
	// Subscribe streams messages posted to the channel until ctx is cancelled,
	// at which point the returned channel is closed.
	Subscribe(ctx context.Context, channelID uuid.UUID) (<-chan domain.Message, error)
}
