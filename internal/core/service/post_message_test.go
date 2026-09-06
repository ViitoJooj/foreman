package service_test

import (
	"context"
	"errors"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/service"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func validPostMessageInput() service.PostMessageInput {
	return service.PostMessageInput{
		ChannelID:   testutil.RandomUUID(),
		Type:        testutil.RandomMessageType(),
		FromAgentID: testutil.RandomUUID(),
		Body:        testutil.RandomText(),
		Payload:     testutil.RandomJSONPayload(),
	}
}

func TestPostMessageExecutePersistsThenBroadcasts(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockMessageRepository(ctrl)
	bus := testutil.NewMockMessageBus(ctrl)

	in := validPostMessageInput()
	stored := domain.Message{
		ID:          testutil.RandomUUID(),
		ChannelID:   in.ChannelID,
		Type:        in.Type,
		FromAgentID: in.FromAgentID,
		Body:        in.Body,
		Payload:     in.Payload,
	}

	gomock.InOrder(
		repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, m domain.Message) (domain.Message, error) {
				require.Equal(t, in.ChannelID, m.ChannelID)
				require.Equal(t, in.Type, m.Type)
				require.Equal(t, in.FromAgentID, m.FromAgentID)
				return stored, nil
			}),
		bus.EXPECT().Publish(gomock.Any(), stored).Return(nil),
	)

	out, err := service.NewPostMessage(repo, bus).Execute(context.Background(), in)
	require.NoError(t, err)
	require.Equal(t, stored, out)
}

func TestPostMessageExecuteDoesNotBroadcastWhenPersistFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockMessageRepository(ctrl)
	bus := testutil.NewMockMessageBus(ctrl)

	wantErr := errors.New(testutil.RandomText())
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(domain.Message{}, wantErr)
	// bus.Publish must not be called.

	_, err := service.NewPostMessage(repo, bus).Execute(context.Background(), validPostMessageInput())
	require.ErrorIs(t, err, wantErr)
}

func TestPostMessageExecuteReturnsStoredMessageWhenBroadcastFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := testutil.NewMockMessageRepository(ctrl)
	bus := testutil.NewMockMessageBus(ctrl)

	in := validPostMessageInput()
	stored := domain.Message{ID: testutil.RandomUUID(), ChannelID: in.ChannelID, Type: in.Type, FromAgentID: in.FromAgentID}
	busErr := errors.New(testutil.RandomText())

	repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(stored, nil)
	bus.EXPECT().Publish(gomock.Any(), stored).Return(busErr)

	out, err := service.NewPostMessage(repo, bus).Execute(context.Background(), in)
	require.ErrorIs(t, err, busErr)
	require.Equal(t, stored, out, "the message is already persisted, so it is still returned")
}

func TestPostMessageExecuteRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*service.PostMessageInput)
	}{
		{"missing channel", func(in *service.PostMessageInput) { in.ChannelID = uuid.UUID{} }},
		{"missing from agent", func(in *service.PostMessageInput) { in.FromAgentID = uuid.UUID{} }},
		{"unknown type", func(in *service.PostMessageInput) { in.Type = domain.MessageType(testutil.RandomSlug()) }},
		{"malformed payload", func(in *service.PostMessageInput) { in.Payload = []byte("{not json") }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := testutil.NewMockMessageRepository(ctrl)
			bus := testutil.NewMockMessageBus(ctrl)
			// Neither dependency should be touched.

			in := validPostMessageInput()
			tc.mutate(&in)

			_, err := service.NewPostMessage(repo, bus).Execute(context.Background(), in)
			require.ErrorIs(t, err, service.ErrInvalidMessageInput)
		})
	}
}
