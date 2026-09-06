package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/testutil"
)

func TestChannelCreateCommand(t *testing.T) {
	key := testutil.RandomAPIKey()
	company := testutil.RandomUUID().String()
	name := testutil.RandomChannelName()

	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/channels", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": testutil.RandomUUID().String(), "name": name})
	}))
	defer srv.Close()

	out, err := executeRoot(t, "channel", "create", "--api-url", srv.URL, "--api-key", key,
		"--company", company, "--name", name)

	require.NoError(t, err)
	require.Contains(t, out, "created channel")
	require.Equal(t, company, gotBody["company_id"])
	require.Equal(t, name, gotBody["name"])
}

func TestChannelListCommand(t *testing.T) {
	key := testutil.RandomAPIKey()
	company := testutil.RandomUUID().String()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/channels", r.URL.Path)
		require.Equal(t, company, r.URL.Query().Get("company"))
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": testutil.RandomUUID().String(), "name": "general", "company_id": company},
		})
	}))
	defer srv.Close()

	out, err := executeRoot(t, "channel", "list", "--api-url", srv.URL, "--api-key", key, "--company", company)

	require.NoError(t, err)
	require.Contains(t, out, "general")
}

func TestChannelGetCommand(t *testing.T) {
	key := testutil.RandomAPIKey()
	id := testutil.RandomUUID().String()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/channels/"+id, r.URL.Path)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": id, "name": "general"})
	}))
	defer srv.Close()

	out, err := executeRoot(t, "channel", "get", id, "--api-url", srv.URL, "--api-key", key)

	require.NoError(t, err)
	require.Contains(t, out, id)
	require.Contains(t, out, "general")
}

func TestChannelTailCommand(t *testing.T) {
	key := testutil.RandomAPIKey()
	channel := testutil.RandomUUID().String()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/channels/"+channel+"/messages/stream", r.URL.Path)
		require.Equal(t, "5", r.URL.Query().Get("history"))
		require.Equal(t, key, r.Header.Get("X-Api-Key"))
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: {\"type\":\"status\",\"body\":\"hello\",\"created_at\":\"2026-09-06T12:00:00Z\"}\n\n")
		fmt.Fprint(w, "data: {\"type\":\"request\",\"body\":\"world\",\"created_at\":\"2026-09-06T12:00:01Z\"}\n\n")
	}))
	defer srv.Close()

	out, err := executeRoot(t, "channel", "tail",
		"--channel", channel, "--history", "5",
		"--api-url", srv.URL, "--api-key", key)

	require.NoError(t, err)
	require.Contains(t, out, "hello")
	require.Contains(t, out, "world")
	require.Contains(t, out, "status")
}

func TestPrintChannelTable(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	printChannelTable(cmd, nil)
	require.Contains(t, buf.String(), "no channels")

	buf.Reset()
	company := testutil.RandomUUID().String()
	printChannelTable(cmd, []channelView{
		{ID: testutil.RandomUUID().String(), Name: "general", CompanyID: company},
	})
	require.Contains(t, buf.String(), "general")
	require.Contains(t, buf.String(), company)
}
