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

func openAIStub(t *testing.T, capture *map[string]any, hdr *http.Header) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/chat/completions", r.URL.Path)
		if hdr != nil {
			*hdr = r.Header.Clone()
		}
		if capture != nil {
			require.NoError(t, json.NewDecoder(r.Body).Decode(capture))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model":   "gpt-4o",
			"choices": []map[string]any{{"message": map[string]any{"role": "assistant", "content": "pong"}}},
			"usage":   map[string]any{"prompt_tokens": 9, "completion_tokens": 4},
		})
	}))
}

func TestOpenAICompatibleRequiresToken(t *testing.T) {
	_, err := llm.NewOpenAICompatible("x", "https://example.com", "")
	require.Error(t, err)
}

func TestOpenAICompatibleCompletePlacesSystemFirst(t *testing.T) {
	var body map[string]any
	var hdr http.Header
	srv := openAIStub(t, &body, &hdr)
	defer srv.Close()

	token := testutil.RandomAPIKey()
	client, err := llm.NewOpenAICompatible("openai", srv.URL, token, llm.WithOpenAIModel("gpt-4o"))
	require.NoError(t, err)

	system := testutil.RandomText()
	resp, err := client.Complete(context.Background(), ports.LLMRequest{
		System:   system,
		Messages: []ports.LLMMessage{{Role: ports.LLMRoleUser, Content: "ping"}},
	})
	require.NoError(t, err)
	require.Equal(t, "pong", resp.Text)
	require.Equal(t, "openai", resp.Provider)
	require.Equal(t, 9, resp.Usage.InputTokens)
	require.Equal(t, 4, resp.Usage.OutputTokens)

	require.Equal(t, "Bearer "+token, hdr.Get("Authorization"))
	msgs := body["messages"].([]any)
	require.Len(t, msgs, 2)
	require.Equal(t, "system", msgs[0].(map[string]any)["role"])
	require.Equal(t, system, msgs[0].(map[string]any)["content"])
	require.Equal(t, "user", msgs[1].(map[string]any)["role"])
}

func TestNewDeepSeekDefaults(t *testing.T) {
	var body map[string]any
	srv := openAIStub(t, &body, nil)
	defer srv.Close()

	client, err := llm.NewDeepSeek(testutil.RandomAPIKey(), llm.WithOpenAIBaseURL(srv.URL))
	require.NoError(t, err)

	_, err = client.Complete(context.Background(), ports.LLMRequest{
		Messages: []ports.LLMMessage{{Role: ports.LLMRoleUser, Content: "hi"}},
	})
	require.NoError(t, err)
	require.Equal(t, "deepseek-chat", body["model"])
}

func TestOpenAICompatibleExtraHeader(t *testing.T) {
	var hdr http.Header
	srv := openAIStub(t, nil, &hdr)
	defer srv.Close()

	accountID := testutil.RandomUUID().String()
	client, err := llm.NewOpenAICompatible("codex", srv.URL, testutil.RandomAPIKey(),
		llm.WithOpenAIExtraHeader("chatgpt-account-id", accountID))
	require.NoError(t, err)

	_, err = client.Complete(context.Background(), ports.LLMRequest{
		Messages: []ports.LLMMessage{{Role: ports.LLMRoleUser, Content: "x"}},
	})
	require.NoError(t, err)
	require.Equal(t, accountID, hdr.Get("Chatgpt-Account-Id"))
}

func TestOpenAICompatiblePropagatesAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"bad key"}`))
	}))
	defer srv.Close()

	client, err := llm.NewOpenAICompatible("deepseek", srv.URL, testutil.RandomAPIKey())
	require.NoError(t, err)

	_, err = client.Complete(context.Background(), ports.LLMRequest{
		Messages: []ports.LLMMessage{{Role: ports.LLMRoleUser, Content: "x"}},
	})
	require.ErrorContains(t, err, "bad key")
	require.ErrorContains(t, err, "401")
}
