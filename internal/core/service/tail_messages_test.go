package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/service"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func TestTailMessagesExecuteReturnsBacklogAndLiveFeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockMessageRepository(ctrl)
	bus := testutil.NewMockMessageBus(ctrl)

	channelID := testutil.RandomUUID()
	backlog := []domain.Message{{ID: testutil.RandomUUID(), ChannelID: channelID, Body: testutil.RandomText()}}
	live := make(chan domain.Message)

	bus.EXPECT().Subscribe(gomock.Any(), channelID).Return((<-chan domain.Message)(live), nil)
	repo.EXPECT().ListByChannel(gomock.Any(), channelID, 5).Return(backlog, nil)

	gotBacklog, gotLive, err := service.NewTailMessages(repo, bus).Execute(context.Background(), channelID, 5)
	require.NoError(t, err)
	require.Equal(t, backlog, gotBacklog)
	require.NotNil(t, gotLive)
}

func TestTailMessagesExecuteSkipsBacklogWhenLimitNonPositive(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockMessageRepository(ctrl)
	bus := testutil.NewMockMessageBus(ctrl)

	channelID := testutil.RandomUUID()
	live := make(chan domain.Message)
	bus.EXPECT().Subscribe(gomock.Any(), channelID).Return((<-chan domain.Message)(live), nil)
	// No ListByChannel call expected.

	backlog, _, err := service.NewTailMessages(repo, bus).Execute(context.Background(), channelID, 0)
	require.NoError(t, err)
	require.Empty(t, backlog)
}

func TestTailMessagesExecutePropagatesSubscribeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockMessageRepository(ctrl)
	bus := testutil.NewMockMessageBus(ctrl)

	bus.EXPECT().Subscribe(gomock.Any(), gomock.Any()).Return(nil, errors.New(testutil.RandomText()))

	_, _, err := service.NewTailMessages(repo, bus).Execute(context.Background(), testutil.RandomUUID(), 10)
	require.Error(t, err)
}
