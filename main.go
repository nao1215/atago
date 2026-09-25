// Command atago is a YAML-based black-box behavior spec runner for CLIs, APIs,
// and generated artifacts. See the README for the format and CLI contract, and
// schema/atago.schema.json for the machine-readable spec file schema.
package main

import (
	"os"
	"runtime/debug"

	"github.com/nao1215/atago/internal/cli"
)

// gcPercent is the garbage collector's target when GOGC is not set. atago is a
// short-lived process whose live heap stays small (a few MB for a suite of a
// thousand scenarios) while it allocates steadily, so the default of 100 spent
// a third of its CPU time collecting. 400 lets the heap grow to five times what
// is live before a collection; past 800 nothing more is saved.
const gcPercent = 400

// memoryLimit bounds the heap that gcPercent lets grow, so a run that holds a
// very large capture collects as often as it did before instead of growing to
// five times it.
const memoryLimit = 2 << 30

func main() {
	tuneGC(os.Getenv)
	os.Exit(cli.Main(os.Args[1:], os.Stdout, os.Stderr))
}

// tuneGC applies atago's collector settings, leaving each one to the user who
// set GOGC or GOMEMLIMIT.
func tuneGC(getenv func(string) string) {
	if getenv("GOGC") == "" {
		debug.SetGCPercent(gcPercent)
	}
	if getenv("GOMEMLIMIT") == "" {
		debug.SetMemoryLimit(memoryLimit)
	}
}
