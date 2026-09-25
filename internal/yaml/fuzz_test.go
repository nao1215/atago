package yaml

import (
	"os"
	"path/filepath"
	"testing"
)

// FuzzParse holds the reader to never panicking: a spec may be malformed, and
// the YAML an assertion inspects is whatever a program printed. Whatever parses
// has to decode into a generic value too, and an error has to render with its
// excerpt.
func FuzzParse(f *testing.F) {
	seeds, _ := filepath.Glob("../../test/e2e/atago/*.atago.yaml")
	for _, s := range seeds {
		if data, err := os.ReadFile(s); err == nil {
			f.Add(data)
		}
	}
	for _, s := range []string{
		"a: &x [1, *x]", "- &a\n  b: 1\n- *a", "k: |+\n  x\n\n", "? a\n: b", "{a: [b, {c: d}]}",
		"a: !!binary aGk=", "<<: {a: 1}\nb: 2", "\"a\\\n  b\"", "%YAML 1.2\n---\na", "--- |\n x\n...\n---\ny",
	} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		file, err := Parse(data)
		if err != nil {
			_ = FormatError(err, data)
			return
		}
		for _, d := range file.Docs {
			if _, err := Value(d.Root); err != nil {
				_ = FormatError(err, file.Source())
			}
		}
	})
}
