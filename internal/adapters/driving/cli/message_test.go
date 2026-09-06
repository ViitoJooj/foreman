package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/testutil"
)

func TestMessagePostCommand(t *testing.T) {
	key := testutil.RandomAPIKey()
	channel := testutil.RandomUUID().String()
	from := testutil.RandomUUID().String()

	var gotPath string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": testutil.RandomUUID().String(), "type": "status", "channel_id": channel,
		})
	}))
	defer srv.Close()

	body := testutil.RandomText()
	out, err := executeRoot(t, "message", "post",
		"--api-url", srv.URL, "--api-key", key,
		"--channel", channel, "--type", "status", "--from", from, "--body", body)

	require.NoError(t, err)
	require.Contains(t, out, "posted")
	require.Equal(t, "/api/v1/channels/"+channel+"/messages", gotPath)
	require.Equal(t, from, gotBody["from_agent_id"])
	require.Equal(t, body, gotBody["body"])
}

func TestMessagePostRejectsInvalidPayload(t *testing.T) {
	_, err := executeRoot(t, "message", "post",
		"--api-key", testutil.RandomAPIKey(),
		"--channel", testutil.RandomUUID().String(),
		"--type", "status", "--from", testutil.RandomUUID().String(),
		"--payload", "{not valid")

	require.ErrorContains(t, err, "valid JSON")
}
