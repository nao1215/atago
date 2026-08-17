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

// TestEngine_RepeatedIterationsKeepTheirOwnArtifacts is the regression for a
// silent lie in a --repeat report: every iteration wrote its payloads to the
// same paths, so the last failing iteration overwrote the earlier ones while the
// folded result kept the FIRST failing iteration's inline diff. A reviewer
// opening the referenced sidecar saw a payload from a different iteration than
// the one the report described. Each iteration's ${workdir} is unique, so the
// actual stdout differs per iteration with no timing dependence.
func TestEngine_RepeatedIterationsKeepTheirOwnArtifacts(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	const src = `
version: "1"
suite:
  name: s
scenarios:
  - name: prints its own workdir
    steps:
      - run:
          shell: true
          command: echo ${workdir}
      - assert:
          stdout: {contains: NOPE}
`
	s, err := loader.LoadBytes("rep.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	eng := New()
	eng.Artifacts = artifact.NewDir(root)
	eng.Repeat = 3
	res := eng.Run(context.Background(), s, "rep.atago.yaml")
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed (every iteration fails)", res.Status)
	}

	cr := failedCheck(t, res)
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
	// The sidecar the report points at must carry the payload the report
	// described, not a later iteration's.
	if strings.TrimSpace(string(got)) != strings.TrimSpace(cr.Actual) {
		t.Errorf("artifact %s = %q, but the report's actual is %q", actualPath, got, cr.Actual)
	}

	// And no iteration's evidence is lost: three failing iterations leave three
	// distinct payload files behind.
	var actuals []string
	err = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), "stdout.actual.txt") {
			actuals = append(actuals, p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk artifacts: %v", err)
	}
	if len(actuals) != 3 {
		t.Errorf("found %d actual payloads, want one per iteration: %v", len(actuals), actuals)
	}
	seen := map[string]bool{}
	for _, p := range actuals {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		if seen[string(b)] {
			t.Errorf("two iterations wrote the same payload %q", b)
		}
		seen[string(b)] = true
	}
}

// TestEngine_RetriedAttemptArtifactsMatchTheReport pins the same contract for
// --retry-failed: the reported attempt is the last one, and the payload it
// references has to be that attempt's.
func TestEngine_RetriedAttemptArtifactsMatchTheReport(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	const src = `
version: "1"
suite:
  name: s
scenarios:
  - name: prints its own workdir
    steps:
      - run:
          shell: true
          command: echo ${workdir}
      - assert:
          stdout: {contains: NOPE}
`
	s, err := loader.LoadBytes("retry.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	eng := New()
	eng.Artifacts = artifact.NewDir(root)
	eng.RetryFailed = 2
	res := eng.Run(context.Background(), s, "retry.atago.yaml")
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed (retries cannot help)", res.Status)
	}
	cr := failedCheck(t, res)
	for _, a := range cr.ArtifactFiles {
		if a.Role != "actual" {
			continue
		}
		got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(a.Path)))
		if err != nil {
			t.Fatalf("read artifact: %v", err)
		}
		if strings.TrimSpace(string(got)) != strings.TrimSpace(cr.Actual) {
			t.Errorf("artifact %s = %q, but the reported attempt's actual is %q", a.Path, got, cr.Actual)
		}
	}
}

// TestEngine_TeardownArtifactsDoNotOverwriteTheStepsEvidence pins the phase
// separation end to end: a scenario whose steps and whose teardown both fail at
// the same index keeps both payloads. The teardown block numbers its steps from
// zero too, so before the phase segment existed the teardown's payload landed on
// the scenario step's filename — the verdict and the diff printed beside it were
// about the step, while the file the report offered held the teardown's bytes.
func TestEngine_TeardownArtifactsDoNotOverwriteTheStepsEvidence(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	res := runSpecWithArtifacts(t, "art.atago.yaml", `
version: "1"
suite:
  name: s
scenarios:
  - name: both phases fail at index 1
    steps:
      - run: {shell: true, command: echo STEPS-ACTUAL}
      - assert:
          stdout: {equals: "STEPS-EXPECTED\n"}
    teardown:
      - run: {shell: true, command: echo TEARDOWN-ACTUAL}
      - assert:
          stdout: {equals: "TEARDOWN-EXPECTED\n"}
`, root)
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed", res.Status)
	}

	// Every payload the run wrote, by content, so the assertion below is about
	// what survived rather than about a path spelling.
	payloads := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		b, rerr := os.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		rel, rerr := filepath.Rel(root, p)
		if rerr != nil {
			return rerr
		}
		payloads[filepath.ToSlash(rel)] = strings.TrimSpace(string(b))
		return nil
	})
	if err != nil {
		t.Fatalf("walk artifacts: %v", err)
	}

	want := map[string]bool{"STEPS-ACTUAL": false, "STEPS-EXPECTED": false, "TEARDOWN-ACTUAL": false, "TEARDOWN-EXPECTED": false}
	for _, content := range payloads {
		if _, ok := want[content]; ok {
			want[content] = true
		}
	}
	for content, found := range want {
		if !found {
			t.Errorf("payload %q was not preserved; artifacts on disk: %v", content, payloads)
		}
	}

	// And the report points at the step's own bytes, not the teardown's.
	cr := failedCheck(t, res)
	for _, a := range cr.ArtifactFiles {
		got := payloads[a.Path]
		if strings.HasPrefix(got, "TEARDOWN-") {
			t.Errorf("the failed step's %s sidecar %s holds the teardown's payload %q", a.Role, a.Path, got)
		}
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
