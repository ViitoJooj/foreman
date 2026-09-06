package llm_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/adapters/driven/llm"
	"github.com/ViitoJooj/foreman/internal/core/ports"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

type fakeClient struct {
	name string
	err  error
}

func (f fakeClient) Complete(context.Context, ports.LLMRequest) (ports.LLMResponse, error) {
	if f.err != nil {
		return ports.LLMResponse{}, f.err
	}
	return ports.LLMResponse{Text: "ok", Provider: f.name}, nil
}

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestRouterFallsThroughToWorkingProvider(t *testing.T) {
	r := llm.NewRouter(quietLogger(),
		fakeClient{name: "a", err: errors.New("down")},
		fakeClient{name: "b"},
	)

	resp, err := r.Complete(context.Background(), ports.LLMRequest{})
	require.NoError(t, err)
	require.Equal(t, "b", resp.Provider)
}

func TestRouterReturnsUnavailableWhenAllFail(t *testing.T) {
	r := llm.NewRouter(quietLogger(),
		fakeClient{name: "a", err: errors.New("down")},
		fakeClient{name: "b", err: errors.New("also down")},
	)

	_, err := r.Complete(context.Background(), ports.LLMRequest{})
	require.ErrorIs(t, err, ports.ErrLLMUnavailable)
	require.ErrorContains(t, err, "also down")
}

func TestRouterEmptyIsUnavailable(t *testing.T) {
	_, err := llm.NewRouter(quietLogger()).Complete(context.Background(), ports.LLMRequest{})
	require.ErrorIs(t, err, ports.ErrLLMUnavailable)
}

func TestBuildRouterRequiresACredential(t *testing.T) {
	_, err := llm.BuildRouter(llm.ProviderConfig{}, quietLogger())
	require.Error(t, err)
}

func TestBuildRouterHonoursOrderAndAvailability(t *testing.T) {
	cfg := llm.ProviderConfig{
		DeepSeekAPIKey: testutil.RandomAPIKey(),
		OpenAIAPIKey:   testutil.RandomAPIKey(),
		Order:          []string{"deepseek", "anthropic", "openai"},
	}
	r, err := llm.BuildRouter(cfg, quietLogger())
	require.NoError(t, err)
	// anthropic has no credential, so only deepseek + openai are wired.
	require.Equal(t, 2, r.Providers())
}

func TestBuildRouterRejectsUnknownProvider(t *testing.T) {
	_, err := llm.BuildRouter(llm.ProviderConfig{
		OpenAIAPIKey: testutil.RandomAPIKey(),
		Order:        []string{"gemini"},
	}, quietLogger())
	require.ErrorContains(t, err, "unknown provider")
}

func TestLoadProviderConfigFromEnv(t *testing.T) {
	key := testutil.RandomAPIKey()
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("ANTHROPIC_OAUTH_TOKEN", "")
	t.Setenv("CLAUDE_CODE_OAUTH_TOKEN", key)
	t.Setenv("DEEPSEEK_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("LLM_PROVIDER_ORDER", "anthropic, deepseek")
	t.Setenv("LLM_MODEL_ANTHROPIC", "claude-opus-5")

	cfg := llm.LoadProviderConfigFromEnv()
	require.Equal(t, key, cfg.AnthropicOAuthToken, "falls back to CLAUDE_CODE_OAUTH_TOKEN")
	require.Equal(t, []string{"anthropic", "deepseek"}, cfg.Order)
	require.Equal(t, "claude-opus-5", cfg.Models["anthropic"])
}
