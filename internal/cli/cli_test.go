package cli

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSpec(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

// The specs below run through the shell (`shell: true`), which maps to /bin/sh
// on POSIX and cmd.exe on Windows; echo and exit are builtins of both, so the
// same YAML runs everywhere.
const passingSpec = `
version: "1"
suite:
  name: sample
scenarios:
  - name: exit 0 passes
    steps:
      - run: {shell: true, command: "exit 0"}
      - assert: {exit_code: 0}
`

const failingSpec = `
version: "1"
suite:
  name: sample
scenarios:
  - name: exit 1 should pass
    steps:
      - run: {shell: true, command: "exit 1"}
      - assert: {exit_code: 0}
`

func TestRunCmd_Passing(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "ok.atago.yaml", passingSpec)
	var out, errb bytes.Buffer
	if got := Main([]string{"run", p}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitOK, errb.String())
	}
	if !strings.Contains(out.String(), "PASS") {
		t.Errorf("stdout = %q, want PASS", out.String())
	}
}

// TestRunCmd_SelectionMatchesNothingWarns proves a --filter/--tag that selects
// zero scenarios warns on stderr (a typo'd selection in CI must not greenlight
// in silence), while still exiting 0.
func TestRunCmd_SelectionMatchesNothingWarns(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "ok.atago.yaml", passingSpec)
	var out, errb bytes.Buffer
	if got := Main([]string{"run", "--filter", "no-such-name", p}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitOK, errb.String())
	}
	if !strings.Contains(errb.String(), `no scenarios matched --filter "no-such-name"`) {
		t.Errorf("stderr = %q, want a no-match warning", errb.String())
	}
	// A matching selection stays quiet.
	errb.Reset()
	out.Reset()
	if got := Main([]string{"run", "--filter", "exit 0", p}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitOK, errb.String())
	}
	if strings.Contains(errb.String(), "no scenarios matched") {
		t.Errorf("stderr = %q, want no warning for a matching filter", errb.String())
	}
}

func TestRunCmd_Failing(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "fail.atago.yaml", failingSpec)
	var out, errb bytes.Buffer
	if got := Main([]string{"run", p}, &out, &errb); got != ExitFailures {
		t.Fatalf("exit = %d, want %d", got, ExitFailures)
	}
	if !strings.Contains(out.String(), "FAILED") {
		t.Errorf("stdout = %q, want FAILED block", out.String())
	}
}

func TestRunCmd_ArtifactsDir(t *testing.T) {
	dir := t.TempDir()
	spec := `
version: "1"
suite:
  name: sample
scenarios:
  - name: stdout drift
    steps:
      - run: {shell: true, command: echo hello}
      - assert:
          stdout: {contains: goodbye}
`
	p := writeSpec(t, dir, "fail.atago.yaml", spec)
	arts := filepath.Join(dir, "arts")
	var out, errb bytes.Buffer
	if got := Main([]string{"run", "--report", "json", "--artifacts-dir", arts, p}, &out, &errb); got != ExitFailures {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
	}
	// The JSON report references the sidecar, and the sidecar exists on disk.
	if !strings.Contains(out.String(), "\"artifacts\"") {
		t.Errorf("JSON report missing artifacts field:\n%s", out.String())
	}
	var found string
	err := filepath.WalkDir(arts, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(path, ".actual.txt") {
			found = path
		}
		return nil
	})
	if err != nil || found == "" {
		t.Fatalf("no actual artifact written under %s (err=%v)", arts, err)
	}
	data, err := os.ReadFile(found)
	if err != nil {
		t.Fatalf("read artifact: %v", err)
	}
	if !strings.Contains(string(data), "hello") {
		t.Errorf("artifact = %q, want full stdout", data)
	}
}

func TestRunCmd_ParseErrorExit2(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "broken.atago.yaml", "version: \"1\"\nsuite:\n  : invalid")
	var out, errb bytes.Buffer
	if got := Main([]string{"run", p}, &out, &errb); got != ExitParse {
		t.Fatalf("exit = %d, want %d", got, ExitParse)
	}
}

// Issue #21: a spec-content error — here a schema/semantic validation failure (a
// db runner missing its dsn) — exits 2 (spec error), not 3. Exit 3 is reserved
// for CLI-invocation problems.
func TestRunCmd_ValidationErrorExit2(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "cfg.atago.yaml",
		"version: \"1\"\nsuite:\n  name: x\nrunners:\n  d: {type: db}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}")
	var out, errb bytes.Buffer
	if got := Main([]string{"run", p}, &out, &errb); got != ExitParse {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitParse, errb.String())
	}
}

// Issue #21: exit 3 (ExitConfig) is returned for CLI-invocation problems, e.g. a
// target that matches no spec files.
func TestRunCmd_NoFilesExit3(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if got := Main([]string{"run", dir}, &out, &errb); got != ExitConfig {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitConfig, errb.String())
	}
}

func TestRunCmd_JSONReportValid(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "ok.atago.yaml", passingSpec)
	var out, errb bytes.Buffer
	if got := Main([]string{"run", "--report", "json", p}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d, want %d", got, ExitOK)
	}
	var parsed struct {
		SchemaVersion string `json:"schema_version"`
		Suites        []struct {
			Status string `json:"status"`
		} `json:"suites"`
	}
	if err := json.Unmarshal(out.Bytes(), &parsed); err != nil {
		t.Fatalf("report is not valid JSON: %v\n%s", err, out.String())
	}
	// The JSON report has one stable top-level shape regardless of suite count (#43).
	if parsed.SchemaVersion != "1" {
		t.Errorf("schema_version = %q, want 1", parsed.SchemaVersion)
	}
	if len(parsed.Suites) != 1 || parsed.Suites[0].Status != "passed" {
		t.Errorf("suites = %+v, want one passed suite", parsed.Suites)
	}
}

// A bare directory is searched recursively (no Go-style "dir/..." needed).
func TestRunCmd_DirectoryIsRecursive(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "nested", "deep")
	if err := os.MkdirAll(sub, 0o750); err != nil {
		t.Fatal(err)
	}
	writeSpec(t, dir, "top.atago.yaml", passingSpec)
	writeSpec(t, sub, "deep.atago.yaml", passingSpec)

	var out, errb bytes.Buffer
	if got := Main([]string{"run", dir}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitOK, errb.String())
	}
	// Both the top-level and the deeply-nested spec must have run: the aggregate
	// summary should report two passing scenarios.
	if !strings.Contains(out.String(), "2 passed") {
		t.Errorf("bare directory was not searched recursively:\n%s", out.String())
	}
}

// A trailing "..." is tolerated but no longer required.
func TestRunCmd_TrailingDotsTolerated(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "a.atago.yaml", passingSpec)
	var out, errb bytes.Buffer
	if got := Main([]string{"run", filepath.Join(dir, "...")}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitOK, errb.String())
	}
}

func TestRunCmd_JUnitValidXML(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "ok.atago.yaml", passingSpec)
	var out, errb bytes.Buffer
	if got := Main([]string{"run", "--report", "junit", p}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d, want %d", got, ExitOK)
	}
	var v struct {
		XMLName xml.Name `xml:"testsuites"`
		Tests   int      `xml:"tests,attr"`
	}
	if err := xml.Unmarshal(out.Bytes(), &v); err != nil {
		t.Fatalf("not valid XML: %v\n%s", err, out.String())
	}
	if v.Tests != 1 {
		t.Errorf("tests = %d, want 1", v.Tests)
	}
}

func TestRunCmd_GHA(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "fail.atago.yaml", failingSpec)
	var out, errb bytes.Buffer
	if got := Main([]string{"run", "--report", "gha", p}, &out, &errb); got != ExitFailures {
		t.Fatalf("exit = %d, want %d", got, ExitFailures)
	}
	if !strings.Contains(out.String(), "::error title=") {
		t.Errorf("missing GHA error annotation:\n%s", out.String())
	}
}

const snapshotSpec = `
version: "1"
suite:
  name: snap
scenarios:
  - name: stdout snapshot
    steps:
      - run: {shell: true, command: echo stable output}
      - assert:
          stdout:
            snapshot: snapshots/out.txt
`

func TestSnapshot_UpdateThenRun(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "snap.atago.yaml", snapshotSpec)

	// First run without a snapshot must fail with a helpful hint.
	var out, errb bytes.Buffer
	if got := Main([]string{"run", p}, &out, &errb); got != ExitFailures {
		t.Fatalf("pre-update exit = %d, want %d", got, ExitFailures)
	}
	if !strings.Contains(out.String(), "--update-snapshots") {
		t.Errorf("missing update hint:\n%s", out.String())
	}

	// snapshot update creates the file.
	out.Reset()
	errb.Reset()
	if got := Main([]string{"snapshot", "update", p}, &out, &errb); got != ExitOK {
		t.Fatalf("update exit = %d, want %d (stderr=%s)", got, ExitOK, errb.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "snapshots", "out.txt")); err != nil {
		t.Fatalf("snapshot file not created: %v", err)
	}

	// Now the run passes.
	out.Reset()
	errb.Reset()
	if got := Main([]string{"run", p}, &out, &errb); got != ExitOK {
		t.Fatalf("post-update exit = %d, want %d (out=%s)", got, ExitOK, out.String())
	}
}

func TestMain_VersionAndUnknown(t *testing.T) {
	var out, errb bytes.Buffer
	if got := Main([]string{"version"}, &out, &errb); got != ExitOK {
		t.Errorf("version exit = %d, want 0", got)
	}
	out.Reset()
	errb.Reset()
	if got := Main([]string{"bogus"}, &out, &errb); got != ExitConfig {
		t.Errorf("unknown exit = %d, want %d", got, ExitConfig)
	}
}
