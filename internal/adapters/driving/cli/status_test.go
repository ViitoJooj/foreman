package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/testutil"
)

func statusStubServer(t *testing.T, companyID string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/companies":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": companyID, "name": "Demo", "slug": "demo", "repo_owner": "me", "repo_name": "demo-repo"},
			})
		case "/api/v1/tasks":
			require.Equal(t, companyID, r.URL.Query().Get("company"))
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": testutil.RandomUUID().String(), "title": "a", "state": "queued", "risk": "low"},
				{"id": testutil.RandomUUID().String(), "title": "b", "state": "merged", "risk": "low"},
				{"id": testutil.RandomUUID().String(), "title": "c", "state": "merged", "risk": "low"},
			})
		case "/api/v1/agents":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": testutil.RandomUUID().String(), "name": "Ada", "role": "coder", "status": "working"},
			})
		case "/api/v1/commands":
			require.Equal(t, "pending", r.URL.Query().Get("status"))
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": testutil.RandomUUID().String(), "kind": "kill", "status": "pending", "reason": "stop"},
			})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
}

func TestStatusCommandText(t *testing.T) {
	companyID := testutil.RandomUUID().String()
	srv := statusStubServer(t, companyID)
	defer srv.Close()

	out, err := executeRoot(t, "status", "--api-url", srv.URL, "--api-key", testutil.RandomAPIKey())
	require.NoError(t, err)

	require.Contains(t, out, "Demo (demo)  me/demo-repo")
	require.Contains(t, out, "merged 2")
	require.Contains(t, out, "queued 1")
	require.Contains(t, out, "Ada")
	require.Contains(t, out, "working")
	require.Contains(t, out, "pending commands: 1")
}

func TestStatusCommandJSON(t *testing.T) {
	companyID := testutil.RandomUUID().String()
	srv := statusStubServer(t, companyID)
	defer srv.Close()

	out, err := executeRoot(t, "status", "--json", "--api-url", srv.URL, "--api-key", testutil.RandomAPIKey())
	require.NoError(t, err)

	var report statusReport
	require.NoError(t, json.Unmarshal([]byte(out), &report))
	require.Len(t, report.Companies, 1)
	require.Equal(t, 2, report.Companies[0].States["merged"])
	require.Len(t, report.PendingCommands, 1)
}

func TestFormatStates(t *testing.T) {
	require.Equal(t, "(none)", formatStates(map[string]int{}))
	require.Equal(t, "coding 1  merged 3", formatStates(map[string]int{"merged": 3, "coding": 1}))
}

func TestPrintStatusNoCompanies(t *testing.T) {
	cmd := &cobra.Command{}
	var buf strings.Builder
	cmd.SetOut(&buf)

	printStatus(cmd, statusReport{})
	require.Contains(t, buf.String(), "no companies")
	require.Contains(t, buf.String(), "pending commands: 0")
}
