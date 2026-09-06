// Package llm implements the ports.LLMClient port for several providers. Each
// adapter is a thin HTTP client; auth is either an API key or a bearer/OAuth
// token supplied by the caller.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ViitoJooj/foreman/internal/core/ports"
)

const (
	anthropicDefaultBaseURL = "https://api.anthropic.com"
	anthropicVersion        = "2023-06-01"
	anthropicOAuthBeta      = "oauth-2025-04-20"
	anthropicDefaultModel   = "claude-sonnet-5"
	defaultMaxTokens        = 1024
	defaultTimeout          = 120 * time.Second
)

var _ ports.LLMClient = (*Anthropic)(nil)

// AnthropicAuth selects how the adapter authenticates. Exactly one field must be
// set: APIKey for a standard key, OAuthToken for a Claude Code / subscription
// token (sent as a bearer with the oauth beta header).
type AnthropicAuth struct {
	APIKey     string
	OAuthToken string
}

// Anthropic is the ports.LLMClient backed by the Claude Messages API.
type Anthropic struct {
	baseURL string
	auth    AnthropicAuth
	model   string
	http    *http.Client
}

// AnthropicOption customises an Anthropic client.
type AnthropicOption func(*Anthropic)

// WithAnthropicBaseURL overrides the API base URL (useful for a gateway or tests).
func WithAnthropicBaseURL(u string) AnthropicOption {
	return func(a *Anthropic) { a.baseURL = strings.TrimRight(u, "/") }
}

// WithAnthropicModel sets the default model used when a request omits one.
func WithAnthropicModel(m string) AnthropicOption {
	return func(a *Anthropic) { a.model = m }
}

// WithAnthropicHTTPClient supplies a custom HTTP client.
func WithAnthropicHTTPClient(c *http.Client) AnthropicOption {
	return func(a *Anthropic) { a.http = c }
}

// NewAnthropic builds a Claude client. It fails if no credential is supplied.
func NewAnthropic(auth AnthropicAuth, opts ...AnthropicOption) (*Anthropic, error) {
	if auth.APIKey == "" && auth.OAuthToken == "" {
		return nil, errors.New("llm: anthropic requires an api key or an oauth token")
	}
	a := &Anthropic{
		baseURL: anthropicDefaultBaseURL,
		auth:    auth,
		model:   anthropicDefaultModel,
		http:    &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(a)
	}
	return a, nil
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	System      string             `json:"system,omitempty"`
	Temperature float64            `json:"temperature,omitempty"`
	Messages    []anthropicMessage `json:"messages"`
}

type anthropicResponse struct {
	Model   string `json:"model"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// Complete sends the request to the Messages API and returns the assistant text.
func (a *Anthropic) Complete(ctx context.Context, req ports.LLMRequest) (ports.LLMResponse, error) {
	model := req.Model
	if model == "" {
		model = a.model
	}

	payload, err := json.Marshal(anthropicRequest{
		Model:       model,
		MaxTokens:   maxTokensOrDefault(req.MaxTokens),
		System:      req.System,
		Temperature: req.Temperature,
		Messages:    toAnthropicMessages(req.Messages),
	})
	if err != nil {
		return ports.LLMResponse{}, fmt.Errorf("anthropic: encode request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/v1/messages", bytes.NewReader(payload))
	if err != nil {
		return ports.LLMResponse{}, fmt.Errorf("anthropic: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Anthropic-Version", anthropicVersion)
	if a.auth.OAuthToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+a.auth.OAuthToken)
		httpReq.Header.Set("Anthropic-Beta", anthropicOAuthBeta)
	} else {
		httpReq.Header.Set("X-Api-Key", a.auth.APIKey)
	}

	resp, err := a.http.Do(httpReq)
	if err != nil {
		return ports.LLMResponse{}, fmt.Errorf("anthropic: call: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ports.LLMResponse{}, fmt.Errorf("anthropic: %s: %s", resp.Status, strings.TrimSpace(string(data)))
	}

	var decoded anthropicResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		return ports.LLMResponse{}, fmt.Errorf("anthropic: decode response: %w", err)
	}

	return ports.LLMResponse{
		Text:     firstText(decoded),
		Model:    decoded.Model,
		Provider: "anthropic",
		Usage: ports.LLMUsage{
			InputTokens:  decoded.Usage.InputTokens,
			OutputTokens: decoded.Usage.OutputTokens,
		},
	}, nil
}

func toAnthropicMessages(in []ports.LLMMessage) []anthropicMessage {
	out := make([]anthropicMessage, 0, len(in))
	for _, m := range in {
		out = append(out, anthropicMessage{Role: string(m.Role), Content: m.Content})
	}
	return out
}

func firstText(r anthropicResponse) string {
	for _, c := range r.Content {
		if c.Type == "text" {
			return c.Text
		}
	}
	return ""
}

func maxTokensOrDefault(n int) int {
	if n <= 0 {
		return defaultMaxTokens
	}
	return n
}
