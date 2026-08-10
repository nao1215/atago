package loader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTree materializes a map of relative path -> content under a fresh temp
// dir and returns it.
func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

const minimalSpec = `version: "1"
suite:
  name: s
scenarios:
  - name: one
    steps:
      - run:
          command: echo hi
`

// TestFindProject_NearestAncestor pins the discovery rule (#392): a spec resolves
// the nearest manifest at or above it, so running a whole tree and re-running one
// failing spec inside it cannot produce different configuration — which is the
// way a developer would otherwise get a green local run and a red CI one.
func TestFindProject_NearestAncestor(t *testing.T) {
	t.Parallel()
	root := writeTree(t, map[string]string{
		"atago.project.yaml":              "env:\n  WHICH: outer\n",
		"deep/atago.project.yaml":         "env:\n  WHICH: inner\n",
		"deep/nested/spec.atago.yaml":     minimalSpec,
		"shallow/nested/spec.atago.yaml":  minimalSpec,
		"unrelated/other/spec.atago.yaml": minimalSpec,
	})

	deep, err := Load(filepath.Join(root, "deep/nested/spec.atago.yaml"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if deep.Suite.Env["WHICH"] != "inner" {
		t.Errorf("nested spec used %q, want the nearest manifest (inner)", deep.Suite.Env["WHICH"])
	}
	shallow, err := Load(filepath.Join(root, "shallow/nested/spec.atago.yaml"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if shallow.Suite.Env["WHICH"] != "outer" {
		t.Errorf("spec with no closer manifest used %q, want outer", shallow.Suite.Env["WHICH"])
	}
	if shallow.ProjectPath == "" {
		t.Error("the applied manifest must be recorded, or hidden configuration is unattributable")
	}
}

// TestLoad_NoManifestIsUnchanged keeps the common case honest: a spec with no
// manifest above it behaves exactly as before, carrying neither a project path
// nor a fixtures dir.
func TestLoad_NoManifestIsUnchanged(t *testing.T) {
	t.Parallel()
	root := writeTree(t, map[string]string{"spec.atago.yaml": minimalSpec})
	s, err := Load(filepath.Join(root, "spec.atago.yaml"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if s.ProjectPath != "" || s.FixturesDir != "" {
		t.Errorf("no manifest must mean no project state (path=%q fixtures=%q)", s.ProjectPath, s.FixturesDir)
	}
}

// TestApplyProject_SpecOwnValuesWin pins the precedence in both directions: a
// key only the manifest sets arrives, and a key the spec also sets keeps the
// spec's value.
func TestApplyProject_SpecOwnValuesWin(t *testing.T) {
	t.Parallel()
	root := writeTree(t, map[string]string{
		"atago.project.yaml": "env:\n  SHARED: manifest\n  ONLY_HERE: yes\n",
		"spec.atago.yaml": `version: "1"
suite:
  name: s
  env:
    SHARED: spec
scenarios:
  - name: one
    steps:
      - run:
          command: echo hi
`,
	})
	s, err := Load(filepath.Join(root, "spec.atago.yaml"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if s.Suite.Env["SHARED"] != "spec" {
		t.Errorf("suite.env SHARED = %q, want the spec's own value", s.Suite.Env["SHARED"])
	}
	if s.Suite.Env["ONLY_HERE"] != "yes" {
		t.Errorf("a manifest-only key must arrive, got %q", s.Suite.Env["ONLY_HERE"])
	}
}

// TestApplyProject_DefaultsReachStepsAndYieldToTheFile covers the reason the
// manifest exists at all: `sandbox_home` written once for a tree, instead of
// repeated in every file where the next file silently forgets it.
func TestApplyProject_DefaultsReachStepsAndYieldToTheFile(t *testing.T) {
	t.Parallel()
	root := writeTree(t, map[string]string{
		"atago.project.yaml":  "defaults:\n  run:\n    sandbox_home: true\n    shell: true\n",
		"inherits.atago.yaml": minimalSpec,
		"overrides.atago.yaml": `version: "1"
suite:
  name: s
defaults:
  run:
    shell: false
scenarios:
  - name: one
    steps:
      - run:
          command: echo hi
`,
	})
	inherits, err := Load(filepath.Join(root, "inherits.atago.yaml"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	run := inherits.Scenarios[0].Steps[0].Run
	if run.SandboxHome == nil || !*run.SandboxHome {
		t.Error("a manifest default must reach a spec that sets nothing")
	}
	overrides, err := Load(filepath.Join(root, "overrides.atago.yaml"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	orun := overrides.Scenarios[0].Steps[0].Run
	if orun.Shell == nil || *orun.Shell {
		t.Error("the file's own defaults must beat the manifest's")
	}
	if orun.SandboxHome == nil || !*orun.SandboxHome {
		t.Error("a key the file did not set must still come from the manifest")
	}
}

// TestLoadProject_Rejections covers the shapes that must fail at load time
// rather than mid-run.
func TestLoadProject_Rejections(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		manifest string
		want     string
	}{
		"unknown key":            {"nope: 1\n", "nope"},
		"missing fixtures dir":   {"fixtures_dir: not-there\n", "does not exist"},
		"fixtures dir is a file": {"fixtures_dir: afile.txt\n", "is not a directory"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := writeTree(t, map[string]string{
				"atago.project.yaml": tt.manifest,
				"afile.txt":          "x",
				"spec.atago.yaml":    minimalSpec,
			})
			_, err := Load(filepath.Join(root, "spec.atago.yaml"))
			if err == nil {
				t.Fatal("expected a load error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error %v should mention %q", err, tt.want)
			}
		})
	}
}

// TestLoadProject_FixturesDirResolvesAgainstTheManifest pins that the path is
// relative to the MANIFEST, not to the spec that uses it or to the working
// directory — the spec may sit several levels below, and cwd is whatever the
// developer happened to be in.
func TestLoadProject_FixturesDirResolvesAgainstTheManifest(t *testing.T) {
	t.Parallel()
	root := writeTree(t, map[string]string{
		"atago.project.yaml":          "fixtures_dir: corpus\n",
		"corpus/data.txt":             "x",
		"deep/nested/spec.atago.yaml": minimalSpec,
	})
	s, err := Load(filepath.Join(root, "deep/nested/spec.atago.yaml"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	want := filepath.Join(root, "corpus")
	if s.FixturesDir != want {
		t.Errorf("FixturesDir = %q, want %q", s.FixturesDir, want)
	}
}
