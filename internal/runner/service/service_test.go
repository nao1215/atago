package service

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/spec"
)

func skipWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("relies on POSIX process groups / signals; Windows teardown is covered by the portable tests")
	}
}

// The helpers below spell service commands for both shells (service.Shell maps
// to /bin/sh on POSIX and cmd.exe on Windows), so most tests run on every OS.

// sleepCmd returns a silent shell command that blocks ~sec seconds. Windows has
// no sleep; ping -n waits ~1s between echoes and >nul keeps it quiet.
func sleepCmd(sec int) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("ping -n %d 127.0.0.1 >nul", sec+1)
	}
	return fmt.Sprintf("sleep %d", sec)
}

// publishThenIdle writes content into the workdir-relative file, then idles —
// the way a server publishes its bound address and keeps serving. cmd.exe's
// echo adds a trailing space + CRLF, which ready.store trims.
func publishThenIdle(content, file string, sec int) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("echo %s >%s& %s", content, file, sleepCmd(sec))
	}
	return fmt.Sprintf("printf '%s' > %s; %s", content, file, sleepCmd(sec))
}

// echoThenIdle prints msg, then idles.
func echoThenIdle(msg string, sec int) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("echo %s& %s", msg, sleepCmd(sec))
	}
	return fmt.Sprintf("echo %s; %s", msg, sleepCmd(sec))
}

// TestStart_FileReadinessCapturesContent covers the canonical pattern: a server
// publishes its (ephemeral) address to a file, and the spec captures it.
func TestStart_FileReadinessCapturesContent(t *testing.T) {
	wd := t.TempDir()
	svc := &spec.Service{
		Name:    "addr",
		Shell:   spec.Bool(true),
		Command: publishThenIdle("127.0.0.1:54321", "ready.txt", 5),
		Ready:   &spec.Ready{File: "ready.txt", Store: "addr", Timeout: "5s"},
	}
	p, captured, err := Start(context.Background(), svc, wd)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer p.Stop()
	if captured != "127.0.0.1:54321" {
		t.Errorf("captured = %q, want 127.0.0.1:54321", captured)
	}
}

// TestStart_ReadyFileTraversalRejected proves a ready.file path may not point
// outside the scenario workdir.
func TestStart_ReadyFileTraversalRejected(t *testing.T) {
	wd := t.TempDir()
	svc := &spec.Service{
		Name:    "s",
		Shell:   spec.Bool(true),
		Command: sleepCmd(5),
		Ready:   &spec.Ready{File: "../ready.txt", Timeout: "5s"},
	}
	p, _, err := Start(context.Background(), svc, wd)
	if p != nil {
		p.Stop()
	}
	if err == nil {
		t.Fatal("ready.file escaping the workdir was accepted")
	}
	if !strings.Contains(err.Error(), "escapes the scenario workdir") {
		t.Errorf("error %q should explain the containment failure", err)
	}
}

// Regression for issue #19: a silent service that never becomes ready must not
// produce a dangling "--- service output ---" header with nothing after it.
func TestStart_ReadinessErrorOmitsEmptyOutput(t *testing.T) {
	wd := t.TempDir()
	svc := &spec.Service{
		Name:    "silent",
		Shell:   spec.Bool(true),
		Command: sleepCmd(5), // prints nothing, never creates the ready file
		Ready:   &spec.Ready{File: "never.txt", Timeout: "150ms"},
	}
	_, _, err := Start(context.Background(), svc, wd)
	if err == nil {
		t.Fatal("Start() error = nil, want readiness timeout")
	}
	if strings.Contains(err.Error(), "--- service output ---") {
		t.Errorf("readiness error has a dangling output header for a silent service:\n%s", err)
	}
	if !strings.Contains(err.Error(), "not ready") {
		t.Errorf("error should mention 'not ready': %v", err)
	}
}

// TestStart_DelayReadyDetectsEarlyExit is a regression: a `ready.delay` service
// whose command exits during the delay window must fail readiness, not be
// reported ready — otherwise the scenario runs against a dead peer. The
// file/port/log probes already detect early exit; the delay branch did not.
func TestStart_DelayReadyDetectsEarlyExit(t *testing.T) {
	wd := t.TempDir()
	svc := &spec.Service{
		Name:    "crasher",
		Shell:   spec.Bool(true),
		Command: "exit 3", // exits immediately, well within the delay
		Ready:   &spec.Ready{Delay: "2s"},
	}
	_, _, err := Start(context.Background(), svc, wd)
	if err == nil {
		t.Fatal("Start() error = nil, want 'exited before it became ready'")
	}
	if !strings.Contains(err.Error(), "exited before it became ready") {
		t.Errorf("error = %v, want it to report the early exit", err)
	}
}

// TestStart_BarePortReadyRejected is a regression: a host-less ready.port must
// fail fast with a clear message, not swallow "missing port in address" and run
// to the full readiness timeout.
func TestStart_BarePortReadyRejected(t *testing.T) {
	wd := t.TempDir()
	svc := &spec.Service{
		Name:    "svc",
		Shell:   spec.Bool(true),
		Command: sleepCmd(5),
		Ready:   &spec.Ready{Port: "9997", Timeout: "5s"}, // no host
	}
	start := time.Now()
	_, _, err := Start(context.Background(), svc, wd)
	if err == nil {
		t.Fatal("Start() error = nil, want an invalid-port error")
	}
	if !strings.Contains(err.Error(), "invalid ready.port") {
		t.Errorf("error = %v, want it to name the invalid ready.port", err)
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("took %s; a bare port must fail fast, not run to the readiness timeout", elapsed)
	}
}

// On a readiness failure, Start returns the stopped Proc so the caller can still
// read its captured output to preserve it as a log artifact (#51).
func TestStart_ReadinessFailureReturnsProcForLogCapture(t *testing.T) {
	wd := t.TempDir()
	svc := &spec.Service{
		Name:    "chatty",
		Shell:   spec.Bool(true),
		Command: echoThenIdle("booting-up", 5),
		Ready:   &spec.Ready{File: "never.txt", Timeout: "150ms"},
	}
	p, _, err := Start(context.Background(), svc, wd)
	if err == nil {
		t.Fatal("Start() error = nil, want readiness timeout")
	}
	if p == nil {
		t.Fatal("Start() proc = nil on readiness failure, want the stopped proc for log capture")
	}
	if p.Name() != "chatty" {
		t.Errorf("proc name = %q", p.Name())
	}
	if !strings.Contains(p.Output(), "booting-up") {
		t.Errorf("proc output = %q, want the captured output", p.Output())
	}
}

// A service that DID print output keeps the output block in its readiness error.
func TestStart_ReadinessErrorKeepsNonEmptyOutput(t *testing.T) {
	wd := t.TempDir()
	svc := &spec.Service{
		Name:    "chatty",
		Shell:   spec.Bool(true),
		Command: echoThenIdle("starting up", 5),
		Ready:   &spec.Ready{File: "never.txt", Timeout: "150ms"},
	}
	_, _, err := Start(context.Background(), svc, wd)
	if err == nil {
		t.Fatal("Start() error = nil, want readiness timeout")
	}
	if !strings.Contains(err.Error(), "--- service output ---") || !strings.Contains(err.Error(), "starting up") {
		t.Errorf("readiness error should include the captured output:\n%s", err)
	}
}

// TestStart_FileReadinessWithoutStore waits but captures nothing.
func TestStart_FileReadinessWithoutStore(t *testing.T) {
	wd := t.TempDir()
	svc := &spec.Service{
		Name:    "f",
		Shell:   spec.Bool(true),
		Command: publishThenIdle("up", "ready.txt", 5),
		Ready:   &spec.Ready{File: "ready.txt"},
	}
	p, captured, err := Start(context.Background(), svc, wd)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer p.Stop()
	if captured != "" {
		t.Errorf("captured = %q, want empty", captured)
	}
}

// TestStart_LogReadiness waits for a regexp on combined output.
func TestStart_LogReadiness(t *testing.T) {
	wd := t.TempDir()
	svc := &spec.Service{
		Name:    "log",
		Shell:   spec.Bool(true),
		Command: echoThenIdle("now listening on 127.0.0.1", 5),
		Ready:   &spec.Ready{Log: "listening on", Timeout: "5s"},
	}
	p, _, err := Start(context.Background(), svc, wd)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer p.Stop()
	if got := p.Output(); got == "" {
		t.Error("expected captured output, got empty")
	}
}

// TestStart_DelayReadiness waits a fixed duration.
func TestStart_DelayReadiness(t *testing.T) {
	wd := t.TempDir()
	start := time.Now()
	svc := &spec.Service{
		Name:    "d",
		Shell:   spec.Bool(true),
		Command: sleepCmd(5),
		Ready:   &spec.Ready{Delay: "100ms"},
	}
	p, _, err := Start(context.Background(), svc, wd)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer p.Stop()
	if elapsed := time.Since(start); elapsed < 100*time.Millisecond {
		t.Errorf("returned after %s, want >= 100ms", elapsed)
	}
}

// TestStart_NoReadinessProbe returns as soon as the process is spawned.
func TestStart_NoReadinessProbe(t *testing.T) {
	wd := t.TempDir()
	svc := &spec.Service{Name: "n", Shell: spec.Bool(true), Command: sleepCmd(5)}
	p, captured, err := Start(context.Background(), svc, wd)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer p.Stop()
	if captured != "" || p.Name() != "n" {
		t.Errorf("unexpected captured=%q name=%q", captured, p.Name())
	}
}

// TestStart_TimeoutWhenNeverReady fails when the readiness signal never arrives.
func TestStart_TimeoutWhenNeverReady(t *testing.T) {
	wd := t.TempDir()
	svc := &spec.Service{
		Name:    "slow",
		Shell:   spec.Bool(true),
		Command: sleepCmd(5),
		Ready:   &spec.Ready{File: "never.txt", Timeout: "150ms"},
	}
	_, _, err := Start(context.Background(), svc, wd)
	if err == nil {
		t.Fatal("Start() error = nil, want timeout error")
	}
}

// TestStart_ExitedBeforeReady fails fast (does not wait the full timeout) when
// the process dies before publishing readiness.
func TestStart_ExitedBeforeReady(t *testing.T) {
	wd := t.TempDir()
	start := time.Now()
	svc := &spec.Service{
		Name:    "dies",
		Shell:   spec.Bool(true),
		Command: `exit 1`,
		Ready:   &spec.Ready{File: "ready.txt", Timeout: "10s"},
	}
	_, _, err := Start(context.Background(), svc, wd)
	if err == nil {
		t.Fatal("Start() error = nil, want exited-before-ready error")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("took %s; should fail fast when the process exits", elapsed)
	}
}

// TestStart_InvalidCommand reports a start failure.
func TestStart_InvalidCommand(t *testing.T) {
	wd := t.TempDir()
	svc := &spec.Service{Name: "bad", Command: "definitely-not-a-real-binary-xyz"}
	_, _, err := Start(context.Background(), svc, wd)
	if err == nil {
		t.Fatal("Start() error = nil, want failure for a missing binary")
	}
}

// TestStart_BadDuration rejects an unparseable readiness duration.
func TestStart_BadDuration(t *testing.T) {
	wd := t.TempDir()
	svc := &spec.Service{
		Name:    "x",
		Shell:   spec.Bool(true),
		Command: sleepCmd(5),
		Ready:   &spec.Ready{Delay: "notaduration"},
	}
	if _, _, err := Start(context.Background(), svc, wd); err == nil {
		t.Fatal("expected error for bad delay duration")
	}
}

// TestStart_PortReadiness waits until a TCP port accepts a connection. A Go
// listener stands in for the service's socket; the service process itself just
// idles, so readiness is decided purely by the port probe.
func TestStart_PortReadiness(t *testing.T) {
	var lc net.ListenConfig
	ln, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			_ = c.Close()
		}
	}()

	wd := t.TempDir()
	svc := &spec.Service{
		Name:    "port",
		Shell:   spec.Bool(true),
		Command: sleepCmd(5),
		Ready:   &spec.Ready{Port: ln.Addr().String(), Timeout: "5s"},
	}
	p, _, err := Start(context.Background(), svc, wd)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer p.Stop()
}

// TestStart_PortReadinessTimeout times out when nothing listens.
func TestStart_PortReadinessTimeout(t *testing.T) {
	wd := t.TempDir()
	svc := &spec.Service{
		Name:    "noport",
		Shell:   spec.Bool(true),
		Command: sleepCmd(5),
		// Port 1 on localhost is not listening in the sandbox.
		Ready: &spec.Ready{Port: "127.0.0.1:1", Timeout: "150ms"},
	}
	if _, _, err := Start(context.Background(), svc, wd); err == nil {
		t.Fatal("expected timeout waiting for an unbound port")
	}
}

// TestStop_TerminatesChildren verifies the whole process group is torn down: a
// shell that backgrounds a child writing to a file must stop writing after Stop.
func TestStop_TerminatesChildren(t *testing.T) {
	skipWindows(t)
	wd := t.TempDir()
	marker := filepath.Join(wd, "tick")
	// A backgrounded grandchild appends to a file every 50ms. If Stop only killed
	// the shell leader, the grandchild would keep ticking.
	svc := &spec.Service{
		Name:    "ticker",
		Shell:   spec.Bool(true),
		Command: `( while true; do echo tick >> tick; sleep 0.05; done ) & echo ready; wait`,
		Ready:   &spec.Ready{Log: "ready", Timeout: "5s"},
	}
	p, _, err := Start(context.Background(), svc, wd)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	time.Sleep(150 * time.Millisecond)
	p.Stop()
	time.Sleep(100 * time.Millisecond) // let any survivor tick
	before, _ := os.ReadFile(marker)
	time.Sleep(300 * time.Millisecond)
	after, _ := os.ReadFile(marker)
	if len(after) != len(before) {
		t.Errorf("grandchild kept writing after Stop: %d -> %d bytes", len(before), len(after))
	}
}

// TestStop_EscalatesToKill covers the hard-kill path: a service that traps and
// ignores SIGTERM must still be torn down (Stop escalates to SIGKILL after the
// grace period).
func TestStop_EscalatesToKill(t *testing.T) {
	skipWindows(t)
	wd := t.TempDir()
	svc := &spec.Service{
		Name:  "stubborn",
		Shell: spec.Bool(true),
		// Ignore SIGTERM, then idle. Only SIGKILL can stop it.
		Command: `trap '' TERM; echo ready; while true; do sleep 1; done`,
		Ready:   &spec.Ready{Log: "ready", Timeout: "5s"},
	}
	p, _, err := Start(context.Background(), svc, wd)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	done := make(chan struct{})
	go func() { p.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(8 * time.Second):
		t.Fatal("Stop() did not escalate to SIGKILL for a SIGTERM-ignoring service")
	}
}

// TestStop_Idempotent is a no-op on a nil/stopped proc.
func TestStop_Idempotent(t *testing.T) {
	var nilProc *Proc
	nilProc.Stop() // must not panic

	wd := t.TempDir()
	svc := &spec.Service{Name: "n", Shell: spec.Bool(true), Command: sleepCmd(5)}
	p, _, err := Start(context.Background(), svc, wd)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	p.Stop()
	p.Stop() // second call is a no-op
}

// TestStart_ContextCancellation tears the service down when ctx is canceled.
func TestStart_ContextCancellation(t *testing.T) {
	wd := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	svc := &spec.Service{Name: "c", Shell: spec.Bool(true), Command: sleepCmd(30)}
	p, _, err := Start(ctx, svc, wd)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	cancel()
	done := make(chan struct{})
	go func() { p.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(8 * time.Second):
		t.Fatal("Stop() did not return after context cancellation")
	}
}
