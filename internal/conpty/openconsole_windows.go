//go:build windows

package conpty

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// The pseudo console that ships with Windows cannot host a program that draws
// images with the kitty graphics protocol (#676). Its console host re-renders
// the program's output instead of forwarding it, and an APC string
// (`ESC _ G ... ESC \`) has no place in that rendering, so every image and the
// capability query itself are dropped before they reach atago. That is true of
// every version the GitHub runners carry (Windows Server 2022 and 2025, build
// 20348 and 26100), and the PSEUDOCONSOLE_PASSTHROUGH_MODE flag an older
// Windows Terminal once had is ignored there.
//
// The console host of Windows Terminal 1.22 and later forwards output it does
// not understand. Microsoft publishes it as the Microsoft.Windows.Console.ConPTY
// NuGet package (MIT licensed, see openconsole/LICENSE): conpty.dll, which
// exports the pseudo-console API under a Conpty prefix, and OpenConsole.exe,
// the host it starts from its own directory. atago embeds both for the
// architectures it releases and writes them to the user cache on first use.
//
// To update, download the package from
// https://www.nuget.org/packages/Microsoft.Windows.Console.ConPTY, copy
// runtimes/win-<arch>/native/conpty.dll and
// build/native/runtimes/<arch>/OpenConsole.exe into openconsole/<goarch>/, and
// bump openConsoleVersion.
const openConsoleVersion = "1.24.260710001"

// ErrNoOpenConsole reports a Windows architecture atago carries no console
// host for.
var ErrNoOpenConsole = errors.New("atago bundles the Windows Terminal console host only for amd64 and arm64")

var openConsole struct {
	once sync.Once
	api  consoleAPI
	err  error
}

// StartOpenConsole is Start with the bundled Windows Terminal console host in
// place of the one in Windows. Output the program writes reaches the reader
// unchanged, APC strings included; a reply the host's input parser would drop
// has to be written through EncodeReply.
func StartOpenConsole(commandLine, workDir string, env []string, rows, cols int) (*PseudoConsole, error) {
	openConsole.once.Do(func() {
		openConsole.api, openConsole.err = loadOpenConsole()
	})
	if openConsole.err != nil {
		return nil, openConsole.err
	}
	c, err := start(openConsole.api, 0, commandLine, workDir, env, rows, cols)
	if err != nil {
		return nil, err
	}
	c.win32Input = true
	return c, nil
}

// EncodeReply prepares a terminal's reply to the program for Write. Behind the
// bundled host it is sent as key presses (Win32InputKeys), so a reply the host
// does not parse still arrives; behind the in-box host it is sent as is.
func (c *PseudoConsole) EncodeReply(p []byte) []byte {
	if !c.win32Input {
		return p
	}
	return Win32InputKeys(p)
}

// EncodeInput prepares what a terminal user types, as an xterm-class terminal
// sends it, for Write. Behind the bundled host it is sent as key presses
// (Win32InputSend), because that host's input parser holds a lone ESC and drops
// characters outside the BMP; behind the in-box host it is sent as is.
func (c *PseudoConsole) EncodeInput(p []byte) []byte {
	if !c.win32Input {
		return p
	}
	return Win32InputSend(p)
}

// loadOpenConsole writes the bundled host to disk and binds conpty.dll.
func loadOpenConsole() (consoleAPI, error) {
	if openConsoleDLL == nil {
		return consoleAPI{}, ErrNoOpenConsole
	}
	dir, err := installOpenConsole()
	if err != nil {
		return consoleAPI{}, fmt.Errorf("install the bundled console host: %w", err)
	}
	dll := windows.NewLazyDLL(filepath.Join(dir, "conpty.dll"))
	create := dll.NewProc("ConptyCreatePseudoConsole")
	resize := dll.NewProc("ConptyResizePseudoConsole")
	closeProc := dll.NewProc("ConptyClosePseudoConsole")
	for _, p := range []*windows.LazyProc{create, resize, closeProc} {
		if err := p.Find(); err != nil {
			return consoleAPI{}, fmt.Errorf("load the bundled console host: %w", err)
		}
	}
	return consoleAPI{
		create: func(size windows.Coord, in, out windows.Handle, flags uint32, hpc *windows.Handle) error {
			r, _, _ := create.Call(uintptr(coordArg(size)), uintptr(in), uintptr(out), uintptr(flags), uintptr(unsafe.Pointer(hpc))) //nolint:gosec // Win32 out-parameter: the callee writes the HPCON through this pointer during the call
			return hresult(r)
		},
		resize: func(hpc windows.Handle, size windows.Coord) error {
			r, _, _ := resize.Call(uintptr(hpc), uintptr(coordArg(size)))
			return hresult(r)
		},
		close: func(hpc windows.Handle) { _, _, _ = closeProc.Call(uintptr(hpc)) },
	}, nil
}

// coordArg packs a COORD the way the Win32 ABI passes one by value.
func coordArg(c windows.Coord) uint32 {
	return uint32(uint16(c.X)) | uint32(uint16(c.Y))<<16 //nolint:gosec // a bit-for-bit reinterpretation of the two int16 fields, which termDim keeps positive
}

// hresult turns a failed HRESULT into an error.
func hresult(r uintptr) error {
	if int32(r) >= 0 { //nolint:gosec // an HRESULT is a 32-bit value; the sign bit is the failure bit
		return nil
	}
	return fmt.Errorf("HRESULT 0x%08x: %w", uint32(r), syscall.Errno(r&0xffff)) //nolint:gosec // an HRESULT is a 32-bit value
}

// installOpenConsole writes conpty.dll and OpenConsole.exe to a versioned
// directory in the user cache and returns it. Files already there with the
// right content are kept: another atago may be running the host from them, and
// Windows does not let a running executable be replaced.
func installOpenConsole() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		base = os.TempDir()
	}
	dir := filepath.Join(base, "atago", "openconsole-"+openConsoleVersion+"-"+runtimeArch)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	for name, content := range map[string][]byte{"conpty.dll": openConsoleDLL, "OpenConsole.exe": openConsoleExe} {
		if err := installFile(filepath.Join(dir, name), content); err != nil {
			return "", err
		}
	}
	return dir, nil
}

func installFile(path string, content []byte) error {
	if have, err := os.ReadFile(path); err == nil && bytes.Equal(have, content) { //nolint:gosec // path is atago's own cache directory
		return nil
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		_ = os.Remove(tmp.Name())
		// Another atago may have installed the same file and be running it.
		if have, rerr := os.ReadFile(path); rerr == nil && bytes.Equal(have, content) { //nolint:gosec // path is atago's own cache directory
			return nil
		}
		return err
	}
	return nil
}
