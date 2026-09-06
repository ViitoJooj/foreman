// Package browser implements ports.BrowserClient. Phase 1 will back this with
// Claude in Chrome and phase 2 with computer-use; until then NotImplemented is
// the wired placeholder so the rest of the system can depend on the port.
package browser

import (
	"context"
	"errors"

	"github.com/ViitoJooj/foreman/internal/core/ports"
)

// ErrNotImplemented is returned by NotImplemented.Research.
var ErrNotImplemented = errors.New("browser: research is not implemented in this build")

var _ ports.BrowserClient = NotImplemented{}

// NotImplemented is a ports.BrowserClient that always fails. It lets the
// orchestrator and Task Creator wire the port before a real driver exists.
type NotImplemented struct{}

// Research always returns ErrNotImplemented.
func (NotImplemented) Research(context.Context, string) (string, error) {
	return "", ErrNotImplemented
}
