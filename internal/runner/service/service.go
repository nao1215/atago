// Package service runs the background processes a scenario declares under
// `services`: a long-lived peer (a TCP server, an API stub) started
// before the scenario's steps and torn down — with its whole process group —
// when the scenario ends. It reuses the command runner's tokenization and
// environment handling so a service spawns identically to a run step.
package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/nao1215/atago/internal/runner/cmd"
	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
)

// errExitedEarly is the readiness failure when the service process exits before
// its probe (delay/file/port/log) succeeds. Shared so every wait path words it
// identically.
var errExitedEarly = errors.New("service exited before it became ready")

// defaultReadyTimeout bounds a readiness probe when the spec omits one.
const defaultReadyTimeout = 5 * time.Second

// pollInterval is how often a file/port/log probe re-checks.
const pollInterval = 20 * time.Millisecond

// Proc is a running background service. Stop terminates it and its children.
type Proc struct {
	name string
	c    *processCmd
	out  *syncBuffer
	done chan struct{}
}

// Name returns the service's declared name.
func (p *Proc) Name() string { return p.name }

// Output returns whatever the service has written to stdout/stderr so far. It is
// used to enrich a readiness failure message.
func (p *Proc) Output() string { return p.out.String() }

// Start spawns svc in workdir and blocks until its readiness probe passes. The
// returned captured value is the trimmed content of the ready file when
// Ready.File and Ready.Store are both set (otherwise ""); the caller stores it
// as ${Ready.Store}. On any error the (partially started) process is stopped
// before returning.
func Start(ctx context.Context, svc *spec.Service, workdir string) (*Proc, string, error) {
	name, args, err := cmd.CommandLine(svc.Command, svc.ShellEnabled())
	if err != nil {
		return nil, "", fmt.Errorf("service %q: %w", svc.Name, err)
	}

	out := &syncBuffer{}
	pc := newProcessCmd(ctx, name, args)
	if svc.ShellEnabled() {
		// On Windows this hands cmd.exe the raw command line; Go's default argv
		// escaping would corrupt embedded quotes (see cmd.ConfigureShell).
		cmd.ConfigureShell(pc.cmd, svc.Command)
	}
	pc.cmd.Dir = cmd.ResolveDir(workdir, svc.Cwd)
	pc.cmd.Env = cmd.BuildEnv(svc.Env, svc.ClearEnvEnabled(), svc.PassEnv, nil)
	pc.cmd.Stdout = out
	pc.cmd.Stderr = out

	if err := pc.cmd.Start(); err != nil {
		return nil, "", fmt.Errorf("service %q: failed to start %q: %w", svc.Name, svc.Command, err)
	}
	// Tie the process to its teardown mechanism now that it has a pid: a job
	// object on Windows, a no-op on POSIX (the process group was set at spawn).
	// A failure is non-fatal — the service runs and teardown degrades to a
	// single-process kill — so it must not abort a working start.
	_ = pc.started()

	p := &Proc{name: svc.Name, c: pc, out: out, done: make(chan struct{})}
	go func() {
		_ = pc.cmd.Wait()
		close(p.done)
	}()

	captured, err := p.waitReady(ctx, svc.Ready, workdir)
	if err != nil {
		p.Stop()
		// Return the stopped Proc alongside the error so the caller can still read
		// its captured output — e.g. to preserve it as a durable log artifact
		// (#51). Its output buffer survives Stop.
		// Only append the service-output block when the service actually produced
		// output; a dangling "--- service output ---" header with nothing after it
		// is noise for a service that printed nothing (issue #19).
		if output := strings.TrimSpace(out.String()); output != "" {
			return p, "", fmt.Errorf("service %q not ready: %w\n--- service output ---\n%s", svc.Name, err, out.String())
		}
		return p, "", fmt.Errorf("service %q not ready: %w", svc.Name, err)
	}
	return p, captured, nil
}

// Stop terminates the service's process group: a graceful signal first, then a
// hard kill if it does not exit within a short grace period. It is safe to call
// once; subsequent calls are no-ops because the process has already exited.
func (p *Proc) Stop() {
	if p == nil || p.c == nil {
		return
	}
	select {
	case <-p.done:
		return // already exited
	default:
	}
	p.c.terminate() // graceful (SIGTERM to the group on unix)
	select {
	case <-p.done:
	case <-time.After(2 * time.Second):
		p.c.kill() // hard (SIGKILL to the group on unix)
		<-p.done
	}
}

// Signal delivers the named POSIX signal (TERM, INT, HUP, USR1, USR2, KILL)
// to the service's whole process group (#23), consistent with the teardown
// kill semantics. Signaling a service that already exited is an error naming
// the service — a graceful-shutdown spec that signals a dead process is
// asserting against nothing.
func (p *Proc) Signal(name string) error {
	if p == nil || p.c == nil {
		return fmt.Errorf("service is not running")
	}
	select {
	case <-p.done:
		return fmt.Errorf("service %q already exited", p.name)
	default:
	}
	return p.c.signalByName(name)
}

// WaitExit blocks until the service's process exits or timeout elapses,
// reporting whether it exited (#23). The exit is observed through the same
// done channel Stop uses, so a later teardown Stop is a clean no-op.
func (p *Proc) WaitExit(timeout time.Duration) bool {
	if p == nil || p.c == nil {
		return true
	}
	select {
	case <-p.done:
		return true
	case <-time.After(timeout):
		return false
	}
}

// waitReady blocks until the readiness probe passes or its timeout elapses. A
// nil probe means "ready as soon as it is spawned".
func (p *Proc) waitReady(ctx context.Context, r *spec.Ready, workdir string) (string, error) {
	if r == nil {
		return "", nil
	}
	timeout := defaultReadyTimeout
	if r.Timeout != "" {
		d, err := time.ParseDuration(r.Timeout)
		if err != nil {
			return "", fmt.Errorf("invalid ready.timeout %q: %w", r.Timeout, err)
		}
		timeout = d
	}

	switch {
	case r.Delay != "":
		d, err := time.ParseDuration(r.Delay)
		if err != nil {
			return "", fmt.Errorf("invalid ready.delay %q: %w", r.Delay, err)
		}
		// Timeout is the ceiling on any readiness wait (documented on ready), so a
		// delay longer than the timeout can never be reached — wait only up to the
		// timeout and report the misconfiguration instead of stalling for the full
		// delay (a CI-hang hazard). A non-positive timeout means "unbounded".
		wait, cappedByTimeout := d, false
		if timeout > 0 && timeout < d {
			wait, cappedByTimeout = timeout, true
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-p.done:
			// The process exited during the delay window: it is not ready, matching
			// how the file/port/log probes detect an early exit via p.done. Without
			// this, a command that crashes (e.g. `exit 3`) during the delay was
			// reported READY and the scenario ran against a dead peer.
			return "", errExitedEarly
		case <-time.After(wait):
			if cappedByTimeout {
				return "", fmt.Errorf("timed out after %s waiting for readiness (ready.delay %s exceeds ready.timeout)", timeout, r.Delay)
			}
			// Delay elapsed with the process still running — unless it exited in the
			// same instant; check once more so a crash at the boundary is not missed.
			select {
			case <-p.done:
				return "", errExitedEarly
			default:
				return "", nil
			}
		}
	case r.File != "":
		path, err := security.ResolveWorkdirPath("service.ready.file", workdir, r.File)
		if err != nil {
			return "", err
		}
		return p.waitFile(ctx, path, r.Store, timeout)
	case r.Port != "":
		// A bare port ("9997") makes every dial fail with "missing port in
		// address", which the probe swallows and then runs to the full timeout —
		// a slow, misleading failure. Reject a host-less value up front with a
		// clear message (":9997" for any host, "127.0.0.1:9997" for loopback).
		if _, _, err := net.SplitHostPort(r.Port); err != nil {
			return "", fmt.Errorf("invalid ready.port %q (use host:port, e.g. 127.0.0.1:8080 or :8080): %w", r.Port, err)
		}
		dialer := net.Dialer{Timeout: pollInterval}
		return "", p.poll(ctx, timeout, func() bool {
			conn, err := dialer.DialContext(ctx, "tcp", r.Port)
			if err != nil {
				return false
			}
			_ = conn.Close()
			return true
		})
	case r.Log != "":
		re, err := regexp.Compile(r.Log)
		if err != nil {
			return "", fmt.Errorf("invalid ready.log regexp %q: %w", r.Log, err)
		}
		return "", p.poll(ctx, timeout, func() bool { return re.MatchString(p.out.String()) })
	default:
		return "", nil
	}
}

// waitFile waits until path exists and is non-empty, then returns its trimmed
// content when store is set (so the caller can bind a dynamic address).
func (p *Proc) waitFile(ctx context.Context, path, store string, timeout time.Duration) (string, error) {
	if err := p.poll(ctx, timeout, func() bool {
		fi, err := os.Stat(path)
		return err == nil && fi.Size() > 0
	}); err != nil {
		return "", err
	}
	if store == "" {
		return "", nil
	}
	data, err := os.ReadFile(path) //nolint:gosec // path is the user-declared ready file under the scenario workdir
	if err != nil {
		return "", fmt.Errorf("read ready file %q: %w", path, err)
	}
	return strings.TrimSpace(string(data)), nil
}

// poll calls check every pollInterval until it returns true, the service exits,
// the timeout elapses, or ctx is canceled. A non-positive timeout means
// "unbounded" (matching the documented ready semantics and the delay probe): the
// deadline channel stays nil so it never fires, and the wait is bounded only by
// the process staying alive or the scenario context. Without this, ready.timeout
// "0" made the file/port/log probes fail immediately via a zero-duration timer,
// contradicting the delay probe which already treated 0 as unbounded.
func (p *Proc) poll(ctx context.Context, timeout time.Duration, check func() bool) error {
	var deadlineC <-chan time.Time
	if timeout > 0 {
		deadline := time.NewTimer(timeout)
		defer deadline.Stop()
		deadlineC = deadline.C
	}
	tick := time.NewTicker(pollInterval)
	defer tick.Stop()
	for {
		if check() {
			return nil
		}
		select {
		case <-p.done:
			// One last check: the signal we waited for may have appeared as the
			// process exited (e.g. a single-shot server that writes its ready file
			// then keeps serving, or that already finished).
			if check() {
				return nil
			}
			return errExitedEarly
		case <-deadlineC:
			return fmt.Errorf("timed out after %s waiting for readiness", timeout)
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
		}
	}
}

// syncBuffer is a bytes.Buffer safe for concurrent writes (the process's output
// goroutine) and reads (the log readiness probe).
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}
