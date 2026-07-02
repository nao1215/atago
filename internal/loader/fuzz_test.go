package loader

import (
	"strings"
	"testing"
)

// FuzzLoadBytes feeds arbitrary bytes to the loader (issue #46). The invariants:
//   - loading never panics, however malformed the YAML or schema, and
//   - a successful load returns a non-nil spec (so callers never get a nil spec
//     with a nil error), and that spec re-validates cleanly — the one-of step
//     rules and semantic checks are internally consistent.
func FuzzLoadBytes(f *testing.F) {
	seeds := []string{
		"",
		"not: valid: yaml: [",
		"version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}",
		"version: \"1\"\nsuite: {name: x}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n        assert: {exit_code: 0}",
		"version: \"1\"\nsuite: {name: x}\nrunners:\n  d: {type: db, dsn: \"sqlite::memory:\"}\nscenarios:\n  - name: a\n    steps: [{query: {runner: d, sql: SELECT 1}}, {assert: {rows: {empty: false}}}]",
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		s, err := LoadBytes("fuzz.atago.yaml", data)
		if err != nil {
			return
		}
		if s == nil {
			t.Fatal("LoadBytes returned nil spec with nil error")
		}
		// A spec that loaded cleanly must have no residual validation errors.
		if errs := validate(s); len(errs) != 0 {
			t.Fatalf("loaded spec still has validation errors: %s", strings.Join(errs, "; "))
		}
	})
}
