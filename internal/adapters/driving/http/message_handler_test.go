package http

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func TestMessageHandlerPost(t *testing.T) {
	a := newTestAPI(t)
	channelID := testutil.RandomUUID()
	fromID := testutil.RandomUUID()

	a.msgRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, m domain.Message) (domain.Message, error) {
			require.Equal(t, channelID, m.ChannelID)
			require.Equal(t, fromID, m.FromAgentID)
			require.Equal(t, domain.MessageStatus, m.Type)
			m.ID = testutil.RandomUUID()
			m.CreatedAt = time.Now()
			return m, nil
		})
	a.msgBus.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil)

	rec := a.do(t, http.MethodPost, "/api/v1/channels/"+channelID.String()+"/messages", map[string]any{
		"type":          string(domain.MessageStatus),
		"from_agent_id": fromID.String(),
		"body":          testutil.RandomText(),
	})

	require.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, channelID.String(), resp["channel_id"])
	require.Equal(t, "status", resp["type"])
}

func TestMessageHandlerPostRejectsInvalidChannelID(t *testing.T) {
	a := newTestAPI(t)
	// No repository or bus call: the path param fails to parse first.

	rec := a.do(t, http.MethodPost, "/api/v1/channels/"+testutil.RandomSlug()+"/messages", map[string]any{
		"type":          string(domain.MessageStatus),
		"from_agent_id": testutil.RandomUUID().String(),
	})

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMessageHandlerPostRejectsUnknownType(t *testing.T) {
	a := newTestAPI(t)

	rec := a.do(t, http.MethodPost, "/api/v1/channels/"+testutil.RandomUUID().String()+"/messages", map[string]any{
		"type":          testutil.RandomSlug(),
		"from_agent_id": testutil.RandomUUID().String(),
	})

	require.Equal(t, http.StatusBadRequest, rec.Code)
}
