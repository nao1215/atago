// Package browser implements the browser (CDP) runner: a `cdp` step drives a
// headless Chrome via the Chrome DevTools Protocol and captures a value from the
// page as a runner.Result (ADR-0029). It is the atago counterpart
// to runn's CDP runner and builds on github.com/chromedp/chromedp — the same
// library runn uses. One browser session is shared across the cdp steps of a
// scenario, so navigation state persists between them.
package browser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"

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
	// CI/container hardening: in constrained CI environments headless Chrome
	// often fails to publish its DevTools websocket within chromedp's default 20s
	// ("websocket url timeout reached"), which flakes the browser suite.
	// --disable-dev-shm-usage avoids the tiny /dev/shm that makes Chrome hang, and
	// a longer WS-URL read timeout tolerates a slow cold start. Both are safe
	// no-ops on a fast local machine.
	opts = append(opts,
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.WSURLReadTimeout(60*time.Second),
	)
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
	if err := chromedp.Run(browserCtx); err != nil {
		browserStop()
		allocCancel()
		return nil, fmt.Errorf("launching headless browser: %w", err)
	}
	return &Runner{
		allocCtx:    allocCtx,
		allocCancel: allocCancel,
		browserCtx:  browserCtx,
		browserStop: browserStop,
		timeout:     cfg.Timeout,
	}, nil
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
	if err := chromedp.Run(runCtx, tasks...); err != nil {
		return nil, fmt.Errorf("cdp run: %w", err)
	}

	// Persist any screenshots after the run completes, so the captured bytes are
	// on disk before the scenario's file/image assertions read them (#50).
	for _, s := range shots {
		if err := os.WriteFile(s.path, *s.buf, 0o600); err != nil {
			return nil, fmt.Errorf("cdp screenshot: writing %q: %w", s.path, err)
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

func buildTasks(actions []spec.CDPAction, workdir string) (chromedp.Tasks, []capture, []shot, error) {
	var tasks chromedp.Tasks
	var caps []capture
	var shots []shot
	for i, a := range actions {
		switch {
		case a.Navigate != "":
			tasks = append(tasks, chromedp.Navigate(a.Navigate))
		case a.WaitVisible != "":
			tasks = append(tasks, chromedp.WaitVisible(a.WaitVisible, chromedp.ByQuery))
		case a.WaitHidden != "":
			tasks = append(tasks, chromedp.WaitNotVisible(a.WaitHidden, chromedp.ByQuery))
		case a.Click != "":
			tasks = append(tasks, chromedp.Click(a.Click, chromedp.ByQuery))
		case a.Press != nil:
			key, err := resolveKey(a.Press.Key)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("cdp action %d: %w", i, err)
			}
			tasks = append(tasks, chromedp.SendKeys(a.Press.Selector, key, chromedp.ByQuery))
		case a.Select != nil:
			tasks = append(tasks, chromedp.SetValue(a.Select.Selector, a.Select.Value, chromedp.ByQuery))
		case a.Check != "":
			tasks = append(tasks, setChecked(a.Check, true))
		case a.Uncheck != "":
			tasks = append(tasks, setChecked(a.Uncheck, false))
		case a.Screenshot != nil:
			buf := new([]byte)
			if a.Screenshot.Selector != "" {
				tasks = append(tasks, chromedp.Screenshot(a.Screenshot.Selector, buf, chromedp.ByQuery))
			} else {
				tasks = append(tasks, chromedp.FullScreenshot(buf, 100))
			}
			path, err := security.ResolveWorkdirPath("cdp.screenshot.path", workdir, a.Screenshot.Path)
			if err != nil {
				return nil, nil, nil, err
			}
			shots = append(shots, shot{path: path, buf: buf})
		case a.SendKeys != nil:
			tasks = append(tasks, chromedp.SendKeys(a.SendKeys.Selector, a.SendKeys.Value, chromedp.ByQuery))
		case a.Upload != nil:
			// Set a file on an <input type=file> (#75). The file must exist inside the
			// scenario workdir, so a spec cannot upload arbitrary host files.
			file, err := security.ResolveWorkdirPath("cdp.upload.file", workdir, a.Upload.File)
			if err != nil {
				return nil, nil, nil, err
			}
			if _, statErr := os.Stat(file); statErr != nil {
				return nil, nil, nil, fmt.Errorf("cdp action %d: upload file %q: %w", i, a.Upload.File, statErr)
			}
			tasks = append(tasks, chromedp.SetUploadFiles(a.Upload.Selector, []string{file}, chromedp.ByQuery))
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
			tasks = append(tasks, chromedp.Text(a.Text, s, chromedp.ByQuery, chromedp.NodeVisible))
			caps = append(caps, capture{str: s})
		case a.Title:
			s := new(string)
			tasks = append(tasks, chromedp.Title(s))
			caps = append(caps, capture{str: s})
		case a.Attribute != nil:
			s := new(string)
			ok := new(bool)
			tasks = append(tasks, chromedp.AttributeValue(a.Attribute.Selector, a.Attribute.Name, s, ok, chromedp.ByQuery))
			caps = append(caps, capture{str: s})
		case a.Eval != "":
			j := new(json.RawMessage)
			tasks = append(tasks, chromedp.Evaluate(a.Eval, j))
			caps = append(caps, capture{js: j})
		default:
			return nil, nil, nil, fmt.Errorf("cdp action %d sets no recognized action", i)
		}
	}
	if len(tasks) == 0 {
		return nil, nil, nil, errors.New("cdp step has no actions")
	}
	return tasks, caps, shots, nil
}

// setChecked ticks or unticks a checkbox/radio by setting its checked property
// and dispatching a change event, so listeners react as they would to a click.
// Selecting by property (not a click) keeps the action deterministic regardless
// of the element's current state.
func setChecked(selector string, checked bool) chromedp.Action {
	sel, _ := json.Marshal(selector)
	js := fmt.Sprintf(`(() => {
	const el = document.querySelector(%s);
	if (!el) throw new Error('check: no element for selector ' + %s);
	el.checked = %t;
	el.dispatchEvent(new Event('change', {bubbles: true}));
	return el.checked;
})()`, sel, sel, checked)
	return chromedp.Evaluate(js, nil)
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
	return "", fmt.Errorf("unsupported press key %q (use a single character or one of Enter/Tab/Escape/Backspace/Delete/Arrow*/Space)", name)
}
