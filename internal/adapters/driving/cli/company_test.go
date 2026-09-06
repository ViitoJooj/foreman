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

func TestCompanyCreateCommand(t *testing.T) {
	key := testutil.RandomAPIKey()
	name := testutil.RandomCompanyName()

	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/api/v1/companies", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": testutil.RandomUUID().String(), "slug": gotBody["slug"]})
	}))
	defer srv.Close()

	slug := testutil.RandomSlug()
	out, err := executeRoot(t, "company", "create", "--api-url", srv.URL, "--api-key", key,
		"--name", name, "--slug", slug,
		"--repo-owner", testutil.RandomRepoOwner(), "--repo-name", testutil.RandomRepoName())

	require.NoError(t, err)
	require.Contains(t, out, "created company")
	require.Equal(t, name, gotBody["name"])
	require.Equal(t, slug, gotBody["slug"])
}

func TestCompanyListCommand(t *testing.T) {
	key := testutil.RandomAPIKey()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/companies", r.URL.Path)
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": testutil.RandomUUID().String(), "name": "Acme", "slug": "acme", "repo_owner": "acme", "repo_name": "site"},
		})
	}))
	defer srv.Close()

	out, err := executeRoot(t, "company", "list", "--api-url", srv.URL, "--api-key", key)

	require.NoError(t, err)
	require.Contains(t, out, "Acme")
	require.Contains(t, out, "acme/site")
}

func TestCompanyGetCommand(t *testing.T) {
	key := testutil.RandomAPIKey()
	id := testutil.RandomUUID().String()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/companies/"+id, r.URL.Path)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": id, "name": "Acme", "repo_owner": "acme", "repo_name": "site",
		})
	}))
	defer srv.Close()

	out, err := executeRoot(t, "company", "get", id, "--api-url", srv.URL, "--api-key", key)

	require.NoError(t, err)
	require.Contains(t, out, id)
	require.Contains(t, out, "acme/site")
}

func TestPrintCompanyTable(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	printCompanyTable(cmd, nil)
	require.Contains(t, buf.String(), "no companies")

	buf.Reset()
	printCompanyTable(cmd, []companyView{
		{ID: testutil.RandomUUID().String(), Name: "Acme", Slug: "acme", RepoOwner: "acme", RepoName: "site"},
	})
	require.Contains(t, buf.String(), "Acme")
	require.Contains(t, buf.String(), "acme/site")
}
