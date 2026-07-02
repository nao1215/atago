package engine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/loader"
)

// Regression for issue #11: a secret printed by the tool under test must be
// masked in a committed snapshot golden file written under --update-snapshots,
// not written verbatim.
func TestEngine_SnapshotMasksSecrets(t *testing.T) {
	const secret = "ghp_super_secret_token_value"
	t.Setenv("ATAGO_TEST_SECRET", secret)

	dir := t.TempDir()
	specPath := filepath.Join(dir, "s.atago.yaml")
	src := `
version: "1"
suite:
  name: snap
secrets:
  - ATAGO_TEST_SECRET
scenarios:
  - name: prints a secret then snapshots stdout
    steps:
      - run:
          shell: true
          command: echo Authorization Bearer ` + envRef("ATAGO_TEST_SECRET") + `
      - assert:
          stdout:
            snapshot: out.snap
`
	if err := os.WriteFile(specPath, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := loader.LoadBytes(specPath, []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	eng := New()
	eng.UpdateSnapshots = true
	res := eng.Run(context.Background(), s, specPath)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed", res.Status)
	}

	golden, err := os.ReadFile(filepath.Join(dir, "out.snap"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if strings.Contains(string(golden), secret) {
		t.Errorf("snapshot golden contains the raw secret:\n%s", golden)
	}
	if !strings.Contains(string(golden), "***") {
		t.Errorf("snapshot golden did not mask the secret with ***:\n%s", golden)
	}
}

// Regression for issue #12: a failed service-readiness probe embeds the service's
// raw output in its error; a secret in that output must be masked before it
// reaches the report.
func TestEngine_ReadinessErrorMasksSecrets(t *testing.T) {
	const secret = "hunter2_super_secret"
	t.Setenv("ATAGO_TEST_SECRET", secret)

	src := `
version: "1"
suite:
  name: svc
secrets:
  - ATAGO_TEST_SECRET
scenarios:
  - name: service leaks a secret on a failed boot
    services:
      - name: leaky
        shell: true
        command: '` + echoThenIdle("connecting token="+envRef("ATAGO_TEST_SECRET"), 5) + `'
        ready:
          file: never-appears.txt
          timeout: 200ms
    steps:
      - run: {shell: true, command: "exit 0"}
`
	res := runSpec(t, src)
	if res.Scenarios[0].Status != StatusError {
		t.Fatalf("status = %s, want error (readiness should fail)", res.Scenarios[0].Status)
	}
	for _, st := range res.Scenarios[0].Steps {
		if strings.Contains(st.ErrMsg, secret) {
			t.Errorf("readiness error leaked the raw secret: %q", st.ErrMsg)
		}
	}
}
