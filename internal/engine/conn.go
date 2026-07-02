package engine

import (
	"fmt"
	"io"
	"time"

	"github.com/nao1215/atago/internal/spec"
)

// resolveConn returns the scenario's cached connection for a named runner, or
// opens one on first use and caches it. It centralizes the five steps every
// per-runner helper shared — cache lookup, unknown-runner error, wrong-type
// rejection, timeout parsing, and open+store — so a fix to that logic lives in
// one place instead of four near-identical copies (issue #24).
//
// stepVerb names the step kind for the unknown-runner error (e.g. "query step");
// wantType is the required runner.Type; open builds and opens the typed
// connection from the runner definition and its parsed timeout.
func resolveConn[T io.Closer](
	name, stepVerb, wantType string,
	rc runConfig,
	conns map[string]T,
	open func(rdef spec.Runner, timeout time.Duration) (T, error),
) (T, error) {
	var zero T
	if c, ok := conns[name]; ok {
		return c, nil
	}
	rdef, ok := rc.runners[name]
	if !ok {
		return zero, fmt.Errorf("%s references unknown runner %q", stepVerb, name)
	}
	if rdef.Type != wantType {
		return zero, fmt.Errorf("runner %q is not a %s runner (type %q)", name, wantType, rdef.Type)
	}
	var timeout time.Duration
	if rdef.Timeout != "" {
		d, err := time.ParseDuration(rdef.Timeout)
		if err != nil {
			return zero, fmt.Errorf("runner %q has invalid timeout %q: %w", name, rdef.Timeout, err)
		}
		timeout = d
	}
	conn, err := open(rdef, timeout)
	if err != nil {
		return zero, fmt.Errorf("runner %q: %w", name, err)
	}
	conns[name] = conn
	return conn, nil
}
