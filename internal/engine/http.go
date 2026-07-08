package engine

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/runner"
	httprunner "github.com/nao1215/atago/internal/runner/http"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

// runHTTPStep executes an http step, applying retry/until polling when
// requested — the http counterpart of runStep. The
// request is re-issued until until passes or the attempt budget is spent; the
// last response is what later steps observe. It returns the final response, the
// until CheckResult (nil when no retry is configured), whether a failure was a
// network-policy violation, and an execution error.
func (e *Engine) runHTTPStep(ctx context.Context, h *spec.HTTP, st *store.Store, rc runConfig, workdir, specDir string) (*runner.Result, []*assert.CheckResult, bool, error) {
	if h.Retry == nil {
		r, secViolation, err := e.runHTTP(ctx, h, st, rc, workdir)
		return r, nil, secViolation, err
	}
	// The policy-violation flag only matters alongside an error, so the shared
	// poll loop carries it inside the closure.
	secViolation := false
	env := assert.Env{Workdir: workdir, SpecDir: specDir, UpdateSnapshots: e.UpdateSnapshots, Secrets: rc.masker.MaskBytes, Scrub: rc.scrubber.Apply}
	last, checks, err := pollUntil(ctx, h.Retry, st, env, func(ctx context.Context) (*runner.Result, error) {
		r, sv, rerr := e.runHTTP(ctx, h, st, rc, workdir)
		secViolation = sv
		return r, rerr
	})
	if err != nil {
		return nil, nil, secViolation, err
	}
	return last, checks, false, nil
}

// runHTTP executes an http step. It returns the captured response, whether the
// failure was a network-policy violation (mapped to exit 6 by the caller), and
// any execution error. The http step is already ${name}-expanded by the caller.
// workdir confines the step's file payloads (body_file/files) and downloads
// (body_to) to the scenario's isolated directory.
func (e *Engine) runHTTP(ctx context.Context, h *spec.HTTP, st *store.Store, rc runConfig, workdir string) (*runner.Result, bool, error) {
	cfg, err := resolveHTTPConfig(h, st, rc)
	if err != nil {
		return nil, false, err
	}
	cfg.Workdir = workdir
	res, err := httprunner.New(cfg).Do(ctx, h)
	if err != nil {
		var pe *httprunner.PolicyError
		if errors.As(err, &pe) {
			return nil, true, err
		}
		return nil, false, err
	}
	return res, false, nil
}

// resolveHTTPConfig builds the http runner config for a step from its named
// runner (if any) and the spec's network allowlist. base_url is ${name}-expanded
// so it can reference stored values or built-ins.
func resolveHTTPConfig(h *spec.HTTP, st *store.Store, rc runConfig) (httprunner.Config, error) {
	cfg := httprunner.Config{Allow: rc.allow}
	var runnerTimeout string
	if h.Runner != "" {
		r, ok := rc.runners[h.Runner]
		if !ok {
			return cfg, fmt.Errorf("http step references unknown runner %q", h.Runner)
		}
		if r.Type != "http" {
			return cfg, fmt.Errorf("runner %q is not an http runner (type %q)", h.Runner, r.Type)
		}
		cfg.BaseURL = st.Expand(r.BaseURL)
		runnerTimeout = r.Timeout
	}
	// http requests join the step-timeout precedence chain (#17): runner
	// timeout > defaults.run.timeout > suite.timeout > built-in 60s ("0"
	// disables). An http step has no step-level timeout of its own.
	timeoutStr, _ := resolveTimeout("", runnerTimeout, rc.defaultsRunTimeout, rc.suiteTimeout)
	d, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return cfg, fmt.Errorf("runner %q has invalid timeout %q: %w", h.Runner, timeoutStr, err)
	}
	cfg.Timeout = d
	return cfg, nil
}

// allowedHosts returns the hostnames permitted by the spec's network policy, or
// nil when no allowlist is declared (no run-time restriction). Each entry is
// normalized to a bare host so a user may write a full URL, a host:port, or a
// bare host in `permissions.network.allow`.
func allowedHosts(s *spec.Spec) []string {
	if s.Permissions == nil || s.Permissions.Network == nil {
		return nil
	}
	var out []string
	for _, a := range s.Permissions.Network.Allow {
		out = append(out, normalizeHost(a))
	}
	return out
}

// normalizeHost reduces an allowlist entry to the value the runner compares
// against url.Hostname()/url.Host: a full URL becomes its host, otherwise the
// entry is used as-is (covering bare host and host:port forms).
func normalizeHost(entry string) string {
	if u, err := url.Parse(entry); err == nil && u.Host != "" {
		return u.Host
	}
	return entry
}
