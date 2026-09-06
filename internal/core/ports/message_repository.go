package ports

import (
	"context"
	"errors"
	"uuid"

	"github.com/ViitoJooj/foreman/internal/core/domain"
)

// ErrMessageNotFound is returned when no message matches the given identifier.
var ErrMessageNotFound = errors.New("message not found")

// MessageRepository persists channel messages. Messages are immutable once
// created, so there is no Update.
type MessageRepository interface {
	Create(ctx context.Context, m domain.Message) (domain.Message, error)
	Get(ctx context.Context, id uuid.UUID) (domain.Message, error)
	ListByChannel(ctx context.Context, channelID uuid.UUID, limit int) ([]domain.Message, error)
}
