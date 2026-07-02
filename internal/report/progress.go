package report

import (
	"fmt"
	"io"
	"sync"

	"github.com/nao1215/atago/internal/engine"
)

// Progress streams a live one-character-per-scenario indicator, RSpec/Bats
// style: a dot for every pass, so a run visibly "zips along" as it executes.
//
//	.  passed    F  failed    E  errored    s  skipped
type Progress struct {
	mu    sync.Mutex
	w     io.Writer
	color bool
	count int
}

// NewProgress returns a Progress that writes to w. Color is enabled only when w
// is a terminal.
func NewProgress(w io.Writer) *Progress {
	return &Progress{w: w, color: isTTY(w)}
}

// Scenario prints the single marker for one finished scenario. It is safe to use
// directly as engine.Engine.OnScenario, including from concurrent suites.
func (p *Progress) Scenario(res engine.ScenarioResult) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.count++
	fmt.Fprint(p.w, p.marker(res.Status))
}

// Done terminates the marker line if any markers were printed.
func (p *Progress) Done() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.count > 0 {
		fmt.Fprintln(p.w)
	}
}

func (p *Progress) marker(status engine.Status) string {
	switch status {
	case engine.StatusPassed:
		return colorize(p.color, cGreen, ".")
	case engine.StatusFailed:
		return colorize(p.color, cRed, "F")
	case engine.StatusError:
		return colorize(p.color, cBold+cRed, "E")
	case engine.StatusSkipped:
		return colorize(p.color, cYellow, "s")
	default:
		return "?"
	}
}
