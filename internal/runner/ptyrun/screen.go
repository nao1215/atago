package ptyrun

import (
	"strings"

	"github.com/hinshun/vt10x"
	"github.com/nao1215/atago/internal/spec"
)

// RenderScreen replays a pty transcript through a vt10x terminal emulator and
// returns the final rendered screen as plain text (#27): what the user
// actually SEES after every cursor move, overwrite, and erase — the signal a
// raw transcript scatters across redraws. Trailing whitespace is stripped per
// line and trailing blank lines are dropped; colors and attributes are out of
// scope in v1.
func RenderScreen(transcript []byte, p *spec.PTY) string {
	rows, cols := defaultRows, defaultCols
	if p.Rows > 0 {
		rows = p.Rows
	}
	if p.Cols > 0 {
		cols = p.Cols
	}
	term := vt10x.New(vt10x.WithSize(cols, rows))
	_, _ = term.Write(transcript)

	lines := strings.Split(term.String(), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	// Drop trailing blank rows: a 24-row screen showing two lines snapshots
	// as two lines, not twenty-four.
	end := len(lines)
	for end > 0 && lines[end-1] == "" {
		end--
	}
	return strings.Join(lines[:end], "\n")
}
