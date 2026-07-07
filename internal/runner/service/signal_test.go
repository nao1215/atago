package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/spec"
)

// TestSignal_TrapReceivesTERM proves a named signal reaches the service: a
// trap handler writes a marker and exits, and WaitExit observes the exit so
// a later Stop is a clean no-op (#23).
func TestSignal_TrapReceivesTERM(t *testing.T) {
	skipWindows(t)
	wd := t.TempDir()
	svc := &spec.Service{
		Name:    "graceful",
		Shell:   spec.Bool(true),
		Command: `trap 'echo graceful-shutdown > marker.txt; exit 0' TERM; echo ready; while true; do sleep 0.1; done`,
		Ready:   &spec.Ready{Log: "ready", Timeout: "5s"},
	}
	p, _, err := Start(context.Background(), svc, wd)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer p.Stop()
	if err := p.Signal("TERM"); err != nil {
		t.Fatalf("Signal(TERM) error = %v", err)
	}
	if !p.WaitExit(5 * time.Second) {
		t.Fatal("service did not exit within 5s after SIGTERM")
	}
	b, err := os.ReadFile(filepath.Join(wd, "marker.txt"))
	if err != nil {
		t.Fatalf("marker not written by the trap handler: %v", err)
	}
	if !strings.Contains(string(b), "graceful-shutdown") {
		t.Errorf("marker = %q, want graceful-shutdown", b)
	}
	p.Stop() // double-teardown safety after a signaled exit
}

// TestSignal_GroupDelivery proves the signal reaches backgrounded children,
// not just the shell leader (#23).
func TestSignal_GroupDelivery(t *testing.T) {
	skipWindows(t)
	wd := t.TempDir()
	svc := &spec.Service{
		Name:  "family",
		Shell: spec.Bool(true),
		// The backgrounded child writes its own marker when TERM arrives — it only
		// can if the signal went to the whole group, not just the shell leader.
		//
		// The child prints the readiness marker itself, AFTER installing its trap,
		// so "ready" proves the handler is armed. Emitting it from the parent (before
		// the backgrounded subshell had run `trap`) raced: a TERM delivered in that
		// window hit the default action and the child died without writing child.txt,
		// flaking the test on a loaded runner.
		Command: `( trap 'echo child-got-term > child.txt; exit 0' TERM; echo ready; while true; do sleep 0.1; done ) & wait`,
		Ready:   &spec.Ready{Log: "ready", Timeout: "5s"},
	}
	p, _, err := Start(context.Background(), svc, wd)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer p.Stop()
	if err := p.Signal("TERM"); err != nil {
		t.Fatalf("Signal(TERM) error = %v", err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		if b, err := os.ReadFile(filepath.Join(wd, "child.txt")); err == nil && strings.Contains(string(b), "child-got-term") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("backgrounded child never saw SIGTERM (group delivery broken)")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// TestSignal_KILLStopsATermIgnorer proves the KILL path works where TERM is
// trapped and ignored (#23).
func TestSignal_KILLStopsATermIgnorer(t *testing.T) {
	skipWindows(t)
	wd := t.TempDir()
	svc := &spec.Service{
		Name:    "stubborn",
		Shell:   spec.Bool(true),
		Command: `trap '' TERM; echo ready; while true; do sleep 0.2; done`,
		Ready:   &spec.Ready{Log: "ready", Timeout: "5s"},
	}
	p, _, err := Start(context.Background(), svc, wd)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer p.Stop()
	if err := p.Signal("TERM"); err != nil {
		t.Fatalf("Signal(TERM) error = %v", err)
	}
	if p.WaitExit(300 * time.Millisecond) {
		t.Fatal("TERM-ignoring service exited on TERM; the fixture is broken")
	}
	if err := p.Signal("KILL"); err != nil {
		t.Fatalf("Signal(KILL) error = %v", err)
	}
	if !p.WaitExit(5 * time.Second) {
		t.Fatal("service did not exit within 5s after SIGKILL")
	}
}

// TestSignal_AlreadyExited proves signaling a dead service is a named error,
// not a stray ESRCH (#23).
func TestSignal_AlreadyExited(t *testing.T) {
	skipWindows(t)
	wd := t.TempDir()
	svc := &spec.Service{Name: "gone", Shell: spec.Bool(true), Command: "echo bye"}
	p, _, err := Start(context.Background(), svc, wd)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if !p.WaitExit(5 * time.Second) {
		t.Fatal("echo service did not exit")
	}
	err = p.Signal("TERM")
	if err == nil || !strings.Contains(err.Error(), `service "gone" already exited`) {
		t.Errorf("Signal on exited service = %v, want an already-exited error naming the service", err)
	}
}

// TestSignal_UnknownName proves an unmapped signal name errors with the
// accepted set (defense in depth behind the loader's validation).
func TestSignal_UnknownName(t *testing.T) {
	skipWindows(t)
	wd := t.TempDir()
	svc := &spec.Service{Name: "s", Shell: spec.Bool(true), Command: sleepCmd(5)}
	p, _, err := Start(context.Background(), svc, wd)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer p.Stop()
	if err := p.Signal("WINCH"); err == nil || !strings.Contains(err.Error(), "accepted") {
		t.Errorf("Signal(WINCH) = %v, want an unknown-signal error listing the accepted set", err)
	}
}
