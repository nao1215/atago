// Package security implements atago's security model: masking
// secret values in reports and logs, and the safe defaults enabled by --ci.
package security

import (
	"fmt"
	"net"
	"os"
	"sort"
	"strings"

	"github.com/nao1215/atago/internal/spec"
)

// PolicyError reports a network egress that permissions.network.allow does not
// permit. It is returned by CheckHost and lets the engine flag a
// security-policy violation for grpc/ssh steps, mirroring the HTTP runner.
type PolicyError struct {
	Host  string
	Allow []string
}

func (e *PolicyError) Error() string {
	return fmt.Sprintf("network policy denies host %q (allowed: %s)", e.Host, strings.Join(e.Allow, ", "))
}

// CheckHost reports whether hostport (a "host" or "host:port") is permitted by
// the allowlist, returning a *PolicyError otherwise. An empty allowlist means no
// network policy is configured and everything is permitted.
func CheckHost(allow []string, hostport string) error {
	if len(allow) == 0 {
		return nil
	}
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}
	for _, a := range allow {
		if a == host || a == hostport {
			return nil
		}
	}
	return &PolicyError{Host: host, Allow: allow}
}

// minSecretLen avoids masking very short values that would garble unrelated
// output (e.g. a one-character token would replace every occurrence of it).
const minSecretLen = 4

// Masker replaces known secret values with a fixed placeholder.
type Masker struct {
	values []string
}

// NewMasker builds a Masker from literal secret values. Empty or very short
// values are ignored.
func NewMasker(values []string) *Masker {
	seen := map[string]bool{}
	var v []string
	for _, x := range values {
		if len(x) >= minSecretLen && !seen[x] {
			seen[x] = true
			v = append(v, x)
		}
	}
	// Order the values longest-first for deterministic, stable output. Mask now
	// unions the byte ranges of every occurrence over the original string, so
	// correctness no longer depends on order (a substring or an overlapping
	// secret is masked regardless), but a stable order keeps output reproducible.
	sort.SliceStable(v, func(i, j int) bool { return len(v[i]) > len(v[j]) })
	return &Masker{values: v}
}

// NewMaskerForSpec collects secret values for a spec from the process
// environment and from EVERY env-bearing location that can inject one of the
// names listed under `secrets:`: suite.env, defaults.scenario.env, scenario and
// service env, and the env of every run/pty step in suite.setup, suite.teardown,
// a scenario's steps, and a scenario's teardown. A secret that reaches any of
// these places can surface in a report or log, so all of them must feed the
// masker — collecting only run-step and scenario/service env leaks the rest.
func NewMaskerForSpec(s *spec.Spec) *Masker {
	if len(s.Secrets) == 0 {
		return NewMasker(nil)
	}
	var vals []string
	for _, name := range s.Secrets {
		if v := os.Getenv(name); v != "" {
			vals = append(vals, v)
		}
	}
	collect := func(env map[string]string) {
		for _, name := range s.Secrets {
			if v, ok := env[name]; ok && v != "" {
				vals = append(vals, v)
			}
		}
	}
	collectSteps := func(steps []spec.Step) {
		for i := range steps {
			st := &steps[i]
			if st.Run != nil {
				collect(st.Run.Env)
			}
			if st.PTY != nil {
				collect(st.PTY.Env)
			}
			if st.Service != nil {
				collect(st.Service.Env)
			}
		}
	}
	// Suite-wide env is exported to every scenario, setup, and teardown step.
	collect(s.Suite.Env)
	// defaults.scenario.env is merged into each scenario by the loader, but
	// collect it directly so masking does not depend on that merge order.
	if s.Defaults != nil && s.Defaults.Scenario != nil {
		collect(s.Defaults.Scenario.Env)
	}
	collectSteps(s.Suite.Setup)
	collectSteps(s.Suite.Teardown)
	for i := range s.Scenarios {
		sc := &s.Scenarios[i]
		// Scenario-level env is a first-class secret source (#38).
		collect(sc.Env)
		// Service env can carry credentials the service needs (#38).
		for k := range sc.Services {
			collect(sc.Services[k].Env)
		}
		collectSteps(sc.Steps)
		collectSteps(sc.Teardown)
	}
	return NewMasker(vals)
}

// Empty reports whether the masker has nothing to mask.
func (m *Masker) Empty() bool { return m == nil || len(m.values) == 0 }

// Mask replaces every known secret value in s with "***". It marks the byte
// ranges of every secret occurrence over the ORIGINAL string in one pass, then
// collapses each maximal covered run to a single placeholder. Scanning the
// original (rather than a sequence of ReplaceAll that mutates s) is what keeps
// two overlapping secrets — one ending with the bytes the next begins with, e.g.
// "abcXYZ" and "XYZdef" in "abcXYZdef" — both masked; a sequential replace would
// consume the shared bytes and leak the second secret's tail.
func (m *Masker) Mask(s string) string {
	if m.Empty() {
		return s
	}
	covered := make([]bool, len(s))
	any := false
	for _, v := range m.values {
		for i := 0; ; {
			j := strings.Index(s[i:], v)
			if j < 0 {
				break
			}
			start := i + j
			for k := start; k < start+len(v); k++ {
				covered[k] = true
			}
			any = true
			// Advance by one, not len(v), so overlapping occurrences of the same
			// secret are all covered — "aaaa" occurs three times in "aaaaaa".
			i = start + 1
		}
	}
	if !any {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		if covered[i] {
			b.WriteString("***")
			for i < len(s) && covered[i] {
				i++
			}
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// MaskBytes is Mask for byte slices, returning the input unchanged when there is
// nothing to mask.
func (m *Masker) MaskBytes(b []byte) []byte {
	if m.Empty() {
		return b
	}
	return []byte(m.Mask(string(b)))
}
