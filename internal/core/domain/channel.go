package domain

import (
	"time"
	"uuid"
)

// Channel is a Slack-style message stream that a company's agents post to.
type Channel struct {
	ID        uuid.UUID
	CompanyID uuid.UUID
	Name      string // e.g. "general"
	CreatedAt time.Time
}
