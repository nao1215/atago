package engine

import (
	"runtime"
	"strings"
	"testing"
)

// The pty step (#8) runs one command inside a real pseudo-terminal and drives
// it with a declarative expect/send session — the gap that kept sqly's
// interactive-shell coverage on a hand-written Go PTY harness when its
// ShellSpec suite moved to atago. The transcript feeds the stdout assertion
// target, so all stream matchers work unchanged.

func skipOnWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("pty steps are POSIX-only for now (ConPTY later)")
	}
}

// TestEngine_PTY_AllocatesARealTerminal is the reason the step exists: a
// program that branches on TTY-ness must see a terminal. A plain run step
// pipes stdin and would print "pipe".
func TestEngine_PTY_AllocatesARealTerminal(t *testing.T) {
	skipOnWindows(t)
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: stdin is a tty under pty
    steps:
      - pty:
          shell: true
          command: 'if [ -t 0 ]; then echo is-a-tty; else echo is-a-pipe; fi'
      - assert:
          exit_code: 0
          stdout:
            contains: is-a-tty
  - name: stdin is a pipe under plain run
    steps:
      - run:
          shell: true
          command: 'if [ -t 0 ]; then echo is-a-tty; else echo is-a-pipe; fi'
      - assert:
          stdout:
            contains: is-a-pipe
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}

// TestEngine_PTY_ExpectSendSession drives a cat "REPL": terminal echo means
// every sent line lands in the transcript, expect gates each send, and an EOF
// (^D) ends the session with exit 0.
func TestEngine_PTY_ExpectSendSession(t *testing.T) {
	skipOnWindows(t)
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: drive cat interactively
    steps:
      - pty:
          command: cat
          timeout: 10s
          session:
            - send: "hello-repl\n"
            - expect: "hello-repl"
            - send: "second-line\n"
            - expect: "second-line"
            - send: ""
      - assert:
          exit_code: 0
          stdout:
            contains:
              - hello-repl
              - second-line
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}

// TestEngine_PTY_ExpectTimeoutFailsTheStep proves a never-matching expect
// fails like an assertion (scenario failed, not errored), names the pattern,
// and does not leak the child process.
func TestEngine_PTY_ExpectTimeoutFailsTheStep(t *testing.T) {
	skipOnWindows(t)
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: expect never matches
    steps:
      - pty:
          command: cat
          timeout: 2s
          session:
            - send: "something-else\n"
            - expect: "never-going-to-appear"
`)
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed: %+v", res.Status, res.Scenarios)
	}
	checks := res.Scenarios[0].Steps[0].Checks
	if len(checks) == 0 || checks[0].OK {
		t.Fatalf("want a failing expect check, got %+v", checks)
	}
	if !strings.Contains(checks[0].Desc, "never-going-to-appear") {
		t.Errorf("check desc = %q, want it to name the pattern", checks[0].Desc)
	}
}

// TestEngine_PTY_ExpectDoesNotRematchStale is a regression: each expect scans
// only the transcript AFTER the previous match, so a pattern that appeared once
// is not matched again from the stale buffer. "AAA" is printed exactly once, so
// the second expect for it must NOT match and the step times out. With the
// pre-fix whole-transcript match, the second expect matched the stale "AAA"
// instantly and the step passed falsely. (printf, not cat+send, so terminal echo
// does not duplicate the pattern.)
func TestEngine_PTY_ExpectDoesNotRematchStale(t *testing.T) {
	skipOnWindows(t)
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: a once-only pattern is not re-matched
    steps:
      - pty:
          shell: true
          command: 'printf "AAA\n"'
          timeout: 1s
          session:
            - expect: "AAA"
            - expect: "AAA"
`)
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed (the second AAA expect must not match the stale one): %+v", res.Status, res.Scenarios)
	}
}

// TestEngine_PTY_ExpectMatchesEachRecurringOccurrence proves the offset does not
// over-consume: when a pattern genuinely recurs, consecutive expects match the
// successive occurrences in order, so a normal prompt/response loop still passes.
func TestEngine_PTY_ExpectMatchesEachRecurringOccurrence(t *testing.T) {
	skipOnWindows(t)
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: two occurrences match two expects
    steps:
      - pty:
          shell: true
          command: 'printf "TICK\nTICK\n"'
          timeout: 5s
          session:
            - expect: "TICK"
            - expect: "TICK"
      - assert:
          exit_code: 0
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}

// TestEngine_PTY_SeesSuiteEnv proves suite.env reaches pty commands like it
// reaches run steps (CodeRabbit finding on #12: the pty step dropped it).
func TestEngine_PTY_SeesSuiteEnv(t *testing.T) {
	skipOnWindows(t)
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
  env:
    PTY_SUITE_FLAG: from-suite-env
scenarios:
  - name: pty inherits the suite environment
    steps:
      - pty:
          shell: true
          command: echo flag=$PTY_SUITE_FLAG
      - assert:
          exit_code: 0
          stdout:
            contains: flag=from-suite-env
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}

// TestEngine_PTY_SetsDefaultTERM proves a pty step exports TERM=xterm-256color
// by default. Without a sane TERM, full-screen TUIs (less, vim, htop) refuse to
// draw ("terminal is not fully functional"), so a pty/screen assertion can never
// see the real UI. atago renders through an xterm-compatible vt10x emulator, so
// the default TERM matches the emulator and is deterministic regardless of the
// host's own TERM (unset in CI, tmux/screen locally).
func TestEngine_PTY_SetsDefaultTERM(t *testing.T) {
	skipOnWindows(t)
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: pty exports a usable TERM
    steps:
      - pty:
          shell: true
          command: echo "TERM=[$TERM]"
      - assert:
          exit_code: 0
          stdout:
            contains: "TERM=[xterm-256color]"
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}

// TestEngine_PTY_RespectsExplicitTERM proves an author-set TERM wins over the
// default, so a spec can pin a specific terminal type when a program's behavior
// depends on it.
func TestEngine_PTY_RespectsExplicitTERM(t *testing.T) {
	skipOnWindows(t)
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: explicit TERM wins
    steps:
      - pty:
          shell: true
          command: echo "TERM=[$TERM]"
          env:
            TERM: dumb
      - assert:
          exit_code: 0
          stdout:
            contains: "TERM=[dumb]"
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}

// TestEngine_PTY_StoreAndVariablesFlow proves ${name} expansion reaches send
// payloads and the transcript feeds `store from.stdout` like any run step.
func TestEngine_PTY_StoreAndVariablesFlow(t *testing.T) {
	skipOnWindows(t)
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: variables flow through the session
    steps:
      - run: {shell: true, command: echo token-777}
      - store:
          name: tok
          from:
            stdout:
              matches: "token-[0-9]+"
      - pty:
          command: cat
          timeout: 10s
          session:
            - send: "use ${tok}\n"
            - expect: "use token-777"
            - send: ""
      - store:
          name: fromtty
          from:
            stdout:
              matches: "use token-[0-9]+"
      - run: {shell: true, command: "echo captured ${fromtty}"}
      - assert:
          stdout: {contains: captured use token-777}
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}

// TestEngine_PTY_UnresolvedSendVarErrors: a send referencing a ${name} nothing
// defines is an explained step error before any I/O, not the literal reference
// typed into the program (#78) — mirroring the run.command guard so a typo'd
// store name fails at the mistake, not garbled downstream.
func TestEngine_PTY_UnresolvedSendVarErrors(t *testing.T) {
	skipOnWindows(t)
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: v
scenarios:
  - name: typo in a send variable
    steps:
      - pty:
          command: cat
          timeout: 10s
          session:
            - send: "${no_such_var}\n"
`)
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error: %+v", res.Status, res.Scenarios)
	}
	msg := res.Scenarios[0].Steps[0].ErrMsg
	if !strings.Contains(msg, "${no_such_var}") || !strings.Contains(msg, "$${no_such_var}") {
		t.Errorf("error should name the reference and the $${...} literal escape, got %q", msg)
	}
}

// TestEngine_PTY_UnresolvedEnvVarErrors: an unset ${env:NAME} in a send fails
// the same way — a forgotten `env:` wiring must not feed a placeholder to a
// password prompt (#78).
func TestEngine_PTY_UnresolvedEnvVarErrors(t *testing.T) {
	skipOnWindows(t)
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: v
scenarios:
  - name: unset env in a send
    steps:
      - pty:
          command: cat
          timeout: 10s
          session:
            - send: "${env:ATAGO_SURELY_UNSET_78}\n"
`)
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error: %+v", res.Status, res.Scenarios)
	}
	msg := res.Scenarios[0].Steps[0].ErrMsg
	if !strings.Contains(msg, "ATAGO_SURELY_UNSET_78") {
		t.Errorf("error should name the env variable, got %q", msg)
	}
}

// TestEngine_PTY_UnresolvedExpectVarErrors: the guard covers expect patterns
// too — an unresolved reference in an expect would silently match the literal
// text and pass, so it must fail identically (#78).
func TestEngine_PTY_UnresolvedExpectVarErrors(t *testing.T) {
	skipOnWindows(t)
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: v
scenarios:
  - name: typo in an expect variable
    steps:
      - pty:
          command: cat
          timeout: 10s
          session:
            - send: "hi\n"
            - expect: "${no_such_var}"
`)
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error: %+v", res.Status, res.Scenarios)
	}
	msg := res.Scenarios[0].Steps[0].ErrMsg
	if !strings.Contains(msg, "${no_such_var}") {
		t.Errorf("error should name the reference, got %q", msg)
	}
}

// TestEngine_PTY_EscapedLiteralSendTypesLiteral: a $${...} escaped reference in
// a send still types the literal ${...} into the program (this is how recorded
// sessions carry literal `${`), so the guard must not touch it (#78).
func TestEngine_PTY_EscapedLiteralSendTypesLiteral(t *testing.T) {
	skipOnWindows(t)
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: v
scenarios:
  - name: escaped literal is typed verbatim
    steps:
      - pty:
          command: cat
          timeout: 10s
          session:
            - send: "$${literal}\n"
            - expect: '\$\{literal\}'
            - send: ""
      - assert:
          exit_code: 0
          stdout:
            contains: "${literal}"
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios[0].Steps)
	}
}

// TestEngine_PTY_NamedKeys proves named keys transmit their documented xterm
// bytes (#26): `cat -v` renders the down arrow it received as ^[[B, ctrl-d
// ends the stream, and a trap-based shell observes ctrl-c as SIGINT.
func TestEngine_PTY_NamedKeys(t *testing.T) {
	skipOnWindows(t)
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: cat -v shows the arrow bytes and ctrl-d ends input
    steps:
      - pty:
          command: cat -v
          session:
            - send: { key: down }
            - send: { key: enter }
            - expect: '\^\[\[B'
            - send: { key: ctrl-d }
      - assert:
          exit_code: 0
          stdout:
            contains: "^[[B"
  - name: ctrl-c delivers SIGINT
    steps:
      - pty:
          shell: true
          command: "trap 'exit 130' INT; echo waiting; while true; do sleep 0.1; done"
          session:
            - expect: waiting
            - send: { key: ctrl-c }
      - assert:
          exit_code: 130
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}

// TestEngine_PTY_ScreenAssert proves the screen target (#27): an overwritten
// line asserts on its FINAL rendered text (the transcript still carries
// both), line.n addresses screen rows, and a screen assert without a pty
// step is a load-time error.
func TestEngine_PTY_ScreenAssert(t *testing.T) {
	skipOnWindows(t)
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: the screen shows only the final overwrite
    steps:
      - pty:
          shell: true
          command: "printf 'loading...\r'; printf 'done.      \r\n'; printf 'row two\r\n'"
      - assert:
          screen:
            contains: "done."
      - assert:
          screen:
            not_contains: "loading"
      - assert:
          screen:
            line: 2
            equals: "row two"
      # The raw transcript keeps BOTH versions - the emulator adds the value.
      - assert:
          stdout:
            contains: "loading"
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}
