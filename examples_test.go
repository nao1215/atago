package main

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/engine"
	"github.com/nao1215/atago/internal/loader"
)

// exampleSpecs categorizes every spec under examples/: hermetic examples run
// green with no external dependency and are executed here on every OS;
// non-hermetic ones (a live API, an SSH host, a gRPC server, a browser) are
// loaded and validated only. The README links to these files as the syntax
// reference, so this test is what keeps them from drifting away from the
// implementation.
var exampleSpecs = map[string]bool{ // path -> hermetic (run, not just validate)
	"examples/browser.atago.yaml":             false,
	"examples/db.atago.yaml":                  true,
	"examples/defaults.atago.yaml":            true,
	"examples/files_and_fixtures.atago.yaml":  true,
	"examples/grpc.atago.yaml":                false,
	"examples/hermetic_env.atago.yaml":        true,
	"examples/http.atago.yaml":                false,
	"examples/image_and_pdf.atago.yaml":       true,
	"examples/json_and_yaml.atago.yaml":       true,
	"examples/matrix.atago.yaml":              true,
	"examples/mock_server.atago.yaml":         true,
	"examples/pty.atago.yaml":                 true,
	"examples/retry.atago.yaml":               true,
	"examples/run_and_assert.atago.yaml":      true,
	"examples/select_skip_only.atago.yaml":    true,
	"examples/services.atago.yaml":            true,
	"examples/shell_and_redirect.atago.yaml":  true,
	"examples/signal.atago.yaml":              true,
	"examples/snapshot.atago.yaml":            true,
	"examples/ssh.atago.yaml":                 false,
	"examples/stdin.atago.yaml":               true,
	"examples/store_and_variables.atago.yaml": true,
	"examples/suite_setup.atago.yaml":         false,
	"examples/teardown.atago.yaml":            true,
	"examples/timeouts.atago.yaml":            true,
}

// TestExamples_EveryFileCategorized fails when a spec is added to examples/
// without being registered above, so a new example cannot ship untested.
func TestExamples_EveryFileCategorized(t *testing.T) {
	t.Parallel()
	found := map[string]bool{}
	err := filepath.WalkDir("examples", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".atago.yaml") {
			found[filepath.ToSlash(path)] = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk examples: %v", err)
	}
	for path := range found {
		if _, ok := exampleSpecs[path]; !ok {
			t.Errorf("%s is not categorized in exampleSpecs; add it (hermetic or validate-only)", path)
		}
	}
	for path := range exampleSpecs {
		if !found[path] {
			t.Errorf("%s is categorized but does not exist", path)
		}
	}
}

// TestExamples_Valid loads and validates every example, hermetic or not.
func TestExamples_Valid(t *testing.T) {
	t.Parallel()
	for path := range exampleSpecs {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			if _, err := loader.Load(path); err != nil {
				t.Errorf("example does not load/validate: %v", err)
			}
		})
	}
}

// TestExamples_HermeticRunGreen executes every hermetic example through the
// real engine. OS-gated scenarios (skip/only) may be skipped, but nothing may
// fail or error: an example the README points at must actually work.
func TestExamples_HermeticRunGreen(t *testing.T) {
	t.Parallel()
	for path, hermetic := range exampleSpecs {
		if !hermetic {
			continue
		}
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			s, err := loader.Load(path)
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			res := engine.New().Run(context.Background(), s, path)
			if res.Status != engine.StatusPassed && res.Status != engine.StatusSkipped {
				t.Errorf("status = %s, want passed (or skipped by an OS gate): %+v", res.Status, res.Scenarios)
			}
		})
	}
}
