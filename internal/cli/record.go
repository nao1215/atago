package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nao1215/atago/internal/record"
	runnercmd "github.com/nao1215/atago/internal/runner/cmd"
	"github.com/nao1215/atago/internal/snapshot"
	"github.com/nao1215/atago/internal/spec"
)

// recordCmd implements `atago record -- <command> [args...]` (#30): run the
// command once in a fresh scratch dir (never the user's cwd), observe its
// exit code, streams, and created files, and emit a ready-to-edit spec
// skeleton with conservative matchers.
func recordCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("atago record", flag.ContinueOnError)
	fs.SetOutput(stderr)
	out := fs.String("out", "", "write the generated spec to this file (default: stdout)")
	force := fs.Bool("force", false, "overwrite --out if it already exists")
	shell := fs.Bool("shell", false, "record the command line verbatim with shell: true")
	snap := fs.Bool("snapshot", false, "assert stdout against a snapshot golden (requires --out; the golden is written next to it)")
	fs.Usage = func() {
		fmt.Fprint(stderr, `Usage: atago record [--out FILE] [--force] [--shell] [--snapshot] -- <command> [args...]

Runs the command once in a scratch directory and prints a spec skeleton
derived from what it observed: exit code, first stdout line, empty stderr,
and created files. Interactive (pty) and HTTP recording are non-goals for
now — write those steps by hand.
`)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitConfig
	}
	cmdArgs := fs.Args()
	if len(cmdArgs) == 0 {
		fmt.Fprintln(stderr, "atago record: no command given (usage: atago record [flags] -- <command> [args...])")
		return ExitConfig
	}
	if *snap && *out == "" {
		fmt.Fprintln(stderr, "atago record: --snapshot needs --out (the golden is written next to the spec)")
		return ExitConfig
	}

	// Preserve argv boundaries: the runner (and later the generated spec)
	// re-tokenizes the command string, so arguments containing spaces or
	// quotes must be re-quoted or the recorded run diverges from what the
	// user typed. In --shell mode the raw command line IS the input — the
	// shell does its own tokenization — so it joins verbatim.
	command := shellJoin(cmdArgs)
	if *shell {
		command = strings.Join(cmdArgs, " ")
	}
	workdir, err := os.MkdirTemp("", "atago-record-")
	if err != nil {
		fmt.Fprintf(stderr, "atago record: could not create scratch dir: %v\n", err)
		return ExitExec
	}
	defer func() { _ = os.RemoveAll(workdir) }() // best-effort scratch cleanup

	run := &spec.Run{Command: command}
	if *shell {
		run.Shell = spec.Bool(true)
	}
	res, err := runnercmd.New().Run(context.Background(), run, workdir)
	if err != nil {
		fmt.Fprintf(stderr, "atago record: %v\n", err)
		return ExitExec
	}

	created, err := listFiles(workdir)
	if err != nil {
		fmt.Fprintf(stderr, "atago record: could not inspect the scratch dir: %v\n", err)
		return ExitExec
	}

	obs := record.Observation{
		Command:      command,
		Shell:        *shell,
		ExitCode:     res.ExitCode,
		Stdout:       res.Stdout,
		Stderr:       res.Stderr,
		CreatedFiles: created,
	}
	opts := record.Options{SuiteName: suiteNameFor(cmdArgs[0])}
	if *snap {
		opts.Snapshot = true
		opts.SnapshotPath = "snapshots/" + strings.TrimSuffix(filepath.Base(*out), ".atago.yaml") + ".stdout.txt"
	}
	generated, err := record.Generate(obs, opts)
	if err != nil {
		fmt.Fprintf(stderr, "atago record: %v\n", err)
		return ExitInternal
	}

	fmt.Fprintf(stderr, "recorded: exit %d, %d stdout line(s), %d file(s) created\n",
		res.ExitCode, countLines(res.Stdout), len(created))

	if *out == "" {
		fmt.Fprint(stdout, string(generated))
		return ExitOK
	}
	if _, err := os.Stat(*out); err == nil && !*force {
		fmt.Fprintf(stderr, "atago record: %q already exists (use --force to overwrite)\n", *out)
		return ExitConfig
	}
	if dir := filepath.Dir(*out); dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			fmt.Fprintf(stderr, "atago record: %v\n", err)
			return ExitExec
		}
	}
	if *snap {
		golden := filepath.Join(filepath.Dir(*out), filepath.FromSlash(opts.SnapshotPath))
		if err := snapshot.Update(golden, res.Stdout, snapshot.Options{Workdir: workdir}); err != nil {
			fmt.Fprintf(stderr, "atago record: could not write snapshot golden: %v\n", err)
			return ExitExec
		}
		fmt.Fprintf(stderr, "wrote %s\n", golden)
	}
	if err := os.WriteFile(*out, generated, 0o644); err != nil { //nolint:gosec // a spec file the user asked for
		fmt.Fprintf(stderr, "atago record: %v\n", err)
		return ExitExec
	}
	fmt.Fprintf(stderr, "wrote %s\n", *out)
	return ExitOK
}

// listFiles walks the scratch dir and returns every file as a sorted,
// /-separated relative path — the created-files diff (the dir starts empty).
func listFiles(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, rerr := filepath.Rel(root, p)
		if rerr != nil {
			return rerr
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(out)
	return out, nil
}

// shellJoin renders argv as one command string that re-tokenizes to the same
// argv through the cmd runner (go-shellwords on POSIX, native argv splitting
// on Windows). Only whitespace splits a token, so a token is left verbatim
// unless it carries whitespace, a quote, or is empty; those are wrapped in
// DOUBLE quotes — the one quoting both tokenizers accept — with embedded
// backslashes and double-quotes escaped. A Windows path like C:\tool.exe has
// no whitespace, so it passes through unquoted (single-quoting it would leave
// Windows treating the quotes as part of the filename).
func shellJoin(args []string) string {
	quoted := make([]string, len(args))
	for i, a := range args {
		if !needsQuoting(a) {
			quoted[i] = a
			continue
		}
		// Escape only embedded double-quotes: a backslash inside double
		// quotes is a POSIX escape but a literal on Windows, so escaping it
		// would round-trip on one platform and double up on the other.
		// Leaving it literal keeps a spaced Windows path (C:\Program
		// Files\tool.exe) intact on both.
		quoted[i] = `"` + strings.ReplaceAll(a, `"`, `\"`) + `"`
	}
	return strings.Join(quoted, " ")
}

// needsQuoting reports whether a token would lose its argv boundary when
// re-tokenized unquoted: the empty string (would vanish) or anything with
// whitespace or a quote character.
func needsQuoting(s string) bool {
	if s == "" {
		return true
	}
	return strings.ContainsAny(s, " \t\n'\"")
}

// suiteNameFor derives a suite name from the command's base name.
func suiteNameFor(argv0 string) string {
	name := filepath.Base(filepath.FromSlash(argv0))
	name = strings.TrimSuffix(name, ".exe")
	if name == "" || name == "." {
		return "recorded"
	}
	return name
}

func countLines(stream []byte) int {
	n := 0
	for _, l := range strings.Split(string(stream), "\n") {
		if strings.TrimSpace(l) != "" {
			n++
		}
	}
	return n
}
