package engine

import (
	"context"
	"time"

	"github.com/nao1215/atago/internal/runner"
	dbrunner "github.com/nao1215/atago/internal/runner/db"
	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

// runQuery executes a query step against its named db runner, opening (and
// caching) the connection within the scenario on first use. The SQL is
// ${name}-expanded so it can reference stored values from earlier steps.
func (e *Engine) runQuery(ctx context.Context, q *spec.Query, st *store.Store, rc runConfig, conns map[string]*dbrunner.Runner) (*runner.Result, error) {
	conn, err := dbConn(q.Runner, st, rc, conns)
	if err != nil {
		return nil, err
	}
	return conn.Query(ctx, spec.WalkQueryStrings(q, st.Expand).SQL)
}

// dbConn returns the scenario's connection for a named db runner, opening it on
// first use. Connections are scoped to a scenario (closed when it ends) so a dsn
// referencing ${workdir} yields a fresh, isolated database per scenario.
func dbConn(name string, st *store.Store, rc runConfig, conns map[string]*dbrunner.Runner) (*dbrunner.Runner, error) {
	return resolveConn(name, "query step", "db", rc, conns, true, func(rdef spec.Runner, timeout time.Duration) (*dbrunner.Runner, error) {
		cfg, err := dbrunner.Resolve(rdef.Driver, st.Expand(rdef.DSN))
		if err != nil {
			return nil, err
		}
		cfg.Timeout = timeout
		// Enforce the network allowlist before opening (issue #17): db egress is
		// confined to permissions.network.allow just like HTTP, grpc, and ssh.
		// Before the pool, because database/sql connects lazily — a check after
		// it would run once the denied host had already been dialed.
		// Every peer the dsn names is checked, not just the first: a libpq dsn
		// can name a failover list, or a host beside the hostaddr the driver
		// actually dials, and an unchecked one is a hole (#497).
		for _, host := range cfg.Hosts {
			if err := security.CheckHost(rc.allow, host); err != nil {
				return nil, err
			}
		}
		return dbrunner.Open(cfg)
	})
}
