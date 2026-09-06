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

	"github.com/ViitoJooj/foreman/internal/core/ports"
)

const (
	openAIBaseURL        = "https://api.openai.com/v1"
	openAIDefaultModel   = "gpt-4o"
	deepSeekBaseURL      = "https://api.deepseek.com"
	deepSeekDefaultModel = "deepseek-chat"
)

var _ ports.LLMClient = (*OpenAICompatible)(nil)

// OpenAICompatible is a ports.LLMClient for any provider that speaks the OpenAI
// chat-completions protocol: OpenAI/Codex (API key), DeepSeek, or a gateway.
type OpenAICompatible struct {
	provider     string
	baseURL      string
	token        string
	model        string
	http         *http.Client
	extraHeaders map[string]string
}

// OpenAIOption customises an OpenAICompatible client.
type OpenAIOption func(*OpenAICompatible)

// WithOpenAIBaseURL overrides the API base URL.
func WithOpenAIBaseURL(u string) OpenAIOption {
	return func(c *OpenAICompatible) { c.baseURL = strings.TrimRight(u, "/") }
}

// WithOpenAIModel sets the default model used when a request omits one.
func WithOpenAIModel(m string) OpenAIOption {
	return func(c *OpenAICompatible) { c.model = m }
}

// WithOpenAIHTTPClient supplies a custom HTTP client.
func WithOpenAIHTTPClient(h *http.Client) OpenAIOption {
	return func(c *OpenAICompatible) { c.http = h }
}

// WithOpenAIExtraHeader adds a header to every request (e.g. chatgpt-account-id
// for a Codex subscription token).
func WithOpenAIExtraHeader(key, value string) OpenAIOption {
	return func(c *OpenAICompatible) {
		if c.extraHeaders == nil {
			c.extraHeaders = map[string]string{}
		}
		c.extraHeaders[key] = value
	}
}

// NewOpenAICompatible builds a client for an arbitrary OpenAI-compatible endpoint.
// token is sent as a bearer credential (an API key or an OAuth access token).
func NewOpenAICompatible(provider, baseURL, token string, opts ...OpenAIOption) (*OpenAICompatible, error) {
	if token == "" {
		return nil, errors.New("llm: openai-compatible client requires a token")
	}
	c := &OpenAICompatible{
		provider: provider,
		baseURL:  strings.TrimRight(baseURL, "/"),
		token:    token,
		http:     &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.baseURL == "" {
		return nil, errors.New("llm: openai-compatible client requires a base url")
	}
	return c, nil
}

// NewOpenAI builds a client for the OpenAI API using an API key.
func NewOpenAI(apiKey string, opts ...OpenAIOption) (*OpenAICompatible, error) {
	base := []OpenAIOption{WithOpenAIBaseURL(openAIBaseURL), WithOpenAIModel(openAIDefaultModel)}
	return NewOpenAICompatible("openai", openAIBaseURL, apiKey, append(base, opts...)...)
}

// NewDeepSeek builds a client for the DeepSeek API using an API key.
func NewDeepSeek(apiKey string, opts ...OpenAIOption) (*OpenAICompatible, error) {
	base := []OpenAIOption{WithOpenAIBaseURL(deepSeekBaseURL), WithOpenAIModel(deepSeekDefaultModel)}
	return NewOpenAICompatible("deepseek", deepSeekBaseURL, apiKey, append(base, opts...)...)
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
}

type openAIResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

// Complete sends the request to the chat-completions endpoint.
func (c *OpenAICompatible) Complete(ctx context.Context, req ports.LLMRequest) (ports.LLMResponse, error) {
	model := req.Model
	if model == "" {
		model = c.model
	}

	messages := make([]openAIMessage, 0, len(req.Messages)+1)
	if req.System != "" {
		messages = append(messages, openAIMessage{Role: "system", Content: req.System})
	}
	for _, m := range req.Messages {
		messages = append(messages, openAIMessage{Role: string(m.Role), Content: m.Content})
	}

	payload, err := json.Marshal(openAIRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	})
	if err != nil {
		return ports.LLMResponse{}, fmt.Errorf("%s: encode request: %w", c.provider, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return ports.LLMResponse{}, fmt.Errorf("%s: build request: %w", c.provider, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.token)
	for k, v := range c.extraHeaders {
		httpReq.Header.Set(k, v)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return ports.LLMResponse{}, fmt.Errorf("%s: call: %w", c.provider, err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ports.LLMResponse{}, fmt.Errorf("%s: %s: %s", c.provider, resp.Status, strings.TrimSpace(string(data)))
	}

	var decoded openAIResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		return ports.LLMResponse{}, fmt.Errorf("%s: decode response: %w", c.provider, err)
	}

	var text string
	if len(decoded.Choices) > 0 {
		text = decoded.Choices[0].Message.Content
	}
	return ports.LLMResponse{
		Text:     text,
		Model:    decoded.Model,
		Provider: c.provider,
		Usage: ports.LLMUsage{
			InputTokens:  decoded.Usage.PromptTokens,
			OutputTokens: decoded.Usage.CompletionTokens,
		},
	}, nil
}
