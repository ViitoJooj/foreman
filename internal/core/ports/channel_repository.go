package ports

import (
	"context"
	"errors"
	"uuid"

	"github.com/ViitoJooj/foreman/internal/core/domain"
)

// ErrChannelNotFound is returned when no channel matches the given identifier.
var ErrChannelNotFound = errors.New("channel not found")

// ChannelRepository persists the message streams belonging to each company.
type ChannelRepository interface {
	Create(ctx context.Context, ch domain.Channel) (domain.Channel, error)
	Get(ctx context.Context, id uuid.UUID) (domain.Channel, error)
	ListByCompany(ctx context.Context, companyID uuid.UUID) ([]domain.Channel, error)
}
