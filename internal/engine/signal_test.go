package engine

import (
	"runtime"
	"strings"
	"testing"
)

// TestEngine_SignalGracefulShutdown proves the full #23 flow: a trap-based
// service receives SIGTERM from a signal step, the wait observes its exit,
// and the marker file it wrote is assertable — all race-free under the
// default parallelism because the target is atago's own service handle.
func TestEngine_SignalGracefulShutdown(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("signal steps are POSIX-only")
	}
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: graceful shutdown on SIGTERM
    services:
      - name: server
        shell: true
        command: "trap 'echo graceful shutdown complete > server.log; exit 0' TERM; echo ready; while true; do sleep 0.1; done"
        ready: {log: ready, timeout: 5s}
    steps:
      - signal:
          service: server
          signal: TERM
          wait:
            timeout: 5s
      - assert:
          file:
            path: server.log
            contains: "graceful shutdown complete"
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}

// TestEngine_SignalWaitTimeoutFails proves a TERM-ignoring service fails the
// step with the documented message when wait elapses (#23).
func TestEngine_SignalWaitTimeoutFails(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("signal steps are POSIX-only")
	}
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: stubborn service outlives the wait
    services:
      - name: stubborn
        shell: true
        command: "trap '' TERM; echo ready; while true; do sleep 0.2; done"
        ready: {log: ready, timeout: 5s}
    steps:
      - signal:
          service: stubborn
          signal: TERM
          wait:
            timeout: 300ms
`)
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error: %+v", res.Status, res.Scenarios)
	}
	var msg string
	for _, sc := range res.Scenarios {
		for _, st := range sc.Steps {
			if st.ErrMsg != "" {
				msg = st.ErrMsg
			}
		}
	}
	if !strings.Contains(msg, `service "stubborn" did not exit within 300ms after SIGTERM`) {
		t.Errorf("ErrMsg = %q, want the documented did-not-exit message", msg)
	}
}

// TestEngine_SignalSuiteService proves a scenario's signal step can target a
// suite-wide service started by suite.setup (#23).
func TestEngine_SignalSuiteService(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("signal steps are POSIX-only")
	}
	res := runSpec(t, `
version: "1"
suite:
  name: s
  setup:
    - service:
        name: shared
        shell: true
        command: "trap 'exit 0' TERM; echo ready; while true; do sleep 0.1; done"
        ready: {log: ready, timeout: 5s}
scenarios:
  - name: signal the suite service
    steps:
      - signal:
          service: shared
          signal: TERM
          wait:
            timeout: 5s
      - run: {shell: true, command: echo done}
      - assert:
          exit_code: 0
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}
