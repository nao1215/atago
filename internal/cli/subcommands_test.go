package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocCmd_Stdout(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "ok.atago.yaml", passingSpec)
	var out, errb bytes.Buffer
	if got := Main([]string{"doc", p}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d (stderr=%s)", got, errb.String())
	}
	if !strings.Contains(out.String(), "# atago Behavior Specs") {
		t.Errorf("doc stdout missing header:\n%s", out.String())
	}
}

func TestDocCmd_OutFile(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "ok.atago.yaml", passingSpec)
	outPath := filepath.Join(dir, "out.md")
	var out, errb bytes.Buffer
	if got := Main([]string{"doc", "--out", outPath, p}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d (stderr=%s)", got, errb.String())
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "# atago Behavior Specs") {
		t.Errorf("doc file missing header:\n%s", data)
	}
	if !strings.Contains(out.String(), "Wrote") {
		t.Errorf("stdout missing Wrote confirmation: %q", out.String())
	}
}

func TestDocCmd_ParseError(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "bad.atago.yaml", "version: \"1\"\nsuite: {}\nscenarios: []")
	var out, errb bytes.Buffer
	if got := Main([]string{"doc", p}, &out, &errb); got != ExitParse {
		t.Fatalf("exit = %d, want %d", got, ExitParse)
	}
}

func TestDocCmd_NoFiles(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if got := Main([]string{"doc", dir}, &out, &errb); got != ExitConfig {
		t.Fatalf("exit = %d, want %d", got, ExitConfig)
	}
}

// TestDocCmd_OutWithSplitRejected is a regression: --out and --split-by-spec are
// contradictory (split writes into --out-dir), so passing both must fail loudly
// instead of silently ignoring --out.
func TestDocCmd_OutWithSplitRejected(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "ok.atago.yaml", passingSpec)
	var out, errb bytes.Buffer
	got := Main([]string{"doc", "--out", filepath.Join(dir, "ignored.md"), "--split-by-spec", "--out-dir", filepath.Join(dir, "d"), p}, &out, &errb)
	if got != ExitConfig {
		t.Fatalf("exit = %d, want %d (mutually-exclusive flags)", got, ExitConfig)
	}
	if !strings.Contains(errb.String(), "mutually exclusive") {
		t.Errorf("stderr = %q, want a mutually-exclusive message", errb.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "ignored.md")); !os.IsNotExist(err) {
		t.Errorf("the ignored --out file was written despite the rejected combination")
	}
}

func TestManifestCmd_Stdout(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "ok.atago.yaml", passingSpec)
	var out, errb bytes.Buffer
	if got := Main([]string{"manifest", p}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d (stderr=%s)", got, errb.String())
	}
	// Output must be valid JSON carrying a schema_version and the suite name.
	var doc struct {
		SchemaVersion string `json:"schema_version"`
		Specs         []struct {
			Suite string `json:"suite"`
		} `json:"specs"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("manifest output is not valid JSON: %v\n%s", err, out.String())
	}
	if doc.SchemaVersion != "1" {
		t.Errorf("schema_version = %q", doc.SchemaVersion)
	}
	if len(doc.Specs) != 1 || doc.Specs[0].Suite != "sample" {
		t.Errorf("specs = %+v", doc.Specs)
	}
}

func TestManifestCmd_OutFile(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "ok.atago.yaml", passingSpec)
	outPath := filepath.Join(dir, "manifest.json")
	var out, errb bytes.Buffer
	if got := Main([]string{"manifest", "--out", outPath, p}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d (stderr=%s)", got, errb.String())
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(data) {
		t.Errorf("written manifest is not valid JSON:\n%s", data)
	}
	if !strings.Contains(out.String(), "Wrote") {
		t.Errorf("stdout missing Wrote confirmation: %q", out.String())
	}
}

func TestManifestCmd_ParseError(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "bad.atago.yaml", "version: \"1\"\nsuite: {}\nscenarios: []")
	var out, errb bytes.Buffer
	if got := Main([]string{"manifest", p}, &out, &errb); got != ExitParse {
		t.Fatalf("exit = %d, want %d", got, ExitParse)
	}
}

func TestExplainCmd(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "ok.atago.yaml", passingSpec)
	var out, errb bytes.Buffer
	if got := Main([]string{"explain", p}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d (stderr=%s)", got, errb.String())
	}
	if !strings.Contains(out.String(), "Suite: sample") {
		t.Errorf("explain stdout missing suite:\n%s", out.String())
	}
}

// Issue #18: explain with no args defaults to "." for parity with run/doc,
// instead of erroring that a path is required. Run it in an empty dir so the
// default target resolves to no files rather than a required-path usage error.
func TestExplainCmd_NoArgsDefaultsToDot(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	var out, errb bytes.Buffer
	got := Main([]string{"explain"}, &out, &errb)
	if got != ExitConfig {
		t.Fatalf("exit = %d, want %d", got, ExitConfig)
	}
	if !strings.Contains(errb.String(), "no *.atago.yaml files found") {
		t.Errorf("explain with no args should default to \".\", got stderr: %q", errb.String())
	}
}

// Issue #18: explain must print usage on --help/-h (exit 0) instead of treating
// the flag as a file path. Explicitly-requested help goes to stdout so
// `atago explain --help | grep` works.
func TestExplainHelp(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		var out, errb bytes.Buffer
		got := Main([]string{"explain", flag}, &out, &errb)
		if got != ExitOK {
			t.Errorf("explain %s exit = %d, want %d (stderr=%s)", flag, got, ExitOK, errb.String())
		}
		if !strings.Contains(out.String(), "Usage: atago explain") {
			t.Errorf("explain %s missing usage on stdout, stdout=%q stderr=%q", flag, out.String(), errb.String())
		}
	}
}

// TestSubcommands_HelpToStdout proves an explicit --help/-h on the read
// subcommands writes usage to STDOUT (so it can be piped) and exits 0, matching
// completion/snapshot/top-level help. A genuine parse error still uses stderr.
func TestSubcommands_HelpToStdout(t *testing.T) {
	cmds := map[string]string{
		"explain":  "Usage: atago explain",
		"doc":      "Usage: atago doc",
		"list":     "Usage: atago list",
		"manifest": "Usage: atago manifest",
		"init":     "Usage: atago init",
	}
	for cmd, want := range cmds {
		for _, flag := range []string{"--help", "-h"} {
			var out, errb bytes.Buffer
			if got := Main([]string{cmd, flag}, &out, &errb); got != ExitOK {
				t.Errorf("%s %s exit = %d, want %d (stderr=%s)", cmd, flag, got, ExitOK, errb.String())
			}
			if !strings.Contains(out.String(), want) {
				t.Errorf("%s %s: stdout = %q, want usage line %q on stdout", cmd, flag, out.String(), want)
			}
		}
	}
}

func TestExplainCmd_ParseError(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "bad.atago.yaml", "version: \"9\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}")
	var out, errb bytes.Buffer
	if got := Main([]string{"explain", p}, &out, &errb); got != ExitParse {
		t.Fatalf("exit = %d, want %d", got, ExitParse)
	}
}
