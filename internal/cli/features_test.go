package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/loader"
)

// --- #62 completion --------------------------------------------------------

func TestCompletion_EachShellEmitsScript(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			var out, errb bytes.Buffer
			if got := Main([]string{"completion", shell}, &out, &errb); got != ExitOK {
				t.Fatalf("exit = %d (stderr=%s)", got, errb.String())
			}
			s := out.String()
			if s == "" {
				t.Fatal("empty completion script")
			}
			// Every generated script must mention atago and the run subcommand so a
			// change to the surface is visible in output.
			if !strings.Contains(s, "atago") {
				t.Errorf("%s script does not mention atago:\n%s", shell, s)
			}
			if !strings.Contains(s, "run") {
				t.Errorf("%s script does not mention the run subcommand", shell)
			}
		})
	}
}

func TestCompletion_UnknownShell(t *testing.T) {
	var out, errb bytes.Buffer
	if got := Main([]string{"completion", "tcsh"}, &out, &errb); got != ExitConfig {
		t.Fatalf("exit = %d, want %d", got, ExitConfig)
	}
	if !strings.Contains(errb.String(), "unknown shell") {
		t.Errorf("stderr = %q, want unknown shell error", errb.String())
	}
}

func TestCompletion_MissingArg(t *testing.T) {
	var out, errb bytes.Buffer
	if got := Main([]string{"completion"}, &out, &errb); got != ExitConfig {
		t.Fatalf("exit = %d, want %d", got, ExitConfig)
	}
}

// TestCompletion_Golden guards the deterministic completion output so adding or
// removing a subcommand/flag is an intentional, reviewable diff.
func TestCompletion_Golden(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		script, ok := completionScript(shell)
		if !ok {
			t.Fatalf("completionScript(%q) not ok", shell)
		}
		golden := filepath.Join("testdata", "completion", shell+".txt")
		if os.Getenv("UPDATE_GOLDEN") == "1" {
			if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(golden, []byte(script), 0o600); err != nil {
				t.Fatal(err)
			}
			continue
		}
		want, err := os.ReadFile(golden)
		if err != nil {
			t.Fatalf("read golden %s: %v (regenerate with UPDATE_GOLDEN=1)", golden, err)
		}
		if script != string(want) {
			t.Errorf("%s completion drifted from %s; regenerate with UPDATE_GOLDEN=1", shell, golden)
		}
	}
}

// --- #63 list --------------------------------------------------------------

const listSpec = `version: "1"
suite:
  name: listsuite
scenarios:
  - name: alpha scenario
    tags: [smoke, fast]
    steps:
      - run: {command: "true"}
      - assert: {exit_code: 0}
  - name: beta scenario
    skip: {os: windows}
    steps:
      - run: {shell: true, command: "echo hi > out.txt"}
      - assert: {file: {path: out.txt, exists: true}}
`

func TestListCmd_Table(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "s.atago.yaml", listSpec)
	var out, errb bytes.Buffer
	if got := Main([]string{"list", p}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d (stderr=%s)", got, errb.String())
	}
	s := out.String()
	for _, want := range []string{"SUITE", "SCENARIO", "listsuite", "alpha scenario", "beta scenario", "smoke", "skip:os=windows", "out.txt"} {
		if !strings.Contains(s, want) {
			t.Errorf("list table missing %q:\n%s", want, s)
		}
	}
}

func TestListCmd_JSON(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "s.atago.yaml", listSpec)
	var out, errb bytes.Buffer
	if got := Main([]string{"list", "--json", p}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d (stderr=%s)", got, errb.String())
	}
	var doc listDocument
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out.String())
	}
	if doc.SchemaVersion != ListSchemaVersion {
		t.Errorf("schema_version = %q, want %q", doc.SchemaVersion, ListSchemaVersion)
	}
	if len(doc.Scenarios) != 2 {
		t.Fatalf("scenarios = %d, want 2", len(doc.Scenarios))
	}
	if doc.Scenarios[0].Scenario != "alpha scenario" {
		t.Errorf("first scenario = %q", doc.Scenarios[0].Scenario)
	}
	if got := doc.Scenarios[0].Tags; len(got) != 2 || got[0] != "smoke" {
		t.Errorf("tags = %v", got)
	}
	if got := doc.Scenarios[1].Gates; len(got) != 1 || got[0] != "skip:os=windows" {
		t.Errorf("gates = %v", got)
	}
	if got := doc.Scenarios[1].Artifacts; len(got) == 0 {
		t.Errorf("beta scenario should report a generated artifact, got %v", got)
	}
}

func TestListCmd_Deterministic(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "s.atago.yaml", listSpec)
	run := func() string {
		var out, errb bytes.Buffer
		if got := Main([]string{"list", "--json", p}, &out, &errb); got != ExitOK {
			t.Fatalf("exit = %d", got)
		}
		return out.String()
	}
	first := run()
	second := run()
	if first != second {
		t.Error("list --json is not deterministic across runs")
	}
}

func TestListCmd_NoFiles(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if got := Main([]string{"list", dir}, &out, &errb); got != ExitConfig {
		t.Fatalf("exit = %d, want %d", got, ExitConfig)
	}
}

// --- #65 init templates ----------------------------------------------------

func TestInit_EveryTemplateIsSchemaValid(t *testing.T) {
	for _, name := range initTemplateNames() {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			outPath := filepath.Join(dir, "gen.atago.yaml")
			var out, errb bytes.Buffer
			if got := Main([]string{"init", "--template", name, outPath}, &out, &errb); got != ExitOK {
				t.Fatalf("init --template %s exit = %d (stderr=%s)", name, got, errb.String())
			}
			if _, err := loader.Load(outPath); err != nil {
				t.Fatalf("template %s does not load/validate: %v", name, err)
			}
		})
	}
}

func TestInit_UnknownTemplate(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if got := Main([]string{"init", "--template", "nope", filepath.Join(dir, "x.yaml")}, &out, &errb); got != ExitConfig {
		t.Fatalf("exit = %d, want %d", got, ExitConfig)
	}
}

func TestInit_ListTemplates(t *testing.T) {
	var out, errb bytes.Buffer
	if got := Main([]string{"init", "--list-templates"}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d", got)
	}
	for _, name := range []string{"cli", "http", "db", "grpc", "ssh", "browser", "services"} {
		if !strings.Contains(out.String(), name) {
			t.Errorf("--list-templates missing %q:\n%s", name, out.String())
		}
	}
	// Each template line carries a description, not just a bare name, so a
	// user can pick a template without generating and opening each one first.
	// The listing ends with a blank line and a how-to-scaffold footer.
	if !strings.Contains(out.String(), "Scaffold one with: atago init --template <name>") {
		t.Errorf("--list-templates output has no scaffold hint:\n%s", out.String())
	}
	for line := range strings.Lines(out.String()) {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "Scaffold one with:") {
			continue
		}
		name, desc, found := strings.Cut(line, " ")
		if !found || strings.TrimSpace(desc) == "" {
			t.Errorf("--list-templates line %q has no description", trimmed)
		}
		if _, ok := initTemplates[name]; !ok {
			t.Errorf("--list-templates line %q does not start with a template name", trimmed)
		}
	}
}

// runInitTemplate scaffolds the named template into a temp dir and returns the
// generated spec path.
func runInitTemplate(t *testing.T, name string) string {
	t.Helper()
	outPath := filepath.Join(t.TempDir(), name+".atago.yaml")
	var out, errb bytes.Buffer
	if got := Main([]string{"init", "--template", name, outPath}, &out, &errb); got != ExitOK {
		t.Fatalf("init --template %s exit = %d (stderr=%s)", name, got, errb.String())
	}
	return outPath
}

// TestInit_RunnableTemplatesRunGreen runs every template that advertises
// "runs as-is" in its description, so the first-run promise is enforced, not
// just schema validity (#65 follow-up).
func TestInit_RunnableTemplatesRunGreen(t *testing.T) {
	for name, tmpl := range initTemplates {
		if !strings.Contains(tmpl.desc, "runs as-is") {
			continue
		}
		t.Run(name, func(t *testing.T) {
			if runtime.GOOS == "windows" && name != "db" {
				t.Skip("template uses POSIX shell commands")
			}
			outPath := runInitTemplate(t, name)
			var out, errb bytes.Buffer
			if got := Main([]string{"run", outPath}, &out, &errb); got != ExitOK {
				t.Fatalf("run of scaffolded %s template exit = %d (stdout=%s stderr=%s)",
					name, got, out.String(), errb.String())
			}
		})
	}
}

func TestInit_DefaultCliRunnable(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "example.atago.yaml")
	var out, errb bytes.Buffer
	if got := Main([]string{"init", outPath}, &out, &errb); got != ExitOK {
		t.Fatalf("init exit = %d", got)
	}
	out.Reset()
	errb.Reset()
	if got := Main([]string{"run", outPath}, &out, &errb); got != ExitOK {
		t.Fatalf("run of scaffolded cli template exit = %d (stderr=%s)", got, errb.String())
	}
}

// --- #64 rerun-failed ------------------------------------------------------

const twoScenarioSpec = `version: "1"
suite:
  name: rerun
scenarios:
  - name: passes
    steps:
      - run: {shell: true, command: "exit 0"}
      - assert: {exit_code: 0}
  - name: fails
    steps:
      - run: {shell: true, command: "exit 1"}
      - assert: {exit_code: 0}
`

// withWorkdir runs fn with the process cwd set to dir, restoring it afterward.
// The rerun state file is written relative to cwd, so tests isolate it in a temp
// dir to avoid touching the repo.
func withWorkdir(t *testing.T, dir string, fn func()) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(orig); err != nil {
			t.Fatal(err)
		}
	}()
	fn()
}

func TestRerunFailed_RecordsAndReruns(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "s.atago.yaml", twoScenarioSpec)

	withWorkdir(t, dir, func() {
		// First full run: one scenario fails, so it is recorded.
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("first run exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		st, err := loadRerunState()
		if err != nil {
			t.Fatal(err)
		}
		if len(st.Failed) != 1 || st.Failed[0].Scenario != "fails" {
			t.Fatalf("recorded failures = %+v, want the 'fails' scenario", st.Failed)
		}

		// Rerun only the failed scenario: still fails, so it stays recorded.
		out.Reset()
		errb.Reset()
		if got := Main([]string{"run", "--rerun-failed", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("rerun exit = %d, want %d", got, ExitFailures)
		}
		// The report should mention only the previously-failed scenario, not the
		// passing one.
		if strings.Contains(out.String(), "passes") {
			t.Errorf("rerun ran the passing scenario:\n%s", out.String())
		}
	})
}

// TestRerunFailed_NoMatchKeepsStateAndWarns is a regression: when the recorded
// failing scenario no longer exists in the spec (renamed/removed while still
// broken), --rerun-failed must NOT report a false green and must NOT clear the
// recorded failures — otherwise the still-failing work is silently forgotten.
func TestRerunFailed_NoMatchKeepsStateAndWarns(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "s.atago.yaml", twoScenarioSpec)
	withWorkdir(t, dir, func() {
		// First run records the failing "fails" scenario.
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("first run exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		// Rename the failing scenario so the recorded name no longer matches, but
		// keep the spec path (and keep it broken).
		renamed := strings.ReplaceAll(twoScenarioSpec, "name: fails", "name: fails-renamed")
		writeSpec(t, dir, "s.atago.yaml", renamed)

		out.Reset()
		errb.Reset()
		got := Main([]string{"run", "--rerun-failed", "."}, &out, &errb)
		if got == ExitOK {
			t.Errorf("exit = %d (ExitOK); a rerun that matched no recorded failure must not greenlight", got)
		}
		if !strings.Contains(errb.String(), "no recorded failing scenarios matched") {
			t.Errorf("stderr = %q, want the no-match warning", errb.String())
		}
		// The recorded failures must survive so a later, correct rerun can find them.
		st, err := loadRerunState()
		if err != nil {
			t.Fatalf("rerun state was removed or unreadable: %v", err)
		}
		if len(st.Failed) != 1 || st.Failed[0].Scenario != "fails" {
			t.Errorf("recorded failures = %+v, want the original 'fails' preserved", st.Failed)
		}
	})
}

func TestRerunFailed_NothingRecorded(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "s.atago.yaml", passingSpec)
	withWorkdir(t, dir, func() {
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "--rerun-failed", "."}, &out, &errb); got != ExitOK {
			t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitOK, errb.String())
		}
		if !strings.Contains(errb.String(), "nothing to rerun") {
			t.Errorf("stderr = %q, want nothing-to-rerun note", errb.String())
		}
	})
}

func TestRerunFailed_GreenRunClearsState(t *testing.T) {
	dir := t.TempDir()
	withWorkdir(t, dir, func() {
		// Seed a state file, then a fully-green run should remove it.
		if err := saveRerunState([]failedEntry{{SpecPath: "x.atago.yaml", Scenario: "old"}}); err != nil {
			t.Fatal(err)
		}
		writeSpec(t, dir, "ok.atago.yaml", passingSpec)
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "ok.atago.yaml"}, &out, &errb); got != ExitOK {
			t.Fatalf("exit = %d (stderr=%s)", got, errb.String())
		}
		if _, err := os.Stat(rerunStatePath()); !os.IsNotExist(err) {
			t.Errorf("green run did not clear rerun state file (err=%v)", err)
		}
	})
}

// TestCompletion_HelpFlag proves --help behaves like every other subcommand's
// --help (usage on stdout, exit 0) instead of being mistaken for a shell name.
func TestCompletion_HelpFlag(t *testing.T) {
	for _, flag := range []string{"-h", "-help", "--help"} {
		var out, errb bytes.Buffer
		if got := Main([]string{"completion", flag}, &out, &errb); got != ExitOK {
			t.Errorf("completion %s: exit = %d, want %d", flag, got, ExitOK)
		}
		if !strings.Contains(out.String(), "Usage: atago completion") {
			t.Errorf("completion %s: stdout = %q, want a usage line", flag, out.String())
		}
	}
}

// TestSnapshot_HelpFlag proves `atago snapshot --help` prints usage and exits 0
// rather than reporting a bad invocation.
func TestSnapshot_HelpFlag(t *testing.T) {
	for _, flag := range []string{"-h", "-help", "--help"} {
		var out, errb bytes.Buffer
		if got := Main([]string{"snapshot", flag}, &out, &errb); got != ExitOK {
			t.Errorf("snapshot %s: exit = %d, want %d", flag, got, ExitOK)
		}
		if !strings.Contains(out.String(), "Usage: atago snapshot update") {
			t.Errorf("snapshot %s: stdout = %q, want a usage line", flag, out.String())
		}
	}
}
