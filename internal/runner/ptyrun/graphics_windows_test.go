//go:build windows

package ptyrun

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/windows"

	"github.com/nao1215/atago/internal/spec"
)

// graphicsChildEnv selects what the test binary does when a test runs it as
// the program under test.
const graphicsChildEnv = "ATAGO_PTYRUN_GRAPHICS_CHILD"

func TestMain(m *testing.M) {
	switch os.Getenv(graphicsChildEnv) {
	case "probe":
		os.Exit(graphicsProbeChild())
	case "image":
		os.Exit(graphicsImageChild())
	case "keys":
		os.Exit(keysChild())
	case "vtinput":
		os.Exit(vtInputChild())
	}
	os.Exit(m.Run())
}

// childConsole opens the console the way a TUI does: raw VT input, so the
// terminal's answers arrive as bytes, and VT output.
func childConsole() (in, out windows.Handle, err error) {
	open := func(name string) (windows.Handle, error) {
		return windows.CreateFile(windows.StringToUTF16Ptr(name),
			windows.GENERIC_READ|windows.GENERIC_WRITE,
			windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, 0, 0)
	}
	if in, err = open("CONIN$"); err != nil {
		return 0, 0, err
	}
	if out, err = open("CONOUT$"); err != nil {
		return 0, 0, err
	}
	if err = windows.SetConsoleMode(in, windows.ENABLE_VIRTUAL_TERMINAL_INPUT); err != nil {
		return 0, 0, err
	}
	err = windows.SetConsoleMode(out, windows.ENABLE_PROCESSED_OUTPUT|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING|windows.DISABLE_NEWLINE_AUTO_RETURN)
	return in, out, err
}

func childWrite(out windows.Handle, s string) {
	var n uint32
	_ = windows.WriteFile(out, []byte(s), &n, nil)
}

// graphicsProbeChild is how a program that needs images finds out whether it
// can draw them: the kitty graphics query, the cell-size query, and DA1 as the
// end of the answers. It prints what came back with ESC shown as "E", as the
// POSIX probe in test/e2e/atago/pty_graphics.atago.yaml does.
func graphicsProbeChild() int {
	in, out, err := childConsole()
	if err != nil {
		fmt.Printf("console: %v\r\n", err)
		return 2
	}
	var mu sync.Mutex
	var got []byte
	done := make(chan struct{})
	go func() {
		buf := make([]byte, 256)
		for {
			var n uint32
			if err := windows.ReadFile(in, buf, &n, nil); err != nil {
				return
			}
			mu.Lock()
			got = append(got, buf[:n]...)
			// The DA1 reply is the last answer.
			finished := strings.Contains(string(got), "\x1b[?6c")
			mu.Unlock()
			if finished {
				close(done)
				return
			}
		}
	}()
	childWrite(out, "\x1b_Gi=31,s=1,v=1,a=q,t=d,f=24;AAAA\x1b\\\x1b[16t\x1b[c")
	select {
	case <-done:
	case <-time.After(10 * time.Second):
	}
	mu.Lock()
	defer mu.Unlock()
	childWrite(out, "ANSWER:"+strings.ReplaceAll(string(got), "\x1b", "E")+"\r\n")
	return 0
}

// graphicsImageChild draws one red pixel with a direct kitty transmission.
func graphicsImageChild() int {
	_, out, err := childConsole()
	if err != nil {
		fmt.Printf("console: %v\r\n", err)
		return 2
	}
	childWrite(out, "\x1b_Ga=T,i=5,f=24,s=1,v=1;/wAA\x1b\\")
	childWrite(out, "drawn\r\n")
	time.Sleep(500 * time.Millisecond)
	return 0
}

func graphicsChild(t *testing.T, mode string) (string, []string) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return syscall.EscapeArg(exe), append(os.Environ(), graphicsChildEnv+"="+mode)
}

// TestRun_Windows_GraphicsAnswersTheKittyQuery is #676: under the pseudo
// console in Windows the query never reached atago and the answer never reached
// the program, so a program that requires images refused to start. The answers
// must arrive complete and in the order the queries were sent, as on POSIX.
func TestRun_Windows_GraphicsAnswersTheKittyQuery(t *testing.T) {
	t.Parallel()
	command, env := graphicsChild(t, "probe")
	p := &spec.PTY{
		Command:  command,
		Graphics: spec.PTYGraphicsKitty,
		Timeout:  "30s",
		Session:  []spec.PTYAction{{Expect: "ANSWER:"}},
	}
	res, ef, err := Run(context.Background(), p, t.TempDir(), env)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if ef != nil {
		t.Fatalf("expect failure: %+v (transcript %q)", ef, res.Stdout)
	}
	if res.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0 (transcript %q)", res.ExitCode, res.Stdout)
	}
	want := `ANSWER:E_Gi=31;OKE\E[6;20;10tE[?6c`
	if !strings.Contains(string(res.Screen), want) {
		t.Errorf("screen = %q, want it to contain %q (transcript %q)", res.Screen, want, res.Stdout)
	}
}

// TestRun_Windows_GraphicsRecordsATransmittedImage is the other half of #676:
// an image the program transmits must reach atago's screen.images.
func TestRun_Windows_GraphicsRecordsATransmittedImage(t *testing.T) {
	t.Parallel()
	command, env := graphicsChild(t, "image")
	p := &spec.PTY{
		Command:  command,
		Graphics: spec.PTYGraphicsKitty,
		Timeout:  "30s",
		Session:  []spec.PTYAction{{Expect: "drawn"}},
	}
	res, ef, err := Run(context.Background(), p, t.TempDir(), env)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if ef != nil {
		t.Fatalf("expect failure: %+v (transcript %q)", ef, res.Stdout)
	}
	if len(res.Images) != 1 {
		t.Fatalf("images = %d, want 1 (transcript %q)", len(res.Images), res.Stdout)
	}
	if img := res.Images[0]; img.Width != 1 || img.Height != 1 {
		t.Errorf("image = %dx%d, want 1x1", img.Width, img.Height)
	}
}

// TestRun_Windows_GraphicsRunsAShellCommand checks that the bundled console
// host is an ordinary terminal for the rest of the session: a shell command
// behind it prints and exits as it does behind the in-box host.
func TestRun_Windows_GraphicsRunsAShellCommand(t *testing.T) {
	t.Parallel()
	shell := true
	p := &spec.PTY{
		Shell:    &shell,
		Command:  "echo hello-openconsole",
		Graphics: spec.PTYGraphicsKitty,
		Timeout:  "30s",
		Session:  []spec.PTYAction{{Expect: "hello-openconsole"}},
	}
	res, ef, err := Run(context.Background(), p, t.TempDir(), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if ef != nil {
		t.Fatalf("expect failure: %+v (transcript %q)", ef, res.Stdout)
	}
	if res.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0 (transcript %q)", res.ExitCode, res.Stdout)
	}
	if !strings.Contains(string(res.Screen), "hello-openconsole") {
		t.Errorf("screen = %q, want the command's output", res.Screen)
	}
}
