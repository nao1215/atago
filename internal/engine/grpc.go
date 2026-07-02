package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nao1215/atago/internal/runner"
	grpcrunner "github.com/nao1215/atago/internal/runner/grpc"
	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

// runGRPC executes a grpc step against its named grpc runner, opening (and
// caching) the connection within the scenario on first use. The step is already
// ${name}-expanded by the caller; the JSON request body is marshaled here.
func (e *Engine) runGRPC(ctx context.Context, g *spec.GRPC, st *store.Store, rc runConfig, conns map[string]*grpcrunner.Runner) (*runner.Result, error) {
	conn, err := grpcConn(g.Runner, st, rc, conns)
	if err != nil {
		return nil, err
	}
	var body []byte
	if g.JSON != nil {
		body, err = json.Marshal(g.JSON)
		if err != nil {
			return nil, fmt.Errorf("encoding grpc request body: %w", err)
		}
	}
	return conn.Invoke(ctx, g.Method, g.Header, body)
}

// grpcConn returns the scenario's connection for a named grpc runner, opening it
// on first use.
func grpcConn(name string, st *store.Store, rc runConfig, conns map[string]*grpcrunner.Runner) (*grpcrunner.Runner, error) {
	return resolveConn(name, "grpc step", "grpc", rc, conns, func(rdef spec.Runner, timeout time.Duration) (*grpcrunner.Runner, error) {
		cfg := grpcrunner.Config{Target: st.Expand(rdef.Target), TLS: rdef.TLS, Timeout: timeout}
		// Enforce the network allowlist before dialing (issue #17): grpc egress is
		// confined to permissions.network.allow just like HTTP.
		if err := security.CheckHost(rc.allow, cfg.Target); err != nil {
			return nil, err
		}
		return grpcrunner.Open(cfg)
	})
}
