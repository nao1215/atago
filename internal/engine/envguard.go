package engine

import (
	"fmt"
	"strings"

	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

// envRefGuard refuses a suite.env value that references a ${name} nothing
// defines, naming the key, the reference, and what IS defined (#399).
//
// Without it the literal text `${proxy_url}` reaches the child process as the
// value of the variable. That is bad in a way a missing command is not: the
// child does not fail on it, it uses it — `GOPROXY=${proxy_url}` becomes a
// baffling proxy error from the Go toolchain, several layers away from the typo
// in the spec, or a silent pass when the program treats the value as opaque.
//
// It applies to suite.env and not to a scenario's own `env:`, and the asymmetry
// is deliberate. suite.env exists to carry values suite.setup PRODUCED — a
// `store` capture, the ephemeral address a service published through
// `ready.store` — so a name that never resolves there is always an ordering or
// spelling mistake. A scenario's env: is authored literal text, and leaving an
// unresolved ${NAME} verbatim is a documented contract with its own example
// (examples/extend_host_env.atago.yaml): a bare ${NAME} without the `env:`
// prefix stays literal so that adding ${env:...} could not silently change what
// an existing value means.
//
// For the same reason `${env:NAME}` is exempt here too: a suite that passes an
// OPTIONAL host variable through must keep working when the host does not set
// it, exactly as the scenario-level rule promises.
func envRefGuard(st *store.Store, label string, env map[string]string) string {
	for _, key := range spec.SortedKeys(toEnvSet(env)) {
		for _, name := range st.Unresolved(env[key]) {
			if strings.HasPrefix(name, envRefPrefix) {
				continue // an optional host variable, per the documented rule
			}
			return unresolvedEnvRefMsg(st, label, key, name)
		}
	}
	return ""
}

// toEnvSet adapts an env map to the key-set shape SortedKeys takes, so the
// guard reports keys in a stable order rather than a map's.
func toEnvSet(env map[string]string) map[string]bool {
	out := make(map[string]bool, len(env))
	for k := range env {
		out[k] = true
	}
	return out
}

// envRefPrefix marks a reference resolved from the host environment rather than
// from the store.
const envRefPrefix = "env:"

func unresolvedEnvRefMsg(st *store.Store, label, key, name string) string {
	msg := fmt.Sprintf(
		"%s %s references ${%s}, but no variable with that name is defined yet, so the child process would "+
			"receive the literal text %q as the value", label, key, name, "${"+name+"}")
	if defined := st.Names(); len(defined) > 0 {
		msg += fmt.Sprintf(" (defined here: %s)", strings.Join(defined, ", "))
	}
	return msg + fmt.Sprintf("; a `store` or a service `ready.store` that defines ${%s} must run BEFORE the value is used, "+
		"or write $${%s} for the literal text", name, name)
}

// resolvableEnv drops the suite.env entries that cannot resolve yet, for the
// children that run DURING suite setup (#399).
//
// suite.env is scenario configuration, and setup steps only inherit it as a
// side effect of sharing the env-merging path. The values it exists to carry —
// the address a service publishes through `ready.store`, a value a later
// `store` step captures — cannot exist while the step that produces them is
// still running, so guarding a setup step against them would refuse the exact
// pattern suite-level setup is for: the run that produces the value would fail
// because a later key referenced it.
//
// Passing the literal `${name}` on instead is the option that must not happen:
// a child that receives GOPROXY=${proxy_url} does not fail on it, it USES it,
// and the resulting error arrives several layers away from the spec. So a
// not-yet-resolvable entry is simply not part of that child's environment yet.
// Nothing is silently lost: by the time scenarios run, every entry has to
// resolve or envRefGuard fails the scenario by name.
func resolvableEnv(st *store.Store, env map[string]string) map[string]string {
	if len(env) == 0 {
		return env
	}
	out := make(map[string]string, len(env))
	for k, v := range env {
		if len(st.Unresolved(v)) == 0 {
			out[k] = v
		}
	}
	return out
}
