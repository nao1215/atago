//go:build windows

package ptyrun

import (
	"context"
	"fmt"
	"os"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// conPTY is a self-contained Windows pseudo-console (ConPTY). It wires a child
// process's console I/O to a pipe pair through a PseudoConsole and exposes
// Read/Write/Resize/wait/Close. It calls the ConPTY and process-creation APIs
// directly through golang.org/x/sys/windows (already a dependency), so atago
// carries no third-party ConPTY library: the surface is small and the Win32
// calls are stable since Windows 10 (1809).
//
// It follows Microsoft's documented pseudo-console recipe: CreatePseudoConsole
// over a pipe pair, a PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE attribute list, then
// CreateProcess with EXTENDED_STARTUPINFO_PRESENT.
type conPTY struct {
	hpc       windows.Handle // the pseudo console (HPCON)
	inWrite   windows.Handle // parent → child (sends)
	outRead   windows.Handle // child → parent (transcript)
	process   windows.Handle // child process handle, for wait/kill
	pid       uint32
	attrList  *windows.ProcThreadAttributeListContainer
	closeOnce sync.Once
}

// isConPTYAvailable reports whether the host exposes the ConPTY API (Windows 10
// version 1809 and later), so an older host gets a clear error instead of a
// missing-proc failure deep in start.
func isConPTYAvailable() bool {
	return windows.NewLazySystemDLL("kernel32.dll").NewProc("CreatePseudoConsole").Find() == nil
}

// startConPTY launches commandLine inside a fresh pseudo console sized rows×cols,
// in workDir, with env (nil inherits the parent's environment; a non-nil slice,
// even empty, starts the child from exactly that set). The returned conPTY must
// be Closed.
func startConPTY(commandLine, workDir string, env []string, rows, cols int) (*conPTY, error) {
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

	return &conPTY{
		hpc:      hpc,
		inWrite:  inWrite,
		outRead:  outRead,
		process:  pi.Process,
		pid:      pi.ProcessId,
		attrList: attrList,
	}, nil
}

// Read drains the child's output (the transcript source). A broken/closed pipe
// once the child exits is the ConPTY analog of POSIX EIO: it surfaces as an
// error so the reader loop ends cleanly (after appending any final bytes).
func (c *conPTY) Read(p []byte) (int, error) {
	var done uint32
	if err := windows.ReadFile(c.outRead, p, &done, nil); err != nil {
		return int(done), os.ErrClosed
	}
	return int(done), nil
}

// Write delivers a send to the child.
func (c *conPTY) Write(p []byte) (int, error) {
	var done uint32
	if err := windows.WriteFile(c.inWrite, p, &done, nil); err != nil {
		return int(done), err
	}
	return int(done), nil
}

// Resize changes the pseudo console's dimensions.
func (c *conPTY) Resize(rows, cols int) error {
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

// wait blocks until the child exits (or ctx is done) and returns its exit code;
// a wait that cannot read the code, or a ctx that fires first, returns -1.
func (c *conPTY) wait(ctx context.Context) int {
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

// pidValue exposes the child's process id for a tree kill.
func (c *conPTY) pidValue() int { return int(c.pid) }

// Close tears down the pseudo console and every handle exactly once. Closing the
// pseudo console signals the child that its console went away; a caller that
// must not let the child linger kills the tree first.
func (c *conPTY) Close() error {
	c.closeOnce.Do(func() {
		windows.ClosePseudoConsole(c.hpc)
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
// CREATE_UNICODE_ENVIRONMENT. An empty (non-nil) slice yields just the
// terminating NUL — an empty environment.
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
