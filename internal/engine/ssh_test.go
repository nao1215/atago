package engine

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	sshd "github.com/gliderlabs/ssh"

	"github.com/nao1215/atago/internal/loader"
)

// sshTestServer starts an in-process SSH server that interprets a few commands,
// so SSH-runner behavior can be tested hermetically with no external sshd.
func sshTestServer(t *testing.T) string {
	t.Helper()
	handler := func(s sshd.Session) {
		cmd := s.Command()
		if len(cmd) == 0 {
			_ = s.Exit(0)
			return
		}
		switch cmd[0] {
		case "echo":
			_, _ = io.WriteString(s, strings.Join(cmd[1:], " ")+"\n")
			_ = s.Exit(0)
		case "exit":
			code := 0
			if len(cmd) > 1 {
				code, _ = strconv.Atoi(cmd[1])
			}
			_ = s.Exit(code)
		case "fail":
			_, _ = io.WriteString(s.Stderr(), "boom\n")
			_ = s.Exit(2)
		case "sleep":
			d, _ := time.ParseDuration(cmd[1])
			select {
			case <-time.After(d):
				_ = s.Exit(0)
			case <-s.Context().Done():
				return
			}
		default:
			_, _ = io.WriteString(s.Stderr(), "unknown command\n")
			_ = s.Exit(127)
		}
	}
	srv := &sshd.Server{Handler: handler}
	srv.PasswordHandler = func(_ sshd.Context, _ string) bool { return true }

	var lc net.ListenConfig
	ln, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() {
		_ = srv.Close()
		_ = ln.Close()
	})
	return ln.Addr().String()
}

func TestEngine_SSHRun(t *testing.T) {
	t.Parallel()
	addr := sshTestServer(t)
	src := fmt.Sprintf(`
version: "1"
suite:
  name: ssh
runners:
  box:
    type: ssh
    host: %s
    user: tester
    password: secret
    insecure_host_key: true
scenarios:
  - name: remote echo, exit code, and stderr
    steps:
      - run:
          runner: box
          command: echo hello-remote
      - assert:
          exit_code: 0
      - assert:
          stdout:
            contains: hello-remote
      - run:
          runner: box
          command: exit 3
      - assert:
          exit_code: 3
      - run:
          runner: box
          command: fail
      - assert:
          exit_code: 2
      - assert:
          stderr:
            contains: boom
`, addr)
	res := runSpec(t, src)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios[0].Steps)
	}
}

// TestEngine_SSHStepTimeoutHonored closes the silent no-op the loader comment
// promised away: validateSSHRunFields whitelists `timeout` on an ssh run step
// "because it is honored remotely", but the engine only ever applied the
// RUNNER-level timeout captured at dial — a step-level `timeout:` was parsed,
// validated, and ignored. It must bound the remote command, and — mirroring the
// cmd runner (#17) — a fired timeout is an OBSERVABLE TimedOut result naming
// its source, not a hard scenario error.
func TestEngine_SSHStepTimeoutHonored(t *testing.T) {
	t.Parallel()
	addr := sshTestServer(t)
	src := fmt.Sprintf(`
version: "1"
suite:
  name: ssh
runners:
  box:
    type: ssh
    host: %s
    user: tester
    password: secret
    insecure_host_key: true
scenarios:
  - name: a step timeout bounds the remote command
    steps:
      - run:
          runner: box
          command: sleep 10s
          timeout: 300ms
      - assert:
          exit_code: {not: 0}
`, addr)
	res := runSpec(t, src)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed (a timeout is an observable result the assert inspects): %+v", res.Status, res.Scenarios[0].Steps)
	}
	run := res.Scenarios[0].Steps[0].Run
	if run == nil || !run.TimedOut {
		t.Fatalf("step result = %+v, want TimedOut", run)
	}
	if run.TimeoutSource != "run.timeout" {
		t.Errorf("TimeoutSource = %q, want %q so the failure hint names the knob", run.TimeoutSource, "run.timeout")
	}
	if run.Duration >= 5*time.Second {
		t.Errorf("Duration = %s, want the 300ms step timeout to have cut the 10s sleep", run.Duration)
	}
}

// TestEngine_SSHRunnerTimeoutIsObservable pins the parity half: a RUNNER-level
// ssh timeout that fires mid-command also produces a TimedOut result (source
// "runner.timeout"), matching how the cmd runner reports local timeouts,
// instead of the previous hard "ssh command timed out" scenario error.
func TestEngine_SSHRunnerTimeoutIsObservable(t *testing.T) {
	t.Parallel()
	addr := sshTestServer(t)
	src := fmt.Sprintf(`
version: "1"
suite:
  name: ssh
runners:
  box:
    type: ssh
    host: %s
    user: tester
    password: secret
    insecure_host_key: true
    timeout: 300ms
scenarios:
  - name: the runner timeout bounds every remote command
    steps:
      - run:
          runner: box
          command: sleep 10s
      - assert:
          exit_code: {not: 0}
`, addr)
	res := runSpec(t, src)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios[0].Steps)
	}
	run := res.Scenarios[0].Steps[0].Run
	if run == nil || !run.TimedOut {
		t.Fatalf("step result = %+v, want TimedOut", run)
	}
	if run.TimeoutSource != "runner.timeout" {
		t.Errorf("TimeoutSource = %q, want %q", run.TimeoutSource, "runner.timeout")
	}
}

func TestEngine_SSHValueBinding(t *testing.T) {
	t.Parallel()
	addr := sshTestServer(t)
	src := fmt.Sprintf(`
version: "1"
suite:
  name: ssh
runners:
  box:
    type: ssh
    host: %s
    user: tester
    password: secret
    insecure_host_key: true
scenarios:
  - name: capture remote output and reuse it
    steps:
      - run:
          runner: box
          command: echo token42
      - store:
          name: tok
          from:
            stdout:
              matches: "(token[0-9]+)"
      - run:
          runner: box
          command: echo got-${tok}
      - assert:
          stdout:
            contains: got-token42
`, addr)
	res := runSpec(t, src)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios[0].Steps)
	}
}

func TestEngine_SSHUnknownRunner(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: ssh
scenarios:
  - name: run references an undeclared runner
    steps:
      - run:
          runner: missing
          command: echo hi
`
	// A run step naming an unknown runner is rejected at load time rather than
	// silently running locally — a typo'd runner must not pass a remote test by
	// accident.
	if _, err := loader.LoadBytes("t.atago.yaml", []byte(src)); err == nil || !strings.Contains(err.Error(), "is not declared") {
		t.Fatalf("LoadBytes() error = %v, want an undeclared-runner validation error", err)
	}
}
