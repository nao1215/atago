package security

import (
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/loader"
)

func TestMasker_Mask(t *testing.T) {
	t.Parallel()
	m := NewMasker([]string{"supersecret", "tok", "abcd"})
	// "tok" is too short (<4) and must be ignored.
	got := m.Mask("value=supersecret tok=tok end=abcd")
	if strings.Contains(got, "supersecret") {
		t.Errorf("secret leaked: %q", got)
	}
	if !strings.Contains(got, "tok=tok") {
		t.Errorf("short value should not be masked: %q", got)
	}
	if strings.Contains(got, "abcd") {
		t.Errorf("abcd should be masked: %q", got)
	}
}

// TestMasker_OverlappingSecretsNoLeak is a regression: masking must apply the
// LONGEST secret first, or a short secret that is a substring of a longer one
// (e.g. a token and that same token with a suffix) masks only the prefix and
// leaks the remainder into reports and snapshots.
func TestMasker_OverlappingSecretsNoLeak(t *testing.T) {
	t.Parallel()
	// Deliberately supply the short value first, the order a naive masker leaks on.
	m := NewMasker([]string{"abcd", "abcdefgh"})
	got := m.Mask("token is abcdefgh here")
	if strings.Contains(got, "efgh") {
		t.Errorf("longer secret partially leaked: %q", got)
	}
	if strings.Contains(got, "abcd") {
		t.Errorf("secret leaked: %q", got)
	}
}

// TestMasker_AdjacentOverlapNoLeak is a regression: two secrets that overlap
// end-to-start (one ends with the bytes the other begins with) must both be
// masked. Sequential ReplaceAll masked the first and consumed the shared bytes,
// so the second secret's tail leaked though it appeared verbatim in the output.
func TestMasker_AdjacentOverlapNoLeak(t *testing.T) {
	t.Parallel()
	m := NewMasker([]string{"abcXYZ", "XYZdef"})
	got := m.Mask("token=abcXYZdef end")
	if strings.Contains(got, "def") {
		t.Errorf("overlapping secret tail leaked: %q", got)
	}
	if strings.Contains(got, "abc") {
		t.Errorf("overlapping secret head leaked: %q", got)
	}
}

// TestMasker_SelfOverlappingSecret is a regression: a secret that overlaps its
// own repeats must be masked at every occurrence. Advancing past a hit by the
// secret's full length skipped the overlapping repeats, so "aaaa" in "aaaaaa"
// masked only the first four bytes and leaked "aa".
func TestMasker_SelfOverlappingSecret(t *testing.T) {
	t.Parallel()
	m := NewMasker([]string{"aaaa"})
	if got := m.Mask("token=aaaaaa"); got != "token=***" {
		t.Errorf("Mask self-overlap = %q, want %q", got, "token=***")
	}
}

func TestMasker_MaskBytes(t *testing.T) {
	t.Parallel()
	m := NewMasker([]string{"supersecret"})
	got := m.MaskBytes([]byte("here is supersecret data"))
	if strings.Contains(string(got), "supersecret") {
		t.Errorf("secret leaked: %q", got)
	}
	// An empty masker returns the input unchanged.
	in := []byte("nothing to mask")
	if out := NewMasker(nil).MaskBytes(in); string(out) != string(in) {
		t.Errorf("empty masker changed bytes: %q", out)
	}
}

func TestMasker_Empty(t *testing.T) {
	t.Parallel()
	if !NewMasker(nil).Empty() {
		t.Error("nil masker should be empty")
	}
	if NewMasker([]string{"longenough"}).Empty() {
		t.Error("masker with a value should not be empty")
	}
}

func TestNewMaskerForSpec_FromEnvOverride(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: s
secrets:
  - API_TOKEN
scenarios:
  - name: uses token
    steps:
      - run:
          command: echo hi
          env:
            API_TOKEN: very-secret-token
      - assert:
          exit_code: 0
`
	spc, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	m := NewMaskerForSpec(spc)
	if got := m.Mask("here is very-secret-token in output"); strings.Contains(got, "very-secret-token") {
		t.Errorf("token from env override not masked: %q", got)
	}
}

// TestNewMaskerForSpec_FromScenarioEnv verifies that a declared secret injected
// through scenario-level env is masked (issue #38).
func TestNewMaskerForSpec_FromScenarioEnv(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: s
secrets:
  - API_TOKEN
scenarios:
  - name: scenario env carries the secret
    env:
      API_TOKEN: scenario-secret-value
    steps:
      - run:
          command: echo hi
      - assert:
          exit_code: 0
`
	spc, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	m := NewMaskerForSpec(spc)
	if got := m.Mask("leaked scenario-secret-value here"); strings.Contains(got, "scenario-secret-value") {
		t.Errorf("token from scenario env not masked: %q", got)
	}
}

// TestNewMaskerForSpec_FromServiceEnv verifies that a declared secret injected
// through a background service's env is masked (issue #38).
func TestNewMaskerForSpec_FromServiceEnv(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: s
secrets:
  - DB_PASSWORD
scenarios:
  - name: service env carries the secret
    services:
      - name: db
        command: sleep 1
        env:
          DB_PASSWORD: service-secret-value
    steps:
      - run:
          command: echo hi
      - assert:
          exit_code: 0
`
	spc, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	m := NewMaskerForSpec(spc)
	if got := m.Mask("leaked service-secret-value here"); strings.Contains(got, "service-secret-value") {
		t.Errorf("token from service env not masked: %q", got)
	}
}

// TestNewMaskerForSpec_FromEnvSources verifies that a declared secret injected
// through any env-bearing location — a pty step, suite.env, suite.setup /
// suite.teardown steps, scenario teardown steps, and defaults.scenario.env — is
// masked, not just run-step and scenario/service env.
func TestNewMaskerForSpec_FromEnvSources(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		secret string
		src    string
	}{
		{
			name:   "pty step env",
			secret: "pty-secret-value",
			src: `
version: "1"
suite:
  name: s
secrets:
  - TOKEN
scenarios:
  - name: pty carries the secret
    steps:
      - pty:
          command: cat
          env:
            TOKEN: pty-secret-value
          session:
            - send: ""
`,
		},
		{
			name:   "suite env",
			secret: "suite-secret-value",
			src: `
version: "1"
suite:
  name: s
  env:
    TOKEN: suite-secret-value
secrets:
  - TOKEN
scenarios:
  - name: uses suite env
    steps:
      - run:
          command: echo hi
      - assert:
          exit_code: 0
`,
		},
		{
			name:   "suite setup step env",
			secret: "setup-secret-value",
			src: `
version: "1"
suite:
  name: s
  setup:
    - run:
        command: echo hi
        env:
          TOKEN: setup-secret-value
secrets:
  - TOKEN
scenarios:
  - name: needs setup
    steps:
      - run:
          command: echo hi
      - assert:
          exit_code: 0
`,
		},
		{
			name:   "suite teardown step env",
			secret: "suiteteardown-secret-value",
			src: `
version: "1"
suite:
  name: s
  teardown:
    - run:
        command: echo hi
        env:
          TOKEN: suiteteardown-secret-value
secrets:
  - TOKEN
scenarios:
  - name: has suite teardown
    steps:
      - run:
          command: echo hi
      - assert:
          exit_code: 0
`,
		},
		{
			name:   "scenario teardown step env",
			secret: "teardown-secret-value",
			src: `
version: "1"
suite:
  name: s
secrets:
  - TOKEN
scenarios:
  - name: has teardown
    steps:
      - run:
          command: echo hi
      - assert:
          exit_code: 0
    teardown:
      - run:
          command: echo bye
          env:
            TOKEN: teardown-secret-value
`,
		},
		{
			name:   "defaults scenario env",
			secret: "defaults-secret-value",
			src: `
version: "1"
suite:
  name: s
secrets:
  - TOKEN
defaults:
  scenario:
    env:
      TOKEN: defaults-secret-value
scenarios:
  - name: inherits default env
    steps:
      - run:
          command: echo hi
      - assert:
          exit_code: 0
`,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			spc, err := loader.LoadBytes("t.atago.yaml", []byte(tc.src))
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			m := NewMaskerForSpec(spc)
			if got := m.Mask("leaked " + tc.secret + " here"); strings.Contains(got, tc.secret) {
				t.Errorf("secret from %s not masked: %q", tc.name, got)
			}
		})
	}
}
