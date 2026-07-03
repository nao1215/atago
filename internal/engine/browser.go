package engine

import (
	"context"
	"time"

	"github.com/nao1215/atago/internal/runner"
	browserrunner "github.com/nao1215/atago/internal/runner/browser"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

// runCDP executes a cdp step against its named browser runner, launching (and
// caching) the session within the scenario on first use. The step is already
// ${name}-expanded by the caller.
func (e *Engine) runCDP(ctx context.Context, c *spec.CDP, workdir string, st *store.Store, rc runConfig, conns map[string]*browserrunner.Runner) (*runner.Result, error) {
	conn, err := browserConn(c.Runner, st, rc, conns)
	if err != nil {
		return nil, err
	}
	return conn.Run(ctx, c.Actions, workdir)
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
