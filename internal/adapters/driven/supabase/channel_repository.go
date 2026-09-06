package supabase

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/ports"
)

var _ ports.ChannelRepository = (*ChannelRepository)(nil)

// ChannelRepository is the Supabase-backed implementation of ports.ChannelRepository.
type ChannelRepository struct {
	pool *pgxpool.Pool
}

// NewChannelRepository builds a ChannelRepository bound to the given pool.
func NewChannelRepository(pool *pgxpool.Pool) *ChannelRepository {
	return &ChannelRepository{pool: pool}
}

const channelColumns = `id, company_id, name, created_at`

func scanChannel(row pgx.Row) (domain.Channel, error) {
	var ch domain.Channel
	err := row.Scan(&ch.ID, &ch.CompanyID, &ch.Name, &ch.CreatedAt)
	return ch, err
}

// Create inserts a channel and returns it with the database-assigned id and timestamp.
func (r *ChannelRepository) Create(ctx context.Context, ch domain.Channel) (domain.Channel, error) {
	const q = `
		INSERT INTO channels (company_id, name)
		VALUES ($1, $2)
		RETURNING ` + channelColumns
	created, err := scanChannel(r.pool.QueryRow(ctx, q, ch.CompanyID, ch.Name))
	if err != nil {
		return domain.Channel{}, fmt.Errorf("create channel: %w", err)
	}
	return created, nil
}

// Get returns the channel with the given id, or ports.ErrChannelNotFound.
func (r *ChannelRepository) Get(ctx context.Context, id uuid.UUID) (domain.Channel, error) {
	ch, err := scanChannel(r.pool.QueryRow(ctx, `SELECT `+channelColumns+` FROM channels WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Channel{}, ports.ErrChannelNotFound
	}
	if err != nil {
		return domain.Channel{}, fmt.Errorf("get channel: %w", err)
	}
	return ch, nil
}

// ListByCompany returns every channel of a company, oldest first.
func (r *ChannelRepository) ListByCompany(ctx context.Context, companyID uuid.UUID) ([]domain.Channel, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+channelColumns+` FROM channels WHERE company_id = $1 ORDER BY created_at`, companyID)
	if err != nil {
		return nil, fmt.Errorf("list channels by company: %w", err)
	}
	defer rows.Close()

	var channels []domain.Channel
	for rows.Next() {
		ch, err := scanChannel(rows)
		if err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}
