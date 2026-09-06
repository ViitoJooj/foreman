package ports

import (
	"context"
	"errors"
)

// ErrLLMUnavailable is returned when no configured provider could serve a request.
var ErrLLMUnavailable = errors.New("no llm provider available")

// LLMRole is the author of a chat message.
type LLMRole string

// The chat message authors.
const (
	LLMRoleUser      LLMRole = "user"
	LLMRoleAssistant LLMRole = "assistant"
)

// LLMMessage is one turn in a chat exchange.
type LLMMessage struct {
	Role    LLMRole
	Content string
}

// LLMRequest is a provider-agnostic completion request.
type LLMRequest struct {
	Model       string // provider model id; empty lets the adapter pick its default
	System      string
	Messages    []LLMMessage
	MaxTokens   int
	Temperature float64
}

// LLMUsage reports the tokens a completion consumed, for the budget guard.
type LLMUsage struct {
	InputTokens  int
	OutputTokens int
}

// LLMResponse is a provider-agnostic completion result.
type LLMResponse struct {
	Text     string
	Model    string
	Provider string
	Usage    LLMUsage
}

// LLMClient is a single chat-completion provider (Claude, OpenAI/Codex, DeepSeek,
// or a router over several of them).
type LLMClient interface {
	// Complete returns a single assistant message for the request.
	Complete(ctx context.Context, req LLMRequest) (LLMResponse, error)
}
