//go:build windows

// Package conpty is a self-contained Windows pseudo-console (ConPTY) wrapper,
// shared by the pty runner (internal/runner/ptyrun) and interactive recording
// (internal/record). It calls the ConPTY and process-creation APIs directly
// through golang.org/x/sys/windows (already a dependency), so atago carries no
// third-party ConPTY library: the surface is small and the Win32 calls are
// stable since Windows 10 (1809). It follows Microsoft's documented
// pseudo-console recipe — CreatePseudoConsole over a pipe pair, a
// PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE attribute list, then CreateProcess with
// EXTENDED_STARTUPINFO_PRESENT.
package conpty

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"

	runnercmd "github.com/nao1215/atago/internal/runner/cmd"
)

// PseudoConsole wires a child process's console I/O to a pipe pair through a
// Windows pseudo console and exposes Read/Write/Resize/Wait/Close.
type PseudoConsole struct {
	hpc       windows.Handle // the pseudo console (HPCON)
	inWrite   windows.Handle // parent → child (sends)
	outRead   windows.Handle // child → parent (transcript)
	process   windows.Handle // child process handle, for wait/kill
	pid       uint32
	attrList  *windows.ProcThreadAttributeListContainer
	closeOnce sync.Once
}

// IsAvailable reports whether the host exposes the ConPTY API (Windows 10
// version 1809 and later), so an older host gets a clear error instead of a
// missing-proc failure deep in Start.
func IsAvailable() bool {
	return windows.NewLazySystemDLL("kernel32.dll").NewProc("CreatePseudoConsole").Find() == nil
}

// CommandLine builds the single command-line string ConPTY's CreateProcess
// receives. A shell command reuses cmd.exe's documented `/S /C "<command>"`
// contract (strip the outer quotes, run the rest verbatim) so it matches the cmd
// runner's ConfigureShell. A shell-free command is tokenized with the same
// splitter the cmd runner uses, then re-escaped with syscall.EscapeArg so the C
// runtime re-parses it back to the identical argv — keeping pty/record and run
// steps in agreement about how a command line splits on Windows.
func CommandLine(command string, shell bool) (string, error) {
	if shell {
		return `cmd /S /C "` + command + `"`, nil
	}
	name, args, err := runnercmd.CommandLine(command, false)
	if err != nil {
		return "", err
	}
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, syscall.EscapeArg(name))
	for _, a := range args {
		parts = append(parts, syscall.EscapeArg(a))
	}
	return strings.Join(parts, " "), nil
}

// Start launches commandLine inside a fresh pseudo console sized rows×cols, in
// workDir, with env (nil inherits the parent's environment; a non-nil slice,
// even empty, starts the child from exactly that set). The returned
// PseudoConsole must be Closed.
//
// Everything below is scoped to the CHILD. Start never calls AllocConsole,
// AttachConsole, FreeConsole, or SetStdHandle, so atago's own console and
// standard handles are untouched, and a `run` step executing concurrently in
// another goroutine cannot have its stdout rebound by a pty step starting up.
// This was one of the three things #339 listed as worth checking for a Windows
// capture anomaly seen while pty and nested-run specs shared a parallel batch;
// it is recorded here so the question is not re-opened from scratch. The
// remaining handle discipline that keeps the two independent: the pipes below
// are created with nil SECURITY_ATTRIBUTES (not inheritable), and CreateProcess
// is called with bInheritHandles=false, so the child reaches its console only
// through the PSEUDOCONSOLE attribute.
func Start(commandLine, workDir string, env []string, rows, cols int) (*PseudoConsole, error) {
	// Two anonymous pipes: one carries parent→child input, the other
	// child→parent output. CreatePseudoConsole takes the child's ends (inRead,
	// outWrite); the parent keeps inWrite and outRead.
	var inRead, inWrite, outRead, outWrite windows.Handle
	if err := windows.CreatePipe(&inRead, &inWrite, nil, 0); err != nil {
		return nil, fmt.Errorf("create input pipe: %w", err)
	}
	if err := windows.CreatePipe(&outRead, &outWrite, nil, 0); err != nil {
		closeHandles(inRead, inWrite)
		return nil, fmt.Errorf("create output pipe: %w", err)
	}

	var hpc windows.Handle
	size := windows.Coord{X: termDim(cols), Y: termDim(rows)}
	if err := windows.CreatePseudoConsole(size, inRead, outWrite, 0, &hpc); err != nil {
		closeHandles(inRead, inWrite, outRead, outWrite)
		return nil, fmt.Errorf("create pseudo console: %w", err)
	}
	// The child owns inRead/outWrite through the pseudo console now; the parent's
	// copies are done. Leaving outWrite open would keep the read side from ever
	// seeing EOF when the child exits.
	closeHandles(inRead, outWrite)

	attrList, err := windows.NewProcThreadAttributeList(1)
	if err != nil {
		windows.ClosePseudoConsole(hpc)
		closeHandles(inWrite, outRead)
		return nil, fmt.Errorf("alloc attribute list: %w", err)
	}
	// The PSEUDOCONSOLE attribute value IS the HPCON handle itself (passed by
	// value as the documented Win32 idiom), sized as one handle.
	if err := attrList.Update(
		windows.PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE,
		unsafe.Pointer(hpc), //nolint:govet,gosec // Win32 ConPTY idiom: the HPCON handle value IS the attribute payload (a by-value handle, not a Go pointer, so it is GC-safe)
		unsafe.Sizeof(hpc),
	); err != nil {
		attrList.Delete()
		windows.ClosePseudoConsole(hpc)
		closeHandles(inWrite, outRead)
		return nil, fmt.Errorf("set pseudo-console attribute: %w", err)
	}

	si := new(windows.StartupInfoEx)
	si.Cb = uint32(unsafe.Sizeof(*si))
	// STARTF_USESTDHANDLES with the std handles left nil is what stops the child
	// from inheriting the PARENT's console: without it a console app attaches to
	// atago's own console and writes there instead of through the pseudo-console,
	// so nothing reaches the transcript pipe. The pseudo-console attribute below
	// then supplies the child's actual console I/O. This mirrors the established
	// ConPTY wrappers (aymanbagabas/go-pty, UserExistsError/conpty).
	si.Flags |= windows.STARTF_USESTDHANDLES
	si.ProcThreadAttributeList = attrList.List()

	argv, err := windows.UTF16PtrFromString(commandLine)
	if err != nil {
		attrList.Delete()
		windows.ClosePseudoConsole(hpc)
		closeHandles(inWrite, outRead)
		return nil, fmt.Errorf("encode command line: %w", err)
	}
	var dir *uint16
	if workDir != "" {
		if dir, err = windows.UTF16PtrFromString(workDir); err != nil {
			attrList.Delete()
			windows.ClosePseudoConsole(hpc)
			closeHandles(inWrite, outRead)
			return nil, fmt.Errorf("encode workdir: %w", err)
		}
	}
	var envBlock *uint16
	if env != nil {
		envBlock = utf16EnvBlock(env)
	}

	pi := new(windows.ProcessInformation)
	// EXTENDED_STARTUPINFO_PRESENT makes CreateProcess read the attribute list;
	// CREATE_UNICODE_ENVIRONMENT matches the UTF-16 env block. bInheritHandles is
	// false: the child reaches the console through the attribute, not inheritance.
	flags := uint32(windows.EXTENDED_STARTUPINFO_PRESENT | windows.CREATE_UNICODE_ENVIRONMENT)
	if err := windows.CreateProcess(
		nil, argv, nil, nil, false, flags, envBlock, dir, &si.StartupInfo, pi,
	); err != nil {
		attrList.Delete()
		windows.ClosePseudoConsole(hpc)
		closeHandles(inWrite, outRead)
		return nil, fmt.Errorf("create process: %w", err)
	}
	// The primary thread handle is unused; keep the process handle for wait/kill.
	closeHandles(pi.Thread)

	return &PseudoConsole{
		hpc:      hpc,
		inWrite:  inWrite,
		outRead:  outRead,
		process:  pi.Process,
		pid:      pi.ProcessId,
		attrList: attrList,
	}, nil
}

// Read drains the child's output (the transcript source).
//
// The error class is preserved, because the caller has to tell the end of a
// session from a lost transcript (#345). This used to collapse every ReadFile
// failure into os.ErrClosed, so the pty runner could not distinguish them even
// if it wanted to — and a short transcript then produced a confidently wrong
// screen assertion. The mapping:
//
//   - ERROR_BROKEN_PIPE / ERROR_HANDLE_EOF / ERROR_PIPE_NOT_CONNECTED are the
//     ConPTY analog of POSIX EIO once the child is gone: the session ended, so
//     they become io.EOF and the reader loop ends cleanly after appending any
//     final bytes.
//   - ERROR_INVALID_HANDLE / ERROR_OPERATION_ABORTED are what a pending read
//     sees when Close runs on another goroutine — atago's own teardown, not a
//     failure — so they stay os.ErrClosed, which the caller also treats as a
//     normal end.
//   - Anything else is a genuine read failure and is returned wrapped, so the
//     caller can report the transcript as incomplete instead of silently
//     truncating it.
func (c *PseudoConsole) Read(p []byte) (int, error) {
	var done uint32
	if err := windows.ReadFile(c.outRead, p, &done, nil); err != nil {
		return int(done), classifyReadError(err)
	}
	return int(done), nil
}

// classifyReadError implements Read's mapping. It is separate so the
// classification is testable without a live pseudo console, which cannot be
// made to fail on demand.
func classifyReadError(err error) error {
	switch {
	case errors.Is(err, windows.ERROR_BROKEN_PIPE),
		errors.Is(err, windows.ERROR_HANDLE_EOF),
		errors.Is(err, windows.ERROR_PIPE_NOT_CONNECTED):
		return io.EOF
	case errors.Is(err, windows.ERROR_INVALID_HANDLE),
		errors.Is(err, windows.ERROR_OPERATION_ABORTED):
		return os.ErrClosed
	default:
		return fmt.Errorf("conpty: reading the pseudo console: %w", err)
	}
}

// Write delivers a send to the child.
func (c *PseudoConsole) Write(p []byte) (int, error) {
	var done uint32
	if err := windows.WriteFile(c.inWrite, p, &done, nil); err != nil {
		return int(done), err
	}
	return int(done), nil
}

// Resize changes the pseudo console's dimensions.
func (c *PseudoConsole) Resize(rows, cols int) error {
	return windows.ResizePseudoConsole(c.hpc, windows.Coord{X: termDim(cols), Y: termDim(rows)})
}

// termDim clamps a terminal dimension into the positive int16 range a Coord
// carries, so an out-of-range rows/cols can never wrap to a negative or zero
// size (a 24x80 default and small authored sizes are the norm).
func termDim(n int) int16 {
	const maxDim = 0x7fff
	switch {
	case n < 1:
		return 1
	case n > maxDim:
		return maxDim
	default:
		return int16(n)
	}
}

// Wait blocks until the child exits (or ctx is done) and returns its exit code;
// a wait that cannot read the code, or a ctx that fires first, returns -1.
func (c *PseudoConsole) Wait(ctx context.Context) int {
	done := make(chan int, 1)
	go func() {
		if _, err := windows.WaitForSingleObject(c.process, windows.INFINITE); err != nil {
			done <- -1
			return
		}
		var code uint32
		if err := windows.GetExitCodeProcess(c.process, &code); err != nil {
			done <- -1
			return
		}
		done <- int(code)
	}()
	select {
	case code := <-done:
		return code
	case <-ctx.Done():
		return -1
	}
}

// Pid exposes the child's process id for a tree kill.
func (c *PseudoConsole) Pid() int { return int(c.pid) }

// Close tears down the pseudo console and every handle exactly once. Closing the
// pseudo console signals the child that its console went away; a caller that
// must not let the child linger kills the tree first.
//
// CancelIoEx is what actually ends a read already in flight (#406). Read issues
// a plain synchronous ReadFile on the output pipe, and CloseHandle does NOT
// abort one that is already parked in the kernel: the reader goroutine stays
// blocked until the write end goes away on its own, which a surviving conhost —
// or a descendant that outlived the tree kill — can put off indefinitely. The
// recorder's bounded join then has to wait out its whole grace on every timed-out
// capture, and the goroutine leaks for the life of the process holding the pipe
// handle open. CancelIoEx cancels pending I/O on the handle regardless of which
// thread issued it (unlike CancelIo, which is limited to the calling thread), so
// the parked read returns ERROR_OPERATION_ABORTED — which classifyReadError
// already maps to os.ErrClosed, and which isSessionEnd already reads as a normal
// end of session. That mapping was written for this case before anything
// produced it.
//
// It runs after ClosePseudoConsole so a reader that can still finish normally
// sees EOF rather than a cancellation, and before CloseHandle because canceling
// I/O on a closed handle is meaningless. A failure is ignored on purpose: there
// is nothing to cancel when no read is pending, which is the common case.
func (c *PseudoConsole) Close() error {
	c.closeOnce.Do(func() {
		windows.ClosePseudoConsole(c.hpc)
		_ = windows.CancelIoEx(c.outRead, nil)
		closeHandles(c.inWrite, c.outRead, c.process)
		if c.attrList != nil {
			c.attrList.Delete()
		}
	})
	return nil
}

// closeHandles best-effort closes each valid handle.
func closeHandles(handles ...windows.Handle) {
	for _, h := range handles {
		if h != 0 && h != windows.InvalidHandle {
			_ = windows.CloseHandle(h)
		}
	}
}

// utf16EnvBlock encodes env ("KEY=VALUE" entries) as the NUL-separated,
// double-NUL-terminated UTF-16 block CreateProcess wants with
// CREATE_UNICODE_ENVIRONMENT. An empty (non-nil) slice yields just the double
// NUL — an empty environment CreateProcessW still accepts.
func utf16EnvBlock(env []string) *uint16 {
	var buf []uint16
	for _, e := range env {
		enc, err := windows.UTF16FromString(e)
		if err != nil {
			continue // skip an entry with an embedded NUL rather than fail the run
		}
		buf = append(buf, enc...) // UTF16FromString already appends the entry's NUL
	}
	buf = append(buf, 0) // final NUL closes the block
	if len(buf) == 1 {
		// An empty environment still needs the double-NUL terminator, or
		// CreateProcessW rejects the block with ERROR_INVALID_PARAMETER.
		buf = append(buf, 0)
	}
	return &buf[0]
}
