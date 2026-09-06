package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/testutil"
)

func commandServer(t *testing.T, gotBody *map[string]any, gotPath *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if gotPath != nil {
			*gotPath = r.URL.Path + "?" + r.URL.RawQuery
		}
		if r.Method == http.MethodPost && gotBody != nil {
			require.NoError(t, json.NewDecoder(r.Body).Decode(gotBody))
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": testutil.RandomUUID().String(), "kind": "kill", "status": "pending",
		})
	}))
}

func TestKillCommand(t *testing.T) {
	key := testutil.RandomAPIKey()
	company := testutil.RandomUUID().String()

	var gotBody map[string]any
	srv := commandServer(t, &gotBody, nil)
	defer srv.Close()

	out, err := executeRoot(t, "kill", "--api-url", srv.URL, "--api-key", key,
		"--company", company, "--reason", "manual stop")

	require.NoError(t, err)
	require.Contains(t, out, "issued")
	require.Equal(t, "kill", gotBody["kind"])
	require.Equal(t, company, gotBody["company_id"])
	require.Equal(t, "manual stop", gotBody["reason"])
}

func TestKillCommandPanicFlag(t *testing.T) {
	var gotBody map[string]any
	srv := commandServer(t, &gotBody, nil)
	defer srv.Close()

	_, err := executeRoot(t, "kill", "--panic", "--api-url", srv.URL, "--api-key", testutil.RandomAPIKey())

	require.NoError(t, err)
	require.Equal(t, "panic", gotBody["kind"])
}

func TestPauseCommandRequiresAgent(t *testing.T) {
	_, err := executeRoot(t, "pause", "--api-key", testutil.RandomAPIKey())
	require.Error(t, err)
}

func TestPauseAndResumeCommands(t *testing.T) {
	agent := testutil.RandomUUID().String()

	for _, kind := range []string{"pause", "resume"} {
		t.Run(kind, func(t *testing.T) {
			var gotBody map[string]any
			srv := commandServer(t, &gotBody, nil)
			defer srv.Close()

			_, err := executeRoot(t, kind, "--agent", agent,
				"--api-url", srv.URL, "--api-key", testutil.RandomAPIKey())

			require.NoError(t, err)
			require.Equal(t, kind, gotBody["kind"])
			require.Equal(t, agent, gotBody["target_agent_id"])
		})
	}
}

func TestCommandListCommand(t *testing.T) {
	key := testutil.RandomAPIKey()

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": testutil.RandomUUID().String(), "kind": "kill", "status": "pending", "reason": "stop"},
		})
	}))
	defer srv.Close()

	out, err := executeRoot(t, "command", "list", "--status", "pending",
		"--api-url", srv.URL, "--api-key", key)

	require.NoError(t, err)
	require.Contains(t, out, "kill")
	require.Contains(t, gotPath, "status=pending")
}

func TestPrintCommandTable(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	printCommandTable(cmd, nil)
	require.Contains(t, buf.String(), "no commands")

	buf.Reset()
	printCommandTable(cmd, []commandView{
		{ID: testutil.RandomUUID().String(), Kind: "panic", Status: "pending"},
	})
	require.Contains(t, buf.String(), "panic")
	require.Contains(t, buf.String(), "-", "empty reason renders as a dash")
}
