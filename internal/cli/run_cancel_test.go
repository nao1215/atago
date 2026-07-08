//go:build !windows

package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/engine"
)

// TestRunSpecs_CanceledBeforeStart proves the context-aware run helper does not
// execute a long spec when handed an already-canceled context.
func TestRunSpecs_CanceledBeforeStart(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	specPath := filepath.Join(dir, "slow.atago.yaml")
	writeFile(t, specPath, `
version: "1"
suite:
  name: cancel
scenarios:
  - name: slow
    steps:
      - run: {shell: true, command: "sleep 30"}
`)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	results, loadErrs := runSpecs(ctx, engine.New(), []string{specPath})
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("runSpecs took %v with a pre-canceled context; expected a prompt return", elapsed)
	}
	if loadErrs[0] != nil {
		t.Fatalf("unexpected load error: %v", loadErrs[0])
	}
	// The single suite must not report a passing scenario.
	if r := results[0]; r != nil && len(r.Scenarios) > 0 && r.Scenarios[0].Status == "passed" {
		t.Fatalf("scenario passed despite a canceled context: %+v", r.Scenarios)
	}
}

// TestRunSpecs_ParallelCanceledBeforeStart proves the same guarantee on the
// suite-parallel path (`--parallel > 1` sets eng.Sem): with a pre-canceled
// context the dispatch loop must not hand out a single spec, so nothing is
// loaded — not merely "loaded but refused by the engine".
func TestRunSpecs_ParallelCanceledBeforeStart(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := `
version: "1"
suite:
  name: cancel
scenarios:
  - name: slow
    steps:
      - run: {shell: true, command: "sleep 30"}
`
	paths := make([]string, 4)
	for i := range paths {
		paths[i] = filepath.Join(dir, fmt.Sprintf("slow%d.atago.yaml", i))
		writeFile(t, paths[i], src)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	eng := engine.New()
	eng.Parallel = 2
	eng.Sem = make(chan struct{}, 2)

	start := time.Now()
	results, loadErrs := runSpecs(ctx, eng, paths)
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("parallel runSpecs took %v with a pre-canceled context; expected a prompt return", elapsed)
	}
	for i := range paths {
		if results[i] != nil {
			t.Errorf("spec %d was executed despite a pre-canceled context: %+v", i, results[i].Scenarios)
		}
		if loadErrs[i] != nil {
			t.Errorf("spec %d was loaded despite a pre-canceled context: %v", i, loadErrs[i])
		}
	}
}

// TestRunSpecs_ParallelCanceledMidRun cancels while the first spec's scenario
// is sleeping and asserts two things: the run unwinds long before the specs'
// natural completion, and — the point of the worker pool — the many specs
// queued behind the busy worker are never loaded at all once the interrupt
// lands, instead of every path getting a goroutine and a spec load up front.
func TestRunSpecs_ParallelCanceledMidRun(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := `
version: "1"
suite:
  name: cancel
scenarios:
  - name: slow
    steps:
      - run: {shell: true, command: "sleep 30"}
`
	paths := make([]string, 20)
	for i := range paths {
		paths[i] = filepath.Join(dir, fmt.Sprintf("slow%d.atago.yaml", i))
		writeFile(t, paths[i], src)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(300 * time.Millisecond)
		cancel()
	}()

	// One semaphore slot => one suite worker: spec 0 occupies it (sleeping)
	// while the dispatcher is blocked handing out spec 1. Everything after the
	// in-flight pair must stay untouched once the cancel lands.
	eng := engine.New()
	eng.Parallel = 1 // scenario workers per suite; the Sem below is the global cap
	eng.Sem = make(chan struct{}, 1)

	start := time.Now()
	results, loadErrs := runSpecs(ctx, eng, paths)
	if elapsed := time.Since(start); elapsed > 10*time.Second {
		t.Fatalf("parallel runSpecs took %v after mid-run cancellation; expected a prompt unwind", elapsed)
	}
	// A suite whose scenarios were all skipped by the interrupt still reports
	// StatusPassed at the suite level (runCmd adds the interrupt exit code), so
	// the per-scenario invariant is that nothing ran to a passing completion.
	for i, r := range results {
		if r == nil {
			continue
		}
		for _, sc := range r.Scenarios {
			if sc.Status == engine.StatusPassed {
				t.Fatalf("spec %d: scenario passed despite mid-run cancellation: %+v", i, sc)
			}
		}
	}
	// With one worker, only spec 0 (in flight) and at most spec 1 (already
	// handed to the dispatch channel) may have been touched; the rest must
	// never have been loaded or run.
	for i := 2; i < len(paths); i++ {
		if results[i] != nil || loadErrs[i] != nil {
			t.Errorf("spec %d was loaded/run after cancellation (result=%v, loadErr=%v)", i, results[i] != nil, loadErrs[i])
		}
	}
}

// TestRunCmd_InterruptExitsPromptly spawns the real `atago run` binary on a
// long-running spec, sends SIGINT, and asserts it exits well before the spec's
// natural completion — proving Ctrl-C propagates into the run graph.
func TestRunCmd_InterruptExitsPromptly(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a binary; skipped in -short")
	}

	dir := t.TempDir()
	bin := filepath.Join(dir, "atago")
	build := exec.CommandContext(context.Background(), "go", "build", "-o", bin, "github.com/nao1215/atago")
	build.Env = os.Environ()
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building atago: %v\n%s", err, out)
	}

	specPath := filepath.Join(dir, "slow.atago.yaml")
	writeFile(t, specPath, `
version: "1"
suite:
  name: cancel
scenarios:
  - name: slow
    steps:
      - run: {shell: true, command: "sleep 60"}
`)

	// A background context here: the interrupt under test is delivered via an OS
	// signal to the running process, not through context cancellation.
	cmd := exec.CommandContext(context.Background(), bin, "run", specPath)
	// Own process group so we can be sure our SIGINT targets the atago process.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatalf("starting atago: %v", err)
	}

	// Give it a moment to reach the sleeping step, then interrupt.
	time.Sleep(500 * time.Millisecond)
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("sending SIGINT: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		// A cancelled run exits non-zero; the point is that it exited at all,
		// promptly, rather than sleeping for 60s.
		if err == nil {
			t.Fatal("atago run exited 0 after SIGINT; expected a non-zero cancellation exit")
		}
	case <-time.After(15 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("atago run did not exit within 15s of SIGINT")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestRunCmd_InterruptRunsTeardownAndReportsPartials pins the full Ctrl-C
// promise from the README ("in-flight processes, services, and sessions are
// torn down, partial results are reported, and the run exits 4") — previously
// only "exits non-zero promptly" was verified, so a regression that skipped
// teardown on interrupt (leaking servers in every user's aborted CI run) would
// have passed the suite. The sentinel files land OUTSIDE the scenario workdir
// (which atago removes) via an env-passed directory.
func TestRunCmd_InterruptRunsTeardownAndReportsPartials(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a binary; skipped in -short")
	}

	dir := t.TempDir()
	sentinels := t.TempDir()
	bin := filepath.Join(dir, "atago")
	build := exec.CommandContext(context.Background(), "go", "build", "-o", bin, "github.com/nao1215/atago")
	build.Env = os.Environ()
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building atago: %v\n%s", err, out)
	}

	specPath := filepath.Join(dir, "slow.atago.yaml")
	writeFile(t, specPath, `
version: "1"
suite:
  name: cancel
scenarios:
  - name: slow with service and teardown
    services:
      - name: peer
        shell: true
        command: 'echo $$ > "${env:ATAGO_TEST_SENTINELS}/svc.pid"; echo service-armed; sleep 120'
        ready:
          log: service-armed
          timeout: 5s
    steps:
      - run: {shell: true, command: "sleep 60"}
    teardown:
      - run: {shell: true, command: 'echo cleaned > "${env:ATAGO_TEST_SENTINELS}/teardown.txt"'}
`)

	cmd := exec.CommandContext(context.Background(), bin, "run", specPath)
	cmd.Env = append(os.Environ(), "ATAGO_TEST_SENTINELS="+sentinels)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("starting atago: %v", err)
	}

	// Interrupt only after the service is up (its pid file exists) and the
	// scenario has had a moment to reach the sleeping step.
	pidFile := filepath.Join(sentinels, "svc.pid")
	deadline := time.Now().Add(10 * time.Second)
	for {
		if fi, err := os.Stat(pidFile); err == nil && fi.Size() > 0 {
			break
		}
		if time.Now().After(deadline) {
			_ = cmd.Process.Kill()
			t.Fatalf("service pid file never appeared; stderr=%s", stderr.String())
		}
		time.Sleep(20 * time.Millisecond)
	}
	time.Sleep(300 * time.Millisecond)
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("sending SIGINT: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	var waitErr error
	select {
	case waitErr = <-done:
	case <-time.After(15 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("atago run did not exit within 15s of SIGINT")
	}

	// Exit code 4 (ExitExec), per the README's stable contract.
	var exitErr *exec.ExitError
	if !errors.As(waitErr, &exitErr) || exitErr.ExitCode() != 4 {
		t.Errorf("exit = %v, want exit code 4 (ExitExec)\nstdout=%s\nstderr=%s", waitErr, stdout.String(), stderr.String())
	}

	// Teardown ran.
	if _, err := os.Stat(filepath.Join(sentinels, "teardown.txt")); err != nil {
		t.Errorf("teardown sentinel missing (%v): teardown did not run on interrupt\nstderr=%s", err, stderr.String())
	}

	// The service's process is gone (kill(pid, 0) fails once reaped).
	pidBytes, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("read pid file: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil {
		t.Fatalf("parse service pid %q: %v", pidBytes, err)
	}
	// The process may linger a beat while the group signal lands.
	gone := false
	for range 50 {
		if syscall.Kill(pid, 0) != nil {
			gone = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !gone {
		t.Errorf("service pid %d still alive after atago exited: services were not torn down", pid)
	}

	// Partial results were reported: the summary tally line still renders.
	if !strings.Contains(stdout.String(), "scenario") {
		t.Errorf("stdout = %q, want a partial report with the scenario tally", stdout.String())
	}
}
