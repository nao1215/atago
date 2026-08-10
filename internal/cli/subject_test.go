package cli

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// subjectTree writes a manifest declaring a subject whose "build" is a portable
// shell-free command, plus one spec beneath it, and returns the spec's path.
// The build is `go build` nowhere in sight on purpose: nothing about the
// subject feature is Go-specific, and a test that used the Go toolchain would
// quietly encode the opposite.
func subjectTree(t *testing.T, manifest string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "atago.project.yaml"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	spec := `version: "1"
suite:
  name: s
scenarios:
  - name: one
    steps:
      - run:
          command: echo hi
`
	path := filepath.Join(root, "spec.atago.yaml")
	if err := os.WriteFile(path, []byte(spec), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// copyCommand returns a shell-free command that produces a file at ${artifact},
// standing in for a compiler on every OS.
func copyCommand(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the stand-in build command is POSIX; the subject mechanism itself is platform-neutral")
	}
	// chmod +x because the produced file has to be runnable: a subject that
	// cannot execute is caught at build time now, which is the point.
	return "cp go.mod ${artifact} && chmod +x ${artifact}"
}

// TestBuildSubjects_ProducesTheArtifactOnce covers the core promise: the build
// runs once per invocation even when many specs share a manifest, and the
// artifact lands where ${artifact} pointed.
func TestBuildSubjects_ProducesTheArtifactOnce(t *testing.T) {
	cmd := copyCommand(t)
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := "subject:\n  name: mytool\n  artifact: bin/mytool\n  build:\n    shell: true\n    command: \"" + cmd + "\"\n"
	if err := os.WriteFile(filepath.Join(root, "atago.project.yaml"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	specs := []string{}
	for _, name := range []string{"a.atago.yaml", "b.atago.yaml"} {
		p := filepath.Join(root, name)
		if err := os.WriteFile(p, []byte("version: \"1\"\nsuite:\n  name: s\nscenarios:\n  - name: one\n    steps:\n      - run:\n          command: echo hi\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		specs = append(specs, p)
	}

	scratch := t.TempDir()
	built, err := buildSubjects(context.Background(), specs, "", scratch, io.Discard)
	if err != nil {
		t.Fatalf("buildSubjects: %v", err)
	}
	if len(built) != 1 {
		t.Fatalf("built %d subjects for two specs sharing one manifest, want 1", len(built))
	}
	if _, serr := os.Stat(built[0].Artifact); serr != nil {
		t.Errorf("artifact missing at %s: %v", built[0].Artifact, serr)
	}
}

// TestBuildSubjects_NoManifestIsANoOp keeps every existing suite unaffected:
// without a subject there is nothing to build and nothing to report.
func TestBuildSubjects_NoManifestIsANoOp(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	path := filepath.Join(root, "spec.atago.yaml")
	if err := os.WriteFile(path, []byte("version: \"1\"\nsuite:\n  name: s\nscenarios:\n  - name: one\n    steps:\n      - run:\n          command: echo hi\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	built, err := buildSubjects(context.Background(), []string{path}, "", t.TempDir(), io.Discard)
	if err != nil || len(built) != 0 {
		t.Errorf("buildSubjects = %v, %v; want no subjects and no error", built, err)
	}
}

// TestBuildSubjects_UnknownProfileNamesWhatExists turns a typo into a fixable
// message instead of a silent default build, which would be worse: the run
// would go green having tested a binary built the wrong way.
func TestBuildSubjects_UnknownProfileNamesWhatExists(t *testing.T) {
	cmd := copyCommand(t)
	path := subjectTree(t, "subject:\n  name: mytool\n  artifact: bin/mytool\n  build:\n    shell: true\n    command: \""+cmd+"\"\nprofiles:\n  cover:\n    env:\n      X: \"1\"\n")
	if err := os.WriteFile(filepath.Join(filepath.Dir(path), "go.mod"), []byte("module x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := buildSubjects(context.Background(), []string{path}, "race", t.TempDir(), io.Discard)
	if err == nil {
		t.Fatal("an unknown profile must fail the run")
	}
	if !strings.Contains(err.Error(), "cover") {
		t.Errorf("error %v should list the profiles that do exist", err)
	}
}

// TestBuildSubjects_ProfileWithNoSubjectAnywhere refuses a --profile that
// cannot mean anything, rather than running as if it had been ignored.
func TestBuildSubjects_ProfileWithNoSubjectAnywhere(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	path := filepath.Join(root, "spec.atago.yaml")
	if err := os.WriteFile(path, []byte("version: \"1\"\nsuite:\n  name: s\nscenarios:\n  - name: one\n    steps:\n      - run:\n          command: echo hi\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := buildSubjects(context.Background(), []string{path}, "cover", t.TempDir(), io.Discard); err == nil {
		t.Error("--profile with no subject to build must fail rather than be ignored")
	}
}

// TestBuildSubjects_FailedBuildCarriesTheCompilerOutput pins the failure
// report: the build tool's own message is the only useful thing to say, and
// hiding it would send the author to re-run the build by hand.
func TestBuildSubjects_FailedBuildCarriesTheCompilerOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the stand-in build command is POSIX")
	}
	path := subjectTree(t, "subject:\n  name: mytool\n  artifact: bin/mytool\n  build:\n    shell: true\n    command: \"echo BUILD-BROKE >&2; exit 3\"\n")
	_, err := buildSubjects(context.Background(), []string{path}, "", t.TempDir(), io.Discard)
	if err == nil {
		t.Fatal("a failing build must fail the run")
	}
	if !strings.Contains(err.Error(), "BUILD-BROKE") || !strings.Contains(err.Error(), "exit 3") {
		t.Errorf("error %v should carry the build tool's own output and exit code", err)
	}
}

// TestBuildSubjects_SucceedsButWritesNothing catches the build command that
// forgets ${artifact}: it exits 0, so without this check the run would proceed
// and every scenario would fail on a missing binary instead.
func TestBuildSubjects_SucceedsButWritesNothing(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the stand-in build command is POSIX")
	}
	path := subjectTree(t, "subject:\n  name: mytool\n  artifact: bin/mytool\n  build:\n    command: \"true\"\n")
	_, err := buildSubjects(context.Background(), []string{path}, "", t.TempDir(), io.Discard)
	if err == nil || !strings.Contains(err.Error(), "produced no file") {
		t.Errorf("error = %v, want it to name the missing artifact", err)
	}
}

// TestExposeSubjects_PutsTheArtifactFirstOnPath is what lets a spec keep saying
// `mytool convert ...` — the binary is tested the way a user invokes it.
func TestExposeSubjects_PutsTheArtifactFirstOnPath(t *testing.T) {
	dir := t.TempDir()
	artifact := filepath.Join(dir, "mytool")
	t.Setenv("PATH", os.Getenv("PATH"))
	if err := exposeSubjects([]builtSubject{{Name: "mytool", Artifact: artifact, Env: map[string]string{"PROFILE_MARK": "on"}}}); err != nil {
		t.Fatalf("exposeSubjects: %v", err)
	}
	if first := strings.SplitN(os.Getenv("PATH"), string(os.PathListSeparator), 2)[0]; first != dir {
		t.Errorf("PATH starts with %q, want the artifact directory %q", first, dir)
	}
	if os.Getenv("PROFILE_MARK") != "on" {
		t.Error("a profile's env must reach the specs")
	}
}

// TestBuildSubjects_NonExecutableArtifactIsRefused pins the check a `cp` build
// makes necessary: the file exists and the build exited 0, but nothing could
// run it, and without this every scenario fails with "permission denied"
// instead of one message naming the build.
func TestBuildSubjects_NonExecutableArtifactIsRefused(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows has no execute bit; the artifact check there is regular-file only")
	}
	path := subjectTree(t, "subject:\n  name: mytool\n  artifact: bin/mytool\n  build:\n    shell: true\n    command: \"echo x > ${artifact}\"\n")
	_, err := buildSubjects(context.Background(), []string{path}, "", t.TempDir(), io.Discard)
	if err == nil || !strings.Contains(err.Error(), "execute bit") {
		t.Errorf("error = %v, want it to name the missing execute bit", err)
	}
}

// TestBuildSubjects_DuplicateSubjectNameIsRefused covers the ambiguity two
// manifests in one run can create: every artifact directory shares one PATH, so
// the same name would resolve to whichever was prepended last — for both trees.
// Refusing beats silently testing one binary twice.
func TestBuildSubjects_DuplicateSubjectNameIsRefused(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the stand-in build command is POSIX")
	}
	manifest := "subject:\n  name: mytool\n  artifact: bin/mytool\n  build:\n    shell: true\n    command: \"echo x > ${artifact} && chmod +x ${artifact}\"\n"
	var specs []string
	for range 2 {
		specs = append(specs, subjectTree(t, manifest))
	}
	_, err := buildSubjects(context.Background(), specs, "", t.TempDir(), io.Discard)
	if err == nil || !strings.Contains(err.Error(), "declare a subject named") {
		t.Errorf("error = %v, want the duplicate-name refusal", err)
	}
}
