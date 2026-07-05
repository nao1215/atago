package loader

import (
	"encoding/base64"
	"time"

	"github.com/nao1215/atago/internal/spec"
)

// validateRunStep checks a run step wherever it appears. The suite.setup and
// scenario paths share the runner reference, timeout, required command,
// hermetic env, stdin, and retry checks in the same order; full adds the
// scenario-only shell-metacharacter hint (on the command) and the ssh-only
// field checks, so the two call sites stay a single source of truth (their
// only difference was those two additions).
func validateRunStep(add func(string, ...any), where string, r *spec.Run, runners map[string]spec.Runner, full bool) {
	validateRunnerRef(add, where, "run", r.Runner, runners)
	if r.Timeout != "" {
		if _, err := time.ParseDuration(r.Timeout); err != nil {
			add("%s.run.timeout %q is not a valid duration (e.g. \"30s\")", where, r.Timeout)
		}
	}
	if r.Command == "" {
		add("%s.run.command is required", where)
	} else if full && !r.ShellEnabled() && !runnerIsSSH(r.Runner, runners) {
		// Without shell, the command is tokenized into argv, so shell operators
		// (redirects, pipes, sequencing, substitution) are not honored. Rather
		// than silently pass them as literal argv, flag them with a fix-forward
		// hint (#: shell authoring UX). An ssh command is exempt: it travels
		// as one string and the REMOTE login shell always interprets it, so
		// metacharacters are honored there without any shell: opt-in.
		if tok := shellMetachar(r.Command); tok != "" {
			add("%s.run.command contains the shell metacharacter %q but shell is not enabled; "+
				"set `shell: true` to run it through a shell, split it into multiple `run` steps, "+
				"or use `stdout_to` / `stderr_to` for redirection", where, tok)
		}
	}
	validateHermeticEnv(add, where+".run", r.ClearEnv, r.PassEnv)
	validateStdin(add, where+".run", r.Stdin)
	validateRetry(add, where+".run", r.Retry)
	if full {
		validateSSHRunFields(add, where, r, runners)
	}
}

// validateStdin checks a run step's stdin source (#18): the mapping form must
// set exactly one of file/base64, and a base64 payload must decode — at load
// time, so a typo fails with a positioned message instead of mid-run.
func validateStdin(add func(string, ...any), where string, s spec.Stdin) {
	if s.IsMapping() {
		set := 0
		if s.File != "" {
			set++
		}
		if s.Base64 != "" {
			set++
		}
		if set != 1 {
			add("%s.stdin must set exactly one of file/base64 (or be a plain string for inline text)", where)
		}
	}
	if s.Base64 != "" {
		if _, err := base64.StdEncoding.DecodeString(s.Base64); err != nil {
			add("%s.stdin.base64 is not valid base64: %v", where, err)
		}
	}
}

// validateHermeticEnv checks the clear_env/pass_env pairing (#16): pass_env is
// only meaningful when clear_env starts the environment empty, so an
// allowlist without clear_env: true is authoring confusion and is rejected
// instead of silently ignored. Empty variable names are rejected too.
func validateHermeticEnv(add func(string, ...any), where string, clearEnv *bool, passEnv []string) {
	if len(passEnv) == 0 {
		return
	}
	if clearEnv == nil || !*clearEnv {
		add("%s.pass_env requires clear_env: true (pass_env selects host vars for a cleared environment)", where)
	}
	for i, name := range passEnv {
		if name == "" {
			add("%s.pass_env[%d] must not be an empty variable name", where, i)
		}
	}
}

// runnerIsSSH reports whether name references a declared ssh-type runner.
func runnerIsSSH(name string, runners map[string]spec.Runner) bool {
	if name == "" {
		return false
	}
	rdef, ok := runners[name]
	return ok && rdef.Type == "ssh"
}

// validateSSHRunFields rejects run-step fields that only shape LOCAL execution
// when the step names an ssh runner: the command runs on the remote host, so
// env/clear_env/pass_env/sandbox_home/stdin/stdout_to/stderr_to/cwd are silently
// dropped by the engine's remote path (it forwards only the command). Rejecting
// them at load time turns a silent no-op into a clear error. timeout and retry
// are honored remotely and are intentionally absent here.
func validateSSHRunFields(add func(string, ...any), where string, r *spec.Run, runners map[string]spec.Runner) {
	if !runnerIsSSH(r.Runner, runners) {
		return
	}
	// shell gets its own message: it is not merely ignored — the remote login
	// shell ALWAYS interprets the command string, so the knob has nothing to
	// switch and an authored value only misleads.
	if r.Shell != nil {
		add("%s.run.shell has no effect on an ssh runner (the remote login shell always interprets the command)", where)
	}
	fields := []struct {
		set   bool
		field string
	}{
		{r.SandboxHome != nil, "sandbox_home"},
		{r.ClearEnv != nil, "clear_env"},
		{len(r.PassEnv) > 0, "pass_env"},
		{len(r.Env) > 0, "env"},
		{!r.Stdin.IsZero(), "stdin"},
		{r.StdoutTo != "", "stdout_to"},
		{r.StderrTo != "", "stderr_to"},
		{r.Cwd != "", "cwd"},
	}
	for _, f := range fields {
		if f.set {
			add("%s.run.%s has no effect on an ssh runner (the command runs remotely; only command/runner/timeout apply)", where, f.field)
		}
	}
}

// shellMetachar returns the first shell metacharacter found outside quotes in a
// command, or "" if none. It backs the `shell: false` guard: with shell off the
// command is tokenized into argv, so operators like `>` or `|` would be passed
// as literal arguments instead of doing what the author expects. Single- and
// double-quoted regions are skipped so a quoted `">"` argument is not flagged.
func shellMetachar(cmd string) string {
	var quote rune // 0, '\'' or '"' when inside a quoted region
	runes := []rune(cmd)
	for i := range len(runes) {
		c := runes[i]
		if quote != 0 {
			if c == quote {
				quote = 0
			}
			continue
		}
		switch c {
		case '\'', '"':
			quote = c
		case '`':
			return "`"
		case '|':
			if i+1 < len(runes) && runes[i+1] == '|' {
				return "||"
			}
			return "|"
		case '&':
			if i+1 < len(runes) && runes[i+1] == '&' {
				return "&&"
			}
			// A lone `&` (background) is not in the guarded set; ignore it.
		case ';':
			return ";"
		case '<':
			return "<"
		case '>':
			if i+1 < len(runes) && runes[i+1] == '>' {
				return ">>"
			}
			return ">"
		case '$':
			if i+1 < len(runes) && runes[i+1] == '(' {
				return "$("
			}
		}
	}
	return ""
}

// validateRetry validates a retry block; where already names the owning step
// action (".run" or ".http") so messages read e.g. "steps[0].http.retry.times".
func validateRetry(add func(string, ...any), where string, r *spec.Retry) {
	if r == nil {
		return
	}
	if r.Times < 1 {
		add("%s.retry.times must be >= 1 (got %d)", where, r.Times)
	}
	if r.Interval != "" {
		if _, err := time.ParseDuration(r.Interval); err != nil {
			add("%s.retry.interval %q is not a valid duration", where, r.Interval)
		}
	}
	if r.Until == nil {
		add("%s.retry.until is required", where)
		return
	}
	validateAssert(add, where+".retry.until", r.Until, nil)
	// A retry's until is evaluated against the raw exec result of each attempt.
	// changes (the workdir delta) is computed only for a top-level assert
	// directly after the step, and screen renders a pty step's terminal — a
	// run/http result is never a pty. Neither can ever hold here, so the step
	// would only ever exhaust its budget: reject them at load time instead.
	if r.Until.Changes != nil {
		add("%s.retry.until.changes cannot be satisfied in a retry condition (the workdir delta is computed only for a top-level assert directly after the step, never for the exec result a retry polls); move it to an assert after the step", where)
	}
	if r.Until.Screen != nil {
		add("%s.retry.until.screen cannot be satisfied in a retry condition (screen renders a pty step's terminal, and a run/http result is never a pty); move it to an assert after a pty step", where)
	}
}
