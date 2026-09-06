package service

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/ports"
)

// ErrInvalidChannelInput is returned when Channel.Create is given missing fields.
var ErrInvalidChannelInput = errors.New("invalid channel input")

// Channel groups the channel use cases.
type Channel struct {
	repo ports.ChannelRepository
}

// NewChannel wires the use cases to a channel repository.
func NewChannel(repo ports.ChannelRepository) *Channel {
	return &Channel{repo: repo}
}

// CreateChannelInput carries the fields required to open a channel in a company.
type CreateChannelInput struct {
	CompanyID uuid.UUID
	Name      string
}

// Create validates the input and persists the new channel.
func (uc *Channel) Create(ctx context.Context, in CreateChannelInput) (domain.Channel, error) {
	if in.CompanyID == (uuid.UUID{}) {
		return domain.Channel{}, fmt.Errorf("%w: company id is required", ErrInvalidChannelInput)
	}
	if in.Name == "" {
		return domain.Channel{}, fmt.Errorf("%w: name is required", ErrInvalidChannelInput)
	}
	return uc.repo.Create(ctx, domain.Channel{
		CompanyID: in.CompanyID,
		Name:      in.Name,
	})
}

// Get returns the channel with the given id.
func (uc *Channel) Get(ctx context.Context, id uuid.UUID) (domain.Channel, error) {
	return uc.repo.Get(ctx, id)
}

// List returns every channel of a company, oldest first.
func (uc *Channel) List(ctx context.Context, companyID uuid.UUID) ([]domain.Channel, error) {
	return uc.repo.ListByCompany(ctx, companyID)
}
