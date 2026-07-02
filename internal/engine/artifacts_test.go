package engine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/artifact"
	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/loader"
)

// runSpecWithArtifacts runs src as the given spec path with an artifacts dir set,
// returning the suite result and the artifacts root.
func runSpecWithArtifacts(t *testing.T, specPath, src, root string) *SuiteResult {
	t.Helper()
	s, err := loader.LoadBytes(specPath, []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	eng := New()
	eng.Artifacts = artifact.NewDir(root)
	return eng.Run(context.Background(), s, specPath)
}

func TestEngine_ArtifactsWrittenForFailedStdout(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	res := runSpecWithArtifacts(t, "t.atago.yaml", `
version: "1"
suite:
  name: s
scenarios:
  - name: prints hello
    steps:
      - run: {shell: true, command: echo hello world}
      - assert:
          stdout: {contains: goodbye}
`, root)
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed", res.Status)
	}

	// The failed check must reference an "actual" sidecar carrying the full stdout.
	cr := failedCheck(t, res)
	if len(cr.ArtifactFiles) == 0 {
		t.Fatalf("no artifact files recorded on failed check")
	}
	var actualPath string
	for _, a := range cr.ArtifactFiles {
		if a.Role == "actual" {
			actualPath = a.Path
		}
	}
	if actualPath == "" {
		t.Fatalf("no actual artifact recorded: %+v", cr.ArtifactFiles)
	}
	got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(actualPath)))
	if err != nil {
		t.Fatalf("read artifact: %v", err)
	}
	if !strings.Contains(string(got), "hello world") {
		t.Errorf("artifact actual = %q, want the full stdout", got)
	}
}

func TestEngine_ArtifactsMaskSecrets(t *testing.T) {
	const secret = "s3cr3t-value"
	t.Setenv("ATAGO_ARTIFACT_SECRET", secret)
	root := t.TempDir()
	res := runSpecWithArtifacts(t, "sec.atago.yaml", `
version: "1"
suite:
  name: s
secrets:
  - ATAGO_ARTIFACT_SECRET
scenarios:
  - name: leaks
    steps:
      - run:
          shell: true
          command: echo token=`+envRef("ATAGO_ARTIFACT_SECRET")+`
      - assert:
          stdout: {contains: NOPE}
`, root)
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed", res.Status)
	}
	cr := failedCheck(t, res)
	for _, a := range cr.ArtifactFiles {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(a.Path)))
		if err != nil {
			t.Fatalf("read artifact %s: %v", a.Path, err)
		}
		if strings.Contains(string(data), "s3cr3t-value") {
			t.Errorf("artifact %s leaked the secret: %q", a.Path, data)
		}
		if !strings.Contains(string(data), "***") {
			t.Errorf("artifact %s not masked: %q", a.Path, data)
		}
	}
}

// TestEngine_ArtifactFilenamesDoNotCollide is the multi-suite regression from
// #48: two suites whose spec files share a base name, each with a scenario of
// the same name and a failing assertion at the same step index, must write to
// distinct artifact paths.
func TestEngine_ArtifactFilenamesDoNotCollide(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	src := `
version: "1"
suite:
  name: s
scenarios:
  - name: dup
    steps:
      - run: {shell: true, command: echo one}
      - assert:
          stdout: {contains: MISSING}
`
	a := runSpecWithArtifacts(t, "dir-a/same.atago.yaml", src, root)
	b := runSpecWithArtifacts(t, "dir-b/same.atago.yaml", src, root)
	pa := failedCheck(t, a).ArtifactFiles
	pb := failedCheck(t, b).ArtifactFiles
	if len(pa) == 0 || len(pb) == 0 {
		t.Fatalf("expected artifacts for both suites")
	}
	if pa[0].Path == pb[0].Path {
		t.Fatalf("distinct suites collided on artifact path %q", pa[0].Path)
	}
}

func TestEngine_NoArtifactsWithoutDir(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: prints
    steps:
      - run: {shell: true, command: echo hi}
      - assert:
          stdout: {contains: NOPE}
`)
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed", res.Status)
	}
	if fc := failedCheck(t, res); len(fc.ArtifactFiles) != 0 {
		t.Errorf("artifacts recorded without --artifacts-dir: %+v", fc.ArtifactFiles)
	}
}

// failedCheck returns the first failed CheckResult across a suite result.
func failedCheck(t *testing.T, res *SuiteResult) *assert.CheckResult {
	t.Helper()
	for i := range res.Scenarios {
		for _, st := range res.Scenarios[i].Steps {
			for _, ck := range st.Checks {
				if ck != nil && !ck.OK {
					return ck
				}
			}
		}
	}
	t.Fatalf("no failed check found in result")
	return nil
}
