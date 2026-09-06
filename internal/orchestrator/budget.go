package orchestrator

import "sync"

// defaultUSDPerMTok is a blended cost estimate ($/million tokens) used to turn
// token usage into spend. It is deliberately rough; a per-model pricing table
// can replace it later.
const defaultUSDPerMTok = 6.0

// Budget is the token/spend guard. With a non-positive cap it always allows work.
type Budget struct {
	mu         sync.Mutex
	hardCapUSD float64
	spentUSD   float64
}

// NewBudget returns a Budget with the given hard cap in USD. A cap <= 0 disables
// the guard.
func NewBudget(hardCapUSD float64) *Budget {
	return &Budget{hardCapUSD: hardCapUSD}
}

// Allow reports whether more work may be started under the current spend.
func (b *Budget) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.hardCapUSD <= 0 {
		return true
	}
	return b.spentUSD < b.hardCapUSD
}

// Record adds usd to the running spend.
func (b *Budget) Record(usd float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.spentUSD += usd
}

// RecordTokens converts a token count into spend using the blended rate and adds
// it to the running total.
func (b *Budget) RecordTokens(inputTokens, outputTokens int) {
	if inputTokens <= 0 && outputTokens <= 0 {
		return
	}
	b.Record(float64(inputTokens+outputTokens) / 1_000_000.0 * defaultUSDPerMTok)
}

// SpentUSD returns the running spend.
func (b *Budget) SpentUSD() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.spentUSD
}
