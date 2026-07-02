package assert

import (
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
