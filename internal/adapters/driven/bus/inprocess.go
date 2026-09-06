// Package bus provides an in-process implementation of ports.MessageBus. It is a
// stand-in until a Supabase Realtime-backed bus is wired in; it broadcasts only
// within a single running process.
package bus

import (
	"context"
	"sync"
	"uuid"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/ports"
)

var _ ports.MessageBus = (*InProcess)(nil)

// subBuffer is the per-subscriber channel capacity. Sends that would block
// because a subscriber is not keeping up are dropped rather than stalling the
// publisher.
const subBuffer = 32

// InProcess fans out published messages to every active subscriber of the
// target channel. It is safe for concurrent use.
type InProcess struct {
	mu   sync.Mutex
	subs map[uuid.UUID]map[chan domain.Message]struct{}
}

// NewInProcess returns an empty in-process bus.
func NewInProcess() *InProcess {
	return &InProcess{subs: make(map[uuid.UUID]map[chan domain.Message]struct{})}
}

// Publish delivers m to every current subscriber of m.ChannelID. A subscriber
// whose buffer is full is skipped.
func (b *InProcess) Publish(ctx context.Context, m domain.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subs[m.ChannelID] {
		select {
		case ch <- m:
		default:
		}
	}
	return nil
}

// Subscribe registers a new subscriber for channelID. The returned channel is
// closed and the subscription removed when ctx is cancelled.
func (b *InProcess) Subscribe(ctx context.Context, channelID uuid.UUID) (<-chan domain.Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	ch := make(chan domain.Message, subBuffer)

	b.mu.Lock()
	if b.subs[channelID] == nil {
		b.subs[channelID] = make(map[chan domain.Message]struct{})
	}
	b.subs[channelID][ch] = struct{}{}
	b.mu.Unlock()

	go func() {
		<-ctx.Done()
		b.mu.Lock()
		delete(b.subs[channelID], ch)
		if len(b.subs[channelID]) == 0 {
			delete(b.subs, channelID)
		}
		b.mu.Unlock()
		close(ch)
	}()

	return ch, nil
}
