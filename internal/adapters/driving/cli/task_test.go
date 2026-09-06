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

func TestTaskCreateCommand(t *testing.T) {
	key := testutil.RandomAPIKey()
	company := testutil.RandomUUID().String()
	title := testutil.RandomTaskTitle()

	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/api/v1/tasks", r.URL.Path)
		require.Equal(t, key, r.Header.Get("X-Api-Key"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": testutil.RandomUUID().String(), "state": "queued"})
	}))
	defer srv.Close()

	out, err := executeRoot(t, "task", "create",
		"--api-url", srv.URL, "--api-key", key,
		"--company", company, "--title", title, "--risk", "low")

	require.NoError(t, err)
	require.Contains(t, out, "created task")
	require.Equal(t, company, gotBody["company_id"])
	require.Equal(t, title, gotBody["title"])
	require.Equal(t, "low", gotBody["risk"])
}

func TestTaskListCommand(t *testing.T) {
	key := testutil.RandomAPIKey()
	company := testutil.RandomUUID().String()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/tasks", r.URL.Path)
		require.Equal(t, company, r.URL.Query().Get("company"))
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": testutil.RandomUUID().String(), "title": "Ship it", "state": "coding", "risk": "low", "retries": 0, "max_retries": 3, "branch": ""},
		})
	}))
	defer srv.Close()

	out, err := executeRoot(t, "task", "list", "--api-url", srv.URL, "--api-key", key, "--company", company)

	require.NoError(t, err)
	require.Contains(t, out, "Ship it")
	require.Contains(t, out, "coding")
}

func TestPrintTaskTable(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	printTaskTable(cmd, nil)
	require.Contains(t, buf.String(), "no tasks")

	buf.Reset()
	printTaskTable(cmd, []taskView{
		{ID: testutil.RandomUUID().String(), Title: "Hello", State: "coding", Risk: "low", MaxRetries: 3},
	})
	require.Contains(t, buf.String(), "Hello")
	require.Contains(t, buf.String(), "coding")
	require.Contains(t, buf.String(), "-", "empty branch renders as a dash")
}
