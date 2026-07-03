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
