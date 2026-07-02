// Package cmd implements the command Runner: it executes a real process and
// captures its exit code, stdout, and stderr. Shell execution is
// opt-in via run.shell; the default tokenizes the command and execs directly.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	shellwords "github.com/mattn/go-shellwords"
	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
)

// Runner executes commands as child processes.
type Runner struct{}

// New returns a command Runner.
func New() *Runner { return &Runner{} }

// Run executes run.Command inside workdir and returns the observed Result.
func (r *Runner) Run(ctx context.Context, run *spec.Run, workdir string) (*runner.Result, error) {
	name, args, err := commandLine(run)
	if err != nil {
		return nil, err
	}

	timeout, err := parseTimeout(run.Timeout)
	if err != nil {
		return nil, err
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // executing user-declared commands is the purpose of atago
	if run.ShellEnabled() {
		// On Windows this hands cmd.exe the raw command line (see
		// ConfigureShell); Go's default argv escaping follows MSVCRT rules that
		// cmd.exe does not, so an embedded quote would reach the command as \".
		ConfigureShell(cmd, run.Command)
	}
	cmd.Dir = resolveDir(workdir, run.Cwd)
	cmd.Env = buildEnv(run.Env)
	// On cancellation (Ctrl-C / suite cancel / step timeout), kill the whole
	// process group, not just the shell we spawned: `sh -c "sleep 30"` orphans its
	// child, and that orphan keeps the stdout/stderr pipe open, so cmd.Wait would
	// otherwise block until it exits on its own. WaitDelay is a portable backstop
	// that force-closes the pipes if a stray child still lingers.
	configureCancellation(cmd)
	if run.Stdin != "" {
		cmd.Stdin = strings.NewReader(run.Stdin)
	}

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	runErr := cmd.Run()
	elapsed := time.Since(start)

	// Redirect captured stdout/stderr to workdir-relative files when requested
	// (run.stdout_to / run.stderr_to). This lets a `shell: false` step write output
	// to a file without the shell's `>` operator; the streams stay captured above,
	// so stdout/stderr assertions on the same step keep working.
	if err := writeRedirects(run, workdir, stdout.String(), stderr.String()); err != nil {
		return nil, err
	}

	res := &runner.Result{
		Command:  run.Command,
		Stdout:   []byte(stdout.String()),
		Stderr:   []byte(stderr.String()),
		Duration: elapsed,
		Workdir:  workdir,
	}

	// A step's own timeout (run.timeout) is an observable outcome, not an error:
	// mark it TimedOut and let assertions inspect it.
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		res.TimedOut = true
		res.ExitCode = -1
		return res, nil
	}
	// A parent-context cancellation (Ctrl-C / suite cancel) killed the process; it
	// is NOT a normal exit. Surface it as an error so the engine stops the scenario
	// instead of asserting against the killed result (issue #30).
	if err := ctx.Err(); err != nil {
		res.ExitCode = -1
		return res, fmt.Errorf("run %q cancelled: %w", run.Command, err)
	}

	var exitErr *exec.ExitError
	switch {
	case runErr == nil:
		res.ExitCode = 0
	case errors.As(runErr, &exitErr):
		res.ExitCode = exitErr.ExitCode()
	default:
		// Could not start the process at all (not found, permission, ...).
		return nil, fmt.Errorf("failed to execute %q: %w", run.Command, runErr)
	}
	return res, nil
}

// writeRedirects writes the captured stdout/stderr to the workdir-relative files
// named by run.StdoutTo / run.StderrTo (create/truncate). Paths are confined to
// the scenario workdir with the same rule as assertion paths. A no-op when
// neither field is set.
func writeRedirects(run *spec.Run, workdir, stdout, stderr string) error {
	for _, r := range []struct{ field, path, data string }{
		{"run.stdout_to", run.StdoutTo, stdout},
		{"run.stderr_to", run.StderrTo, stderr},
	} {
		if r.path == "" {
			continue
		}
		abs, err := security.ResolveWorkdirPath(r.field, workdir, r.path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(abs, []byte(r.data), 0o644); err != nil { //nolint:gosec // user-declared redirect target inside the workdir
			return fmt.Errorf("%s: %w", r.field, err)
		}
	}
	return nil
}

// commandLine resolves the program and arguments for a run step, honoring the
// explicit shell opt-in.
func commandLine(run *spec.Run) (string, []string, error) {
	return CommandLine(run.Command, run.ShellEnabled())
}

// CommandLine resolves the program and arguments for a command string, honoring
// the explicit shell opt-in. It is shared with the background
// service runner so services tokenize and shell-quote identically to run steps.
//
// A Windows shell command additionally needs ConfigureShell on the built
// *exec.Cmd: the ("cmd", "/c", command) argv returned here would be re-escaped
// by Go with MSVCRT quoting rules that cmd.exe does not follow.
func CommandLine(command string, shell bool) (string, []string, error) {
	if shell {
		if runtime.GOOS == "windows" {
			return "cmd", []string{"/c", command}, nil
		}
		return shellPath(), []string{"-c", command}, nil
	}
	fields, err := splitArgv(command)
	if err != nil {
		return "", nil, fmt.Errorf("cannot parse command %q: %w", command, err)
	}
	if len(fields) == 0 {
		return "", nil, fmt.Errorf("empty command")
	}
	return fields[0], fields[1:], nil
}

// splitArgv tokenizes a no-shell command into argv fields. POSIX uses shell
// word-splitting rules (quotes and backslash escapes); Windows uses
// windowsFields, where a backslash is an ordinary path character — POSIX
// backslash-escape semantics would corrupt every C:\ path (e.g. the expanded
// ${atago} binary path).
func splitArgv(command string) ([]string, error) {
	if runtime.GOOS == "windows" {
		return windowsFields(command)
	}
	return shellwords.Parse(command)
}

// windowsFields splits command on unquoted whitespace. Double quotes group a
// field (and are stripped); every other character — including backslash — is
// literal.
func windowsFields(command string) ([]string, error) {
	var fields []string
	var cur strings.Builder
	inQuote, started := false, false
	for _, r := range command {
		switch {
		case r == '"':
			inQuote = !inQuote
			started = true
		case !inQuote && (r == ' ' || r == '\t'):
			if started {
				fields = append(fields, cur.String())
				cur.Reset()
				started = false
			}
		default:
			cur.WriteRune(r)
			started = true
		}
	}
	if inQuote {
		return nil, fmt.Errorf("unclosed double quote")
	}
	if started {
		fields = append(fields, cur.String())
	}
	return fields, nil
}

// shellPath returns an absolute path to the POSIX shell used for `shell: true`.
//
// It deliberately does NOT trust PATH: atago sets up PATH for the *program
// under test*, and a CLI may legitimately ship its own `sh` applet (e.g.
// mimixbox). If the harness resolved its shell through that PATH, the program
// under test would hijack atago's shell — changing pipe/redirect semantics and
// exit codes. So we prefer a fixed system location (mirroring ShellSpec's
// absolute `--shell /bin/sh`). The ATAGO_SHELL env var allows an explicit
// override; an absolute /bin/sh is the default; only as a last resort do we
// fall back to a PATH lookup.
func shellPath() string {
	if s := os.Getenv("ATAGO_SHELL"); s != "" {
		return s
	}
	if _, err := os.Stat("/bin/sh"); err == nil {
		return "/bin/sh"
	}
	if p, err := exec.LookPath("sh"); err == nil {
		return p
	}
	return "/bin/sh"
}

func parseTimeout(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid timeout %q: %w", s, err)
	}
	return d, nil
}

// ResolveDir returns the working directory for a command run in workdir; cwd is
// interpreted relative to the scenario workdir (shared with the service runner).
func ResolveDir(workdir, cwd string) string { return resolveDir(workdir, cwd) }

// BuildEnv inherits the parent environment and applies per-command overrides
// (shared with the service runner).
func BuildEnv(overrides map[string]string) []string { return buildEnv(overrides) }

// resolveDir returns the working directory for the command. cwd is interpreted
// relative to the scenario workdir.
func resolveDir(workdir, cwd string) string {
	if cwd == "" || cwd == "." {
		return workdir
	}
	if filepath.IsAbs(cwd) {
		return cwd
	}
	return filepath.Join(workdir, cwd)
}

// buildEnv inherits the parent environment and applies per-step overrides.
func buildEnv(overrides map[string]string) []string {
	if len(overrides) == 0 {
		return nil // nil → inherit os.Environ() (exec default)
	}
	env := os.Environ()
	for k, v := range overrides {
		env = append(env, k+"="+v)
	}
	return env
}
