// Package cli implements atago's command-line interface: subcommand dispatch
// and the mapping from results to exit codes.
package cli

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/atago/internal/buildinfo"
	"github.com/nao1215/atago/internal/diag"
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
		fmt.Fprintf(stderr, "atago: %s\n\n", diag.UnknownCommand.Annotate("no subcommand given"))
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
	fmt.Fprintf(stderr, "atago: %s\n\n", diag.UnknownCommand.Annotate(fmt.Sprintf("unknown command %q", cmd)))
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
Documentation: https://nao1215.github.io/atago/
GitHub Sponsors: https://github.com/sponsors/nao1215
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

// parseFlagsAnywhere parses args with fs and returns the operands, accepting a
// flag written after them: `atago run specs/ --report json` means what `atago
// run --report json specs/` means.
//
// The flag package stops at the first non-flag argument, so a trailing flag
// silently became a spec path and the run died with `cannot access "--report"`
// — an error about a file nobody typed, for the invocation order most people
// reach for first (the paths are the subject, the flags are an afterthought).
// Rather than reordering argv by guessing which flags take a value, this
// re-enters Parse after each operand, which asks the FlagSet itself.
//
// A `--` terminator still ends flag parsing for good: everything after it is an
// operand even when it looks like a flag, which is the only way to name a file
// whose name starts with a dash. `atago record` deliberately does NOT use this —
// there the trailing arguments are another program's command line, and its
// flags belong to it.
func parseFlagsAnywhere(fs *flag.FlagSet, args []string) ([]string, error) {
	// The FlagSet's own output is captured rather than left to reach the
	// terminal directly, so the caller can put the diagnostic code in front of
	// the message instead of after the usage block the failure triggers. What
	// the user sees is unchanged apart from that ordering: reportFlagError
	// replays everything the FlagSet wrote.
	var captured bytes.Buffer
	prev := fs.Output()
	fs.SetOutput(&captured)
	defer fs.SetOutput(prev)

	operands, err := parseOperands(fs, args)
	if err != nil {
		return nil, &flagError{err: err, output: captured.String()}
	}
	return operands, nil
}

// flagError carries a flag-parsing failure together with everything the
// FlagSet printed about it, so the caller decides the order and destination.
type flagError struct {
	err    error
	output string
}

func (e *flagError) Error() string { return e.err.Error() }
func (e *flagError) Unwrap() error { return e.err }

// reportFlagError prints a flag-parsing failure: the coded message first, then
// whatever usage the FlagSet produced. The FlagSet writes its message as a line
// of its own before the usage, so dropping that exact line is what keeps the
// message from appearing twice.
func reportFlagError(label string, err error, stderr io.Writer) {
	var fe *flagError
	if !errors.As(err, &fe) {
		fmt.Fprintf(stderr, "%s: %s\n", label, diag.UnknownOption.Annotate(err.Error()))
		return
	}
	fmt.Fprintf(stderr, "%s: %s\n", label, diag.UnknownOption.Annotate(fe.err.Error()))
	fmt.Fprint(stderr, strings.TrimPrefix(fe.output, fe.err.Error()+"\n"))
}

// replayFlagOutput writes what the FlagSet printed to w unchanged. It is the
// help path, where the FlagSet's output is already what the user asked for.
func replayFlagOutput(err error, w io.Writer) {
	var fe *flagError
	if errors.As(err, &fe) {
		fmt.Fprint(w, fe.output)
	}
}

// parseFlagsStrict parses args with fs, capturing the FlagSet's output the way
// parseFlagsAnywhere does. `atago record` cannot use parseFlagsAnywhere — its
// trailing arguments are another program's command line — but a bad flag
// should read the same there as everywhere else.
func parseFlagsStrict(fs *flag.FlagSet, args []string) error {
	var captured bytes.Buffer
	prev := fs.Output()
	fs.SetOutput(&captured)
	defer fs.SetOutput(prev)

	if err := fs.Parse(args); err != nil {
		return &flagError{err: err, output: captured.String()}
	}
	return nil
}

// parseOperands is parseFlagsAnywhere's loop, split out so the output capture
// around it stays readable.
func parseOperands(fs *flag.FlagSet, args []string) ([]string, error) {
	var operands []string
	for {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		rest := fs.Args()
		if len(rest) == 0 {
			return operands, nil
		}
		// Parse stopped either at an operand or at an explicit `--`. The
		// terminator is consumed, so it is the last token before what is left;
		// seeing it there means the caller asked for the rest to stay literal.
		if consumed := len(args) - len(rest); consumed > 0 && args[consumed-1] == "--" {
			return append(operands, rest...), nil
		}
		operands = append(operands, rest[0])
		args = rest[1:]
	}
}

// snapshotCmd implements `atago snapshot <subcommand>`.
func snapshotCmd(args []string, stdout, stderr io.Writer) int {
	if wantsHelp(args) {
		fmt.Fprintln(stdout, "Usage: atago snapshot update <path | dir>...")
		fmt.Fprintln(stdout, "  (records or refreshes the golden files that `snapshot` matchers compare against)")
		return ExitOK
	}
	if len(args) == 0 || args[0] != "update" {
		fmt.Fprintf(stderr, "atago snapshot: %s\n", diag.BadUsage.Annotate("update is the only subcommand"))
		fmt.Fprintln(stderr, "Usage: atago snapshot update <path | dir>...")
		return ExitConfig
	}
	// `snapshot update` is `run` with snapshots written instead of compared.
	return runCmd("atago snapshot update", append([]string{"--update-snapshots"}, args[1:]...), stdout, stderr)
}
