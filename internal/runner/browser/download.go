package browser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/nao1215/atago/internal/diag"
)

// willBeginGrace bounds how long a completed download waits for the
// will-begin event that carries its suggested filename. Chrome sends that event
// before the completion event, so it is normally already queued; the bound only
// matters when the browser sent none, and then the GUID is used as the name.
const willBeginGrace = time.Second

// downloadAction returns a chromedp action that captures a click-triggered
// download into destDir with a deterministic name (#75). It enables Chrome's
// download events into destDir, clicks the trigger selector, waits for the
// download to complete, and renames the GUID-named file to the server-suggested
// filename. The final base name is written to *name so it can be captured as the
// cdp step's value for file/dir/pdf/image assertions.
//
// The action keeps the browser surface intentionally narrow: no scripted dialogs
// or conditional state, just "click this, save what comes down, here".
func downloadAction(clickSelector, destDir string, name *string) chromedp.Action[chromedp.Void] {
	return chromedp.Func(func(ctx context.Context, t *chromedp.Target) error {
		if err := os.MkdirAll(destDir, 0o750); err != nil {
			return fmt.Errorf("cdp download: %w", err)
		}

		willBegin, done := listenDownloadEvents(ctx, t)

		// AllowAndName saves each download under its GUID in destDir and emits the
		// progress events we listen for.
		eventsEnabled := true
		if _, err := cdp.Call(ctx, t, browser.SetDownloadBehavior, browser.SetDownloadBehaviorParams{
			Behavior:      browser.SetDownloadBehaviorBehaviorAllowAndName,
			DownloadPath:  destDir,
			EventsEnabled: &eventsEnabled,
		}); err != nil {
			return diag.BrowserActionFailed.Errorf("cdp download: enabling downloads: %w", err)
		}

		if _, err := chromedp.Click(chromedp.CSS(clickSelector))(ctx, t); err != nil {
			return diag.BrowserActionFailed.Errorf("cdp download: clicking %q: %w", clickSelector, err)
		}

		var guid string
		select {
		case guid = <-done:
		case <-ctx.Done():
			return diag.StepTimeout.Errorf("cdp download: timed out waiting for the download to finish: %w", ctx.Err())
		}

		final := downloadFinalName(guid, willBegin)
		src := filepath.Join(destDir, guid)
		dst := filepath.Join(destDir, final)
		if src != dst {
			if err := os.Rename(src, dst); err != nil {
				return diag.StepFileUnusable.Errorf("cdp download: naming the downloaded file: %w", err)
			}
		}
		*name = final
		return nil
	})
}

// listenDownloadEvents subscribes to the target's download events: willBegin
// receives the server-suggested filename, done the GUID of a completed
// download. The subscriptions start before it returns, so no event that the
// following click triggers is lost, and they end with ctx. Both channels are
// buffered and only the first value is kept, so a reader never blocks the
// forwarding goroutines.
func listenDownloadEvents(ctx context.Context, t *chromedp.Target) (willBegin, done chan string) {
	willBegin = make(chan string, 1)
	done = make(chan string, 1)
	begins := cdp.Events(ctx, t, browser.DownloadWillBegin)
	progress := cdp.Events(ctx, t, browser.DownloadProgress)
	go func() {
		for e, err := range begins {
			if err != nil {
				return
			}
			select {
			case willBegin <- e.SuggestedFilename:
			default:
			}
		}
	}()
	go func() {
		for e, err := range progress {
			if err != nil {
				return
			}
			if e.State == browser.DownloadProgressStateCompleted {
				select {
				case done <- e.GUID:
				default:
				}
			}
		}
	}()
	return willBegin, done
}

// downloadFinalName picks the deterministic base name for a completed
// download: the suggested filename captured from the will-begin event, falling
// back to the GUID when the browser did not provide one. filepath.Base defends
// against a server-suggested name containing path separators.
func downloadFinalName(guid string, willBegin <-chan string) string {
	suggested := guid
	timer := time.NewTimer(willBeginGrace)
	defer timer.Stop()
	select {
	case s := <-willBegin:
		if s != "" {
			suggested = s
		}
	case <-timer.C:
	}
	final := filepath.Base(suggested)
	if final == "." || final == "/" || final == "" {
		final = guid
	}
	return final
}
