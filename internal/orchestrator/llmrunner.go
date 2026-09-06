package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/ports"
)

const (
	llmRunnerMaxTokens   = 800
	llmRunnerTemperature = 0.2
)

// LLMRunner is an AgentRunner that asks an LLM to decide the outcome of a
// pipeline step and to write the channel message, following the inter-agent
// protocol. It does not yet perform real code / git / sandbox work.
type LLMRunner struct {
	role         domain.AgentRole
	systemPrompt string
	client       ports.LLMClient
	model        string
	logger       *slog.Logger
}

// LLMRunnerOption customises an LLMRunner.
type LLMRunnerOption func(*LLMRunner)

// WithLLMRunnerModel pins the model id for this runner.
func WithLLMRunnerModel(model string) LLMRunnerOption {
	return func(r *LLMRunner) { r.model = model }
}

// WithLLMRunnerLogger sets the runner's logger.
func WithLLMRunnerLogger(l *slog.Logger) LLMRunnerOption {
	return func(r *LLMRunner) {
		if l != nil {
			r.logger = l
		}
	}
}

// NewLLMRunner builds an LLM-backed runner for a role.
func NewLLMRunner(role domain.AgentRole, systemPrompt string, client ports.LLMClient, opts ...LLMRunnerOption) *LLMRunner {
	r := &LLMRunner{role: role, systemPrompt: systemPrompt, client: client, logger: slog.Default()}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Role reports which pipeline role this runner fills.
func (r *LLMRunner) Role() domain.AgentRole { return r.role }

type llmDecision struct {
	Outcome string `json:"outcome"`
	Note    string `json:"note"`
}

// Step asks the LLM for a decision constrained to the outcomes legal in the
// task's current state.
func (r *LLMRunner) Step(ctx context.Context, task domain.Task) (StepResult, error) {
	allowed := AllowedOutcomes(task.State)
	if len(allowed) == 0 {
		return StepResult{}, fmt.Errorf("llmrunner: no action for state %q", task.State)
	}

	resp, err := r.client.Complete(ctx, ports.LLMRequest{
		Model:       r.model,
		System:      r.systemPrompt,
		Messages:    []ports.LLMMessage{{Role: ports.LLMRoleUser, Content: buildStepPrompt(task, allowed)}},
		MaxTokens:   llmRunnerMaxTokens,
		Temperature: llmRunnerTemperature,
	})
	if err != nil {
		return StepResult{}, fmt.Errorf("llmrunner: %w", err)
	}

	decision, err := parseDecision(resp.Text)
	if err != nil {
		return StepResult{}, fmt.Errorf("llmrunner: %w", err)
	}

	outcome := Outcome(strings.TrimSpace(decision.Outcome))
	if !outcomeAllowed(outcome, allowed) {
		return StepResult{}, fmt.Errorf("llmrunner: %s returned outcome %q, not allowed in state %q", r.role, decision.Outcome, task.State)
	}

	return StepResult{Outcome: outcome, Note: strings.TrimSpace(decision.Note), Usage: resp.Usage}, nil
}

func buildStepPrompt(task domain.Task, allowed []Outcome) string {
	branch := task.Branch
	if branch == "" {
		branch = "(nenhuma ainda)"
	}
	options := make([]string, len(allowed))
	for i, o := range allowed {
		options[i] = string(o)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Task atual:\n")
	fmt.Fprintf(&b, "- id: %s\n", task.ID)
	fmt.Fprintf(&b, "- titulo: %s\n", task.Title)
	fmt.Fprintf(&b, "- descricao: %s\n", task.Description)
	fmt.Fprintf(&b, "- estado: %s\n", task.State)
	fmt.Fprintf(&b, "- risco: %s\n", task.Risk)
	fmt.Fprintf(&b, "- branch: %s\n", branch)
	fmt.Fprintf(&b, "- tentativas: %d/%d\n\n", task.Retries, task.MaxRetries)
	fmt.Fprintf(&b, "Decida o resultado desta etapa. Responda APENAS com um objeto JSON, sem cercas de codigo:\n")
	fmt.Fprintf(&b, `{"outcome": "<um de: %s>", "note": "<mensagem curta no seu tom para o canal>"}`, strings.Join(options, ", "))
	return b.String()
}

// parseDecision extracts the first {...} block from the model output and decodes it.
func parseDecision(text string) (llmDecision, error) {
	start := strings.IndexByte(text, '{')
	end := strings.LastIndexByte(text, '}')
	if start < 0 || end <= start {
		return llmDecision{}, fmt.Errorf("no json object in model output")
	}
	var d llmDecision
	if err := json.Unmarshal([]byte(text[start:end+1]), &d); err != nil {
		return llmDecision{}, fmt.Errorf("decode decision: %w", err)
	}
	if d.Outcome == "" {
		return llmDecision{}, fmt.Errorf("decision has no outcome")
	}
	return d, nil
}

func outcomeAllowed(o Outcome, allowed []Outcome) bool {
	for _, a := range allowed {
		if a == o {
			return true
		}
	}
	return false
}
