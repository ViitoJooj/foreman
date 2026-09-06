package llm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/ViitoJooj/foreman/internal/core/ports"
)

var _ ports.LLMClient = (*Router)(nil)

// Router is a ports.LLMClient that tries an ordered list of providers, falling
// through to the next one on error.
type Router struct {
	providers []ports.LLMClient
	logger    *slog.Logger
}

// NewRouter builds a Router over the given providers, tried in order.
func NewRouter(logger *slog.Logger, providers ...ports.LLMClient) *Router {
	if logger == nil {
		logger = slog.Default()
	}
	return &Router{providers: providers, logger: logger}
}

// Providers reports how many providers the router will try.
func (r *Router) Providers() int { return len(r.providers) }

// Complete tries each provider until one succeeds. If every provider fails it
// returns ports.ErrLLMUnavailable wrapping the last error.
func (r *Router) Complete(ctx context.Context, req ports.LLMRequest) (ports.LLMResponse, error) {
	if len(r.providers) == 0 {
		return ports.LLMResponse{}, ports.ErrLLMUnavailable
	}

	var lastErr error
	for _, p := range r.providers {
		resp, err := p.Complete(ctx, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		r.logger.Warn("llm provider failed; trying the next one", "err", err)
	}
	return ports.LLMResponse{}, fmt.Errorf("%w: %w", ports.ErrLLMUnavailable, lastErr)
}

// ProviderConfig holds the credentials and preferences for building a Router.
type ProviderConfig struct {
	AnthropicAPIKey     string
	AnthropicOAuthToken string
	OpenAIAPIKey        string
	DeepSeekAPIKey      string
	Order               []string          // fallback order by provider name
	Models              map[string]string // provider name -> model id override
}

var defaultProviderOrder = []string{"anthropic", "openai", "deepseek"}

// LoadProviderConfigFromEnv reads the LLM provider configuration from the
// environment. It recognises ANTHROPIC_API_KEY, ANTHROPIC_OAUTH_TOKEN
// (or CLAUDE_CODE_OAUTH_TOKEN), OPENAI_API_KEY, DEEPSEEK_API_KEY,
// LLM_PROVIDER_ORDER and LLM_MODEL_<PROVIDER>.
func LoadProviderConfigFromEnv() ProviderConfig {
	cfg := ProviderConfig{
		AnthropicAPIKey:     os.Getenv("ANTHROPIC_API_KEY"),
		AnthropicOAuthToken: firstNonEmpty(os.Getenv("ANTHROPIC_OAUTH_TOKEN"), os.Getenv("CLAUDE_CODE_OAUTH_TOKEN")),
		OpenAIAPIKey:        os.Getenv("OPENAI_API_KEY"),
		DeepSeekAPIKey:      os.Getenv("DEEPSEEK_API_KEY"),
		Models:              map[string]string{},
	}
	if raw := os.Getenv("LLM_PROVIDER_ORDER"); raw != "" {
		for _, name := range strings.Split(raw, ",") {
			if n := strings.TrimSpace(name); n != "" {
				cfg.Order = append(cfg.Order, n)
			}
		}
	}
	for _, name := range defaultProviderOrder {
		if m := os.Getenv("LLM_MODEL_" + strings.ToUpper(name)); m != "" {
			cfg.Models[name] = m
		}
	}
	return cfg
}

// BuildRouter constructs a Router from cfg, including only the providers that
// have credentials, in the configured order. It fails if none are configured.
func BuildRouter(cfg ProviderConfig, logger *slog.Logger) (*Router, error) {
	order := cfg.Order
	if len(order) == 0 {
		order = defaultProviderOrder
	}

	var providers []ports.LLMClient
	for _, name := range order {
		client, err := buildProvider(name, cfg)
		if err != nil {
			return nil, err
		}
		if client != nil {
			providers = append(providers, client)
		}
	}
	if len(providers) == 0 {
		return nil, errors.New("llm: no provider configured (set ANTHROPIC_API_KEY, OPENAI_API_KEY or DEEPSEEK_API_KEY)")
	}
	return NewRouter(logger, providers...), nil
}

func buildProvider(name string, cfg ProviderConfig) (ports.LLMClient, error) {
	switch strings.ToLower(name) {
	case "anthropic", "claude":
		if cfg.AnthropicAPIKey == "" && cfg.AnthropicOAuthToken == "" {
			return nil, nil
		}
		opts := modelOpt(cfg.Models["anthropic"], WithAnthropicModel)
		return NewAnthropic(AnthropicAuth{APIKey: cfg.AnthropicAPIKey, OAuthToken: cfg.AnthropicOAuthToken}, opts...)
	case "openai", "codex":
		if cfg.OpenAIAPIKey == "" {
			return nil, nil
		}
		return NewOpenAI(cfg.OpenAIAPIKey, modelOpt(cfg.Models["openai"], WithOpenAIModel)...)
	case "deepseek":
		if cfg.DeepSeekAPIKey == "" {
			return nil, nil
		}
		return NewDeepSeek(cfg.DeepSeekAPIKey, modelOpt(cfg.Models["deepseek"], WithOpenAIModel)...)
	default:
		return nil, fmt.Errorf("llm: unknown provider %q", name)
	}
}

func modelOpt[T any](model string, with func(string) T) []T {
	if model == "" {
		return nil
	}
	return []T{with(model)}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
