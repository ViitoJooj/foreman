package bus_test

import (
	"context"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/adapters/driven/bus"
	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func messageOn(channelID uuid.UUID) domain.Message {
	return domain.Message{
		ID:          testutil.RandomUUID(),
		ChannelID:   channelID,
		Type:        testutil.RandomMessageType(),
		FromAgentID: testutil.RandomUUID(),
		Body:        testutil.RandomText(),
	}
}

func recv(t *testing.T, ch <-chan domain.Message) domain.Message {
	t.Helper()
	select {
	case m := <-ch:
		return m
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for a message")
		return domain.Message{}
	}
}

func TestInProcessDeliversToChannelSubscribers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b := bus.NewInProcess()
	channelID := testutil.RandomUUID()

	sub1, err := b.Subscribe(ctx, channelID)
	require.NoError(t, err)
	sub2, err := b.Subscribe(ctx, channelID)
	require.NoError(t, err)

	msg := messageOn(channelID)
	require.NoError(t, b.Publish(ctx, msg))

	require.Equal(t, msg, recv(t, sub1))
	require.Equal(t, msg, recv(t, sub2))
}

func TestInProcessIsolatesChannels(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b := bus.NewInProcess()
	subscribed := testutil.RandomUUID()
	other := testutil.RandomUUID()

	sub, err := b.Subscribe(ctx, subscribed)
	require.NoError(t, err)

	require.NoError(t, b.Publish(ctx, messageOn(other)))

	select {
	case m := <-sub:
		t.Fatalf("received a message for another channel: %+v", m)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestInProcessPublishWithoutSubscribersIsNoop(t *testing.T) {
	b := bus.NewInProcess()
	require.NoError(t, b.Publish(context.Background(), messageOn(testutil.RandomUUID())))
}

func TestInProcessClosesSubscriptionOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	b := bus.NewInProcess()
	sub, err := b.Subscribe(ctx, testutil.RandomUUID())
	require.NoError(t, err)

	cancel()

	require.Eventually(t, func() bool {
		_, open := <-sub
		return !open
	}, time.Second, 10*time.Millisecond)
}

func TestInProcessRejectsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	b := bus.NewInProcess()

	_, err := b.Subscribe(ctx, testutil.RandomUUID())
	require.ErrorIs(t, err, context.Canceled)

	require.ErrorIs(t, b.Publish(ctx, messageOn(testutil.RandomUUID())), context.Canceled)
}

func TestInProcessDropsMessagesForFullSubscriber(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b := bus.NewInProcess()
	channelID := testutil.RandomUUID()

	sub, err := b.Subscribe(ctx, channelID)
	require.NoError(t, err)

	// Never drain sub: publishing far past its buffer must not block.
	done := make(chan struct{})
	go func() {
		for range 1000 {
			_ = b.Publish(ctx, messageOn(channelID))
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Publish blocked on a slow subscriber")
	}

	require.NotEmpty(t, sub)
}
