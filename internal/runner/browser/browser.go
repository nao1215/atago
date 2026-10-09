// Package browser implements the browser (CDP) runner: a `cdp` step drives a
// headless Chrome via the Chrome DevTools Protocol and captures a value from the
// page as a runner.Result. It is the atago counterpart
// to runn's CDP runner and builds on github.com/chromedp/chromedp — the same
// library runn uses. One browser session is shared across the cdp steps of a
// scenario, so navigation state persists between them.
package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"

	"github.com/nao1215/atago/internal/diag"
	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
)

// Config is the resolved configuration for a browser runner.
type Config struct {
	// Headless runs Chrome without a visible window (default true).
	Headless bool
	// ExecPath, when set, launches a specific Chrome/Chromium binary instead of the
	// one chromedp discovers on PATH.
	ExecPath string
	// Args are extra Chrome launch flags (bare flag names, no leading "--") for CI
	// environments that need them.
	Args []string
	// Timeout bounds a single cdp step (the whole action list); zero means none.
	Timeout time.Duration
}

// Runner owns one Chrome session for a browser runner.
type Runner struct {
	allocCtx    context.Context
	allocCancel context.CancelFunc
	browserCtx  context.Context
	browserStop context.CancelFunc
	timeout     time.Duration
}

// Open launches a headless Chrome session.
func Open(cfg Config) (*Runner, error) {
	opts := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
	// --no-sandbox is required to launch Chrome as root / inside many CI
	// containers, where the setuid sandbox is unavailable.
	opts = append(opts, chromedp.NoSandbox)
	// CI/container hardening: --disable-dev-shm-usage avoids the tiny /dev/shm
	// that makes Chrome hang in constrained CI containers. It is a safe no-op on
	// a fast local machine. A slow cold start is bounded by launchTimeout below.
	opts = append(opts, chromedp.Flag("disable-dev-shm-usage", true))
	if !cfg.Headless {
		opts = append(opts, chromedp.Flag("headless", false))
	}
	if cfg.ExecPath != "" {
		opts = append(opts, chromedp.ExecPath(cfg.ExecPath))
	}
	for _, a := range cfg.Args {
		name, value := splitFlag(a)
		if name == "" {
			continue
		}
		opts = append(opts, chromedp.Flag(name, value))
	}
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	browserCtx, browserStop := chromedp.NewContext(allocCtx)
	// Force the browser to start now so a launch failure surfaces at Open, not on
	// the first action.
	if err := launch(browserCtx); err != nil {
		browserStop()
		allocCancel()
		return nil, diag.ConnectFailed.Errorf("launching headless browser: %w", err)
	}
	return &Runner{
		allocCtx:    allocCtx,
		allocCancel: allocCancel,
		browserCtx:  browserCtx,
		browserStop: browserStop,
		timeout:     cfg.Timeout,
	}, nil
}

// launchTimeout bounds how long Open waits for a cold Chrome to come up. In
// constrained CI environments headless Chrome can be slow to answer its first
// DevTools command, so the bound is generous; it matches the 60s WebSocket-URL
// read timeout atago used before chromedp switched to the pipe transport.
const launchTimeout = 60 * time.Second

// launch starts the browser and opens its first tab. The browser must be
// started with a context that has no deadline (a deadline would stop the whole
// browser when it expires), so the launch bound is a separate timer: when it
// fires, the caller's cancel funcs tear the half-started browser down.
func launch(browserCtx context.Context) error {
	done := make(chan error, 1)
	go func() { done <- chromedp.Do(browserCtx) }()
	timer := time.NewTimer(launchTimeout)
	defer timer.Stop()
	select {
	case err := <-done:
		return err
	case <-timer.C:
		// The buffered channel lets the launch goroutine finish once the
		// caller cancels browserCtx.
		return fmt.Errorf("browser did not start within %v", launchTimeout)
	}
}

// splitFlag turns a bare launch-flag token into the (name, value) pair chromedp
// expects. "disable-gpu" becomes ("disable-gpu", true) — a valueless switch;
// "window-size=1280,720" becomes ("window-size", "1280,720"). A leading "--" is
// tolerated and stripped. An empty or dangling token yields an empty name so the
// caller can skip it.
func splitFlag(token string) (string, any) {
	token = strings.TrimSpace(token)
	token = strings.TrimPrefix(token, "--")
	if token == "" {
		return "", nil
	}
	if name, value, ok := strings.Cut(token, "="); ok {
		if name == "" {
			return "", nil
		}
		return name, value
	}
	return token, true
}

// Close shuts the browser session down.
func (r *Runner) Close() error {
	r.browserStop()
	r.allocCancel()
	return nil
}

// Run executes the action list in order against the session and returns the
// value captured by the last capturing action (text/eval/title/attribute; nil
// when none captured anything). workdir resolves relative screenshot paths.
func (r *Runner) Run(ctx context.Context, actions []spec.CDPAction, workdir string) (*runner.Result, error) {
	// Execute in the persistent browser context (so the chromedp session and its
	// navigation state survive across steps), but tie the run to the caller's
	// context too: propagate a caller cancellation (Ctrl-C / parent cancel /
	// deadline) into runCtx so a browser step — e.g. wait_visible on a selector
	// that never appears with no configured timeout — cannot hang forever (issue
	// #14).
	runCtx, cancel := context.WithCancel(r.browserCtx)
	defer cancel()
	stopPropagate := context.AfterFunc(ctx, cancel)
	defer stopPropagate()
	if r.timeout > 0 {
		var tcancel context.CancelFunc
		runCtx, tcancel = context.WithTimeout(runCtx, r.timeout)
		defer tcancel()
	}

	tasks, caps, shots, err := buildTasks(actions, workdir)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	//nolint:contextcheck // runCtx is rooted at the persistent browser context rather than
	// at ctx on purpose: the chromedp session has to outlive a single step. The caller's
	// cancellation still reaches it through the AfterFunc above, which is what #14 asked for.
	if err := chromedp.Do(runCtx, tasks...); err != nil {
		return nil, fmt.Errorf("cdp run: %w", err)
	}

	// Persist any screenshots after the run completes, so the captured bytes are
	// on disk before the scenario's file/image assertions read them (#50).
	for _, s := range shots {
		// s.path was confined to the workdir when the task was built; the confined
		// write creates its parents, refuses to follow a symlink the page under
		// test may have planted at the target (issue #16), and binds the write to
		// the workdir so an ancestor swapped for a link cannot redirect it (#430).
		if err := security.WriteConfinedFile(workdir, s.path, *s.buf); err != nil {
			return nil, diag.StepFileUnusable.Errorf("cdp screenshot: writing %q: %w", s.path, err)
		}
	}

	out := &runner.Result{Command: "cdp", IsCDP: true, Duration: time.Since(start)}
	if len(caps) > 0 {
		out.CDPValue = caps[len(caps)-1].bytes()
	}
	return out, nil
}

// capture holds a pointer to the value a text/eval action fills in once the run
// completes; exactly one of str/js is set.
type capture struct {
	str *string
	js  *json.RawMessage
}

func (c capture) bytes() []byte {
	if c.str != nil {
		return []byte(*c.str)
	}
	if c.js != nil {
		return []byte(*c.js)
	}
	return nil
}

// shot pairs a pending screenshot's destination path with the buffer chromedp
// fills during the run; the bytes are flushed to disk after Run completes.
type shot struct {
	path string
	buf  *[]byte
}

// storeInto adapts a value-returning chromedp action into a step of the action
// list: the value is written to *dst when the step runs, so the caller reads it
// once the whole list has completed.
func storeInto[T any](a chromedp.Action[T], dst *T) chromedp.Action[chromedp.Void] {
	return chromedp.Func(func(ctx context.Context, t *chromedp.Target) error {
		v, err := a(ctx, t)
		if err != nil {
			return err
		}
		*dst = v
		return nil
	})
}

// attributeValueInto captures an attribute's value (empty when the attribute is
// absent) into *dst.
func attributeValueInto(selector, name string, dst *string) chromedp.Action[chromedp.Void] {
	return chromedp.Func(func(ctx context.Context, t *chromedp.Target) error {
		res, err := chromedp.AttributeValue(chromedp.CSS(selector), name)(ctx, t)
		if err != nil {
			return err
		}
		*dst = res.Value
		return nil
	})
}

// buildTasks turns the spec's action list into chromedp steps. A selector is a
// CSS selector matched with DOM.querySelector (chromedp.CSS, the old ByQuery).
func buildTasks(actions []spec.CDPAction, workdir string) ([]chromedp.Action[chromedp.Void], []capture, []shot, error) {
	var tasks []chromedp.Action[chromedp.Void]
	var caps []capture
	var shots []shot
	for i, a := range actions {
		switch {
		case a.Navigate != "":
			tasks = append(tasks, chromedp.Navigate(a.Navigate))
		case a.WaitVisible != "":
			tasks = append(tasks, chromedp.WaitVisible(chromedp.CSS(a.WaitVisible)))
		case a.WaitHidden != "":
			tasks = append(tasks, chromedp.WaitNotVisible(chromedp.CSS(a.WaitHidden)))
		case a.Click != "":
			tasks = append(tasks, chromedp.Click(chromedp.CSS(a.Click)))
		case a.Press != nil:
			key, err := resolveKey(a.Press.Key)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("cdp action %d: %w", i, err)
			}
			tasks = append(tasks, chromedp.SendKeys(chromedp.CSS(a.Press.Selector), key))
		case a.Select != nil:
			tasks = append(tasks, chromedp.SetValue(chromedp.CSS(a.Select.Selector), a.Select.Value))
		case a.Check != "":
			tasks = append(tasks, setChecked(a.Check, true))
		case a.Uncheck != "":
			tasks = append(tasks, setChecked(a.Uncheck, false))
		case a.Screenshot != nil:
			buf := new([]byte)
			if a.Screenshot.Selector != "" {
				tasks = append(tasks, storeInto(chromedp.Screenshot(chromedp.CSS(a.Screenshot.Selector)), buf))
			} else {
				tasks = append(tasks, storeInto(chromedp.FullScreenshot(100), buf))
			}
			path, err := security.ResolveWorkdirPath("cdp.screenshot.path", workdir, a.Screenshot.Path)
			if err != nil {
				return nil, nil, nil, err
			}
			shots = append(shots, shot{path: path, buf: buf})
		case a.SendKeys != nil:
			tasks = append(tasks, chromedp.SendKeys(chromedp.CSS(a.SendKeys.Selector), a.SendKeys.Value))
		case a.Upload != nil:
			// Set a file on an <input type=file> (#75). The file must exist inside the
			// scenario workdir, so a spec cannot upload arbitrary host files.
			file, err := security.ResolveWorkdirPath("cdp.upload.file", workdir, a.Upload.File)
			if err != nil {
				return nil, nil, nil, err
			}
			if _, statErr := os.Stat(file); statErr != nil {
				return nil, nil, nil, diag.StepFileUnusable.Errorf("cdp action %d: upload file %q: %w", i, a.Upload.File, statErr)
			}
			tasks = append(tasks, chromedp.SetUploadFiles(chromedp.CSS(a.Upload.Selector), []string{file}))
		case a.Download != nil:
			// Capture a click-triggered download into a deterministic scenario
			// directory (#75). The destination is confined to the workdir.
			dir := a.Download.Dir
			if dir == "" {
				dir = "."
			}
			destDir, err := security.ResolveWorkdirPath("cdp.download.dir", workdir, dir)
			if err != nil {
				return nil, nil, nil, err
			}
			name := new(string)
			tasks = append(tasks, downloadAction(a.Download.Click, destDir, name))
			caps = append(caps, capture{str: name})
		case a.Text != "":
			s := new(string)
			tasks = append(tasks, storeInto(chromedp.Text(chromedp.CSS(a.Text), chromedp.NodeVisible), s))
			caps = append(caps, capture{str: s})
		case a.Title:
			s := new(string)
			tasks = append(tasks, storeInto(chromedp.Title(), s))
			caps = append(caps, capture{str: s})
		case a.Attribute != nil:
			s := new(string)
			tasks = append(tasks, attributeValueInto(a.Attribute.Selector, a.Attribute.Name, s))
			caps = append(caps, capture{str: s})
		case a.Eval != "":
			j := new(json.RawMessage)
			tasks = append(tasks, storeInto(chromedp.Evaluate[json.RawMessage](a.Eval), j))
			caps = append(caps, capture{js: j})
		default:
			return nil, nil, nil, diag.InternalError.Errorf("cdp action %d sets no recognized action", i)
		}
	}
	if len(tasks) == 0 {
		return nil, nil, nil, diag.InternalError.Errorf("cdp step has no actions")
	}
	return tasks, caps, shots, nil
}

// jsStringLiteral renders s as a JavaScript string literal for embedding in an
// evaluated snippet. json.Marshal is the correct escaper (a selector may carry
// quotes, backslashes, or newlines) and cannot fail for a string, so the error is
// swallowed here, once, with the reason written down.
func jsStringLiteral(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		// Unreachable: encoding a Go string never fails.
		return `""`
	}
	return string(b)
}

// setChecked ticks or unticks a checkbox/radio by setting its checked property
// and dispatching a change event, so listeners react as they would to a click.
// Selecting by property (not a click) keeps the action deterministic regardless
// of the element's current state.
func setChecked(selector string, checked bool) chromedp.Action[chromedp.Void] {
	sel := jsStringLiteral(selector)
	js := fmt.Sprintf(`(() => {
	const el = document.querySelector(%s);
	if (!el) throw new Error('check: no element for selector ' + %s);
	el.checked = %t;
	el.dispatchEvent(new Event('change', {bubbles: true}));
	return el.checked;
})()`, sel, sel, checked)
	return chromedp.Evaluate[chromedp.Void](js)
}

// pressKeys maps the small set of named keys atago supports for a press action
// to their chromedp key sequences. A single printable character is passed
// through as-is.
var pressKeys = map[string]string{
	"Enter":      kb.Enter,
	"Tab":        kb.Tab,
	"Escape":     kb.Escape,
	"Backspace":  kb.Backspace,
	"Delete":     kb.Delete,
	"ArrowUp":    kb.ArrowUp,
	"ArrowDown":  kb.ArrowDown,
	"ArrowLeft":  kb.ArrowLeft,
	"ArrowRight": kb.ArrowRight,
	"Space":      " ",
}

// resolveKey turns a press key name into the sequence chromedp.SendKeys expects.
func resolveKey(name string) (string, error) {
	if seq, ok := pressKeys[name]; ok {
		return seq, nil
	}
	if len([]rune(name)) == 1 {
		return name, nil
	}
	return "", diag.InputNotSupported.Errorf("unsupported press key %q (use a single character or one of Enter/Tab/Escape/Backspace/Delete/Arrow*/Space)", name)
}
