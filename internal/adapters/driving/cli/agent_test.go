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

func TestAgentCreateCommand(t *testing.T) {
	key := testutil.RandomAPIKey()
	company := testutil.RandomUUID().String()
	name := testutil.RandomAgentName()

	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/agents", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": testutil.RandomUUID().String(), "role": gotBody["role"], "status": "idle",
		})
	}))
	defer srv.Close()

	out, err := executeRoot(t, "agent", "create", "--api-url", srv.URL, "--api-key", key,
		"--company", company, "--name", name, "--role", "coder")

	require.NoError(t, err)
	require.Contains(t, out, "created agent")
	require.Equal(t, company, gotBody["company_id"])
	require.Equal(t, name, gotBody["name"])
	require.Equal(t, "coder", gotBody["role"])
}

func TestAgentListCommand(t *testing.T) {
	key := testutil.RandomAPIKey()
	company := testutil.RandomUUID().String()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/agents", r.URL.Path)
		require.Equal(t, company, r.URL.Query().Get("company"))
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": testutil.RandomUUID().String(), "name": "Ada", "role": "coder", "status": "working"},
		})
	}))
	defer srv.Close()

	out, err := executeRoot(t, "agent", "list", "--api-url", srv.URL, "--api-key", key, "--company", company)

	require.NoError(t, err)
	require.Contains(t, out, "Ada")
	require.Contains(t, out, "working")
}

func TestAgentGetCommand(t *testing.T) {
	key := testutil.RandomAPIKey()
	id := testutil.RandomUUID().String()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/agents/"+id, r.URL.Path)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": id, "name": "Ada", "role": "tester", "status": "idle"})
	}))
	defer srv.Close()

	out, err := executeRoot(t, "agent", "get", id, "--api-url", srv.URL, "--api-key", key)

	require.NoError(t, err)
	require.Contains(t, out, "Ada")
	require.Contains(t, out, "tester")
}

func TestPrintAgentTable(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	printAgentTable(cmd, nil)
	require.Contains(t, buf.String(), "no agents")

	buf.Reset()
	printAgentTable(cmd, []agentView{
		{ID: testutil.RandomUUID().String(), Name: "Ada", Role: "coder", Status: "idle"},
	})
	require.Contains(t, buf.String(), "Ada")
	require.Contains(t, buf.String(), "coder")
}
