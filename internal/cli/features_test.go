package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/engine"
	"github.com/nao1215/atago/internal/loader"
)

// TestCollectFailures_ClassifiesStatuses pins which scenario outcomes are
// recorded for --rerun-failed, in a deterministic order: failed and errored
// scenarios, plus an XPASS, which is what turns an otherwise green run red and
// would otherwise leave `--rerun-failed` with nothing to do after a red run. A
// flaky scenario (one that recovered on retry, #29) stays out even though it
// too fails the run: its rerun most likely passes, which would clear the ledger
// without anything being fixed. Passed, skipped, and XFAIL are green and
// excluded.
func TestCollectFailures_ClassifiesStatuses(t *testing.T) {
	t.Parallel()
	results := []*engine.SuiteResult{
		{SpecPath: "b.atago.yaml", Scenarios: []engine.ScenarioResult{
			{Name: "passes", Status: engine.StatusPassed},
			{Name: "errs", Status: engine.StatusError},
			{Name: "flakes", Status: engine.StatusFlaky},
			{Name: "xfails", Status: engine.StatusXFail},
		}},
		{SpecPath: "a.atago.yaml", Scenarios: []engine.ScenarioResult{
			{Name: "fails", Status: engine.StatusFailed},
			{Name: "skipped", Status: engine.StatusSkipped},
			{Name: "xpasses", Status: engine.StatusXPass},
		}},
		nil, // a nil suite result (e.g. a spec that failed to load) is skipped
	}
	got := collectFailures(results, false)

	want := []failedEntry{
		{SpecPath: "b.atago.yaml", Scenario: "errs"},
		{SpecPath: "a.atago.yaml", Scenario: "fails"},
		{SpecPath: "a.atago.yaml", Scenario: "xpasses"},
	}
	if len(got) != len(want) {
		t.Fatalf("collectFailures = %+v, want the failed/errored/xpass entries %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d = %+v, want %+v (order follows results, statuses filtered)", i, got[i], want[i])
		}
	}
	for _, e := range got {
		switch e.Scenario {
		case "flakes", "passes", "skipped", "xfails":
			t.Errorf("collectFailures recorded a non-failing scenario %q; the rerun loop would never converge", e.Scenario)
		}
	}

	// --allow-xpass makes the XPASS green, and a green run must leave nothing
	// behind to rerun.
	allowed := collectFailures(results, true)
	for _, e := range allowed {
		if e.Scenario == "xpasses" {
			t.Error("collectFailures recorded an XPASS the run was told to allow")
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
// TestCompletion_RunFlagsMatchTheCommand is a regression: runFlags was a
// hand-kept copy of the run command's flag set — its own comment said "keep it
// in sync" — and three flags added later never reached it. `atago run
// --allow-<TAB>` offered nothing, and --profile was invisible to every shell.
// The command's own usage is the source of truth here, so the next flag that
// skips the list fails this test rather than silently going missing.
func TestCompletion_RunFlagsMatchTheCommand(t *testing.T) {
	t.Parallel()
	var out, errb bytes.Buffer
	Main([]string{"run", "--help"}, &out, &errb)

	real := map[string]bool{}
	for _, line := range strings.Split(out.String()+errb.String(), "\n") {
		rest, ok := strings.CutPrefix(line, "  -")
		if !ok {
			continue
		}
		name, _, _ := strings.Cut(rest, " ")
		if name != "" {
			real["--"+name] = true
		}
	}
	if len(real) == 0 {
		t.Fatalf("no flags parsed out of run --help:\n%s%s", out.String(), errb.String())
	}

	offered := map[string]bool{}
	for _, f := range runFlags {
		offered[f] = true
	}
	for f := range real {
		if !offered[f] {
			t.Errorf("`atago run` accepts %s but shell completion does not offer it", f)
		}
	}
	for f := range offered {
		if !real[f] {
			t.Errorf("shell completion offers %s but `atago run` does not accept it", f)
		}
	}

	// The subcommand list is the same kind of hand-kept copy, held to the same
	// rule before it drifts the same way.
	out.Reset()
	errb.Reset()
	Main([]string{"help"}, &out, &errb)
	help := out.String() + errb.String()
	for _, name := range subcommandNames {
		if !strings.Contains(help, "\n  "+name+" ") {
			t.Errorf("shell completion offers the %q subcommand but `atago help` does not list it", name)
		}
	}
}

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

// TestListCmd_ExpectFail is a regression: a scenario documenting a known bug
// was indistinguishable from a healthy one in `atago list`. explain prints the
// marker precisely so a reviewer sees which scenarios are documentation of a
// bug rather than guarantees, doc renders it, and the manifest carries the
// block — the inventory was the one surface where a suite silently read as
// promising more than it does.
func TestListCmd_ExpectFail(t *testing.T) {
	const src = `version: "1"
suite:
  name: xfail
scenarios:
  - name: known bug
    expect_fail:
      reason: "upstream renders the wrong width"
      issue: "https://example.com/issues/42"
    steps:
      - run: {shell: true, command: "exit 1"}
      - assert: {exit_code: 0}
  - name: healthy
    steps:
      - run: {shell: true, command: "true"}
      - assert: {exit_code: 0}
`
	dir := t.TempDir()
	p := writeSpec(t, dir, "s.atago.yaml", src)

	var out, errb bytes.Buffer
	if got := Main([]string{"list", p}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d (stderr=%s)", got, errb.String())
	}
	table := out.String()
	if !strings.Contains(table, "XFAIL") {
		t.Errorf("list table does not mark the expect_fail scenario:\n%s", table)
	}

	out.Reset()
	errb.Reset()
	if got := Main([]string{"list", "--json", p}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d (stderr=%s)", got, errb.String())
	}
	var doc listDocument
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out.String())
	}
	ef := doc.Scenarios[0].ExpectFail
	if ef == nil || ef.Reason != "upstream renders the wrong width" || ef.Issue != "https://example.com/issues/42" {
		t.Errorf("expect_fail = %+v, want the reason and issue", ef)
	}
	if doc.Scenarios[1].ExpectFail != nil {
		t.Errorf("a healthy scenario carries expect_fail: %+v", doc.Scenarios[1].ExpectFail)
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

// TestRerunFailed_PartialMatchWarnsAboutTheRest covers the quiet half of the
// case above: when only SOME recorded failures still exist, the rerun reports
// fewer scenarios than were recorded, which reads as "the others are fixed".
// The unmatched entries survive in the ledger, so the run must say which ones
// it did not verify.
func TestRerunFailed_PartialMatchWarnsAboutTheRest(t *testing.T) {
	const spec = `version: "1"
suite:
  name: rerun
scenarios:
  - name: fails
    steps:
      - run: {shell: true, command: "exit 1"}
      - assert: {exit_code: 0}
  - name: also-fails
    steps:
      - run: {shell: true, command: "exit 1"}
      - assert: {exit_code: 0}
`
	dir := t.TempDir()
	writeSpec(t, dir, "s.atago.yaml", spec)
	withWorkdir(t, dir, func() {
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("first run exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		// One of the two recorded failures is renamed while still broken.
		writeSpec(t, dir, "s.atago.yaml", strings.ReplaceAll(spec, "name: fails\n", "name: fails-renamed\n"))

		out.Reset()
		errb.Reset()
		if got := Main([]string{"run", "--rerun-failed", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("rerun exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		if !strings.Contains(errb.String(), "1 recorded failing scenario did not match") {
			t.Errorf("stderr = %q, want a warning naming the unmatched recorded failure", errb.String())
		}
		if !strings.Contains(errb.String(), "s.atago.yaml / fails") {
			t.Errorf("stderr = %q, want the unmatched scenario named", errb.String())
		}
		// The unmatched entry is still recorded for the next rerun.
		st, err := loadRerunState()
		if err != nil {
			t.Fatalf("rerun state unreadable: %v", err)
		}
		var names []string
		for _, e := range st.Failed {
			names = append(names, e.Scenario)
		}
		sort.Strings(names)
		want := []string{"also-fails", "fails"}
		if !reflect.DeepEqual(names, want) {
			t.Errorf("recorded failures = %v, want %v", names, want)
		}
	})
}

// TestRerunFailed_OutOfTargetEntriesAreNotBlamedOnARename pins the difference
// between "the scenario is gone" and "you asked for a different spec". A rerun
// narrowed to one spec leaves the other spec's recorded failures unexecuted, and
// blaming a rename sends the reader after a spec change that never happened —
// the entry is still there, and so is its spec.
func TestRerunFailed_OutOfTargetEntriesAreNotBlamedOnARename(t *testing.T) {
	const failing = `version: "1"
suite:
  name: rerun
scenarios:
  - name: fails
    steps:
      - run: {shell: true, command: "exit 1"}
      - assert: {exit_code: 0}
`
	dir := t.TempDir()
	writeSpec(t, dir, "a.atago.yaml", failing)
	writeSpec(t, dir, "b.atago.yaml", failing)
	withWorkdir(t, dir, func() {
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("first run exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		out.Reset()
		errb.Reset()
		if got := Main([]string{"run", "--rerun-failed", "a.atago.yaml"}, &out, &errb); got != ExitFailures {
			t.Fatalf("rerun exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		if strings.Contains(errb.String(), "did not match the current specs") {
			t.Errorf("stderr = %q, want no rename/removal blame for a spec that was simply not targeted", errb.String())
		}
		if !strings.Contains(errb.String(), "outside this run's targets") {
			t.Errorf("stderr = %q, want the untouched recorded failure reported as out of scope", errb.String())
		}
		if !strings.Contains(errb.String(), "b.atago.yaml / fails") {
			t.Errorf("stderr = %q, want the out-of-scope entry named", errb.String())
		}
		st, err := loadRerunState()
		if err != nil {
			t.Fatalf("rerun state unreadable: %v", err)
		}
		if len(st.Failed) != 2 {
			t.Errorf("recorded failures = %+v, want both entries preserved", st.Failed)
		}
	})
}

// TestRerunFailed_UnmatchedWarningAgreesInNumberAndIsBounded pins the shape of
// the mismatch warning itself: the verb follows the count, and a ledger that
// accumulated many stale entries is summarized instead of printed in full — the
// warning is read before the run's own result, so it cannot be a wall of text.
func TestRerunFailed_UnmatchedWarningAgreesInNumberAndIsBounded(t *testing.T) {
	scenario := func(name string) string {
		return "  - name: " + name + "\n    steps:\n      - run: {shell: true, command: \"exit 1\"}\n      - assert: {exit_code: 0}\n"
	}
	head := "version: \"1\"\nsuite:\n  name: rerun\nscenarios:\n"
	var before, after strings.Builder
	before.WriteString(head)
	after.WriteString(head)
	before.WriteString(scenario("kept"))
	after.WriteString(scenario("kept"))
	for i := range 8 {
		before.WriteString(scenario("gone" + strconv.Itoa(i)))
		after.WriteString(scenario("renamed" + strconv.Itoa(i)))
	}

	dir := t.TempDir()
	writeSpec(t, dir, "s.atago.yaml", before.String())
	withWorkdir(t, dir, func() {
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("first run exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		// Every recorded failure but one is renamed while still broken.
		writeSpec(t, dir, "s.atago.yaml", after.String())

		out.Reset()
		errb.Reset()
		if got := Main([]string{"run", "--rerun-failed", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("rerun exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		if !strings.Contains(errb.String(), "8 recorded failing scenarios did not match the current specs (renamed or removed?) and were not rerun") {
			t.Errorf("stderr = %q, want a plural subject and a plural verb", errb.String())
		}
		if !strings.Contains(errb.String(), "and 3 more") {
			t.Errorf("stderr = %q, want the named entries bounded with a count of the rest", errb.String())
		}
		if n := strings.Count(errb.String(), "s.atago.yaml / gone"); n != maxNamedRerunEntries {
			t.Errorf("named entries = %d, want %d", n, maxNamedRerunEntries)
		}
	})
}

// TestRerunFailed_SingularWarningKeepsItsVerb guards the other half of the
// agreement fix: one unmatched entry keeps the singular noun and verb.
func TestRerunFailed_SingularWarningKeepsItsVerb(t *testing.T) {
	const spec = `version: "1"
suite:
  name: rerun
scenarios:
  - name: fails
    steps:
      - run: {shell: true, command: "exit 1"}
      - assert: {exit_code: 0}
  - name: also-fails
    steps:
      - run: {shell: true, command: "exit 1"}
      - assert: {exit_code: 0}
`
	dir := t.TempDir()
	writeSpec(t, dir, "s.atago.yaml", spec)
	withWorkdir(t, dir, func() {
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("first run exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		writeSpec(t, dir, "s.atago.yaml", strings.ReplaceAll(spec, "name: fails\n", "name: fails-renamed\n"))

		out.Reset()
		errb.Reset()
		if got := Main([]string{"run", "--rerun-failed", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("rerun exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		if !strings.Contains(errb.String(), "1 recorded failing scenario did not match the current specs (renamed or removed?) and was not rerun") {
			t.Errorf("stderr = %q, want the singular noun and verb", errb.String())
		}
	})
}

// TestRerunFailed_FullMatchIsQuiet pins that the warning is about a real
// mismatch: when every recorded failure still exists, nothing extra is printed.
func TestRerunFailed_FullMatchIsQuiet(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "s.atago.yaml", twoScenarioSpec)
	withWorkdir(t, dir, func() {
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("first run exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		out.Reset()
		errb.Reset()
		if got := Main([]string{"run", "--rerun-failed", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("rerun exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		if strings.Contains(errb.String(), "did not match the current specs") {
			t.Errorf("stderr = %q, want no mismatch warning when every recorded failure ran", errb.String())
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

// TestFailFast_UnrunScenarioKeepsRecordedFailure is a regression: --fail-fast
// stops scheduling, so a scenario after the first red one never runs and is
// reported as "skipped after fail-fast". The ledger read that skip as a verdict
// and dropped the failure it had already recorded, so the next --rerun-failed
// exited 0 with the scenario still broken.
func TestFailFast_UnrunScenarioKeepsRecordedFailure(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "s.atago.yaml", twoFailingSpec)

	withWorkdir(t, dir, func() {
		var out, errb bytes.Buffer
		// A full run records both alpha_fail and beta_fail.
		if got := Main([]string{"run", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("first run exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}

		// --parallel 1 makes fail-fast deterministic: alpha_fail turns the run red
		// before beta_fail is ever scheduled.
		out.Reset()
		errb.Reset()
		if got := Main([]string{"run", "--fail-fast", "--parallel", "1", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("fail-fast run exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		st, err := loadRerunState()
		if err != nil {
			t.Fatal(err)
		}
		names := map[string]bool{}
		for _, e := range st.Failed {
			names[e.Scenario] = true
		}
		if !names["beta_fail"] {
			t.Fatalf("ledger after a fail-fast run = %+v, want beta_fail preserved (never ran, so never re-verified)", st.Failed)
		}

		// Fix only alpha: the red-green loop must still be red, because beta was
		// never re-verified.
		writeSpec(t, dir, "s.atago.yaml", `version: "1"
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
		if got := Main([]string{"run", "--rerun-failed", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("--rerun-failed exit = %d, want %d (beta_fail is still broken); stderr=%s", got, ExitFailures, errb.String())
		}
	})
}

// TestRerunFailed_FailFastUnrunEntryNotBlamedOnRename is a regression: a
// recorded failure --fail-fast never got to still exists in the specs, so the
// run must not blame a rename or a removal for it.
func TestRerunFailed_FailFastUnrunEntryNotBlamedOnRename(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "s.atago.yaml", twoFailingSpec)

	withWorkdir(t, dir, func() {
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("first run exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		out.Reset()
		errb.Reset()
		if got := Main([]string{"run", "--rerun-failed", "--fail-fast", "--parallel", "1", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("fail-fast rerun exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		if strings.Contains(errb.String(), "renamed or removed") {
			t.Errorf("stderr blamed a rename/removal for a scenario --fail-fast never scheduled:\n%s", errb.String())
		}
	})
}

// TestRerunFailed_FailFastUnreachedSpecNotBlamedOnRename is the cross-spec half
// of the case above: --fail-fast also stops scheduling whole spec files, and a
// spec that was never loaded contributes no scenarios at all. Its recorded
// failures are correctly kept, but calling them renamed or removed sends the
// reader looking for a spec change that never happened.
func TestRerunFailed_FailFastUnreachedSpecNotBlamedOnRename(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "a.atago.yaml", singleFailSpec("a", false))
	writeSpec(t, dir, "b.atago.yaml", singleFailSpec("b", false))

	withWorkdir(t, dir, func() {
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("first run exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		// a.atago.yaml fails, so b.atago.yaml is never loaded.
		out.Reset()
		errb.Reset()
		if got := Main([]string{"run", "--rerun-failed", "--fail-fast", "--parallel", "1", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("fail-fast rerun exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		if strings.Contains(errb.String(), "renamed or removed") {
			t.Errorf("stderr blamed a rename/removal for a spec --fail-fast never loaded:\n%s", errb.String())
		}
		st, err := loadRerunState()
		if err != nil {
			t.Fatal(err)
		}
		if len(st.Failed) != 2 {
			t.Errorf("ledger = %+v, want both failures kept (b was never re-verified)", st.Failed)
		}
	})
}

// TestRerunFailed_RenamedScenarioStillWarnsUnderFailFast guards the other side:
// suppressing the warning for specs a fail-fast never reached must not suppress
// it for a spec the run did load, where a recorded scenario really is gone.
func TestRerunFailed_RenamedScenarioStillWarnsUnderFailFast(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, dir, "s.atago.yaml", twoFailingSpec)

	withWorkdir(t, dir, func() {
		var out, errb bytes.Buffer
		if got := Main([]string{"run", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("first run exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		// alpha_fail is renamed while still broken. The rerun loads the spec and
		// runs beta_fail, so alpha_fail really has gone missing.
		writeSpec(t, dir, "s.atago.yaml", `version: "1"
suite:
  name: multi
scenarios:
  - name: alpha_renamed
    steps:
      - run: {shell: true, command: "exit 1"}
      - assert: {exit_code: 0}
  - name: beta_fail
    steps:
      - run: {shell: true, command: "exit 1"}
      - assert: {exit_code: 0}
`)
		out.Reset()
		errb.Reset()
		if got := Main([]string{"run", "--rerun-failed", "--fail-fast", "--parallel", "1", "."}, &out, &errb); got != ExitFailures {
			t.Fatalf("rerun exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
		}
		if !strings.Contains(errb.String(), "renamed or removed") {
			t.Errorf("stderr = %q, want the rename/removal warning for alpha_fail", errb.String())
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

// TestRunCmd_CorruptLedgerLeftUntouchedByPlainRun guards the save-side half of
// the rerun-ledger contract (the read side under --rerun-failed exits 3 and is
// tested elsewhere): a plain `run` that cannot READ .atago/last-failed.json —
// corrupt bytes, or a future schema version a newer atago wrote — must warn and
// leave the file byte-identical. Overwriting it with only this run's outcome
// would destroy recorded failures we cannot see, and a later --rerun-failed
// would silently greenlight while real failures remain. A fully green run of
// unrelated specs must ALSO leave the unreadable ledger in place rather than
// clearing it.
func TestRunCmd_CorruptLedgerLeftUntouchedByPlainRun(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	garbage := []byte(`{"schema_version": 999, "failed": "not-a-list"`)
	if err := os.MkdirAll(".atago", 0o750); err != nil {
		t.Fatal(err)
	}
	ledger := filepath.Join(".atago", "last-failed.json")
	if err := os.WriteFile(ledger, garbage, 0o600); err != nil {
		t.Fatal(err)
	}

	// A failing run: warns, exits 1, ledger byte-identical.
	fp := writeSpec(t, dir, "fail.atago.yaml", failingSpec)
	var out, errb bytes.Buffer
	if got := Main([]string{"run", fp}, &out, &errb); got != ExitFailures {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitFailures, errb.String())
	}
	if !strings.Contains(errb.String(), "leaving it untouched") {
		t.Errorf("stderr = %q, want the leaving-it-untouched warning", errb.String())
	}
	data, err := os.ReadFile(ledger)
	if err != nil {
		t.Fatalf("ledger vanished after a failing run: %v", err)
	}
	if !bytes.Equal(data, garbage) {
		t.Errorf("ledger = %q, want the unreadable bytes preserved verbatim", data)
	}

	// A green run of an unrelated spec: still byte-identical (not cleared).
	pp := writeSpec(t, dir, "ok.atago.yaml", passingSpec)
	out.Reset()
	errb.Reset()
	if got := Main([]string{"run", pp}, &out, &errb); got != ExitOK {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitOK, errb.String())
	}
	data, err = os.ReadFile(ledger)
	if err != nil {
		t.Fatalf("ledger vanished after a green run: %v", err)
	}
	if !bytes.Equal(data, garbage) {
		t.Errorf("ledger = %q, want the unreadable bytes preserved verbatim after a green run", data)
	}
}

// TestSpecTargets_MessagesNameWhatAtagoWasDoing pins the wording a person meets
// when they point atago at the wrong place. A bare "stat x: no such file or
// directory" repeats the path atago already named, and a bare "open
// /etc/credstore: permission denied" reads as if atago wanted that file rather
// than as a search of the directory the user named.
func TestSpecTargets_MessagesNameWhatAtagoWasDoing(t *testing.T) {
	t.Run("a missing target names the reason once", func(t *testing.T) {
		var errb bytes.Buffer
		paths, exit, ok := specTargets("atago run", []string{"no-such.atago.yaml"}, &errb)
		if ok || paths != nil {
			t.Fatalf("ok = %v, paths = %v, want a refusal", ok, paths)
		}
		if exit != ExitConfig {
			t.Errorf("exit = %d, want %d", exit, ExitConfig)
		}
		got := errb.String()
		// The reason itself is the OS's ("no such file or directory" on POSIX,
		// "The system cannot find the file specified." on Windows); what atago
		// controls is that the path is named once, in quotes, with no syscall
		// wrapper repeating it.
		if !strings.Contains(got, `cannot access "no-such.atago.yaml": `) {
			t.Errorf("stderr = %q, want the quoted path followed by the reason", got)
		}
		if strings.Contains(got, "stat ") || strings.Count(got, "no-such.atago.yaml") != 1 {
			t.Errorf("stderr = %q, want the path exactly once and no syscall noise", got)
		}
	})

	t.Run("an empty directory says how to start", func(t *testing.T) {
		dir := t.TempDir()
		var errb bytes.Buffer
		if _, _, ok := specTargets("atago doc", []string{dir}, &errb); ok {
			t.Fatal("an empty directory must be a refusal")
		}
		got := errb.String()
		// The target is quoted, so a Windows path appears with escaped
		// separators — compare against the quoted form rather than the raw one.
		if !strings.Contains(got, "no *.atago.yaml") || !strings.Contains(got, strconv.Quote(dir)) {
			t.Errorf("stderr = %q, want the searched target named", got)
		}
		if !strings.Contains(got, "atago init") {
			t.Errorf("stderr = %q, want a pointer to `atago init`", got)
		}
	})
}
