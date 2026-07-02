package engine

import (
	"context"
	"time"

	"github.com/nao1215/atago/internal/runner"
	dbrunner "github.com/nao1215/atago/internal/runner/db"
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
	return conn.Query(ctx, st.Expand(q.SQL))
}

// dbConn returns the scenario's connection for a named db runner, opening it on
// first use. Connections are scoped to a scenario (closed when it ends) so a dsn
// referencing ${workdir} yields a fresh, isolated database per scenario.
func dbConn(name string, st *store.Store, rc runConfig, conns map[string]*dbrunner.Runner) (*dbrunner.Runner, error) {
	return resolveConn(name, "query step", "db", rc, conns, func(rdef spec.Runner, timeout time.Duration) (*dbrunner.Runner, error) {
		cfg, err := dbrunner.Resolve(rdef.Driver, st.Expand(rdef.DSN))
		if err != nil {
			return nil, err
		}
		cfg.Timeout = timeout
		return dbrunner.Open(cfg)
	})
}
