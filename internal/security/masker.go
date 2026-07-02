// Package security implements atago's security model (spec.md §28): masking
// secret values in reports and logs, and the safe defaults enabled by --ci.
package security

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/nao1215/atago/internal/spec"
)

// PolicyError reports a network egress that permissions.network.allow does not
// permit (spec.md §28.2). It is returned by CheckHost and lets the engine flag a
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
	return &Masker{values: v}
}

// NewMaskerForSpec collects secret values for a spec from the process
// environment and from per-step env overrides of the names listed under
// `secrets:` (spec.md §28.3).
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
	for i := range s.Scenarios {
		sc := &s.Scenarios[i]
		// Scenario-level env is a first-class secret source (#38).
		collect(sc.Env)
		// Service env can carry credentials the service needs (#38).
		for k := range sc.Services {
			collect(sc.Services[k].Env)
		}
		for j := range sc.Steps {
			st := &sc.Steps[j]
			if st.Run == nil {
				continue
			}
			collect(st.Run.Env)
		}
	}
	return NewMasker(vals)
}

// Empty reports whether the masker has nothing to mask.
func (m *Masker) Empty() bool { return m == nil || len(m.values) == 0 }

// Mask replaces every known secret value in s with "***".
func (m *Masker) Mask(s string) string {
	if m.Empty() {
		return s
	}
	for _, v := range m.values {
		s = strings.ReplaceAll(s, v, "***")
	}
	return s
}

// MaskBytes is Mask for byte slices, returning the input unchanged when there is
// nothing to mask.
func (m *Masker) MaskBytes(b []byte) []byte {
	if m.Empty() {
		return b
	}
	return []byte(m.Mask(string(b)))
}
