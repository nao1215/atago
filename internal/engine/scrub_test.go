package engine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/loader"
)

// scrubSpec emits a line carrying a volatile auto-increment id and snapshots
// stdout, with a spec-level scrub rule that normalizes the id to a placeholder.
func scrubSpec(id string, withScrub bool) string {
	scrubBlock := ""
	if withScrub {
		scrubBlock = "scrub:\n  - {pattern: 'id=\\d+', placeholder: 'id=<ID>'}\n"
	}
	return `
version: "1"
suite:
  name: scrub
` + scrubBlock + `scenarios:
  - name: snapshot with a volatile id
    steps:
      - run:
          shell: true
          command: echo "user id=` + id + `"
      - assert:
          stdout:
            snapshot: out.snap
`
}

// TestEngine_ScrubDeterminizesSnapshot proves the #137 scrub layer end-to-end:
// a golden recorded from one run matches a later run whose only difference is a
// different auto-increment id, because the scrub rule normalizes both ids to the
// same placeholder BEFORE the snapshot compare. This is the metamorphic property
// that kills the flake — two observably different outputs, one stable golden.
func TestEngine_ScrubDeterminizesSnapshot(t *testing.T) {
	skipOnWindows(t)
	t.Parallel()
	dir := t.TempDir()
	specPath := filepath.Join(dir, "s.atago.yaml")

	// Record the golden from a run that printed id=4711.
	record := scrubSpec("4711", true)
	if err := os.WriteFile(specPath, []byte(record), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := loader.LoadBytes(specPath, []byte(record))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	eng := New()
	eng.UpdateSnapshots = true
	if res := eng.Run(context.Background(), s, specPath); res.Status != StatusPassed {
		t.Fatalf("record status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
	golden, err := os.ReadFile(filepath.Join(dir, "out.snap"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if !strings.Contains(string(golden), "id=<ID>") || strings.Contains(string(golden), "4711") {
		t.Fatalf("golden should carry the scrubbed placeholder, not the raw id:\n%s", golden)
	}

	// A later run printing a DIFFERENT id still matches the same golden.
	compare := scrubSpec("99999", true)
	s2, err := loader.LoadBytes(specPath, []byte(compare))
	if err != nil {
		t.Fatalf("load compare: %v", err)
	}
	eng2 := New() // UpdateSnapshots off → real comparison
	if res := eng2.Run(context.Background(), s2, specPath); res.Status != StatusPassed {
		t.Fatalf("compare status = %s, want passed (scrub should normalize the new id): %+v", res.Status, res.Scenarios)
	}

	// Control: without the scrub rule, the same differing id fails the compare —
	// proving the scrub rule, not some accident, is what makes it pass.
	noScrub := scrubSpec("99999", false)
	s3, err := loader.LoadBytes(specPath, []byte(noScrub))
	if err != nil {
		t.Fatalf("load no-scrub: %v", err)
	}
	if res := New().Run(context.Background(), s3, specPath); res.Status != StatusFailed {
		t.Fatalf("no-scrub status = %s, want failed (raw id must not match the placeholder golden)", res.Status)
	}
}
