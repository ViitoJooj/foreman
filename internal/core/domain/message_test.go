package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

// TestMessageTypeValues pins the wire values of message_type, which must match
// the Postgres enum in migrations/0001_init.up.sql and the inter-agent protocol.
func TestMessageTypeValues(t *testing.T) {
	types := map[domain.MessageType]string{
		domain.MessageRequest:  "request",
		domain.MessageResponse: "response",
		domain.MessageStatus:   "status",
	}
	for mt, want := range types {
		require.Equal(t, want, string(mt))
	}
}

func TestMessagePayloadIsRawJSON(t *testing.T) {
	payload := testutil.RandomJSONPayload()

	m := domain.Message{
		ID:          testutil.RandomUUID(),
		ChannelID:   testutil.RandomUUID(),
		Type:        testutil.RandomMessageType(),
		FromAgentID: testutil.RandomUUID(),
		Body:        testutil.RandomText(),
		Payload:     payload,
	}

	require.True(t, json.Valid(m.Payload))

	var decoded map[string]string
	require.NoError(t, json.Unmarshal(m.Payload, &decoded))
	require.Len(t, decoded, 1)
}
