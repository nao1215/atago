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
