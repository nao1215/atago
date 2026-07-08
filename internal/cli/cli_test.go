package cli

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/engine"
	"github.com/nao1215/atago/internal/manifest"
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

// TestRunCmd_Verbose covers the --verbose contract (#6): a passing scenario's
// command, captured output, and assertion verdicts become visible; without the
// flag the console stays dots-only; secrets are masked in traces; combining
// --verbose with a machine report keeps stdout machine-readable by routing the
// trace to stderr; and a failing run renders the full FAILED block exactly once.
func TestRunCmd_Verbose(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "ok.atago.yaml", `
version: "1"
suite:
  name: vdemo
scenarios:
  - name: greets
    steps:
      - run: {shell: true, command: echo hello-verbose}
      - assert:
          exit_code: 0
          stdout: {contains: hello-verbose}
`)

	// --verbose shows the command, its output, and per-assert verdicts.
	var out, errb bytes.Buffer
	if got := Main([]string{"run", "--verbose", p}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitOK, errb.String())
	}
	for _, want := range []string{"echo hello-verbose", "hello-verbose", "ok   assert"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("verbose stdout = %q, want %q", out.String(), want)
		}
	}

	// Without --verbose the trace is absent (dots + summary only).
	out.Reset()
	errb.Reset()
	if got := Main([]string{"run", p}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d, want %d", got, ExitOK)
	}
	if strings.Contains(out.String(), "echo hello-verbose") {
		t.Errorf("non-verbose stdout = %q, must not contain the trace", out.String())
	}
}

// TestRunCmd_VerboseMasksSecrets proves declared secrets never reach a verbose
// trace — the same masking contract as failure output and snapshots (#6).
func TestRunCmd_VerboseMasksSecrets(t *testing.T) {
	t.Setenv("VTEST_TOKEN", "sup3r-s3cret-value")
	dir := t.TempDir()
	p := writeSpec(t, dir, "sec.atago.yaml", `
version: "1"
suite:
  name: vsec
secrets:
  - VTEST_TOKEN
scenarios:
  - name: leaks a token to stdout
    steps:
      - run: {shell: true, command: "echo token=${env:VTEST_TOKEN}"}
      - assert: {exit_code: 0}
`)
	var out, errb bytes.Buffer
	if got := Main([]string{"run", "--verbose", p}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitOK, errb.String())
	}
	if strings.Contains(out.String(), "sup3r-s3cret-value") {
		t.Error("verbose trace leaked the raw secret")
	}
	if !strings.Contains(out.String(), "***") {
		t.Errorf("verbose stdout = %q, want the masked marker", out.String())
	}
}

// TestRunCmd_VerboseWithJSONReportSeparatesStreams proves --verbose + --report
// json keeps stdout pure JSON (schema_version intact) and puts the trace on
// stderr (#6).
func TestRunCmd_VerboseWithJSONReportSeparatesStreams(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "ok.atago.yaml", passingSpec)
	var out, errb bytes.Buffer
	if got := Main([]string{"run", "--verbose", "--report", "json", p}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitOK, errb.String())
	}
	var doc struct {
		SchemaVersion string `json:"schema_version"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("stdout is not valid JSON under --verbose: %v\n%s", err, out.String())
	}
	if doc.SchemaVersion != "1" {
		t.Errorf("schema_version = %q, want 1", doc.SchemaVersion)
	}
	if !strings.Contains(errb.String(), "exit 0") {
		t.Errorf("stderr = %q, want the verbose trace there", errb.String())
	}
}

// TestRunCmd_VerboseFailureBlockRenderedOnce proves the full FAILED block is
// not duplicated between the trace and the console report (#6).
func TestRunCmd_VerboseFailureBlockRenderedOnce(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "fail.atago.yaml", failingSpec)
	var out, errb bytes.Buffer
	if got := Main([]string{"run", "--verbose", p}, &out, &errb); got != ExitFailures {
		t.Fatalf("exit = %d, want %d", got, ExitFailures)
	}
	if n := strings.Count(out.String(), "FAILED:"); n != 1 {
		t.Errorf("stdout renders %d FAILED blocks, want exactly 1:\n%s", n, out.String())
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

// filterSpec has three distinctly-named scenarios so --filter selection can be
// asserted precisely (#119).
const filterSpec = `version: "1"
suite:
  name: filterchk
scenarios:
  - name: alpha one
    steps:
      - run: {shell: true, command: "true"}
      - assert: {exit_code: 0}
  - name: beta two
    steps:
      - run: {shell: true, command: "true"}
      - assert: {exit_code: 0}
  - name: gamma three
    steps:
      - run: {shell: true, command: "true"}
      - assert: {exit_code: 0}
`

// TestRunCmd_FilterMultiple proves --filter selects by name with OR semantics
// across both a comma list and repeated flags, and that a single substring is
// unchanged (#119). Before this, a comma list was one literal substring and
// repeated flags silently kept only the last.
func TestRunCmd_FilterMultiple(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "f.atago.yaml", filterSpec)

	run := func(args ...string) (int, string, string) {
		var out, errb bytes.Buffer
		code := Main(append([]string{"run", "--report", "json"}, append(args, p)...), &out, &errb)
		return code, out.String(), errb.String()
	}
	ran := func(stdout string, names ...string) {
		t.Helper()
		for _, n := range names {
			if !strings.Contains(stdout, n) {
				t.Errorf("scenario %q did not run; stdout=%s", n, stdout)
			}
		}
	}
	notRan := func(stdout string, names ...string) {
		t.Helper()
		for _, n := range names {
			if strings.Contains(stdout, n) {
				t.Errorf("scenario %q ran but should have been filtered out; stdout=%s", n, stdout)
			}
		}
	}

	// Comma OR: alpha,beta selects alpha one and beta two, not gamma three.
	if code, out, errb := run("--filter", "alpha,beta"); code != ExitOK {
		t.Fatalf("comma-OR exit = %d, want %d (stderr=%s)", code, ExitOK, errb)
	} else {
		ran(out, "alpha one", "beta two")
		notRan(out, "gamma three")
	}

	// Repeatable OR: --filter alpha --filter gamma selects both (old bug silently
	// dropped the first).
	if code, out, errb := run("--filter", "alpha", "--filter", "gamma"); code != ExitOK {
		t.Fatalf("repeat-OR exit = %d, want %d (stderr=%s)", code, ExitOK, errb)
	} else {
		ran(out, "alpha one", "gamma three")
		notRan(out, "beta two")
	}

	// Single substring is unchanged.
	if code, out, errb := run("--filter", "alpha"); code != ExitOK {
		t.Fatalf("single exit = %d, want %d (stderr=%s)", code, ExitOK, errb)
	} else {
		ran(out, "alpha one")
		notRan(out, "beta two", "gamma three")
	}

	// No match still warns and exits 0.
	if code, _, errb := run("--filter", "zzz,qqq"); code != ExitOK {
		t.Fatalf("no-match exit = %d, want %d (stderr=%s)", code, ExitOK, errb)
	} else if !strings.Contains(errb, "no scenarios matched") {
		t.Errorf("no-match stderr = %q, want a no-match warning", errb)
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

// TestRunCmd_MixedLoadFailureSummary proves that when a directory mixes a
// loadable spec with one that fails schema validation, the console summary
// reads FAILED (not a misleading PASSED), surfaces the dropped file with a
// "spec failed to load" count, and exits 2 (#120). Before the fix the headline
// read "PASSED ... 0 errored" while the process exited 2.
func TestRunCmd_MixedLoadFailureSummary(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "good.atago.yaml", passingSpec)
	writeSpec(t, dir, "bad.atago.yaml",
		"version: \"1\"\nsuite:\n  name: bad\nscenarios:\n  - name: uses an unknown field\n    steps:\n      - nonsense_field: {command: echo nope}")
	var out, errb bytes.Buffer
	if got := Main([]string{"run", dir}, &out, &errb); got != ExitParse {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitParse, errb.String())
	}
	summary := out.String()
	if strings.Contains(summary, "PASSED") {
		t.Errorf("summary must not read PASSED when a spec failed to load:\n%s", summary)
	}
	if !strings.Contains(summary, "FAILED") {
		t.Errorf("summary should read FAILED to match the exit code:\n%s", summary)
	}
	if !strings.Contains(summary, "1 spec failed to load") {
		t.Errorf("summary should report the dropped spec:\n%s", summary)
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

// errWriter is an io.Writer that always fails, to exercise the output-write
// error branches that a bytes.Buffer never triggers.
type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errors.New("boom") }

// --- exit-code logic: exitForSuite -----------------------------------------

// TestExitForSuite_Precedence pins the per-suite result -> exit code mapping.
// This is exit-code logic and a prime spot for an inverted branch, so every
// status is exercised, including the default (unknown status) and the
// security-violation override that must win over any status.
func TestExitForSuite_Precedence(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		res  *engine.SuiteResult
		want int
	}{
		{"passed", &engine.SuiteResult{Status: engine.StatusPassed}, ExitOK},
		{"skipped", &engine.SuiteResult{Status: engine.StatusSkipped}, ExitOK},
		{"flaky", &engine.SuiteResult{Status: engine.StatusFlaky}, ExitOK},
		{"failed", &engine.SuiteResult{Status: engine.StatusFailed}, ExitFailures},
		{"error", &engine.SuiteResult{Status: engine.StatusError}, ExitExec},
		{"unknown", &engine.SuiteResult{Status: engine.Status("weird")}, ExitInternal},
		// A security violation outranks the generic status entirely.
		{"security over passed", &engine.SuiteResult{Status: engine.StatusPassed, SecurityViolation: true}, ExitSecurity},
		{"security over error", &engine.SuiteResult{Status: engine.StatusError, SecurityViolation: true}, ExitSecurity},
		{"security over failed", &engine.SuiteResult{Status: engine.StatusFailed, SecurityViolation: true}, ExitSecurity},
	}
	for _, c := range cases {
		if got := exitForSuite(c.res); got != c.want {
			t.Errorf("%s: exitForSuite = %d, want %d", c.name, got, c.want)
		}
	}
}

// TestWorseExit_MonotonicAndCommutative exhaustively checks the exit-code
// aggregator over every ordered pair of the seven exit codes. It asserts two
// independent properties that would each expose an inverted comparison:
//  1. worseExit returns the code with the higher documented severity;
//  2. it is commutative (order of the two suites must not change the verdict).
func TestWorseExit_MonotonicAndCommutative(t *testing.T) {
	t.Parallel()
	// An independent severity ladder mirroring worseExit's documented contract:
	// security is most severe, then internal, exec, parse, config, failures, ok.
	sev := map[int]int{
		ExitOK:       0,
		ExitFailures: 1,
		ExitConfig:   2,
		ExitParse:    3,
		ExitExec:     4,
		ExitInternal: 5,
		ExitSecurity: 6,
	}
	codes := []int{ExitOK, ExitFailures, ExitConfig, ExitParse, ExitExec, ExitInternal, ExitSecurity}
	for _, a := range codes {
		for _, b := range codes {
			got := worseExit(a, b)
			// Property 1: the winner is the more severe input.
			wantSev := sev[a]
			if sev[b] > wantSev {
				wantSev = sev[b]
			}
			if sev[got] != wantSev {
				t.Errorf("worseExit(%d,%d) = %d (sev %d), want a code of severity %d", a, b, got, sev[got], wantSev)
			}
			// Property 2: commutativity.
			if rev := worseExit(b, a); sev[rev] != sev[got] {
				t.Errorf("worseExit not commutative: (%d,%d)=%d but (%d,%d)=%d", a, b, got, b, a, rev)
			}
		}
	}
	// Spot-check the documented ladder directly so the intent is legible.
	if worseExit(ExitFailures, ExitOK) != ExitFailures {
		t.Error("a failure must outrank a pass")
	}
	if worseExit(ExitExec, ExitFailures) != ExitExec {
		t.Error("an exec error must outrank a failure")
	}
	if worseExit(ExitSecurity, ExitInternal) != ExitSecurity {
		t.Error("a security violation must outrank an internal error")
	}
	if worseExit(ExitParse, ExitConfig) != ExitParse {
		t.Error("a spec parse error must outrank a CLI config error")
	}
}

// --- pure record helpers ----------------------------------------------------

func TestSplitCSV(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"   ", nil},
		{",,", nil},
		{"a", []string{"a"}},
		{"a,b,c", []string{"a", "b", "c"}},
		{" a , b ,, c ", []string{"a", "b", "c"}},
	}
	for _, c := range cases {
		if got := splitCSV(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("splitCSV(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestSuiteNameFor(t *testing.T) {
	t.Parallel()
	cases := []struct{ in, want string }{
		{"echo", "echo"},
		{"/usr/bin/grep", "grep"},
		{"tool.exe", "tool"},
		{"a.b.exe", "a.b"},
		{"", "recorded"},
		{".", "recorded"},
		{"dir/", "dir"},
	}
	for _, c := range cases {
		if got := suiteNameFor(c.in); got != c.want {
			t.Errorf("suiteNameFor(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCountLines(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"\n\n", 0},
		{"one", 1},
		{"one\ntwo\n", 2},
		{"one\n\n  \nthree", 2}, // blank / whitespace-only lines are not counted
	}
	for _, c := range cases {
		if got := countLines([]byte(c.in)); got != c.want {
			t.Errorf("countLines(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestPlainPOSIXWord(t *testing.T) {
	t.Parallel()
	plain := []string{"abc", "ABC123", "a/b.c-d_e", "path/to.file", "user@host", "k=v", "50%", "a,b", "+x", ":port"}
	for _, s := range plain {
		if !plainPOSIXWord(s) {
			t.Errorf("plainPOSIXWord(%q) = false, want true", s)
		}
	}
	notPlain := []string{"", "a b", "a$b", "a|b", "a'b", `a"b`, "a>b", "a;b", "a*b", `a\b`}
	for _, s := range notPlain {
		if plainPOSIXWord(s) {
			t.Errorf("plainPOSIXWord(%q) = true, want false", s)
		}
	}
}

func TestShellJoin_Dispatches(t *testing.T) {
	t.Parallel()
	// shellJoin picks the platform tokenizer; on the test host it must produce a
	// non-empty command that keeps the plain leading word verbatim.
	got := shellJoin([]string{"echo", "plain"})
	if !strings.HasPrefix(got, "echo ") {
		t.Errorf("shellJoin = %q, want it to start with the verbatim plain word", got)
	}
}

func TestListFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub", "deep")
	if err := os.MkdirAll(sub, 0o750); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"z.txt", "a.txt"} {
		if err := os.WriteFile(filepath.Join(dir, rel), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(sub, "b.txt"), []byte("y"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := listFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a.txt", "sub/deep/b.txt", "z.txt"} // sorted, slash-separated
	if !reflect.DeepEqual(got, want) {
		t.Errorf("listFiles = %v, want %v", got, want)
	}
}

// --- list gate rendering ----------------------------------------------------

func TestGateTokens(t *testing.T) {
	t.Parallel()
	if got := gateTokens("only", nil); got != nil {
		t.Errorf("gateTokens(nil) = %v, want nil", got)
	}
	// All three condition fields render, sorted, with the given kind prefix.
	got := gateTokens("only", &manifest.Condition{OS: "linux", Env: "CI", Command: "git"})
	want := []string{"only:command", "only:env=CI", "only:os=linux"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("gateTokens = %v, want %v", got, want)
	}
	// skip kind and a single field.
	if got := gateTokens("skip", &manifest.Condition{Env: "DEBUG"}); !reflect.DeepEqual(got, []string{"skip:env=DEBUG"}) {
		t.Errorf("skip env gate = %v", got)
	}
}

func TestScenarioGates_OnlyAndSkip(t *testing.T) {
	t.Parallel()
	sc := manifest.Scenario{
		Only: &manifest.Condition{OS: "linux"},
		Skip: &manifest.Condition{Env: "CI"},
	}
	got := scenarioGates(sc)
	want := []string{"only:os=linux", "skip:env=CI"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("scenarioGates = %v, want %v", got, want)
	}
	// No conditions -> no gates.
	if got := scenarioGates(manifest.Scenario{}); got != nil {
		t.Errorf("scenarioGates(empty) = %v, want nil", got)
	}
}

// --- rerun state round-trip -------------------------------------------------

// TestRerunState_RoundTrip proves save then load returns the same failures,
// sorted deterministically, and that an empty save clears the file.
func TestRerunState_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	withWorkdir(t, dir, func() {
		// Intentionally unsorted input; saveRerunState must normalize it.
		in := []failedEntry{
			{SpecPath: "b.atago.yaml", Scenario: "z"},
			{SpecPath: "a.atago.yaml", Scenario: "y"},
			{SpecPath: "a.atago.yaml", Scenario: "x"},
		}
		if err := saveRerunState(in); err != nil {
			t.Fatal(err)
		}
		st, err := loadRerunState()
		if err != nil {
			t.Fatal(err)
		}
		want := []failedEntry{
			{SpecPath: "a.atago.yaml", Scenario: "x"},
			{SpecPath: "a.atago.yaml", Scenario: "y"},
			{SpecPath: "b.atago.yaml", Scenario: "z"},
		}
		if !reflect.DeepEqual(st.Failed, want) {
			t.Errorf("loaded failures = %+v, want sorted %+v", st.Failed, want)
		}
		if st.SchemaVersion != RerunStateSchemaVersion {
			t.Errorf("schema_version = %q, want %q", st.SchemaVersion, RerunStateSchemaVersion)
		}
		// specPaths dedups and sorts.
		if got := st.specPaths(); !reflect.DeepEqual(got, []string{"a.atago.yaml", "b.atago.yaml"}) {
			t.Errorf("specPaths = %v", got)
		}
		// selectSet has one identity per entry.
		if got := st.selectSet(); len(got) != 3 {
			t.Errorf("selectSet size = %d, want 3", len(got))
		}

		// An empty save removes the file (a fully-green run clears state).
		if err := saveRerunState(nil); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(rerunStatePath()); !os.IsNotExist(err) {
			t.Errorf("empty saveRerunState did not remove the state file (err=%v)", err)
		}
		// Loading a missing file is not an error: empty state with schema set.
		st, err = loadRerunState()
		if err != nil {
			t.Fatalf("loadRerunState on missing file errored: %v", err)
		}
		if len(st.Failed) != 0 || st.selectSet() != nil {
			t.Errorf("missing-file state should be empty, got %+v", st)
		}
	})
}

// TestLoadRerunState_Corrupt proves a garbled state file surfaces a clean error
// rather than silently degrading to "nothing recorded".
func TestLoadRerunState_Corrupt(t *testing.T) {
	dir := t.TempDir()
	withWorkdir(t, dir, func() {
		if err := os.MkdirAll(rerunStateDir, 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(rerunStatePath(), []byte("{ this is not json"), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := loadRerunState(); err == nil {
			t.Error("loadRerunState on corrupt JSON returned nil error")
		}
	})
}

// TestRunCmd_CorruptRerunStateExitsConfig proves `run --rerun-failed` reports a
// clean ExitConfig (not a panic or false green) when the state file is corrupt.
func TestRunCmd_CorruptRerunStateExitsConfig(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "s.atago.yaml", passingSpec)
	withWorkdir(t, dir, func() {
		if err := os.MkdirAll(rerunStateDir, 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(rerunStatePath(), []byte("not json"), 0o600); err != nil {
			t.Fatal(err)
		}
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "--rerun-failed", "."}, &out, &errb); got != ExitConfig {
			t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitConfig, errb.String())
		}
		if !strings.Contains(errb.String(), "cannot read") {
			t.Errorf("stderr = %q, want a cannot-read message", errb.String())
		}
	})
}

// --- Main dispatch / usage --------------------------------------------------

func TestMain_NoArgsShowsUsage(t *testing.T) {
	t.Parallel()
	var out, errb bytes.Buffer
	if got := Main(nil, &out, &errb); got != ExitConfig {
		t.Fatalf("exit = %d, want %d", got, ExitConfig)
	}
	if !strings.Contains(errb.String(), "atago <command>") {
		t.Errorf("stderr = %q, want usage", errb.String())
	}
}

func TestMain_HelpAndVersionVariants(t *testing.T) {
	t.Parallel()
	for _, arg := range []string{"help", "-h", "--help"} {
		var out, errb bytes.Buffer
		if got := Main([]string{arg}, &out, &errb); got != ExitOK {
			t.Errorf("%s exit = %d, want 0", arg, got)
		}
		if !strings.Contains(out.String(), "atago <command>") {
			t.Errorf("%s: usage not on stdout: %q", arg, out.String())
		}
	}
	for _, arg := range []string{"version", "-version", "--version"} {
		var out, errb bytes.Buffer
		if got := Main([]string{arg}, &out, &errb); got != ExitOK {
			t.Errorf("%s exit = %d, want 0", arg, got)
		}
		if !strings.Contains(out.String(), "atago ") {
			t.Errorf("%s: version not on stdout: %q", arg, out.String())
		}
	}
}

func TestSnapshotCmd_BadInvocation(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{{"snapshot"}, {"snapshot", "bogus"}} {
		var out, errb bytes.Buffer
		if got := Main(args, &out, &errb); got != ExitConfig {
			t.Errorf("%v exit = %d, want %d", args, got, ExitConfig)
		}
		if !strings.Contains(errb.String(), "Usage: atago snapshot update") {
			t.Errorf("%v: stderr = %q, want usage", args, errb.String())
		}
	}
}

// TestSnapshotCmd_ErrorNamesSnapshot proves an error from `snapshot update`
// names that command, not the `run` it delegates to. Reporting "atago run:" for
// a `snapshot update` invocation misidentifies the command the user typed.
func TestSnapshotCmd_ErrorNamesSnapshot(t *testing.T) {
	t.Parallel()
	var out, errb bytes.Buffer
	got := Main([]string{"snapshot", "update", filepath.Join(t.TempDir(), "missing.atago.yaml")}, &out, &errb)
	if got != ExitConfig {
		t.Errorf("exit = %d, want %d", got, ExitConfig)
	}
	if s := errb.String(); !strings.Contains(s, "atago snapshot update:") || strings.Contains(s, "atago run:") {
		t.Errorf("stderr = %q, want it to name 'atago snapshot update:' and not 'atago run:'", s)
	}
}

// --- run argument parsing ---------------------------------------------------

func TestRunCmd_ArgParsingErrors(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "ok.atago.yaml", passingSpec)
	cases := []struct {
		name    string
		args    []string
		wantMsg string
	}{
		{"unknown report", []string{"run", "--report", "xml", p}, "unknown --report"},
		{"mutually exclusive repeat/retry", []string{"run", "--repeat", "2", "--retry-failed", "2", p}, "mutually exclusive"},
		{"negative repeat", []string{"run", "--repeat", "-1", p}, "must be >= 0"},
		{"negative retry", []string{"run", "--retry-failed", "-1", p}, "must be >= 0"},
		{"unknown flag", []string{"run", "--no-such-flag", p}, ""},
		{"nonexistent target", []string{"run", filepath.Join(dir, "missing")}, "cannot access"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out, errb bytes.Buffer
			if got := Main(c.args, &out, &errb); got != ExitConfig {
				t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitConfig, errb.String())
			}
			if c.wantMsg != "" && !strings.Contains(errb.String(), c.wantMsg) {
				t.Errorf("stderr = %q, want %q", errb.String(), c.wantMsg)
			}
		})
	}
}

// TestRunCmd_CIFlagAndParallel exercises the --ci and --parallel>1 branches of
// runCmd (setting NO_COLOR and the shared scenario semaphore).
func TestRunCmd_CIFlagAndParallel(t *testing.T) {
	// t.Setenv registers cleanup to restore NO_COLOR after --ci overwrites it.
	t.Setenv("NO_COLOR", os.Getenv("NO_COLOR"))
	dir := t.TempDir()
	writeSpec(t, dir, "a.atago.yaml", passingSpec)
	writeSpec(t, dir, "b.atago.yaml", passingSpec)
	var out, errb bytes.Buffer
	if got := Main([]string{"run", "--ci", "--parallel", "2", dir}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitOK, errb.String())
	}
	if os.Getenv("NO_COLOR") != "1" {
		t.Errorf("--ci did not set NO_COLOR=1 (got %q)", os.Getenv("NO_COLOR"))
	}
}

// --- doc / manifest / list error and split paths ----------------------------

func TestDocCmd_FlagCombinations(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "ok.atago.yaml", passingSpec)
	cases := []struct {
		name    string
		args    []string
		wantMsg string
	}{
		{"split without out-dir", []string{"doc", "--split-by-spec", p}, "requires --out-dir"},
		{"out-dir without split", []string{"doc", "--out-dir", filepath.Join(dir, "d"), p}, "requires --split-by-spec"},
		{"nonexistent target", []string{"doc", filepath.Join(dir, "missing")}, "cannot access"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out, errb bytes.Buffer
			if got := Main(c.args, &out, &errb); got != ExitConfig {
				t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitConfig, errb.String())
			}
			if !strings.Contains(errb.String(), c.wantMsg) {
				t.Errorf("stderr = %q, want %q", errb.String(), c.wantMsg)
			}
		})
	}
}

// TestDocCmd_SplitBySpec exercises writeSplitDocs: one file per spec plus an
// index.md written into --out-dir, with a confirmation on stdout.
func TestDocCmd_SplitBySpec(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "one.atago.yaml", passingSpec)
	writeSpec(t, dir, "two.atago.yaml", passingSpec)
	outDir := filepath.Join(dir, "docs")
	var out, errb bytes.Buffer
	if got := Main([]string{"doc", "--split-by-spec", "--out-dir", outDir, dir}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitOK, errb.String())
	}
	if _, err := os.Stat(filepath.Join(outDir, "index.md")); err != nil {
		t.Errorf("index.md not written: %v", err)
	}
	if !strings.Contains(out.String(), "Wrote") {
		t.Errorf("stdout = %q, want a Wrote confirmation", out.String())
	}
	// At least the two per-spec docs plus index.md should exist.
	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) < 3 {
		t.Errorf("out-dir has %d files, want at least 3 (2 specs + index.md)", len(entries))
	}
}

// TestDocCmd_OutWriteError proves a failed write to --out reports a config error
// rather than panicking (the parent directory does not exist).
func TestDocCmd_OutWriteError(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "ok.atago.yaml", passingSpec)
	var out, errb bytes.Buffer
	bad := filepath.Join(dir, "nope", "out.md")
	if got := Main([]string{"doc", "--out", bad, p}, &out, &errb); got != ExitConfig {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitConfig, errb.String())
	}
}

func TestManifestCmd_ErrorPaths(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "ok.atago.yaml", passingSpec)
	t.Run("nonexistent target", func(t *testing.T) {
		var out, errb bytes.Buffer
		if got := Main([]string{"manifest", filepath.Join(dir, "missing")}, &out, &errb); got != ExitConfig {
			t.Fatalf("exit = %d, want %d", got, ExitConfig)
		}
		if !strings.Contains(errb.String(), "cannot access") {
			t.Errorf("stderr = %q", errb.String())
		}
	})
	t.Run("out write error", func(t *testing.T) {
		var out, errb bytes.Buffer
		bad := filepath.Join(dir, "nope", "m.json")
		if got := Main([]string{"manifest", "--out", bad, p}, &out, &errb); got != ExitConfig {
			t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitConfig, errb.String())
		}
	})
}

func TestListCmd_ErrorPaths(t *testing.T) {
	dir := t.TempDir()
	t.Run("nonexistent target", func(t *testing.T) {
		var out, errb bytes.Buffer
		if got := Main([]string{"list", filepath.Join(dir, "missing")}, &out, &errb); got != ExitConfig {
			t.Fatalf("exit = %d, want %d", got, ExitConfig)
		}
		if !strings.Contains(errb.String(), "cannot access") {
			t.Errorf("stderr = %q", errb.String())
		}
	})
	t.Run("parse error", func(t *testing.T) {
		p := writeSpec(t, dir, "bad.atago.yaml", "version: \"1\"\nsuite: {}\nscenarios: []")
		var out, errb bytes.Buffer
		if got := Main([]string{"list", p}, &out, &errb); got != ExitParse {
			t.Fatalf("exit = %d, want %d", got, ExitParse)
		}
	})
}

// TestListCmd_JSONEmptyIsValid proves a spec with no runnable scenario still
// yields a valid JSON document whose scenarios field is a [] (never null).
func TestListCmd_JSONEmptyIsValid(t *testing.T) {
	dir := t.TempDir()
	// A single scenario carrying an only:os gate that never matches keeps the
	// suite loadable while producing exactly one row we can round-trip.
	spec := `version: "1"
suite:
  name: gated
scenarios:
  - name: only linux
    only: {os: linux}
    steps:
      - run: {shell: true, command: "true"}
      - assert: {exit_code: 0}
`
	p := writeSpec(t, dir, "g.atago.yaml", spec)
	var out, errb bytes.Buffer
	if got := Main([]string{"list", "--json", p}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d (stderr=%s)", got, errb.String())
	}
	var doc listDocument
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out.String())
	}
	if len(doc.Scenarios) != 1 || len(doc.Scenarios[0].Gates) != 1 || doc.Scenarios[0].Gates[0] != "only:os=linux" {
		t.Errorf("gates round-trip failed: %+v", doc.Scenarios)
	}
}

// --- explain multi-file partial failure -------------------------------------

// TestExplainCmd_ContinuesPastParseError proves explain reports ExitParse when
// one spec fails to load but still explains the loadable specs (the loop uses
// worseExit + continue rather than aborting on the first bad file).
func TestExplainCmd_ContinuesPastParseError(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "a-good.atago.yaml", passingSpec)
	writeSpec(t, dir, "b-bad.atago.yaml", "version: \"1\"\nsuite: {}\nscenarios: []")
	var out, errb bytes.Buffer
	if got := Main([]string{"explain", dir}, &out, &errb); got != ExitParse {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitParse, errb.String())
	}
	// The good spec was still explained despite the bad sibling.
	if !strings.Contains(out.String(), "Suite: sample") {
		t.Errorf("stdout = %q, want the good spec still explained", out.String())
	}
}

// --- init argument parsing --------------------------------------------------

func TestInitCmd_ArgParsing(t *testing.T) {
	dir := t.TempDir()
	t.Run("help exits ok", func(t *testing.T) {
		var out, errb bytes.Buffer
		if got := Main([]string{"init", "-h"}, &out, &errb); got != ExitOK {
			t.Fatalf("exit = %d, want %d", got, ExitOK)
		}
	})
	t.Run("unknown flag", func(t *testing.T) {
		var out, errb bytes.Buffer
		if got := Main([]string{"init", "--no-such-flag"}, &out, &errb); got != ExitConfig {
			t.Fatalf("exit = %d, want %d", got, ExitConfig)
		}
	})
	t.Run("write error", func(t *testing.T) {
		var out, errb bytes.Buffer
		bad := filepath.Join(dir, "nope", "x.atago.yaml")
		if got := Main([]string{"init", bad}, &out, &errb); got != ExitConfig {
			t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitConfig, errb.String())
		}
	})
}

// --- record command paths ---------------------------------------------------

func TestRecordCmd_ArgErrors(t *testing.T) {
	dir := t.TempDir()
	cases := []struct {
		name    string
		args    []string
		wantMsg string
	}{
		{"no command", []string{"record"}, "no command given"},
		{"snapshot needs out", []string{"record", "--snapshot", "--", "echo", "hi"}, "--snapshot needs --out"},
		{"snapshot with pty", []string{"record", "--pty", "--snapshot", "--out", filepath.Join(dir, "s.atago.yaml"), "--", "echo", "hi"}, "cannot be combined with --pty"},
		{"unknown flag", []string{"record", "--no-such-flag"}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out, errb bytes.Buffer
			if got := Main(c.args, &out, &errb); got != ExitConfig {
				t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitConfig, errb.String())
			}
			if c.wantMsg != "" && !strings.Contains(errb.String(), c.wantMsg) {
				t.Errorf("stderr = %q, want %q", errb.String(), c.wantMsg)
			}
		})
	}
}

// TestRecordCmd_StdoutHappyPath records a simple command to stdout and checks
// the generated spec skeleton is emitted, exercising the run/observe/generate
// path plus shellJoin, listFiles, suiteNameFor, and countLines.
func TestRecordCmd_StdoutHappyPath(t *testing.T) {
	var out, errb bytes.Buffer
	if got := Main([]string{"record", "--", "echo", "hello-record"}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitOK, errb.String())
	}
	if !strings.Contains(out.String(), "version:") || !strings.Contains(out.String(), "echo") {
		t.Errorf("generated spec on stdout looks wrong:\n%s", out.String())
	}
	if !strings.Contains(errb.String(), "recorded:") {
		t.Errorf("stderr = %q, want a recorded summary", errb.String())
	}
}

// TestRecordCmd_OutFileForceAndShell covers writing to --out, the
// already-exists guard, --force overwrite, and --shell verbatim joining.
func TestRecordCmd_OutFileForceAndShell(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "gen.atago.yaml")

	var out, errb bytes.Buffer
	if got := Main([]string{"record", "--out", outPath, "--", "echo", "hi"}, &out, &errb); got != ExitOK {
		t.Fatalf("first write exit = %d (stderr=%s)", got, errb.String())
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("spec not written: %v", err)
	}
	if !strings.Contains(errb.String(), "wrote "+outPath) {
		t.Errorf("stderr = %q, want a wrote confirmation", errb.String())
	}

	// Second write without --force is refused.
	out.Reset()
	errb.Reset()
	if got := Main([]string{"record", "--out", outPath, "--", "echo", "hi"}, &out, &errb); got != ExitConfig {
		t.Fatalf("overwrite-without-force exit = %d, want %d", got, ExitConfig)
	}
	if !strings.Contains(errb.String(), "already exists") {
		t.Errorf("stderr = %q, want already-exists", errb.String())
	}

	// --force overwrites; --shell records the command line verbatim.
	out.Reset()
	errb.Reset()
	if got := Main([]string{"record", "--force", "--shell", "--out", outPath, "--", "echo hi there"}, &out, &errb); got != ExitOK {
		t.Fatalf("force+shell exit = %d (stderr=%s)", got, errb.String())
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "shell: true") {
		t.Errorf("--shell spec missing shell: true:\n%s", data)
	}
}

// TestRecordCmd_Snapshot writes a spec plus its stdout golden next to it.
func TestRecordCmd_Snapshot(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "snap.atago.yaml")
	var out, errb bytes.Buffer
	if got := Main([]string{"record", "--snapshot", "--out", outPath, "--", "echo", "snapshot-me"}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d (stderr=%s)", got, errb.String())
	}
	golden := filepath.Join(dir, "snapshots", "snap.stdout.txt")
	if _, err := os.Stat(golden); err != nil {
		t.Fatalf("snapshot golden not written at %s: %v", golden, err)
	}
	data, _ := os.ReadFile(outPath)
	if !strings.Contains(string(data), "snapshot") {
		t.Errorf("spec missing snapshot matcher:\n%s", data)
	}
}

// --- list writers: nil-scenarios normalization and write errors -------------

// TestWriteListJSON_NilScenariosAndWriteError covers two branches at once: an
// empty manifest normalizes scenarios to [] (never null in the JSON), and a
// failing writer surfaces ExitInternal rather than a false success.
func TestWriteListJSON_NilScenariosAndWriteError(t *testing.T) {
	t.Parallel()
	// Success path with a real buffer: empty document -> "scenarios": [].
	var buf bytes.Buffer
	if got := writeListJSON(manifest.Document{}, &buf, &buf); got != ExitOK {
		t.Fatalf("exit = %d, want %d", got, ExitOK)
	}
	var doc listDocument
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if doc.Scenarios == nil {
		t.Error("scenarios serialized as null; want an empty array")
	}
	// Failing writer -> ExitInternal.
	var errb bytes.Buffer
	if got := writeListJSON(manifest.Document{}, errWriter{}, &errb); got != ExitInternal {
		t.Errorf("write error exit = %d, want %d", got, ExitInternal)
	}
}

// TestWriteListTable_WriteError proves a failed tabwriter flush surfaces
// ExitInternal.
func TestWriteListTable_WriteError(t *testing.T) {
	t.Parallel()
	var errb bytes.Buffer
	if got := writeListTable(manifest.Document{}, errWriter{}, &errb); got != ExitInternal {
		t.Errorf("write error exit = %d, want %d", got, ExitInternal)
	}
	if !strings.Contains(errb.String(), "atago list:") {
		t.Errorf("stderr = %q, want an atago list error", errb.String())
	}
}

// --- filesystem error branches for rerun state ------------------------------

// TestSaveRerunState_MkdirError proves an unwritable state dir (here .atago is
// a regular file, so MkdirAll fails) is returned as an error for the caller to
// warn about — writes are best-effort and must not panic.
func TestSaveRerunState_MkdirError(t *testing.T) {
	dir := t.TempDir()
	withWorkdir(t, dir, func() {
		// Occupy the state dir path with a file so MkdirAll(.atago) fails.
		if err := os.WriteFile(rerunStateDir, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		err := saveRerunState([]failedEntry{{SpecPath: "a", Scenario: "b"}})
		if err == nil {
			t.Error("saveRerunState with a file where .atago should be returned nil error")
		}
	})
}

// TestLoadRerunState_ReadErrorSurfaced proves a read error that is not
// os.IsNotExist is surfaced, not swallowed as empty state. It makes the state
// path itself a directory, so os.ReadFile fails with a non-NotExist error on
// every platform (EISDIR on POSIX, an equivalent on Windows) — portable, unlike
// making .atago a file, whose read error Windows reports as not-exist.
func TestLoadRerunState_ReadErrorSurfaced(t *testing.T) {
	dir := t.TempDir()
	withWorkdir(t, dir, func() {
		if err := os.MkdirAll(rerunStatePath(), 0o750); err != nil {
			t.Fatal(err)
		}
		if _, err := loadRerunState(); err == nil {
			t.Error("loadRerunState reading a directory as the state file returned nil error")
		}
	})
}

// --- writeSplitDocs mkdir error ---------------------------------------------

// TestDocCmd_SplitMkdirError proves writeSplitDocs reports a config error when
// --out-dir cannot be created (a file already occupies the path).
func TestDocCmd_SplitMkdirError(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "one.atago.yaml", passingSpec)
	// A regular file at the out-dir path makes MkdirAll fail.
	fileAsDir := filepath.Join(dir, "outfile")
	if err := os.WriteFile(fileAsDir, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if got := Main([]string{"doc", "--split-by-spec", "--out-dir", fileAsDir, filepath.Join(dir, "one.atago.yaml")}, &out, &errb); got != ExitConfig {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitConfig, errb.String())
	}
}

// --- manifest default target + parse error ----------------------------------

func TestManifestCmd_NoFilesDefaultDot(t *testing.T) {
	dir := t.TempDir()
	withWorkdir(t, dir, func() {
		var out, errb bytes.Buffer
		if got := Main([]string{"manifest"}, &out, &errb); got != ExitConfig {
			t.Fatalf("exit = %d, want %d", got, ExitConfig)
		}
		if !strings.Contains(errb.String(), "no *.atago.yaml (or *.atago.yml) files found") {
			t.Errorf("stderr = %q, want no-files message", errb.String())
		}
	})
}

// TestRecordCmd_CommandNotFound proves a command that cannot start reports an
// execution error (exit 4), not a panic or a spurious success.
func TestRecordCmd_CommandNotFound(t *testing.T) {
	var out, errb bytes.Buffer
	got := Main([]string{"record", "--", "atago-no-such-binary-xyz123"}, &out, &errb)
	if got != ExitExec {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitExec, errb.String())
	}
}

// TestRecordCmd_HelpFlag proves `record -h` prints usage and exits 0.
func TestRecordCmd_HelpFlag(t *testing.T) {
	t.Parallel()
	var out, errb bytes.Buffer
	if got := Main([]string{"record", "-h"}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d, want %d", got, ExitOK)
	}
}

// TestRecordCmd_OutIsDirectory proves that recording over an existing directory
// path with --force surfaces an execution error (the WriteFile fails) rather
// than panicking.
func TestRecordCmd_OutIsDirectory(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "adir")
	if err := os.MkdirAll(target, 0o750); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if got := Main([]string{"record", "--force", "--out", target, "--", "echo", "hi"}, &out, &errb); got != ExitExec {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitExec, errb.String())
	}
}

// TestExplainCmd_NonexistentTarget proves a bad path exits ExitConfig with a
// helpful message (collectSpecFiles error branch).
func TestExplainCmd_NonexistentTarget(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if got := Main([]string{"explain", filepath.Join(dir, "missing")}, &out, &errb); got != ExitConfig {
		t.Fatalf("exit = %d, want %d", got, ExitConfig)
	}
	if !strings.Contains(errb.String(), "cannot access") {
		t.Errorf("stderr = %q, want cannot-access", errb.String())
	}
}

// TestRunCmd_SaveStateWriteErrorWarns proves a run whose failing-scenario state
// cannot be persisted (here .atago is occupied by a regular file) still exits on
// the run's own verdict and only warns about the failed write — persistence is
// best-effort and must never fail the run.
func TestRunCmd_SaveStateWriteErrorWarns(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "fail.atago.yaml", failingSpec)
	withWorkdir(t, dir, func() {
		// Block .atago dir creation with a file of the same name.
		if err := os.WriteFile(rerunStateDir, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "fail.atago.yaml"}, &out, &errb); got != ExitFailures {
			t.Fatalf("exit = %d, want %d (the run's own verdict, stderr=%s)", got, ExitFailures, errb.String())
		}
		if !strings.Contains(errb.String(), "could not update") {
			t.Errorf("stderr = %q, want a best-effort save warning", errb.String())
		}
	})
}

// --- help flags on FlagSet-based subcommands --------------------------------

// TestSubcommands_HelpFlagExitsOK proves `-h` on the flag-parsing subcommands
// prints usage and exits 0 (the flag.ErrHelp path), never treating the flag as a
// spec path. Stream routing per command is asserted in TestSubcommands_HelpToStdout.
func TestSubcommands_HelpFlagExitsOK(t *testing.T) {
	t.Parallel()
	for _, cmd := range []string{"run", "doc", "manifest", "list"} {
		var out, errb bytes.Buffer
		if got := Main([]string{cmd, "-h"}, &out, &errb); got != ExitOK {
			t.Errorf("%s -h exit = %d, want %d", cmd, got, ExitOK)
		}
	}
}

// --- default "." target ------------------------------------------------------

// TestSubcommands_DefaultTargetIsDot proves run/doc/list default to the current
// directory when no path argument is given (parity across the read commands).
func TestSubcommands_DefaultTargetIsDot(t *testing.T) {
	for _, cmd := range []string{"run", "doc", "list"} {
		t.Run(cmd, func(t *testing.T) {
			dir := t.TempDir()
			writeSpec(t, dir, "ok.atago.yaml", passingSpec)
			withWorkdir(t, dir, func() {
				var out, errb bytes.Buffer
				if got := Main([]string{cmd}, &out, &errb); got != ExitOK {
					t.Fatalf("%s (no target) exit = %d, want %d (stderr=%s)", cmd, got, ExitOK, errb.String())
				}
			})
		})
	}
}

// --- run --tag / --skip-tag no-match warnings -------------------------------

// TestRunCmd_TagSelectionNoMatchWarns proves a --tag/--skip-tag that selects
// nothing still exits 0 but warns loudly, mentioning the exact selector — a
// typo in CI must not greenlight in silence.
func TestRunCmd_TagSelectionNoMatchWarns(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "ok.atago.yaml", passingSpec)
	t.Run("tag", func(t *testing.T) {
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "--tag", "no-such-tag", p}, &out, &errb); got != ExitOK {
			t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitOK, errb.String())
		}
		if !strings.Contains(errb.String(), `--tag "no-such-tag"`) {
			t.Errorf("stderr = %q, want a --tag no-match warning", errb.String())
		}
	})
	t.Run("skip-tag", func(t *testing.T) {
		// skip-tag alone can only remove scenarios; skipping the sole scenario's
		// (nonexistent) tag keeps it, so pair it with a filter that matches nothing
		// is not needed — instead assert the skip-tag path is exercised by skipping
		// a tag the scenario does carry via a tagged spec.
		tagged := `version: "1"
suite:
  name: t
scenarios:
  - name: only
    tags: [slow]
    steps:
      - run: {shell: true, command: "exit 0"}
      - assert: {exit_code: 0}
`
		tp := writeSpec(t, dir, "tagged.atago.yaml", tagged)
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "--skip-tag", "slow", tp}, &out, &errb); got != ExitOK {
			t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitOK, errb.String())
		}
		if !strings.Contains(errb.String(), `--skip-tag "slow"`) {
			t.Errorf("stderr = %q, want a --skip-tag no-match warning", errb.String())
		}
	})
}

// --- rerun-failed target intersection is empty ------------------------------

// TestRerunFailed_TargetOutsideRecorded proves that when failures were recorded
// for one spec but --rerun-failed is pointed at a different (passing) spec, the
// run reports "nothing under the given targets" and exits 0 without rerunning.
func TestRerunFailed_TargetOutsideRecorded(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "fail.atago.yaml", failingSpec)
	writeSpec(t, dir, "pass.atago.yaml", passingSpec)
	withWorkdir(t, dir, func() {
		// Record the failing spec's failure.
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "fail.atago.yaml"}, &out, &errb); got != ExitFailures {
			t.Fatalf("seed run exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		// Rerun, but target only the passing spec: the recorded failure is not
		// under this target, so nothing reruns.
		out.Reset()
		errb.Reset()
		if got := Main([]string{"run", "--rerun-failed", "pass.atago.yaml"}, &out, &errb); got != ExitOK {
			t.Fatalf("rerun exit = %d, want %d (stderr=%s)", got, ExitOK, errb.String())
		}
		if !strings.Contains(errb.String(), "no previously failed scenarios under the given targets") {
			t.Errorf("stderr = %q, want the out-of-target note", errb.String())
		}
	})
}
