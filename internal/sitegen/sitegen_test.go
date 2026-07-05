package sitegen

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGenerate_WritesEverySiteFile drives Generate against a synthetic repo root
// and asserts it writes exactly the file set Files reports, byte-for-byte. It
// also seeds a doc/e2e page and a schema so the index's link sections
// (writeLinkSection / exists) render their populated branches rather than the
// empty-directory skip.
func TestGenerate_WritesEverySiteFile(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// Seed the sources the index links to so both link sections are exercised.
	mustWrite(t, filepath.Join(root, "doc/e2e/echo.md"), "# echo\n")
	mustWrite(t, filepath.Join(root, "schema/atago.schema.json"), "{}\n")

	if err := Generate(root); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	files, err := Files(root)
	if err != nil {
		t.Fatalf("Files: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("Files returned no site files")
	}
	for name, want := range files {
		got, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Errorf("Generate did not write %s: %v", name, err)
			continue
		}
		if !bytes.Equal(got, want) {
			t.Errorf("written %s differs from Files() content (%d vs %d bytes)", name, len(got), len(want))
		}
	}

	// The index must exist and link the seeded doc page and schema (the populated
	// link-section branches), and the sample gallery must be materialized.
	index := string(files["site/README.md"])
	for _, want := range []string{
		"# atago documentation",
		"../doc/e2e/echo.md",
		"schema/atago.schema.json",
		"samples/report.json",
	} {
		if !strings.Contains(index, want) {
			t.Errorf("site index missing %q\n%s", want, index)
		}
	}
	// A generated binary sample (PNG) must be non-empty on disk.
	if fi, err := os.Stat(filepath.Join(root, "site/samples/imagediff/diff.png")); err != nil || fi.Size() == 0 {
		t.Errorf("diff.png not generated: err=%v", err)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
