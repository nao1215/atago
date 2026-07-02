package engine

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"

	sshd "github.com/gliderlabs/ssh"
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
	res := runHTTPSpec(t, src)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios[0].Steps)
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
	res := runHTTPSpec(t, src)
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
	res := runHTTPSpec(t, src)
	// A run step naming an unknown runner errors rather than silently running
	// locally — a typo'd runner must not pass a remote test by accident.
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error (unknown runner)", res.Status)
	}
}
