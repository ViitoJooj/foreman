package llm_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/adapters/driven/llm"
	"github.com/ViitoJooj/foreman/internal/core/ports"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func anthropicStub(t *testing.T, capture *map[string]any, hdr *http.Header) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/messages", r.URL.Path)
		if hdr != nil {
			*hdr = r.Header.Clone()
		}
		if capture != nil {
			require.NoError(t, json.NewDecoder(r.Body).Decode(capture))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model":   "claude-sonnet-5",
			"content": []map[string]any{{"type": "text", "text": "hi there"}},
			"usage":   map[string]any{"input_tokens": 11, "output_tokens": 7},
		})
	}))
}

func sampleRequest() ports.LLMRequest {
	return ports.LLMRequest{
		System:   testutil.RandomText(),
		Messages: []ports.LLMMessage{{Role: ports.LLMRoleUser, Content: testutil.RandomText()}},
	}
}

func TestAnthropicRequiresCredential(t *testing.T) {
	_, err := llm.NewAnthropic(llm.AnthropicAuth{})
	require.Error(t, err)
}

func TestAnthropicCompleteWithAPIKey(t *testing.T) {
	var body map[string]any
	var hdr http.Header
	srv := anthropicStub(t, &body, &hdr)
	defer srv.Close()

	key := testutil.RandomAPIKey()
	client, err := llm.NewAnthropic(llm.AnthropicAuth{APIKey: key}, llm.WithAnthropicBaseURL(srv.URL))
	require.NoError(t, err)

	resp, err := client.Complete(context.Background(), sampleRequest())
	require.NoError(t, err)
	require.Equal(t, "hi there", resp.Text)
	require.Equal(t, "anthropic", resp.Provider)
	require.Equal(t, 11, resp.Usage.InputTokens)
	require.Equal(t, 7, resp.Usage.OutputTokens)

	require.Equal(t, key, hdr.Get("X-Api-Key"))
	require.Empty(t, hdr.Get("Authorization"))
	require.Equal(t, "claude-sonnet-5", body["model"])
	require.EqualValues(t, 1024, body["max_tokens"])
}

func TestAnthropicCompleteWithOAuthToken(t *testing.T) {
	var hdr http.Header
	srv := anthropicStub(t, nil, &hdr)
	defer srv.Close()

	token := "sk-ant-oat-" + testutil.RandomAPIKey()
	client, err := llm.NewAnthropic(
		llm.AnthropicAuth{OAuthToken: token},
		llm.WithAnthropicBaseURL(srv.URL),
		llm.WithAnthropicModel("claude-opus-5"),
	)
	require.NoError(t, err)

	_, err = client.Complete(context.Background(), sampleRequest())
	require.NoError(t, err)

	require.Equal(t, "Bearer "+token, hdr.Get("Authorization"))
	require.Equal(t, "oauth-2025-04-20", hdr.Get("Anthropic-Beta"))
	require.Empty(t, hdr.Get("X-Api-Key"))
}

func TestAnthropicCompletePropagatesAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"message":"slow down"}}`))
	}))
	defer srv.Close()

	client, err := llm.NewAnthropic(llm.AnthropicAuth{APIKey: testutil.RandomAPIKey()}, llm.WithAnthropicBaseURL(srv.URL))
	require.NoError(t, err)

	_, err = client.Complete(context.Background(), sampleRequest())
	require.ErrorContains(t, err, "slow down")
	require.ErrorContains(t, err, "429")
}
