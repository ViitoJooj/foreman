package ports

import "context"

// BrowserClient drives a browser session for research tasks. Phase 1 is backed
// by Claude in Chrome, phase 2 by computer-use. The surface is intentionally
// minimal and will grow as research needs are pinned down.
type BrowserClient interface {
	// Research runs a browser-driven investigation described by prompt and
	// returns a free-form findings report.
	Research(ctx context.Context, prompt string) (findings string, err error)
}
