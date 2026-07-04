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
	"runtime"
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
	ptyMode := fs.Bool("pty", false, "record an interactive pty session and generate an expect/send spec (POSIX-only)")
	fs.Usage = func() {
		fmt.Fprint(stderr, `Usage: atago record [--out FILE] [--force] [--shell] [--snapshot] -- <command> [args...]
       atago record --pty [--out FILE] [--force] [--shell] -- <command> [args...]

Runs the command once in a scratch directory and prints a spec skeleton
derived from what it observed: exit code, first stdout line, empty stderr,
and created files.

With --pty, runs the command in a real pseudo-terminal wired to your
terminal, lets you drive one interactive session by hand, and generates a
spec whose pty: step replays it as expect/send pairs (POSIX-only). HTTP
recording is a non-goal for now — write those steps by hand.
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
	if *ptyMode && *snap {
		// A hand-driven screen golden is too brittle to auto-record in v1.
		fmt.Fprintln(stderr, "atago record: --snapshot cannot be combined with --pty")
		return ExitConfig
	}
	if *ptyMode {
		return recordPTY(cmdArgs, *shell, *out, *force, stdout, stderr)
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

// recordPTY implements `atago record --pty` (#69): run the command in a real
// pseudo-terminal wired to the developer's own terminal, let them drive one
// interactive session by hand, and generate a spec whose pty: step replays it
// as expect/send pairs. POSIX-only; Windows returns a clear error.
func recordPTY(cmdArgs []string, shell bool, out string, force bool, stdout, stderr io.Writer) int {
	// Refuse an existing --out up front, before driving the session: otherwise
	// a user hand-drives the whole interactive session only to be told the file
	// already exists once it is too late to save the transcript.
	if _, err := os.Stat(out); err == nil && !force {
		fmt.Fprintf(stderr, "atago record: %q already exists (use --force to overwrite)\n", out)
		return ExitConfig
	}

	command := shellJoin(cmdArgs)
	if shell {
		command = strings.Join(cmdArgs, " ")
	}

	fmt.Fprintln(stderr, "recording pty session — drive it by hand; it ends when the program exits")
	rec, err := record.CapturePTY(command, shell, os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintf(stderr, "atago record: %v\n", err)
		return ExitExec
	}

	opts := record.Options{SuiteName: suiteNameFor(cmdArgs[0])}
	generated, err := record.GeneratePTY(rec, opts)
	if err != nil {
		fmt.Fprintf(stderr, "atago record: %v\n", err)
		return ExitInternal
	}

	sends := 0
	for _, seg := range rec.Segments {
		if seg.Input != nil {
			sends++
		}
	}
	fmt.Fprintf(stderr, "recorded: pty session, exit %d, %d send(s)\n", rec.ExitCode, sends)

	if out == "" {
		fmt.Fprint(stdout, string(generated))
		return ExitOK
	}
	if dir := filepath.Dir(out); dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			fmt.Fprintf(stderr, "atago record: %v\n", err)
			return ExitExec
		}
	}
	if err := os.WriteFile(out, generated, 0o644); err != nil { //nolint:gosec // a spec file the user asked for
		fmt.Fprintf(stderr, "atago record: %v\n", err)
		return ExitExec
	}
	fmt.Fprintf(stderr, "wrote %s\n", out)
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
// argv through the cmd runner, which uses go-shellwords on POSIX and native
// argv splitting on Windows. The two tokenizers reinterpret different
// characters, so quoting is platform-specific: it runs on the machine that
// records, which is the machine that replays.
func shellJoin(args []string) string {
	if runtime.GOOS == "windows" {
		return windowsJoin(args)
	}
	return posixJoin(args)
}

// posixJoin single-quotes every token that is not a plain word. Single quotes
// disable ALL go-shellwords interpretation — whitespace, the metacharacters
// it treats as boundaries (; & | < > ( )), globs, $-expansion, and
// backslashes — so `foo|bar`, `a>b`, and `C:\tmp\file` all round-trip. An
// embedded single quote is closed, escaped, and reopened ('\”).
func posixJoin(args []string) string {
	out := make([]string, len(args))
	for i, a := range args {
		if plainPOSIXWord(a) {
			out[i] = a
			continue
		}
		out[i] = "'" + strings.ReplaceAll(a, "'", `'\''`) + "'"
	}
	return strings.Join(out, " ")
}

// windowsJoin double-quotes tokens the native tokenizer would split or
// mis-read. Windows argv splitting breaks only on whitespace (the shell
// metacharacters go-shellwords honors are not argv boundaries there), and
// backslashes are literal, so a path like C:\tool.exe passes through and only
// whitespace/quote-bearing tokens need wrapping.
func windowsJoin(args []string) string {
	out := make([]string, len(args))
	for i, a := range args {
		if a != "" && !strings.ContainsAny(a, " \t\n\"") {
			out[i] = a
			continue
		}
		out[i] = `"` + strings.ReplaceAll(a, `"`, `\"`) + `"`
	}
	return strings.Join(out, " ")
}

// plainPOSIXWord reports whether a token survives go-shellwords tokenization
// unquoted: only alphanumerics and a small set of punctuation that carries no
// shell meaning.
func plainPOSIXWord(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case strings.ContainsRune("_@%+=:,./-", r):
		default:
			return false
		}
	}
	return true
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
