// Package cli implements atago's command-line interface: subcommand dispatch
// and the mapping from results to exit codes.
package cli

import (
	"fmt"
	"io"

	"github.com/nao1215/atago/internal/buildinfo"
)

// Exit codes. These are part of the stable user-facing contract.
const (
	ExitOK       = 0 // all scenarios passed
	ExitFailures = 1 // one or more scenarios failed
	ExitParse    = 2 // spec parse error
	ExitConfig   = 3 // configuration error
	ExitExec     = 4 // execution error
	ExitInternal = 5 // internal error
	ExitSecurity = 6 // security policy violation
)

// subcommand pairs a subcommand name with its handler. The dispatch table
// (see dispatchTable) is the single source of truth for which subcommands
// exist: Main dispatches through it and Subcommands reports its names, so the
// user-facing command inventory and its documentation cannot silently drift.
type subcommand struct {
	name string
	run  func(rest []string, stdout, stderr io.Writer) int
}

// dispatchTable lists every atago subcommand in the order shown in usage. Both
// Main and Subcommands read from it; do not add a command to one without the
// other.
func dispatchTable() []subcommand {
	return []subcommand{
		{"run", func(rest []string, stdout, stderr io.Writer) int { return runCmd("atago run", rest, stdout, stderr) }},
		{"init", initCmd},
		{"record", recordCmd},
		{"list", listCmd},
		{"explain", explainCmd},
		{"doc", docCmd},
		{"manifest", manifestCmd},
		{"completion", completionCmd},
		{"snapshot", snapshotCmd},
		{"version", versionCmd},
		{"help", helpCmd},
	}
}

// Subcommands returns the atago subcommand names in dispatch order. It is
// derived from the same table Main dispatches through, so documentation-drift
// tests can check a doc's advertised subcommand list against the real inventory
// without maintaining a second hand-written list.
func Subcommands() []string {
	table := dispatchTable()
	names := make([]string, len(table))
	for i, sc := range table {
		names[i] = sc.name
	}
	return names
}

// Main is the CLI entry point. It returns the process exit code.
func Main(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return ExitConfig
	}

	cmd, rest := args[0], args[1:]
	// Flag-style aliases for the meta subcommands.
	switch cmd {
	case "-version", "--version":
		cmd = "version"
	case "-h", "--help":
		cmd = "help"
	}
	for _, sc := range dispatchTable() {
		if sc.name == cmd {
			return sc.run(rest, stdout, stderr)
		}
	}
	fmt.Fprintf(stderr, "atago: unknown command %q\n\n", cmd)
	usage(stderr)
	return ExitConfig
}

// versionCmd prints the atago version.
func versionCmd(_ []string, stdout, _ io.Writer) int {
	fmt.Fprintf(stdout, "atago %s\n", buildinfo.Get())
	return ExitOK
}

// helpCmd prints top-level usage to stdout so it can be piped.
func helpCmd(_ []string, stdout, _ io.Writer) int {
	usage(stdout)
	return ExitOK
}

func usage(w io.Writer) {
	fmt.Fprint(w, `atago — black-box behavior spec runner

atago runs the *.atago.yaml specs in a directory (or the files you name):
each spec declares commands, HTTP/DB/SSH/gRPC/browser interactions, and
assertions on what a user observes — exit codes, output, files, responses.

Usage:
  atago <command> [arguments]

Commands:
  run         Run spec files and assert behavior
  init        Scaffold a starter spec file (--template browser|cli|db|grpc|http|mock|services|ssh)
  record      Generate a spec skeleton from one observed command run (record -- <cmd>)
  list        List suites, scenarios, tags, and generated artifacts (--json)
  explain     Describe what a spec does without running it
  doc         Generate Markdown documentation from specs
  manifest    Emit a stable machine-readable JSON summary of specs
  completion  Generate a shell completion script (bash|zsh|fish|powershell)
  snapshot    Manage snapshots (snapshot update <paths>)
  version     Print the atago version
  help        Show this help

Run "atago <command> --help" for a command's options.
Start with "atago init" — it writes a runnable example spec.
Docs and the spec-file reference: https://github.com/nao1215/atago
`)
}

// wantsHelp reports whether the sole argument is a help flag, for subcommands
// that parse their arguments by hand instead of through a FlagSet.
func wantsHelp(args []string) bool {
	if len(args) != 1 {
		return false
	}
	for _, a := range args {
		switch a {
		case "-h", "-help", "--help":
			return true
		}
	}
	return false
}

// snapshotCmd implements `atago snapshot <subcommand>`.
func snapshotCmd(args []string, stdout, stderr io.Writer) int {
	if wantsHelp(args) {
		fmt.Fprintln(stdout, "Usage: atago snapshot update <path | dir>...")
		fmt.Fprintln(stdout, "  (records or refreshes the golden files that `snapshot` matchers compare against)")
		return ExitOK
	}
	if len(args) == 0 || args[0] != "update" {
		fmt.Fprintln(stderr, "Usage: atago snapshot update <path | dir>...")
		return ExitConfig
	}
	// `snapshot update` is `run` with snapshots written instead of compared.
	return runCmd("atago snapshot update", append([]string{"--update-snapshots"}, args[1:]...), stdout, stderr)
}
