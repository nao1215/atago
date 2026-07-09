package assert

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// TestCheck_ArtifactPayloads verifies that failed text assertions expose the full
// compared payload for durable export via --artifacts-dir (#48), and that
// passing assertions expose nothing.
func TestCheck_ArtifactPayloads(t *testing.T) {
	t.Parallel()
	res := &runner.Result{Stdout: []byte("hello world\n")}
	env := Env{Workdir: t.TempDir()}

	t.Run("failed contains carries full actual, no expected", func(t *testing.T) {
		t.Parallel()
		cr := Check(&spec.Assert{Stdout: &spec.StreamAssert{Contains: spec.StringList{"goodbye"}}}, res, env)
		if cr.OK {
			t.Fatal("expected failure")
		}
		if cr.ArtifactKind != "stdout" {
			t.Errorf("ArtifactKind = %q, want stdout", cr.ArtifactKind)
		}
		if string(cr.ArtifactActual) != "hello world\n" {
			t.Errorf("ArtifactActual = %q, want full stdout", cr.ArtifactActual)
		}
		if cr.ArtifactExpected != nil {
			t.Errorf("ArtifactExpected = %q, want nil for contains", cr.ArtifactExpected)
		}
	})

	t.Run("failed equals carries both actual and expected", func(t *testing.T) {
		t.Parallel()
		cr := Check(&spec.Assert{Stdout: &spec.StreamAssert{Equals: strp("nope\n")}}, res, env)
		if cr.OK {
			t.Fatal("expected failure")
		}
		if string(cr.ArtifactActual) != "hello world\n" {
			t.Errorf("ArtifactActual = %q", cr.ArtifactActual)
		}
		if string(cr.ArtifactExpected) != "nope\n" {
			t.Errorf("ArtifactExpected = %q, want the expected text", cr.ArtifactExpected)
		}
	})

	t.Run("passing assertion exposes no artifact", func(t *testing.T) {
		t.Parallel()
		cr := Check(&spec.Assert{Stdout: &spec.StreamAssert{Contains: spec.StringList{"hello"}}}, res, env)
		if !cr.OK {
			t.Fatal("expected pass")
		}
		if cr.ArtifactKind != "" || cr.ArtifactActual != nil {
			t.Errorf("passing check leaked artifact: kind=%q actual=%q", cr.ArtifactKind, cr.ArtifactActual)
		}
	})
}

// TestCheck_FileArtifactPayloads guards the #247 convention: every file matcher
// that reads the file's content attaches the "file" artifact with the file's
// bytes on failure (via checkFile's shared deferred hook), so a new content
// matcher inherits the --artifacts-dir sidecar instead of forgetting it. Matchers
// that never read content (exists/executable) attach nothing.
func TestCheck_FileArtifactPayloads(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	const body = "the actual file body\n"
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "other.txt"), []byte("different\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	env := Env{Workdir: dir}

	// Each of these matchers fails and must carry the file bytes as the artifact.
	content := map[string]*spec.FileAssert{
		"contains":     {Path: "f.txt", Contains: spec.StringList{"absent-substring"}},
		"not_contains": {Path: "f.txt", NotContains: spec.StringList{"actual"}},
		"equals":       {Path: "f.txt", Equals: strp("nope")},
		"equals_file":  {Path: "f.txt", EqualsFile: strp("other.txt")},
		"json":         {Path: "f.txt", JSON: spec.JSONChecks{{Path: "$.x", Equals: "y"}}},
	}
	for name, fa := range content {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cr := Check(&spec.Assert{File: fa}, &runner.Result{}, env)
			if cr.OK {
				t.Fatalf("%s: expected failure", name)
			}
			if cr.ArtifactKind != "file" {
				t.Errorf("%s: ArtifactKind = %q, want file", name, cr.ArtifactKind)
			}
			if string(cr.ArtifactActual) != body {
				t.Errorf("%s: ArtifactActual = %q, want the file body", name, cr.ArtifactActual)
			}
		})
	}

	// A failing exists/executable check reads no content, so it attaches nothing.
	t.Run("exists attaches nothing", func(t *testing.T) {
		t.Parallel()
		cr := Check(&spec.Assert{File: &spec.FileAssert{Path: "missing.txt", Exists: boolp(true)}}, &runner.Result{}, env)
		if cr.OK {
			t.Fatal("expected failure")
		}
		if cr.ArtifactKind != "" || cr.ArtifactActual != nil {
			t.Errorf("exists check leaked artifact: kind=%q actual=%q", cr.ArtifactKind, cr.ArtifactActual)
		}
	})
}
