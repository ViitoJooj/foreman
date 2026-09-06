package orchestrator

import "sync"

// Budget is the token/spend guard. It is a hook: nothing records spend until the
// LLM layer lands, so with a non-positive cap it always allows work.
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

// SpentUSD returns the running spend.
func (b *Budget) SpentUSD() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.spentUSD
}
