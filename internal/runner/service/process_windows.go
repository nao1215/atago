//go:build windows

package service

import (
	"context"
	"errors"
	"os/exec"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// processCmd wraps an *exec.Cmd and, once started, ties the process to a
// kill-on-close job object. Windows has no process groups, so a service launched
// with `shell: true` that forks further children would orphan them on a bare
// process kill (the previous behavior). A job object terminates the whole tree
// at once, and JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE also reaps any survivors if
// atago itself exits without a clean teardown — the closest Windows analog to
// the POSIX runner's process-group kill.
type processCmd struct {
	cmd *exec.Cmd

	mu   sync.Mutex
	job  windows.Handle
	down bool
}

func newProcessCmd(ctx context.Context, name string, args []string) *processCmd {
	c := exec.CommandContext(ctx, name, args...) //nolint:gosec // executing user-declared service commands is the purpose of atago
	c.WaitDelay = 5 * time.Second
	p := &processCmd{cmd: c}
	// On scenario-context cancellation, tear down the whole tree, not just the
	// leader — mirroring the POSIX runner's process-group kill on cancel.
	c.Cancel = func() error { p.killTree(); return nil }
	return p
}

// started ties the freshly-started process to a kill-on-close job object. Start
// calls it right after cmd.Start(); every child the service spawns afterward is
// captured by the job automatically. A child forked in the microseconds between
// Start and this assignment would escape, but real services fork during their
// own initialization, well after this returns. A failure here is non-fatal: the
// service still runs, and killTree falls back to a single-process kill.
func (p *processCmd) started() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cmd.Process == nil || p.down {
		return nil
	}
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return err
	}
	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
		BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
			LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE,
		},
	}
	if _, err := windows.SetInformationJobObject(
		job,
		uint32(windows.JobObjectExtendedLimitInformation),
		uintptr(unsafe.Pointer(&info)), //nolint:gosec // documented x/sys pattern: pass the struct address to a uintptr-typed syscall arg
		uint32(unsafe.Sizeof(info)),
	); err != nil {
		_ = windows.CloseHandle(job)
		return err
	}
	h, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(p.cmd.Process.Pid)) //nolint:gosec // pid fits a uint32 on Windows
	if err != nil {
		_ = windows.CloseHandle(job)
		return err
	}
	defer windows.CloseHandle(h)
	if err := windows.AssignProcessToJobObject(job, h); err != nil {
		_ = windows.CloseHandle(job)
		return err
	}
	p.job = job
	return nil
}

// terminate and kill are identical on Windows: there is no SIGTERM to deliver, so
// a "graceful" stop is the same hard job termination as the escalated one. This
// mirrors the pre-existing Windows behavior (both were a process kill); the only
// change is that the whole tree now goes down, not just the leader.
func (p *processCmd) terminate() { p.killTree() }
func (p *processCmd) kill()      { p.killTree() }

// killTree terminates every process in the job and releases the handle. It is
// idempotent: the graceful-then-hard teardown in service.Stop calls it twice,
// and once the tree is down the second call is a no-op.
func (p *processCmd) killTree() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.down {
		return
	}
	p.down = true
	if p.job != 0 {
		_ = windows.TerminateJobObject(p.job, 1)
		_ = windows.CloseHandle(p.job)
		p.job = 0
		return
	}
	// The job was never assigned (started() failed or ran before the process
	// existed): fall back to a single-process kill so teardown still stops the
	// leader.
	if p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
	}
}

// signalByName rejects signal steps on Windows (#23): there are no POSIX signals
// to deliver. The loader accepts the step everywhere; execution reports a clear
// error, mirroring the pty runner's contract.
func (p *processCmd) signalByName(string) error {
	return errors.New("signal steps are not supported on Windows (POSIX-only; gate the scenario with `skip: {os: windows}`)")
}
