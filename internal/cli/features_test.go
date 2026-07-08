package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/engine"
	"github.com/nao1215/atago/internal/loader"
)

// TestCollectFailures_ClassifiesStatuses pins which scenario outcomes are
// recorded for --rerun-failed: only failed and errored scenarios are, in a
// deterministic order. A flaky scenario (one that recovered on retry, #29) is
// green and must NOT be recorded, or the red-green loop would keep re-running a
// scenario that already passes; passed and skipped are likewise excluded.
func TestCollectFailures_ClassifiesStatuses(t *testing.T) {
	t.Parallel()
	results := []*engine.SuiteResult{
		{SpecPath: "b.atago.yaml", Scenarios: []engine.ScenarioResult{
			{Name: "passes", Status: engine.StatusPassed},
			{Name: "errs", Status: engine.StatusError},
			{Name: "flakes", Status: engine.StatusFlaky},
		}},
		{SpecPath: "a.atago.yaml", Scenarios: []engine.ScenarioResult{
			{Name: "fails", Status: engine.StatusFailed},
			{Name: "skipped", Status: engine.StatusSkipped},
		}},
		nil, // a nil suite result (e.g. a spec that failed to load) is skipped
	}
	got := collectFailures(results)

	want := []failedEntry{
		{SpecPath: "b.atago.yaml", Scenario: "errs"},
		{SpecPath: "a.atago.yaml", Scenario: "fails"},
	}
	if len(got) != len(want) {
		t.Fatalf("collectFailures = %+v, want only the failed/errored entries %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d = %+v, want %+v (order follows results, statuses filtered)", i, got[i], want[i])
		}
	}
	for _, e := range got {
		if e.Scenario == "flakes" || e.Scenario == "passes" || e.Scenario == "skipped" {
			t.Errorf("collectFailures recorded a non-failing scenario %q; the rerun loop would never converge", e.Scenario)
		}
	}
}

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

// TestInit_EmitsResolvableSchemaHeader proves that init writes a resolvable
// `# yaml-language-server: $schema=<url>` header as the first line of the
// generated spec, so a scaffolded spec gets editor completion for the DSL out
// of the box, and that the URL is absolute (not the old repo-relative
// `./schema/...` path that only resolves inside the atago repo) (#121).
func TestInit_EmitsResolvableSchemaHeader(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.atago.yaml")
	var out, errb bytes.Buffer
	if got := Main([]string{"init", "--template", "cli", outPath}, &out, &errb); got != ExitOK {
		t.Fatalf("init exit = %d (stderr=%s)", got, errb.String())
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	firstLine, _, _ := strings.Cut(string(data), "\n")
	if !strings.HasPrefix(firstLine, "# yaml-language-server: $schema=") {
		t.Errorf("first line = %q, want a yaml-language-server schema header", firstLine)
	}
	if !strings.Contains(firstLine, "https://") {
		t.Errorf("schema URL must be absolute, got %q", firstLine)
	}
	if strings.Contains(firstLine, "./schema/") {
		t.Errorf("schema URL must not be the repo-relative ./schema path: %q", firstLine)
	}
	// The header is an ignored YAML comment, so the spec still loads and runs.
	if _, err := loader.Load(outPath); err != nil {
		t.Fatalf("header-carrying spec does not load: %v", err)
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
		// A spec that fails, records its failure, then is fixed and re-run green:
		// re-running THE SAME spec re-decides its scenario and clears the ledger.
		writeSpec(t, dir, "s.atago.yaml", singleFailSpec("s", false))
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "s.atago.yaml"}, &out, &errb); got != ExitFailures {
			t.Fatalf("failing run exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		writeSpec(t, dir, "s.atago.yaml", singleFailSpec("s", true))
		out.Reset()
		errb.Reset()
		if got := Main([]string{"run", "s.atago.yaml"}, &out, &errb); got != ExitOK {
			t.Fatalf("fixed run exit = %d (stderr=%s)", got, errb.String())
		}
		if _, err := os.Stat(rerunStatePath()); !os.IsNotExist(err) {
			t.Errorf("re-running the fixed spec did not clear the rerun state file (err=%v)", err)
		}
	})
}

// TestRun_UnrelatedGreenRunPreservesRecordedFailures is a regression: a green run
// that does not execute a recorded failure — running an unrelated spec, or a
// --filter that excludes the failing scenario — must not clear that failure from
// the ledger. Overwriting the ledger with only what ran forgot still-failing work
// and let the next --rerun-failed exit 0 while the failure was still real. The
// preserve rule the narrowed-rerun path always used now applies to every run.
func TestRun_UnrelatedGreenRunPreservesRecordedFailures(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "fail.atago.yaml", singleFailSpec("f", false))
	writeSpec(t, dir, "ok.atago.yaml", passingSpec)

	withWorkdir(t, dir, func() {
		var out, errb bytes.Buffer
		// Record f_fail.
		if got := Main([]string{"run", "fail.atago.yaml"}, &out, &errb); got != ExitFailures {
			t.Fatalf("first run exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		// Run an unrelated green spec: it must not touch fail.atago.yaml's record.
		out.Reset()
		errb.Reset()
		if got := Main([]string{"run", "ok.atago.yaml"}, &out, &errb); got != ExitOK {
			t.Fatalf("unrelated green run exit = %d (stderr=%s)", got, errb.String())
		}
		st, err := loadRerunState()
		if err != nil {
			t.Fatal(err)
		}
		if len(st.Failed) != 1 || st.Failed[0].Scenario != "f_fail" {
			t.Fatalf("ledger = %+v, want f_fail preserved after an unrelated green run", st.Failed)
		}
		// The still-real failure is therefore still caught by --rerun-failed.
		out.Reset()
		errb.Reset()
		if got := Main([]string{"run", "--rerun-failed", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("--rerun-failed exit = %d, want %d (the preserved failure must still be caught); stderr=%s", got, ExitFailures, errb.String())
		}
	})
}

// singleFailSpec renders a one-scenario spec named <name> whose only scenario
// <name>_fail asserts exit_code 0. When passes is false the command exits 1 (the
// assertion fails); when true it exits 0 (passes). Used to build multi-spec
// rerun-ledger fixtures.
func singleFailSpec(name string, passes bool) string {
	code := "1"
	if passes {
		code = "0"
	}
	return `version: "1"
suite:
  name: ` + name + `
scenarios:
  - name: ` + name + `_fail
    steps:
      - run: {shell: true, command: "exit ` + code + `"}
      - assert: {exit_code: 0}
`
}

// TestRerunFailed_NarrowedTargetPreservesOtherFailures is a regression: a
// `--rerun-failed` narrowed to a subset of the recorded specs must not drop the
// recorded failures in the specs it did not run. Rewriting the whole ledger from
// only the narrowed subset forgot still-failing work elsewhere — a red-green
// loop that silently loses a broken scenario the moment you rerun a single spec.
func TestRerunFailed_NarrowedTargetPreservesOtherFailures(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "a.atago.yaml", singleFailSpec("a", false))
	writeSpec(t, dir, "b.atago.yaml", singleFailSpec("b", false))

	withWorkdir(t, dir, func() {
		var out, errb bytes.Buffer
		// A full run records both a_fail and b_fail.
		if got := Main([]string{"run", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("first run exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		st, err := loadRerunState()
		if err != nil {
			t.Fatal(err)
		}
		if len(st.Failed) != 2 {
			t.Fatalf("recorded failures = %+v, want both a_fail and b_fail", st.Failed)
		}

		// Rerun only a (still failing). b was not re-verified and is still broken,
		// so it must survive in the ledger alongside the freshly-recorded a_fail.
		out.Reset()
		errb.Reset()
		if got := Main([]string{"run", "--rerun-failed", "a.atago.yaml"}, &out, &errb); got != ExitFailures {
			t.Fatalf("narrowed rerun exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		st, err = loadRerunState()
		if err != nil {
			t.Fatal(err)
		}
		names := map[string]bool{}
		for _, e := range st.Failed {
			names[e.Scenario] = true
		}
		if !names["a_fail"] || !names["b_fail"] {
			t.Errorf("ledger after narrowed rerun = %+v, want both a_fail (re-verified) and b_fail (preserved)", st.Failed)
		}
	})
}

// TestRerunFailed_NarrowedGreenKeepsOtherFailures is the greenlight variant: a
// narrowed `--rerun-failed` whose target now passes must not wipe the ledger
// and exit green while another recorded spec is still broken. Only the specs it
// actually re-ran may be cleared.
func TestRerunFailed_NarrowedGreenKeepsOtherFailures(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "a.atago.yaml", singleFailSpec("a", false))
	writeSpec(t, dir, "b.atago.yaml", singleFailSpec("b", false))

	withWorkdir(t, dir, func() {
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("first run exit = %d (stderr=%s)", got, errb.String())
		}
		// Fix a so its narrowed rerun passes; leave b broken and un-run.
		writeSpec(t, dir, "a.atago.yaml", singleFailSpec("a", true))

		out.Reset()
		errb.Reset()
		if got := Main([]string{"run", "--rerun-failed", "a.atago.yaml"}, &out, &errb); got != ExitOK {
			t.Fatalf("narrowed green rerun exit = %d, want %d (stderr=%s)", got, ExitOK, errb.String())
		}
		st, err := loadRerunState()
		if err != nil {
			t.Fatal(err)
		}
		if len(st.Failed) != 1 || st.Failed[0].Scenario != "b_fail" {
			t.Errorf("ledger = %+v, want only b_fail preserved (a cleared, b kept)", st.Failed)
		}
	})
}

// twoFailingSpec has two failing scenarios in ONE spec file, so a --filter can
// exclude one while the other reruns — the shape that exposes a filtered
// --rerun-failed silently dropping the excluded (still-failing) scenario.
const twoFailingSpec = `version: "1"
suite:
  name: multi
scenarios:
  - name: alpha_fail
    steps:
      - run: {shell: true, command: "exit 1"}
      - assert: {exit_code: 0}
  - name: beta_fail
    steps:
      - run: {shell: true, command: "exit 1"}
      - assert: {exit_code: 0}
`

// TestRerunFailed_FilterExcludedFailurePreserved is a regression: a
// `--rerun-failed --filter` that excludes a recorded failure must not drop it
// from the ledger. Rewriting the ledger from only the scenarios that ran forgot
// the filter-excluded failure — a false green the next time the filter is gone.
func TestRerunFailed_FilterExcludedFailurePreserved(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "s.atago.yaml", twoFailingSpec)

	withWorkdir(t, dir, func() {
		var out, errb bytes.Buffer
		// Full run records both alpha_fail and beta_fail.
		if got := Main([]string{"run", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("first run exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		st, err := loadRerunState()
		if err != nil {
			t.Fatal(err)
		}
		if len(st.Failed) != 2 {
			t.Fatalf("recorded failures = %+v, want both alpha_fail and beta_fail", st.Failed)
		}

		// Rerun only alpha (still failing). beta was excluded by the filter, not
		// re-verified, and is still broken — it must survive in the ledger.
		out.Reset()
		errb.Reset()
		if got := Main([]string{"run", "--rerun-failed", "--filter", "alpha", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("filtered rerun exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		st, err = loadRerunState()
		if err != nil {
			t.Fatal(err)
		}
		names := map[string]bool{}
		for _, e := range st.Failed {
			names[e.Scenario] = true
		}
		if !names["beta_fail"] {
			t.Errorf("ledger after filtered rerun = %+v, want beta_fail preserved (excluded by --filter, never re-verified)", st.Failed)
		}
	})
}

// TestRerunFailed_FilterExcludesAllNoRenamedWarning is a regression: when a
// user's own --filter excludes every recorded failure, the run must not blame a
// rename/removal — that diagnostic is wrong and contradicts the filter warning.
// The recorded failures must also survive so a later unfiltered rerun finds them.
func TestRerunFailed_FilterExcludesAllNoRenamedWarning(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "s.atago.yaml", twoFailingSpec)

	withWorkdir(t, dir, func() {
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("first run exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}

		// A filter that matches no recorded scenario: nothing runs because of the
		// filter, not because anything was renamed or removed.
		out.Reset()
		errb.Reset()
		Main([]string{"run", "--rerun-failed", "--filter", "no-such-scenario", "."}, &out, &errb)
		if strings.Contains(errb.String(), "renamed or removed") {
			t.Errorf("stderr claimed a rename/removal when the user's --filter excluded everything:\n%s", errb.String())
		}
		// The recorded failures must not be silently forgotten.
		st, err := loadRerunState()
		if err != nil {
			t.Fatalf("rerun state unreadable: %v", err)
		}
		if len(st.Failed) != 2 {
			t.Errorf("ledger = %+v, want both recorded failures preserved when a filter excludes them all", st.Failed)
		}
	})
}

// TestRerunFailed_AbsolutePathMatchesRelativeLedger is a regression: a rerun
// addressed by an absolute path must match a recorded relative spec_path (and
// vice versa). Comparing raw path strings meant an equivalent-but-differently
// spelled target found "nothing" and greenlit despite real failures.
func TestRerunFailed_AbsolutePathMatchesRelativeLedger(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "s.atago.yaml", twoScenarioSpec)

	withWorkdir(t, dir, func() {
		var out, errb bytes.Buffer
		// Record a RELATIVE spec_path by running with a relative target.
		if got := Main([]string{"run", "s.atago.yaml"}, &out, &errb); got != ExitFailures {
			t.Fatalf("first run exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		st, err := loadRerunState()
		if err != nil {
			t.Fatal(err)
		}
		if len(st.Failed) != 1 || filepath.IsAbs(st.Failed[0].SpecPath) {
			t.Fatalf("recorded state = %+v, want one relative-path failure", st.Failed)
		}

		// Rerun with the ABSOLUTE spelling of the same spec: the recorded failure
		// must still be selected and re-run (exit fails), not treated as "nothing".
		out.Reset()
		errb.Reset()
		abs := filepath.Join(dir, "s.atago.yaml")
		if got := Main([]string{"run", "--rerun-failed", abs}, &out, &errb); got != ExitFailures {
			t.Fatalf("absolute-path rerun exit = %d, want %d (recorded failure was not selected); stderr=%s", got, ExitFailures, errb.String())
		}
	})
}

// TestRun_FilteredGreenRunPreservesExcludedFailure proves a plain run whose
// --filter excludes a recorded failing scenario does not clear that failure, the
// same way a narrowed --rerun-failed preserves it. Only scenarios that actually
// ran are re-decided.
func TestRun_FilteredGreenRunPreservesExcludedFailure(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "m.atago.yaml", twoFailingSpec)

	withWorkdir(t, dir, func() {
		var out, errb bytes.Buffer
		// Record alpha_fail and beta_fail.
		if got := Main([]string{"run", "m.atago.yaml"}, &out, &errb); got != ExitFailures {
			t.Fatalf("first run exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		// Fix alpha, then run with a --filter that only touches alpha: beta_fail was
		// not re-run and is still broken, so it must survive.
		writeSpec(t, dir, "m.atago.yaml", `version: "1"
suite:
  name: multi
scenarios:
  - name: alpha_fail
    steps:
      - run: {shell: true, command: "exit 0"}
      - assert: {exit_code: 0}
  - name: beta_fail
    steps:
      - run: {shell: true, command: "exit 1"}
      - assert: {exit_code: 0}
`)
		out.Reset()
		errb.Reset()
		if got := Main([]string{"run", "--filter", "alpha_fail", "m.atago.yaml"}, &out, &errb); got != ExitOK {
			t.Fatalf("filtered run exit = %d (stderr=%s)", got, errb.String())
		}
		st, err := loadRerunState()
		if err != nil {
			t.Fatal(err)
		}
		if len(st.Failed) != 1 || st.Failed[0].Scenario != "beta_fail" {
			t.Errorf("ledger = %+v, want only beta_fail (alpha re-verified green, beta preserved)", st.Failed)
		}
	})
}

// TestRerunFailed_LedgerStaysRelativeAfterRerun proves a --rerun-failed does not
// rewrite the ledger's spec paths to absolute. Persisting the absolute form (used
// only in memory to match across spellings) made the next rerun after the project
// moved find nothing and silently greenlight still-failing work.
func TestRerunFailed_LedgerStaysRelativeAfterRerun(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "s.atago.yaml", twoScenarioSpec)

	withWorkdir(t, dir, func() {
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "s.atago.yaml"}, &out, &errb); got != ExitFailures {
			t.Fatalf("first run exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		// Rerun once (still failing): the ledger must keep the relative spelling.
		out.Reset()
		errb.Reset()
		if got := Main([]string{"run", "--rerun-failed", "s.atago.yaml"}, &out, &errb); got != ExitFailures {
			t.Fatalf("rerun exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		st, err := loadRerunState()
		if err != nil {
			t.Fatal(err)
		}
		if len(st.Failed) == 0 {
			t.Fatal("ledger empty after a still-failing rerun")
		}
		for _, e := range st.Failed {
			if filepath.IsAbs(e.SpecPath) {
				t.Errorf("ledger stored an absolute spec_path %q; it must stay portable", e.SpecPath)
			}
		}
	})
}

// TestLoadRerunState_UnknownSchemaVersion proves a state file written by a future
// atago (a schema_version this build does not understand) is rejected rather than
// read under v1 assumptions, which could silently drop recorded failures.
func TestLoadRerunState_UnknownSchemaVersion(t *testing.T) {
	dir := t.TempDir()
	withWorkdir(t, dir, func() {
		if err := os.MkdirAll(rerunStateDir, 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(rerunStatePath(), []byte(`{"schema_version":"999","failed":[]}`), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := loadRerunState(); err == nil {
			t.Error("loadRerunState accepted an unknown schema_version; want an error")
		}
	})
}

// TestRunCmd_ArtifactsDirNotADirectory proves --artifacts-dir pointing at an
// existing regular file is a clean config error, not a run that silently writes
// no artifacts and leaves the user believing there were no failures to review.
func TestRunCmd_ArtifactsDirNotADirectory(t *testing.T) {
	dir := t.TempDir()
	spec := writeSpec(t, dir, "fail.atago.yaml", singleFailSpec("f", false))
	afile := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(afile, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if got := Main([]string{"run", "--artifacts-dir", afile, spec}, &out, &errb); got != ExitConfig {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitConfig, errb.String())
	}
	if !strings.Contains(errb.String(), "not usable") {
		t.Errorf("stderr should explain the unusable artifacts dir, got: %s", errb.String())
	}
}

// TestRunCmd_NegativeParallelRejected proves a negative --parallel is a config
// error, matching --repeat/--retry-failed, rather than being silently coerced to
// sequential and exiting 0.
func TestRunCmd_NegativeParallelRejected(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "ok.atago.yaml", passingSpec)

	var out, errb bytes.Buffer
	if got := Main([]string{"run", "--parallel", "-1", p}, &out, &errb); got != ExitConfig {
		t.Errorf("--parallel -1 exit = %d, want %d (stderr=%s)", got, ExitConfig, errb.String())
	}
	// A valid positive value still runs.
	out.Reset()
	errb.Reset()
	if got := Main([]string{"run", "--parallel", "2", p}, &out, &errb); got != ExitOK {
		t.Errorf("--parallel 2 exit = %d, want %d (stderr=%s)", got, ExitOK, errb.String())
	}
}

// TestRunCmd_Repeat1WithRetryFailedAccepted proves the mutual-exclusion guard
// fires only for an ACTIVE --repeat (> 1): --repeat 1 is a documented no-op and
// must not be rejected alongside --retry-failed, while --repeat 2 still is.
func TestRunCmd_Repeat1WithRetryFailedAccepted(t *testing.T) {
	dir := t.TempDir()
	p := writeSpec(t, dir, "ok.atago.yaml", passingSpec)

	var out, errb bytes.Buffer
	if got := Main([]string{"run", "--repeat", "1", "--retry-failed", "3", p}, &out, &errb); got == ExitConfig {
		t.Errorf("--repeat 1 --retry-failed 3 was rejected (exit %d); repeat 1 is a no-op (stderr=%s)", got, errb.String())
	}
	// An active --repeat (> 1) with --retry-failed is still mutually exclusive.
	out.Reset()
	errb.Reset()
	if got := Main([]string{"run", "--repeat", "2", "--retry-failed", "1", p}, &out, &errb); got != ExitConfig {
		t.Errorf("--repeat 2 --retry-failed 1 exit = %d, want %d (must stay mutually exclusive)", got, ExitConfig)
	}
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

// tagSelectSpec has one smoke-tagged scenario that FAILS and one slow-tagged
// scenario that passes, so tag selection is observable through the exit code.
const tagSelectSpec = `version: "1"
suite:
  name: tagged
scenarios:
  - name: alpha
    tags: [smoke]
    steps:
      - run: {shell: true, command: "exit 1"}
      - assert: {exit_code: 0}
  - name: beta
    tags: [slow]
    steps:
      - run: {command: "true"}
      - assert: {exit_code: 0}
`

// TestRunTags_RepeatableFlagOrSemantics is a regression: --tag and --skip-tag
// must be repeatable and OR their values, like --filter (#119). The old
// single-string flags kept only the last occurrence, silently dropping earlier
// selections.
func TestRunTags_RepeatableFlagOrSemantics(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "s.atago.yaml", tagSelectSpec)
	withWorkdir(t, dir, func() {
		// --tag smoke --tag slow selects BOTH; alpha (smoke) fails, so the run
		// fails. Last-flag-wins would have kept only slow (beta) and gone green.
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "--tag", "smoke", "--tag", "slow", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("--tag OR exit = %d, want %d; a repeated --tag dropped the smoke selection\n%s", got, ExitFailures, out.String())
		}
		// --skip-tag smoke --skip-tag slow skips BOTH, so the failing alpha never
		// runs and the run is green. Last-flag-wins would skip only slow, run alpha,
		// and fail.
		out.Reset()
		errb.Reset()
		if got := Main([]string{"run", "--skip-tag", "smoke", "--skip-tag", "slow", "."}, &out, &errb); got != ExitOK {
			t.Fatalf("--skip-tag OR exit = %d, want %d; a repeated --skip-tag dropped the smoke skip\n%s", got, ExitOK, out.String())
		}
	})
}

// TestFailFast_StopsSubsequentSpecFiles is a regression: --fail-fast must stop
// scheduling across spec files, not only within one suite. The first spec fails;
// the second (which would pass) must never run.
func TestFailFast_StopsSubsequentSpecFiles(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "1first.atago.yaml", singleFailSpec("firstsuite", false))
	writeSpec(t, dir, "2second.atago.yaml", singleFailSpec("secondsuite", true))
	withWorkdir(t, dir, func() {
		var out, errb bytes.Buffer
		got := Main([]string{"run", "--fail-fast", "--parallel", "1", "--report", "json", "."}, &out, &errb)
		if got != ExitFailures {
			t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		if strings.Contains(out.String(), "secondsuite") {
			t.Errorf("--fail-fast scheduled a spec after the first failure:\n%s", out.String())
		}
	})
}
