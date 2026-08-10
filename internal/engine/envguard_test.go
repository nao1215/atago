package engine

import (
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/store"
)

// TestEnvRefGuard_UnresolvedReferenceIsRefused is the #399 regression: an env
// value referencing a name nothing defines used to reach the child process as
// the literal text `${name}`. A child does not fail on that — it uses it — so
// the mistake surfaced as a baffling error from the program under test rather
// than as the spec error it is.
func TestEnvRefGuard_UnresolvedReferenceIsRefused(t *testing.T) {
	t.Parallel()
	st := store.New()
	st.Set("proxy_addr", "127.0.0.1:1234")

	msg := envRefGuard(st, "suite.env", map[string]string{"GOPROXY": "http://${proxy_url}"})
	if msg == "" {
		t.Fatal("an unresolved env reference must be refused")
	}
	for _, want := range []string{"suite.env", "GOPROXY", "${proxy_url}", "literal text"} {
		if !strings.Contains(msg, want) {
			t.Errorf("message %q should mention %q", msg, want)
		}
	}
	// Naming what IS defined is the difference between "go find your typo" and
	// "you meant proxy_addr".
	if !strings.Contains(msg, "proxy_addr") {
		t.Errorf("message %q should list the names that are defined", msg)
	}
}

// TestEnvRefGuard_ResolvableValuesPass keeps the guard from getting in the way
// of every ordinary env value, including a deliberate $${literal}.
func TestEnvRefGuard_ResolvableValuesPass(t *testing.T) {
	t.Parallel()
	st := store.New()
	st.Set("addr", "127.0.0.1:9999")
	env := map[string]string{
		"PLAIN":   "value",
		"USES":    "http://${addr}/mod",
		"LITERAL": "$${not_a_reference}",
	}
	if msg := envRefGuard(st, "suite.env", env); msg != "" {
		t.Errorf("resolvable env must pass, got %q", msg)
	}
}

// TestEnvRefGuard_OptionalHostVariableIsExempt pins the deliberate asymmetry:
// a ${env:NAME} the host does not set is left alone here, because a suite that
// passes an OPTIONAL host variable through must keep working — that is the
// documented rule examples/extend_host_env.atago.yaml exists to state, and the
// guard must not quietly overrule it.
func TestEnvRefGuard_OptionalHostVariableIsExempt(t *testing.T) {
	t.Parallel()
	msg := envRefGuard(store.New(), "suite.env", map[string]string{"TOKEN": "${env:ATAGO_DEFINITELY_UNSET_9f2}"})
	if msg != "" {
		t.Errorf("an unset host variable must stay literal, got %q", msg)
	}
}

// TestEnvRefGuard_ReportsAStableKey pins that the guard reports keys in sorted
// order: a map's iteration order would make the same broken spec report a
// different key from run to run.
func TestEnvRefGuard_ReportsAStableKey(t *testing.T) {
	t.Parallel()
	env := map[string]string{"ZULU": "${nope}", "ALPHA": "${nope}"}
	for range 20 {
		if msg := envRefGuard(store.New(), "suite.env", env); !strings.Contains(msg, "ALPHA") {
			t.Fatalf("expected the first key in sorted order, got %q", msg)
		}
	}
}

// TestResolvableEnv_DropsWhatCannotResolveYet covers the suite-setup rule: the
// value a setup step is about to produce cannot resolve while that step runs,
// so it is left out of that child's environment rather than passed on as
// literal text — and guarding it instead would refuse the very pattern
// suite-level setup exists for.
func TestResolvableEnv_DropsWhatCannotResolveYet(t *testing.T) {
	t.Parallel()
	st := store.New()
	st.Set("suitedir", "/tmp/suite")

	got := resolvableEnv(st, map[string]string{
		"READY":   "${suitedir}/x",
		"PENDING": "${proxy_url}",
	})
	if _, ok := got["PENDING"]; ok {
		t.Error("a not-yet-resolvable entry must not reach the child at all")
	}
	if got["READY"] != "${suitedir}/x" {
		t.Errorf("a resolvable entry must be kept verbatim for later expansion, got %q", got["READY"])
	}
}
