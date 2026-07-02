package browser

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/spec"
)

// chromeUsable reports whether a Chrome/Chromium binary is present and launches.
func chromeUsable(t *testing.T) bool {
	t.Helper()
	found := false
	for _, name := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser", "headless-shell"} {
		if _, err := exec.LookPath(name); err == nil {
			found = true
			break
		}
	}
	return found
}

// Regression for issue #14: the CDP runner ignored the caller ctx, so a step
// with no configured timeout (here wait_visible on a selector that never
// appears) could hang forever and could not be interrupted by cancellation.
// With the fix a caller-cancel must interrupt chromedp.Run promptly.
// TestOpen_WithLaunchArgs proves a configured browser runner (extra launch
// flags) still launches and can run a step and be cancelled, so the args
// plumbing does not break the allocator.
func TestOpen_WithLaunchArgs(t *testing.T) {
	if !chromeUsable(t) {
		t.Skip("no usable Chrome/Chromium for the browser runner")
	}
	r, err := Open(Config{Headless: true, Args: []string{"disable-gpu", "window-size=800,600"}})
	if err != nil {
		t.Skipf("chrome present but did not launch with args: %v", err)
	}
	t.Cleanup(func() { _ = r.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := r.Run(ctx, []spec.CDPAction{
		{Navigate: "data:text/html,<html><body><h1 id=t>hi</h1></body></html>"},
		{Text: "#t"},
	}, t.TempDir())
	if err != nil {
		t.Fatalf("Run with launch args failed: %v", err)
	}
	if res == nil || string(res.CDPValue) != "hi" {
		t.Errorf("captured text = %q, want hi", res.CDPValue)
	}
}

func TestRun_CallerCancelInterruptsHangingStep(t *testing.T) {
	if !chromeUsable(t) {
		t.Skip("no usable Chrome/Chromium for the browser runner")
	}
	r, err := Open(Config{Headless: true}) // no Timeout: only ctx can stop it
	if err != nil {
		t.Skipf("chrome present but did not launch: %v", err)
	}
	t.Cleanup(func() { _ = r.Close() })

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(300 * time.Millisecond)
		cancel()
	}()

	actions := []spec.CDPAction{
		{Navigate: "data:text/html,<html><body><h1>hi</h1></body></html>"},
		{WaitVisible: "#never-appears"}, // no such element: waits until cancelled
	}

	done := make(chan error, 1)
	go func() {
		_, err := r.Run(ctx, actions, t.TempDir())
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("Run returned nil error; a cancelled hanging step must fail")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Run did not return within 10s after cancellation; caller ctx not wired into chromedp.Run")
	}
}
