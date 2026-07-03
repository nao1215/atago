package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/spec"
)

// argvCommand returns a command string for the no-shell (argv-tokenized) path
// that works on every OS: POSIX gets the bare utility, Windows spells out the
// cmd.exe invocation explicitly (echo/exit are cmd builtins, not executables).
func argvCommand(posix, windows string) string {
	if runtime.GOOS == "windows" {
		return windows
	}
	return posix
}

func TestRun_BasicCapture(t *testing.T) {
	t.Parallel()
	r := New()
	res, err := r.Run(context.Background(), &spec.Run{Command: argvCommand("echo hello", "cmd /c echo hello")}, t.TempDir())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if res.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", res.ExitCode)
	}
	if strings.TrimSpace(string(res.Stdout)) != "hello" {
		t.Errorf("stdout = %q, want hello", res.Stdout)
	}
}

// TestRun_StdoutToStderrTo proves stdout_to / stderr_to write the captured
// streams to workdir-relative files (no shell needed) while the streams remain
// captured for assertions on the same step.
func TestRun_StdoutToStderrTo(t *testing.T) {
	t.Parallel()
	wd := t.TempDir()
	r := New()
	res, err := r.Run(context.Background(), &spec.Run{
		Command:  argvCommand("sh -c 'echo out; echo err 1>&2'", "echo out& echo err 1>&2"),
		Shell:    spec.Bool(true),
		StdoutTo: "out.txt",
		StderrTo: "err.txt",
	}, wd)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	// Files are written with the captured stream contents.
	if got := readTrim(t, filepath.Join(wd, "out.txt")); got != "out" {
		t.Errorf("out.txt = %q, want out", got)
	}
	if got := readTrim(t, filepath.Join(wd, "err.txt")); got != "err" {
		t.Errorf("err.txt = %q, want err", got)
	}
	// The streams are still captured internally for assertions.
	if strings.TrimSpace(string(res.Stdout)) != "out" {
		t.Errorf("captured stdout = %q, want out", res.Stdout)
	}
}

// TestRun_StdoutToConfinedToWorkdir proves a redirect path may not escape the
// scenario workdir.
func TestRun_StdoutToConfinedToWorkdir(t *testing.T) {
	t.Parallel()
	r := New()
	_, err := r.Run(context.Background(), &spec.Run{
		Command:  argvCommand("echo hi", "cmd /c echo hi"),
		StdoutTo: "../escape.txt",
	}, t.TempDir())
	if err == nil {
		t.Fatal("expected a workdir-confinement error, got nil")
	}
}

func readTrim(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	return strings.TrimSpace(string(b))
}

func TestRun_NonZeroExit(t *testing.T) {
	t.Parallel()
	r := New()
	res, err := r.Run(context.Background(), &spec.Run{Command: argvCommand("false", "cmd /c exit 1")}, t.TempDir())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if res.ExitCode != 1 {
		t.Errorf("exit code = %d, want 1", res.ExitCode)
	}
}

func TestRun_ShellAndEnvAndCwd(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o750); err != nil {
		t.Fatal(err)
	}
	r := New()
	res, err := r.Run(context.Background(), &spec.Run{
		Shell: spec.Bool(true),
		// Both print "<env value> in <cwd>"; POSIX prints the basename, cmd.exe
		// the absolute path, so the assertion checks prefix and suffix.
		Command: argvCommand("echo $GREETING in $(basename $PWD)", "echo %GREETING% in %CD%"),
		Cwd:     "sub",
		Env:     map[string]string{"GREETING": "hi"},
	}, dir)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	got := strings.TrimSpace(string(res.Stdout))
	if !strings.HasPrefix(got, "hi in ") || !strings.HasSuffix(got, "sub") {
		t.Errorf("stdout = %q, want %q", got, "hi in .../sub")
	}
}

// envMap converts an exec-style env slice into a map for assertions (last
// entry wins, matching exec.Cmd's dedup semantics).
func envMap(env []string) map[string]string {
	m := make(map[string]string, len(env))
	for _, kv := range env {
		k, v, _ := strings.Cut(kv, "=")
		m[k] = v
	}
	return m
}

// TestBuildEnv_ClearEnvDropsHostEnv proves clear_env: true starts from an
// empty environment: a canary host var must not reach the child (#16).
func TestBuildEnv_ClearEnvDropsHostEnv(t *testing.T) {
	t.Setenv("ATAGO_TEST_CANARY", "leaked")
	env := BuildEnv(nil, true, nil)
	if env == nil {
		t.Fatal("BuildEnv(clear) = nil, want a non-nil slice (nil inherits os.Environ)")
	}
	if _, ok := envMap(env)["ATAGO_TEST_CANARY"]; ok {
		t.Errorf("canary host var leaked into cleared env: %v", env)
	}
}

// TestBuildEnv_PassEnvCopiesListedVars proves pass_env copies exactly the
// listed host vars and skips unset ones (#16).
func TestBuildEnv_PassEnvCopiesListedVars(t *testing.T) {
	t.Setenv("ATAGO_TEST_KEEP", "kept")
	t.Setenv("ATAGO_TEST_DROP", "dropped")
	env := envMap(BuildEnv(nil, true, []string{"ATAGO_TEST_KEEP", "ATAGO_TEST_UNSET"}))
	if got := env["ATAGO_TEST_KEEP"]; got != "kept" {
		t.Errorf("passed var = %q, want kept", got)
	}
	if _, ok := env["ATAGO_TEST_DROP"]; ok {
		t.Error("unlisted host var leaked into cleared env")
	}
	if _, ok := env["ATAGO_TEST_UNSET"]; ok {
		t.Error("unset host var should be skipped, not set")
	}
}

// TestBuildEnv_OverridesLayerOnTop proves explicit env overrides win over
// passed-through host vars, preserving the existing layering order (#16).
func TestBuildEnv_OverridesLayerOnTop(t *testing.T) {
	t.Setenv("ATAGO_TEST_LAYER", "host")
	env := envMap(BuildEnv(map[string]string{"ATAGO_TEST_LAYER": "step"}, true, []string{"ATAGO_TEST_LAYER"}))
	if got := env["ATAGO_TEST_LAYER"]; got != "step" {
		t.Errorf("override = %q, want step (explicit env wins over pass_env)", got)
	}
}

// TestBuildEnv_NoClearKeepsInheritance proves the pre-#16 behavior is
// untouched when clear_env is off.
func TestBuildEnv_NoClearKeepsInheritance(t *testing.T) {
	if env := BuildEnv(nil, false, nil); env != nil {
		t.Errorf("BuildEnv(no overrides) = %v, want nil (inherit os.Environ)", env)
	}
	t.Setenv("ATAGO_TEST_INHERIT", "here")
	env := envMap(BuildEnv(map[string]string{"EXTRA": "1"}, false, nil))
	if got := env["ATAGO_TEST_INHERIT"]; got != "here" {
		t.Errorf("inherited var = %q, want here", got)
	}
	if got := env["EXTRA"]; got != "1" {
		t.Errorf("override = %q, want 1", got)
	}
}

// TestBuildEnv_WindowsKeepsSystemCriticalVars proves the documented
// system-critical set survives clear_env on Windows so processes can start.
func TestBuildEnv_WindowsKeepsSystemCriticalVars(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only behavior")
	}
	env := envMap(BuildEnv(nil, true, nil))
	for _, name := range []string{"SystemRoot", "SystemDrive", "PATHEXT"} {
		if _, ok := env[name]; !ok {
			t.Errorf("system-critical var %s missing from cleared env", name)
		}
	}
}

// TestRun_ClearEnvEndToEnd proves a run step with clear_env does not see a
// canary host var, and sees it again without clear_env (#16).
func TestRun_ClearEnvEndToEnd(t *testing.T) {
	t.Setenv("ATAGO_E2E_CANARY", "leaky")
	r := New()
	cmdLine := argvCommand("env", "cmd /c set")
	hermetic, err := r.Run(context.Background(), &spec.Run{
		Command:  cmdLine,
		ClearEnv: spec.Bool(true),
		PassEnv:  []string{"PATH"},
	}, t.TempDir())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if strings.Contains(string(hermetic.Stdout), "ATAGO_E2E_CANARY") {
		t.Errorf("clear_env child saw the canary var:\n%s", hermetic.Stdout)
	}
	plain, err := r.Run(context.Background(), &spec.Run{Command: cmdLine}, t.TempDir())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.Contains(string(plain.Stdout), "ATAGO_E2E_CANARY") {
		t.Errorf("non-hermetic child should inherit the canary var:\n%s", plain.Stdout)
	}
}

// TestRun_StdinFromFile proves stdin: {file: ...} feeds a workdir file's bytes
// to the child (#18).
func TestRun_StdinFromFile(t *testing.T) {
	t.Parallel()
	wd := t.TempDir()
	if err := os.WriteFile(filepath.Join(wd, "in.txt"), []byte("piped\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	r := New()
	res, err := r.Run(context.Background(), &spec.Run{
		Command: argvCommand("cat", "findstr piped"),
		Stdin:   spec.Stdin{File: "in.txt"},
	}, wd)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := strings.TrimSpace(string(res.Stdout)); got != "piped" {
		t.Errorf("stdout = %q, want piped", got)
	}
}

// TestRun_StdinFileConfinedToWorkdir proves the stdin file path may not escape
// the scenario workdir (#18).
func TestRun_StdinFileConfinedToWorkdir(t *testing.T) {
	t.Parallel()
	r := New()
	_, err := r.Run(context.Background(), &spec.Run{
		Command: argvCommand("cat", "findstr x"),
		Stdin:   spec.Stdin{File: "../escape.txt"},
	}, t.TempDir())
	if err == nil {
		t.Fatal("expected a workdir-confinement error, got nil")
	}
}

// TestRun_StdinFileMissing proves a missing stdin file is a named execution
// error instead of a silent empty stdin (#18).
func TestRun_StdinFileMissing(t *testing.T) {
	t.Parallel()
	r := New()
	_, err := r.Run(context.Background(), &spec.Run{
		Command: argvCommand("cat", "findstr x"),
		Stdin:   spec.Stdin{File: "nope.txt"},
	}, t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "run.stdin.file") {
		t.Fatalf("error = %v, want it to name run.stdin.file", err)
	}
}

// TestRun_StdinBase64Binary proves base64 stdin delivers the decoded bytes
// exactly, including NUL and non-UTF8 bytes (#18).
func TestRun_StdinBase64Binary(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("cat-based byte-exact check is POSIX-only; loader/schema cover Windows")
	}
	want := []byte{0x00, 0x01, 0x02, 0xff}
	r := New()
	res, err := r.Run(context.Background(), &spec.Run{
		Command: "cat",
		Stdin:   spec.Stdin{Base64: base64.StdEncoding.EncodeToString(want)},
	}, t.TempDir())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !bytes.Equal(res.Stdout, want) {
		t.Errorf("stdout = %x, want %x (byte-exact binary delivery)", res.Stdout, want)
	}
}

func TestRun_Stdin(t *testing.T) {
	t.Parallel()
	r := New()
	// cat / findstr both copy the matching stdin lines to stdout.
	res, err := r.Run(context.Background(), &spec.Run{
		Command: argvCommand("cat", "findstr piped"),
		Stdin:   spec.Stdin{Inline: "piped\n"},
	}, t.TempDir())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := strings.TrimSpace(string(res.Stdout)); got != "piped" {
		t.Errorf("stdout = %q, want %q", got, "piped")
	}
}

func TestRun_Timeout(t *testing.T) {
	t.Parallel()
	r := New()
	// ping -n 6 pauses ~5s between its echoes — the portable stand-in for sleep.
	res, err := r.Run(context.Background(), &spec.Run{
		Command: argvCommand("sleep 5", "ping -n 6 127.0.0.1"),
		Timeout: "50ms",
	}, t.TempDir())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !res.TimedOut {
		t.Errorf("TimedOut = false, want true")
	}
}

// Regression for issue #30: a parent-context cancellation (Ctrl-C / suite cancel)
// killed the process but was reported as a normal exit (ExitCode -1, nil error).
// It must now surface as an error and NOT be flagged as a step timeout.
func TestRun_ParentCancelIsNotNormalExit(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	res, err := New().Run(ctx, &spec.Run{Command: argvCommand("sleep 5", "ping -n 6 127.0.0.1")}, t.TempDir())
	if err == nil {
		t.Fatal("Run() error = nil; a parent-cancel kill must be reported as an error, not a normal exit")
	}
	if res != nil && res.TimedOut {
		t.Error("TimedOut = true; a parent cancel is not a step-timeout deadline")
	}
}

// TestRun_ShellCancelKillsChildPromptly is the regression for the pipe-orphan
// hang: a cancelled `sh -c "sleep 30"` used to block cmd.Wait for the full 30s
// because the orphaned sleep kept the captured stdout pipe open. Killing the
// whole process group (POSIX) or force-closing the pipes after WaitDelay
// (Windows) must let Run return in well under the sleep duration.
func TestRun_ShellCancelKillsChildPromptly(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	// Not t.TempDir(): on Windows the orphaned child (no process groups) may
	// briefly outlive the kill while holding the workdir open, and TempDir's
	// cleanup would fail the test on that unrelated race. Best-effort cleanup.
	wd, err := os.MkdirTemp("", "atago-cancel-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(wd) })
	start := time.Now()
	_, err = New().Run(ctx, &spec.Run{Shell: spec.Bool(true), Command: argvCommand("sleep 30", "ping -n 31 127.0.0.1")}, wd)
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("Run() error = nil; a parent-cancel kill must be reported as an error")
	}
	if elapsed > 5*time.Second {
		t.Fatalf("Run() took %v; the orphaned child kept the pipe open — process group was not killed", elapsed)
	}
}

// TestRun_ShellEmbeddedQuotes: a shell command carrying double quotes must
// reach the shell verbatim. On Windows, Go's MSVCRT-style argv escaping used to
// turn the quotes into literal \" because cmd.exe never unescapes them; the raw
// command line (ConfigureShell) fixes that.
func TestRun_ShellEmbeddedQuotes(t *testing.T) {
	t.Parallel()
	res, err := New().Run(context.Background(), &spec.Run{
		Shell:   spec.Bool(true),
		Command: `echo {"id":7}`,
	}, t.TempDir())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	got := strings.TrimSpace(string(res.Stdout))
	// Each shell's own quote handling applies — sh strips the quotes, cmd.exe
	// echoes them verbatim — but neither may see injected backslashes.
	want := "{id:7}"
	if runtime.GOOS == "windows" {
		want = `{"id":7}`
	}
	if got != want {
		t.Errorf("stdout = %q, want %q (argv escaping must not alter the shell command)", got, want)
	}
}

// TestWindowsFields pins the Windows argv tokenizer: backslashes are literal
// (they are path separators, not escapes) and double quotes group fields.
func TestWindowsFields(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want []string
	}{
		{`C:\bin\atago.exe run spec.yaml`, []string{`C:\bin\atago.exe`, "run", "spec.yaml"}},
		{`tool "a b" c`, []string{"tool", "a b", "c"}},
		{`tool --path "C:\Program Files\x"`, []string{"tool", "--path", `C:\Program Files\x`}},
		{"  spaced   out  ", []string{"spaced", "out"}},
		{`empty ""`, []string{"empty", ""}},
	}
	for _, tt := range tests {
		got, err := windowsFields(tt.in)
		if err != nil {
			t.Errorf("windowsFields(%q) error = %v", tt.in, err)
			continue
		}
		if len(got) != len(tt.want) {
			t.Errorf("windowsFields(%q) = %#v, want %#v", tt.in, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("windowsFields(%q)[%d] = %q, want %q", tt.in, i, got[i], tt.want[i])
			}
		}
	}
	if _, err := windowsFields(`broken "quote`); err == nil {
		t.Error("windowsFields with an unclosed quote should error")
	}
}

func TestRun_CommandNotFound(t *testing.T) {
	t.Parallel()
	r := New()
	_, err := r.Run(context.Background(), &spec.Run{Command: "definitely-not-a-real-binary-xyz"}, t.TempDir())
	if err == nil {
		t.Fatal("Run() error = nil, want execution error")
	}
}

// TestRun_ShellNotShadowedByPath verifies that a `sh` placed first on PATH does
// not hijack the harness shell (ADR-0018). A sabotaged `sh` that always prints
// HIJACKED and exits 0 sits in the workdir; with PATH pointing only at it, a
// PATH-resolved shell would run it, but the absolute /bin/sh must be used.
func TestRun_ShellNotShadowedByPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shellPath applies only to POSIX; Windows runs shell steps via cmd.exe")
	}
	dir := t.TempDir()
	fake := filepath.Join(dir, "sh")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\necho HIJACKED\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	t.Setenv("ATAGO_SHELL", "") // ensure no override interferes
	r := New()
	res, err := r.Run(context.Background(), &spec.Run{Shell: spec.Bool(true), Command: "echo real"}, dir)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := strings.TrimSpace(string(res.Stdout)); got != "real" {
		t.Errorf("stdout = %q, want %q (the PATH-resident sh hijacked the harness shell)", got, "real")
	}
}

func TestShellPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shellPath applies only to POSIX; Windows runs shell steps via cmd.exe")
	}
	t.Run("honors ATAGO_SHELL override", func(t *testing.T) {
		t.Setenv("ATAGO_SHELL", "/custom/shell")
		if got := shellPath(); got != "/custom/shell" {
			t.Errorf("shellPath() = %q, want /custom/shell", got)
		}
	})
	t.Run("defaults to an absolute path", func(t *testing.T) {
		t.Setenv("ATAGO_SHELL", "")
		if got := shellPath(); !filepath.IsAbs(got) {
			t.Errorf("shellPath() = %q, want an absolute path", got)
		}
	})
}
