// Package cmd implements the command Runner: it executes a real process and
// captures its exit code, stdout, and stderr. Shell execution is
// opt-in via run.shell; the default tokenizes the command and execs directly.
package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
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
	// sandbox_home (#71) redirects the child's home and per-OS config/cache dirs
	// at ${workdir}/.atago-home. The overlay sits between pass_env and the step's
	// own env in precedence and composes with clear_env.
	var sandbox map[string]string
	if run.SandboxHomeEnabled() {
		sandbox, err = EnsureSandboxHome(workdir)
		if err != nil {
			return nil, err
		}
	}
	cmd.Env = buildEnv(run.Env, run.ClearEnvEnabled(), run.PassEnv, sandbox)
	// On cancellation (Ctrl-C / suite cancel / step timeout), kill the whole
	// process group, not just the shell we spawned: `sh -c "sleep 30"` orphans its
	// child, and that orphan keeps the stdout/stderr pipe open, so cmd.Wait would
	// otherwise block until it exits on its own. WaitDelay is a portable backstop
	// that force-closes the pipes if a stray child still lingers.
	configureCancellation(cmd)
	stdin, err := stdinReader(run, workdir)
	if err != nil {
		return nil, err
	}
	if stdin != nil {
		cmd.Stdin = stdin
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
	// mark it TimedOut and let assertions inspect it. TimeoutSource carries the
	// engine-resolved level (step/runner/defaults/suite/built-in, #17) into the
	// failure hint.
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		res.TimedOut = true
		res.TimeoutSource = run.TimeoutSource
		res.ExitCode = -1
		return res, nil
	}
	// A parent-context cancellation (Ctrl-C / suite cancel) killed the process; it
	// is NOT a normal exit. Surface it as an error so the engine stops the scenario
	// instead of asserting against the killed result (issue #30).
	if err := ctx.Err(); err != nil {
		res.ExitCode = -1
		return res, fmt.Errorf("run %q canceled: %w", run.Command, err)
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

// stdinReader builds the reader fed to the command's standard input (#18):
// inline text verbatim, a workdir-confined file's bytes, or decoded base64
// for binary input. It returns nil when no stdin was authored (the child
// inherits no stdin, exec's default). The one-of shape and the base64 payload
// are validated at load time; the file read stays a runtime concern because
// the file typically comes from an earlier fixture or run step.
func stdinReader(run *spec.Run, workdir string) (io.Reader, error) {
	s := run.Stdin
	switch {
	case s.File != "":
		abs, err := security.ResolveWorkdirPath("run.stdin.file", workdir, s.File)
		if err != nil {
			return nil, err
		}
		data, err := os.ReadFile(abs) //nolint:gosec // confined to the scenario workdir above
		if err != nil {
			return nil, fmt.Errorf("run.stdin.file: %w", err)
		}
		return bytes.NewReader(data), nil
	case s.Base64 != "":
		data, err := base64.StdEncoding.DecodeString(s.Base64)
		if err != nil {
			// Unreachable for loader-validated specs; kept for direct API users.
			return nil, fmt.Errorf("run.stdin.base64: %w", err)
		}
		return bytes.NewReader(data), nil
	case s.Inline != "":
		return strings.NewReader(s.Inline), nil
	}
	return nil, nil
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
		// The program under test may have planted a symlink at the redirect target
		// pointing outside the workdir; write without following it (issue #16).
		if err := security.WriteFileNoFollow(abs, []byte(r.data), 0o644); err != nil {
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

// splitArgv tokenizes a no-shell command into argv fields. On POSIX it uses
// go-shellwords for quote handling; note its backslash escapes are C-style, not
// sh-style — `\t`/`\n` become a real tab/newline (sh would drop the backslash
// and keep the letter), and a trailing `\` is an error rather than a line
// continuation. Quote your command or set shell: true if you need a literal
// backslash. Windows uses windowsFields, which groups the same single/double
// quotes as go-shellwords (so a single-quoted argument tokenizes identically on
// both OSes) but keeps a backslash literal — sh backslash-escape semantics would
// corrupt every C:\ path (e.g. the expanded ${atago} binary path).
func splitArgv(command string) ([]string, error) {
	if runtime.GOOS == "windows" {
		return windowsFields(command)
	}
	return shellwords.Parse(command)
}

// windowsFields splits command on unquoted whitespace. Both double and single
// quotes group a field (and are stripped): inside a double-quoted group a single
// quote is literal, and inside a single-quoted group a double quote is literal —
// so a shell-free command tokenizes to the SAME argv on Windows as on POSIX
// (go-shellwords), where a single-quoted segment groups and strips too (#154).
// Before this, a single quote was an ordinary character on Windows only, so a
// cross-platform spec passing single-quoted inline JSON (`'{"k":"v"}'`) reached
// the CLI as non-JSON on Windows and broke there alone.
//
// Unquoted whitespace is space, tab, carriage return, and newline — the same set
// go-shellwords treats as a field separator. A command authored as a YAML block
// scalar carries newlines (a trailing one from `>`, interior ones from `|`) and a
// CRLF-authored spec carries `\r`; splitting on them keeps the argv identical to
// POSIX instead of gluing a stray `\n` onto the final argument on Windows alone.
//
// A backslash stays literal OUTSIDE quotes so a bare C:\ path survives (sh
// backslash-escape semantics would corrupt it); it is also literal inside either
// quote. This keeps the deliberate Windows-path behavior while removing the
// single-quote divergence, which was an unintended side effect, not a design goal.
func windowsFields(command string) ([]string, error) {
	var fields []string
	var cur strings.Builder
	inDouble, inSingle, started := false, false, false
	for _, r := range command {
		switch {
		case r == '"' && !inSingle:
			inDouble = !inDouble
			started = true
		case r == '\'' && !inDouble:
			inSingle = !inSingle
			started = true
		case !inDouble && !inSingle && isFieldSpace(r):
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
	if inDouble {
		return nil, fmt.Errorf("unclosed double quote")
	}
	if inSingle {
		return nil, fmt.Errorf("unclosed single quote")
	}
	if started {
		fields = append(fields, cur.String())
	}
	return fields, nil
}

// isFieldSpace reports whether r separates argv fields outside quotes. It matches
// go-shellwords' whitespace set (space, tab, carriage return, newline) so the
// no-shell tokenizer agrees with POSIX on where fields begin and end.
func isFieldSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
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

// BuildEnv composes a child process environment: it inherits the parent
// environment (or starts empty when clearEnv is set, re-admitting only the
// passEnv allowlist) and applies per-command overrides on top (shared with the
// service and pty runners) (#16). sandbox is the optional sandbox_home overlay
// (#71), layered above pass_env/host but below the step's own overrides; pass
// nil when no isolated home is requested.
func BuildEnv(overrides map[string]string, clearEnv bool, passEnv []string, sandbox map[string]string) []string {
	return buildEnv(overrides, clearEnv, passEnv, sandbox)
}

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

// windowsCriticalEnv is the documented set of system-critical variables always
// retained under clear_env on Windows so processes can start at all (#16).
var windowsCriticalEnv = []string{"SystemRoot", "SystemDrive", "TEMP", "TMP", "PATHEXT"}

// buildEnv composes the child environment. Without clearEnv it inherits the
// parent environment and appends the sandbox_home overlay and per-step
// overrides (a nil return means "inherit os.Environ()", exec's default). With
// clearEnv it starts empty: only the passEnv allowlist (plus the Windows
// system-critical set) is copied from the host, then the sandbox overlay, then
// overrides — so the precedence step env > sandbox > pass_env > host holds
// (os/exec keeps the last value for a duplicated key) (#16, #71).
func buildEnv(overrides map[string]string, clearEnv bool, passEnv []string, sandbox map[string]string) []string {
	if !clearEnv {
		if len(overrides) == 0 && len(sandbox) == 0 {
			return nil // nil → inherit os.Environ() (exec default)
		}
		env := os.Environ()
		for k, v := range sandbox {
			env = append(env, k+"="+v)
		}
		for k, v := range overrides {
			env = append(env, k+"="+v)
		}
		return env
	}
	env := make([]string, 0, len(passEnv)+len(sandbox)+len(overrides)+len(windowsCriticalEnv))
	if runtime.GOOS == "windows" {
		for _, name := range windowsCriticalEnv {
			if v, ok := lookupHostEnv(name); ok {
				env = append(env, name+"="+v)
			}
		}
	}
	for _, name := range passEnv {
		if v, ok := lookupHostEnv(name); ok {
			env = append(env, name+"="+v)
		}
	}
	// The sandbox overlay is injected after the environment is cleared and after
	// pass_env, so `pass_env: [HOME]` cannot leak the host home past an enabled
	// sandbox (#71).
	for k, v := range sandbox {
		env = append(env, k+"="+v)
	}
	for k, v := range overrides {
		env = append(env, k+"="+v)
	}
	return env
}

// lookupHostEnv reads a host variable for pass_env. Windows environment names
// are case-insensitive (PATH arrives as "Path"), so fall back to a
// case-insensitive scan there; os.LookupEnv is case-sensitive everywhere.
func lookupHostEnv(name string) (string, bool) {
	if v, ok := os.LookupEnv(name); ok {
		return v, true
	}
	if runtime.GOOS == "windows" {
		for _, kv := range os.Environ() {
			k, v, found := strings.Cut(kv, "=")
			if found && strings.EqualFold(k, name) {
				return v, true
			}
		}
	}
	return "", false
}
