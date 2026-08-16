package engine

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/nao1215/atago/internal/runner"
	browserrunner "github.com/nao1215/atago/internal/runner/browser"
	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

// runCDP executes a cdp step against its named browser runner, launching (and
// caching) the session within the scenario on first use. The step is already
// ${name}-expanded by the caller.
func (e *Engine) runCDP(ctx context.Context, c *spec.CDP, workdir string, st *store.Store, rc runConfig, conns map[string]*browserrunner.Runner) (*runner.Result, error) {
	// Before the session is launched: a denied navigation should cost nothing,
	// and the answer does not depend on the browser.
	if err := checkNavigation(rc.allow, c.Actions); err != nil {
		return nil, err
	}
	conn, err := browserConn(c.Runner, st, rc, conns)
	if err != nil {
		return nil, err
	}
	return conn.Run(ctx, c.Actions, workdir)
}

// checkNavigation holds every `navigate:` action to permissions.network.allow,
// the way the http, grpc, ssh, and db runners are held to it (#486).
//
// It checks the URLs the SPEC names, and nothing else. A page loaded from an
// allowed host is free to fetch scripts, images, and XHRs from anywhere, and
// none of that passes through a navigate action — so the denial says which half
// it covers rather than letting the exit code imply a guarantee the browser
// runner does not make.
//
// A navigation naming no host reaches no network (about:blank, a file:// page
// serving the spec's own fixtures, a data: URL) and is left alone, like a
// sqlite dsn. So is a URL Chrome could not parse: it names no host to check,
// and the navigation fails on its own.
func checkNavigation(allow []string, actions []spec.CDPAction) error {
	if len(allow) == 0 {
		return nil
	}
	for _, a := range actions {
		if a.Navigate == "" {
			continue
		}
		u, err := url.Parse(a.Navigate)
		if err != nil || u.Host == "" {
			continue
		}
		if err := security.CheckHost(allow, u.Host); err != nil {
			return fmt.Errorf("%w — the browser runner holds `navigate:` to the policy, not the requests the loaded page then makes", err)
		}
	}
	return nil
}

// browserConn returns the scenario's session for a named browser runner,
// launching it on first use.
func browserConn(name string, _ *store.Store, rc runConfig, conns map[string]*browserrunner.Runner) (*browserrunner.Runner, error) {
	return resolveConn(name, "cdp step", "browser", rc, conns, false, func(rdef spec.Runner, timeout time.Duration) (*browserrunner.Runner, error) {
		// Headless defaults to true; an explicit `headless: false` runs headed.
		headless := rdef.Headless == nil || *rdef.Headless
		return browserrunner.Open(browserrunner.Config{
			Headless: headless,
			ExecPath: rdef.ExecPath,
			Args:     rdef.BrowserArgs,
			Timeout:  timeout,
		})
	})
}
