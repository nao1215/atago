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

// Main is the CLI entry point. It returns the process exit code.
func Main(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return ExitConfig
	}

	cmd, rest := args[0], args[1:]
	switch cmd {
	case "run":
		return runCmd("atago run", rest, stdout, stderr)
	case "init":
		return initCmd(rest, stdout, stderr)
	case "record":
		return recordCmd(rest, stdout, stderr)
	case "explain":
		return explainCmd(rest, stdout, stderr)
	case "doc":
		return docCmd(rest, stdout, stderr)
	case "manifest":
		return manifestCmd(rest, stdout, stderr)
	case "list":
		return listCmd(rest, stdout, stderr)
	case "completion":
		return completionCmd(rest, stdout, stderr)
	case "snapshot":
		return snapshotCmd(rest, stdout, stderr)
	case "version", "-version", "--version":
		fmt.Fprintf(stdout, "atago %s\n", buildinfo.Get())
		return ExitOK
	case "help", "-h", "--help":
		usage(stdout)
		return ExitOK
	default:
		fmt.Fprintf(stderr, "atago: unknown command %q\n\n", cmd)
		usage(stderr)
		return ExitConfig
	}
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
