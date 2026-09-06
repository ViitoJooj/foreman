package cli

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/testutil"
)

func TestAPIClientDoDecodesSuccess(t *testing.T) {
	key := testutil.RandomAPIKey()
	type payload struct {
		Name string `json:"name"`
	}
	want := payload{Name: testutil.RandomText()}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, key, r.Header.Get("X-Api-Key"))
		_ = json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	var got payload
	err := newAPIClient(srv.URL, key).do(context.Background(), http.MethodGet, "/x", nil, &got)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestAPIClientDoSurfacesAPIError(t *testing.T) {
	message := testutil.RandomText()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
	}))
	defer srv.Close()

	err := newAPIClient(srv.URL, testutil.RandomAPIKey()).do(context.Background(), http.MethodGet, "/x", nil, nil)
	require.ErrorContains(t, err, message)
	require.ErrorContains(t, err, "400")
}

func TestAPIClientDoReportsUnreachableServer(t *testing.T) {
	err := newAPIClient("http://127.0.0.1:1", testutil.RandomAPIKey()).
		do(context.Background(), http.MethodGet, "/x", nil, nil)
	require.ErrorContains(t, err, "calling")
}

func TestAPIClientDoRejectsInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "{not json")
	}))
	defer srv.Close()

	var got map[string]any
	err := newAPIClient(srv.URL, testutil.RandomAPIKey()).do(context.Background(), http.MethodGet, "/x", nil, &got)
	require.ErrorContains(t, err, "decoding response")
}

func TestAPIErrorMessage(t *testing.T) {
	require.Contains(t, apiErrorMessage("400 Bad Request", []byte(`{"error":"nope"}`)), "nope")
	require.Contains(t, apiErrorMessage("500 Internal Server Error", []byte("raw body")), "raw body")
	require.Contains(t, apiErrorMessage("503 Service Unavailable", nil), "503")
}
