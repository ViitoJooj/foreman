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

var _ ports.MessageRepository = (*MessageRepository)(nil)

// MessageRepository is the Supabase-backed implementation of ports.MessageRepository.
type MessageRepository struct {
	pool *pgxpool.Pool
}

// NewMessageRepository builds a MessageRepository bound to the given pool.
func NewMessageRepository(pool *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{pool: pool}
}

const messageColumns = `id, channel_id, type, from_agent_id, to_agent_id, in_reply_to,
	requires_response, body, payload, created_at`

func scanMessage(row pgx.Row) (domain.Message, error) {
	var m domain.Message
	err := row.Scan(
		&m.ID, &m.ChannelID, &m.Type, &m.FromAgentID, &m.ToAgentID, &m.InReplyTo,
		&m.RequiresResponse, &m.Body, &m.Payload, &m.CreatedAt,
	)
	return m, err
}

// Create inserts a message and returns it with the database-assigned id and timestamp.
func (r *MessageRepository) Create(ctx context.Context, m domain.Message) (domain.Message, error) {
	payload := m.Payload
	if len(payload) == 0 {
		payload = []byte("{}")
	}
	const q = `
		INSERT INTO messages (channel_id, type, from_agent_id, to_agent_id, in_reply_to,
			requires_response, body, payload)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING ` + messageColumns
	created, err := scanMessage(r.pool.QueryRow(ctx, q,
		m.ChannelID, m.Type, m.FromAgentID, nullableUUID(m.ToAgentID), nullableUUID(m.InReplyTo),
		m.RequiresResponse, m.Body, payload,
	))
	if err != nil {
		return domain.Message{}, fmt.Errorf("create message: %w", err)
	}
	return created, nil
}

// Get returns the message with the given id, or ports.ErrMessageNotFound.
func (r *MessageRepository) Get(ctx context.Context, id uuid.UUID) (domain.Message, error) {
	m, err := scanMessage(r.pool.QueryRow(ctx, `SELECT `+messageColumns+` FROM messages WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Message{}, ports.ErrMessageNotFound
	}
	if err != nil {
		return domain.Message{}, fmt.Errorf("get message: %w", err)
	}
	return m, nil
}

// ListByChannel returns the most recent messages of a channel in chronological
// order. A non-positive limit returns every message in the channel.
func (r *MessageRepository) ListByChannel(ctx context.Context, channelID uuid.UUID, limit int) ([]domain.Message, error) {
	sql := `SELECT ` + messageColumns + ` FROM messages WHERE channel_id = $1 ORDER BY created_at`
	args := []any{channelID}
	if limit > 0 {
		sql = `SELECT * FROM (` + sql + ` DESC LIMIT $2) AS recent ORDER BY created_at`
		args = append(args, limit)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list messages by channel: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		m, err := scanMessage(rows)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}
