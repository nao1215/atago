package report

import (
	"io"
	"os"
)

// ANSI color codes used for console output.
const (
	cReset  = "\x1b[0m"
	cBold   = "\x1b[1m"
	cGreen  = "\x1b[32m"
	cRed    = "\x1b[31m"
	cYellow = "\x1b[33m"
	cDim    = "\x1b[2m"
)

// isTTY reports whether w is a terminal, so color is only emitted when a human
// is watching (and never when output is piped or captured into reports). The
// standard NO_COLOR convention (and thus `--ci`, which sets it) forces color
// off regardless.
func isTTY(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// colorize wraps s in code when on, otherwise returns s unchanged.
func colorize(on bool, code, s string) string {
	if !on || code == "" {
		return s
	}
	return code + s + cReset
}
