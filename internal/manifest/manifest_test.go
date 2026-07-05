package manifest

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/loader"
	"github.com/nao1215/atago/internal/spec"
)

// mixedSpec exercises run, http, query, grpc, cdp, services, store, and matrix so
// the manifest builder is covered end to end (#49).
const mixedSpec = `
version: "1"
suite:
  name: mixed
secrets:
  - TOKEN
permissions:
  network:
    allow:
      - api.example.com
runners:
  api:
    type: http
    base_url: https://api.example.com
  data:
    type: db
    dsn: sqlite:./app.db
  svc:
    type: grpc
    target: localhost:50051
  web:
    type: browser
scenarios:
  - name: "greets ${who}"
    tags: [smoke]
    matrix:
      - { who: Alice }
      - { who: Bob }
    services:
      - name: peer
        command: ./peer --addr ${workdir}/peer.addr
        ready:
          file: peer.addr
          store: addr
    steps:
      - run: {command: "echo ${who}"}
      - http: {runner: api, method: GET, path: /users}
      - query: {runner: data, sql: "SELECT 1"}
      - grpc: {runner: svc, method: pkg.S/M}
      - cdp:
          runner: web
          actions:
            - navigate: https://example.com
            - text: h1
      - store:
          name: uid
          from:
            body: {json: {path: $.id}}
      - assert: {stdout: {contains: "${who}"}}
      - assert: {file: {path: out.txt, exists: true}}
`

func buildMixed(t *testing.T) Document {
	t.Helper()
	s, err := loader.LoadBytes("mixed.atago.yaml", []byte(mixedSpec))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	return Build([]Input{{Spec: s, Path: "mixed.atago.yaml"}})
}

func TestBuild_MixedSpec(t *testing.T) {
	t.Parallel()
	doc := buildMixed(t)
	if doc.SchemaVersion != "1" {
		t.Errorf("schema_version = %q, want 1", doc.SchemaVersion)
	}
	if len(doc.Specs) != 1 {
		t.Fatalf("specs = %d, want 1", len(doc.Specs))
	}
	sp := doc.Specs[0]
	if sp.Suite != "mixed" {
		t.Errorf("suite = %q", sp.Suite)
	}
	if len(sp.Secrets) != 1 || sp.Secrets[0] != "TOKEN" {
		t.Errorf("secrets = %v", sp.Secrets)
	}
	if sp.Network.Policy != "allowlist" || len(sp.Network.Allow) != 1 {
		t.Errorf("network = %+v", sp.Network)
	}
	// Runners are sorted by name.
	if got := runnerNames(sp.Runners); strings.Join(got, ",") != "api,data,svc,web" {
		t.Errorf("runner order = %v", got)
	}
	for _, r := range sp.Runners {
		if r.Name == "data" && !r.HasDSN {
			t.Errorf("db runner should flag has_dsn: %+v", r)
		}
	}
	// Matrix expands into two scenarios, each carrying its bound row as vars.
	if len(sp.Scenarios) != 2 {
		t.Fatalf("scenarios = %d, want 2 (matrix expanded)", len(sp.Scenarios))
	}
	sc := sp.Scenarios[0]
	if sc.Vars["who"] != "Alice" {
		t.Errorf("scenario vars = %v, want who=Alice", sc.Vars)
	}
	// Steps preserve definition order and kinds.
	wantKinds := []string{"run", "http", "query", "grpc", "cdp", "store", "assert", "assert"}
	if len(sc.Steps) != len(wantKinds) {
		t.Fatalf("steps = %d, want %d", len(sc.Steps), len(wantKinds))
	}
	for i, k := range wantKinds {
		if sc.Steps[i].Kind != k {
			t.Errorf("step %d kind = %q, want %q", i, sc.Steps[i].Kind, k)
		}
		if sc.Steps[i].Index != i {
			t.Errorf("step %d index = %d", i, sc.Steps[i].Index)
		}
	}
	// A service and its readiness signal are captured.
	if len(sc.Services) != 1 || sc.Services[0].Ready != "file" || sc.Services[0].Store != "addr" {
		t.Errorf("services = %+v", sc.Services)
	}
	// Variable references are collected and sorted (who + workdir).
	if strings.Join(sc.Variables, ",") != "who,workdir" {
		t.Errorf("variables = %v", sc.Variables)
	}
	// Generated artifact and security notes are surfaced.
	if len(sc.Generates) != 1 || sc.Generates[0] != "out.txt" {
		t.Errorf("generates = %v", sc.Generates)
	}
	if !containsSub(sc.Security, "network access: HTTP request") ||
		!containsSub(sc.Security, "browser automation (CDP)") ||
		!containsSub(sc.Security, "network access: gRPC") {
		t.Errorf("security = %v", sc.Security)
	}
}

func TestBuild_DeterministicJSON(t *testing.T) {
	t.Parallel()
	a, err := json.MarshalIndent(buildMixed(t), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.MarshalIndent(buildMixed(t), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if string(a) != string(b) {
		t.Errorf("manifest JSON not deterministic across builds")
	}
	// The dsn must never leak into the serialized manifest.
	if strings.Contains(string(a), "app.db") {
		t.Errorf("dsn leaked into manifest:\n%s", a)
	}
}

func TestBuild_SecretsNeverExposeValues(t *testing.T) {
	t.Parallel()
	doc := buildMixed(t)
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	// Only the secret name is declared; there is no value to leak, but assert the
	// manifest carries the name so tooling can flag secret-bearing specs.
	if !strings.Contains(string(out), "TOKEN") {
		t.Errorf("manifest missing declared secret name")
	}
}

// TestBuild_GeneratedArtifactsAcrossKinds keeps the manifest aligned with the
// shared spec model: image outputs and cdp screenshots are generated artifacts,
// just like file exists:true (#56).
func TestBuild_GeneratedArtifactsAcrossKinds(t *testing.T) {
	t.Parallel()
	const src = `
version: "1"
suite:
  name: gen
runners:
  web: {type: browser}
scenarios:
  - name: produces artifacts
    steps:
      - assert:
          image:
            path: thumb.png
            similar_to: baseline.png
            max_diff: 0.02
      - assert:
          file:
            path: out.txt
            exists: true
      - cdp:
          runner: web
          actions:
            - navigate: http://localhost:8080
            - screenshot: {path: home.png}
`
	s, err := loader.LoadBytes("gen.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	doc := Build([]Input{{Spec: s, Path: "gen.atago.yaml"}})
	got := doc.Specs[0].Scenarios[0].Generates
	if strings.Join(got, ",") != "thumb.png,out.txt,home.png" {
		t.Errorf("generates = %v, want [thumb.png out.txt home.png]", got)
	}
}

// TestBuild_BrowserRunnerConfig proves the manifest surfaces the browser-runner
// configuration so tooling sees the same runtime knobs the engine honors.
func TestBuild_BrowserRunnerConfig(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: b
runners:
  web:
    type: browser
    headless: false
    exec_path: /usr/bin/chromium
    browser_args: ["disable-gpu", "window-size=1280,720"]
scenarios:
  - name: s
    steps:
      - cdp: {runner: web, actions: [{navigate: https://example.com}]}
`
	s, err := loader.LoadBytes("b.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	doc := Build([]Input{{Spec: s, Path: "b.atago.yaml"}})
	var web *Runner
	for i := range doc.Specs[0].Runners {
		if doc.Specs[0].Runners[i].Name == "web" {
			web = &doc.Specs[0].Runners[i]
		}
	}
	if web == nil {
		t.Fatal("web runner missing from manifest")
	}
	if web.Headless == nil || *web.Headless {
		t.Errorf("manifest headless = %v, want explicit false", web.Headless)
	}
	if web.ExecPath != "/usr/bin/chromium" {
		t.Errorf("manifest exec_path = %q", web.ExecPath)
	}
	if len(web.BrowserArgs) != 2 {
		t.Errorf("manifest browser_args = %v, want two entries", web.BrowserArgs)
	}
	// The config must not leak into unrelated runner types when absent.
	out, _ := json.Marshal(doc)
	if strings.Count(string(out), "\"headless\"") != 1 {
		t.Errorf("headless should appear exactly once in the manifest JSON:\n%s", out)
	}
}

func runnerNames(rs []Runner) []string {
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		out = append(out, r.Name)
	}
	return out
}

func containsSub(list []string, sub string) bool {
	for _, s := range list {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func bp(b bool) *bool { return &b }
func ip(i int) *int   { return &i }

// TestBuildStep_EveryKind drives the manifest builder over every step kind,
// including the ones the loader only allows in specific positions (service,
// mock_server), by constructing the spec model directly. It asserts the kind,
// action summary, and the salient declarative fields the manifest promises for
// each kind, so a silently dropped or mislabeled kind is caught (#49).
func TestBuildStep_EveryKind(t *testing.T) {
	t.Parallel()
	s := &spec.Spec{
		Version: "1",
		Suite: spec.Suite{
			Name: "everykind",
			// service and mock_server steps are valid only in suite.setup; put
			// them there so buildStep's StepService / StepMockServer arms run.
			Setup: []spec.Step{
				{Service: &spec.Service{Name: "peer", Command: "./peer", Shell: spec.Bool(true), ClearEnv: bp(true), PassEnv: []string{"PATH"}}},
				{MockServer: &spec.MockServer{Name: "api", Routes: []spec.MockRoute{{Method: "GET", Path: "/a"}, {Method: "GET", Path: "/b"}}}},
			},
		},
		Scenarios: []spec.Scenario{{
			Name:        "all",
			SourceIndex: 0,
			Steps: []spec.Step{
				{Fixture: &spec.Fixture{File: "in.txt", Content: "hi"}},
				{Run: &spec.Run{Command: "echo hi", Shell: spec.Bool(true), ClearEnv: bp(true), PassEnv: []string{"HOME"}, Runner: "cmd", Retry: &spec.Retry{Times: 3, Interval: "1s"}}},
				{HTTP: &spec.HTTP{Runner: "api", Method: "POST", Path: "/x", Retry: &spec.Retry{Times: 2, Interval: "500ms"}}},
				{Query: &spec.Query{Runner: "db", SQL: "SELECT 1"}},
				{GRPC: &spec.GRPC{Runner: "g", Method: "pkg.S/M"}},
				{CDP: &spec.CDP{Runner: "web", Actions: []spec.CDPAction{{Navigate: "http://x"}, {Text: "#h"}}}},
				{PTY: &spec.PTY{Command: "top", Shell: spec.Bool(true), ClearEnv: bp(true), PassEnv: []string{"TERM"}}},
				{Signal: &spec.Signal{Service: "peer", Signal: "sigterm", Wait: &spec.SignalWait{Timeout: "3s"}}},
				{Assert: &spec.Assert{ExitCode: &spec.ExitCode{Equals: ip(0)}, Stdout: &spec.StreamAssert{Contains: spec.StringList{"ok"}}}},
				{Store: &spec.Store{Name: "uid"}},
			},
		}},
	}

	doc := Build([]Input{{Spec: s, Path: "e.atago.yaml"}})
	sp := doc.Specs[0]

	// Suite setup carries the service and mock_server steps.
	if len(sp.SuiteSetup) != 2 {
		t.Fatalf("suite setup steps = %d, want 2", len(sp.SuiteSetup))
	}
	svcStep := sp.SuiteSetup[0]
	if svcStep.Kind != "service" || svcStep.Target != "peer" || !svcStep.Shell || !svcStep.ClearEnv {
		t.Errorf("service step = %+v", svcStep)
	}
	if svcStep.Action != "start suite service peer" {
		t.Errorf("service action = %q", svcStep.Action)
	}
	msStep := sp.SuiteSetup[1]
	if msStep.Kind != "mock_server" || msStep.Target != "api" || !strings.Contains(msStep.Action, "2 routes") {
		t.Errorf("mock_server step = %+v", msStep)
	}

	steps := sp.Scenarios[0].Steps
	byKind := map[string]Step{}
	for _, st := range steps {
		byKind[st.Kind] = st
	}
	for _, want := range []string{"fixture", "run", "http", "query", "grpc", "cdp", "pty", "signal", "assert", "store"} {
		if _, ok := byKind[want]; !ok {
			t.Errorf("manifest dropped step kind %q; got kinds %v", want, keysOf(byKind))
		}
	}

	if got := byKind["fixture"]; got.File != "in.txt" || got.Action != "write fixture in.txt" {
		t.Errorf("fixture step = %+v", got)
	}
	if got := byKind["run"]; got.Command != "echo hi" || !got.Shell || !got.ClearEnv || got.Runner != "cmd" || got.Retry == nil || got.Retry.Times != 3 || got.Retry.Interval != "1s" {
		t.Errorf("run step = %+v (retry %+v)", got, got.Retry)
	}
	if got := byKind["http"]; got.Method != http.MethodPost || got.Path != "/x" || got.Retry == nil || got.Retry.Times != 2 {
		t.Errorf("http step = %+v", got)
	}
	if got := byKind["query"]; got.SQL != "SELECT 1" || got.Action != "SQL query via db" {
		t.Errorf("query step = %+v", got)
	}
	if got := byKind["grpc"]; got.Runner != "g" || got.Action != "gRPC pkg.S/M via g" {
		t.Errorf("grpc step = %+v", got)
	}
	if got := byKind["cdp"]; got.Runner != "web" || !strings.Contains(got.Action, "navigate http://x") {
		t.Errorf("cdp step = %+v", got)
	}
	if got := byKind["pty"]; got.Command != "top" || !got.Shell || !got.ClearEnv || got.Action != "interactive (pty) top" {
		t.Errorf("pty step = %+v", got)
	}
	// The signal step normalizes SIGTERM and reports the wait timeout.
	if got := byKind["signal"]; got.Target != "peer" || got.Action != "signal SIGTERM to service peer, wait up to 3s for exit" {
		t.Errorf("signal step action = %q", byKind["signal"].Action)
	}
	// A multi-target assert joins its target names with "+".
	if got := byKind["assert"]; got.Target != "exit_code+stdout" || got.Action != "assert exit_code+stdout" {
		t.Errorf("assert step target = %q", byKind["assert"].Target)
	}
	if got := byKind["store"]; got.Target != "uid" || got.Action != "store uid" {
		t.Errorf("store step = %+v", got)
	}
}

// TestBuildStep_SignalDefaultWaitTimeout covers the signal step's default-wait
// branch: `wait:` with no timeout reports the documented 5s default.
func TestBuildStep_SignalDefaultWaitTimeout(t *testing.T) {
	t.Parallel()
	s := &spec.Spec{
		Suite: spec.Suite{Name: "s"},
		Scenarios: []spec.Scenario{{
			Name: "sig",
			Steps: []spec.Step{
				{Signal: &spec.Signal{Service: "peer", Signal: "INT", Wait: &spec.SignalWait{}}},
				// A signal without wait: omits the wait clause entirely.
				{Signal: &spec.Signal{Service: "peer", Signal: "HUP"}},
			},
		}},
	}
	steps := Build([]Input{{Spec: s, Path: "s.atago.yaml"}}).Specs[0].Scenarios[0].Steps
	if steps[0].Action != "signal SIGINT to service peer, wait up to 5s for exit" {
		t.Errorf("default-wait action = %q", steps[0].Action)
	}
	if steps[1].Action != "signal SIGHUP to service peer" {
		t.Errorf("no-wait action = %q", steps[1].Action)
	}
}

// TestBuildService_AllReadinessModes exercises every readiness kind buildService
// can label, plus the no-readiness case and the store capture.
func TestBuildService_AllReadinessModes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		ready *spec.Ready
		want  string
		store string
	}{
		{"file", &spec.Ready{File: "sock", Store: "addr"}, "file", "addr"},
		{"port", &spec.Ready{Port: "127.0.0.1:9000"}, "port", ""},
		{"log", &spec.Ready{Log: "listening"}, "log", ""},
		{"delay", &spec.Ready{Delay: "200ms"}, "delay", ""},
		{"none", nil, "", ""},
		// An empty (all-zero) Ready block is non-nil but names no signal: the
		// switch falls through and Ready stays empty.
		{"empty ready block", &spec.Ready{}, "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out := buildService(&spec.Service{Name: "svc", Command: "run", Ready: tc.ready})
			if out.Ready != tc.want {
				t.Errorf("Ready = %q, want %q", out.Ready, tc.want)
			}
			if out.Store != tc.store {
				t.Errorf("Store = %q, want %q", out.Store, tc.store)
			}
		})
	}
}

// TestBuildService_EnvControls proves the hermetic-env controls surface.
func TestBuildService_EnvControls(t *testing.T) {
	t.Parallel()
	out := buildService(&spec.Service{Name: "s", Command: "c", Shell: spec.Bool(true), ClearEnv: bp(true), PassEnv: []string{"PATH", "HOME"}})
	if !out.Shell || !out.ClearEnv || len(out.PassEnv) != 2 {
		t.Errorf("service env controls = %+v", out)
	}
}

// TestAssertTarget_Invalid covers the empty-assert branch: an assert with no
// target family reduces to "invalid" rather than an empty or panicking label.
func TestAssertTarget_Invalid(t *testing.T) {
	t.Parallel()
	if got := assertTarget(&spec.Assert{}); got != "invalid" {
		t.Errorf("assertTarget(empty) = %q, want invalid", got)
	}
	// A single-target assert is named directly.
	if got := assertTarget(&spec.Assert{Status: ip(200)}); got != "status" {
		t.Errorf("assertTarget(status) = %q", got)
	}
}

// TestSourceFrom covers the position-to-Source reduction: an unknown (<=0) line
// yields nil so the field is omitted; a known line keeps its column.
func TestSourceFrom(t *testing.T) {
	t.Parallel()
	if got := sourceFrom(0, 0); got != nil {
		t.Errorf("sourceFrom(0,0) = %+v, want nil", got)
	}
	if got := sourceFrom(-1, 4); got != nil {
		t.Errorf("sourceFrom(-1,4) = %+v, want nil", got)
	}
	if got := sourceFrom(7, 0); got == nil || got.Line != 7 || got.Column != 0 {
		t.Errorf("sourceFrom(7,0) = %+v, want line 7 col 0", got)
	}
	if got := sourceFrom(7, 3); got == nil || got.Line != 7 || got.Column != 3 {
		t.Errorf("sourceFrom(7,3) = %+v, want line 7 col 3", got)
	}
}

// TestBuildSpec_SuiteLifecycle covers buildSpec's suite-lifecycle handling:
// suite timeout, sorted env key names (values never leaked), and setup/teardown
// step summaries.
func TestBuildSpec_SuiteLifecycle(t *testing.T) {
	t.Parallel()
	s := &spec.Spec{
		Suite: spec.Suite{
			Name:    "life",
			Timeout: "30s",
			Env:     map[string]string{"ZED": "secretz", "ALPHA": "secreta"},
			Setup:   []spec.Step{{Run: &spec.Run{Command: "build"}}},
			Teardown: []spec.Step{
				{Run: &spec.Run{Command: "cleanup"}},
			},
		},
		Scenarios: []spec.Scenario{{Name: "s", Steps: []spec.Step{{Run: &spec.Run{Command: "x"}}}}},
	}
	sp := Build([]Input{{Spec: s, Path: "life.atago.yaml"}}).Specs[0]
	if sp.SuiteTimeout != "30s" {
		t.Errorf("suite timeout = %q", sp.SuiteTimeout)
	}
	// Env is emitted as sorted key names only.
	if strings.Join(sp.SuiteEnv, ",") != "ALPHA,ZED" {
		t.Errorf("suite env keys = %v, want [ALPHA ZED]", sp.SuiteEnv)
	}
	if len(sp.SuiteSetup) != 1 || sp.SuiteSetup[0].Command != "build" {
		t.Errorf("suite setup = %+v", sp.SuiteSetup)
	}
	if len(sp.SuiteTeardown) != 1 || sp.SuiteTeardown[0].Command != "cleanup" {
		t.Errorf("suite teardown = %+v", sp.SuiteTeardown)
	}
	// Env values must never appear in the serialized manifest.
	out, err := json.Marshal(sp)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "secret") {
		t.Errorf("suite env value leaked into manifest:\n%s", out)
	}
}

// TestBuildScenario_TeardownAndConditions covers scenario teardown steps and the
// only/skip condition reduction.
func TestBuildScenario_TeardownAndConditions(t *testing.T) {
	t.Parallel()
	s := &spec.Spec{
		Suite: spec.Suite{Name: "s"},
		Scenarios: []spec.Scenario{{
			Name: "sc",
			Only: &spec.Condition{OS: "linux"},
			Skip: &spec.Condition{Env: "CI"},
			Steps: []spec.Step{
				{Run: &spec.Run{Command: "echo ${who}"}},
			},
			Teardown: []spec.Step{
				{Run: &spec.Run{Command: "rm ${tmp}"}},
			},
		}},
	}
	sc := Build([]Input{{Spec: s, Path: "s.atago.yaml"}}).Specs[0].Scenarios[0]
	if sc.Only == nil || sc.Only.OS != "linux" {
		t.Errorf("only = %+v", sc.Only)
	}
	if sc.Skip == nil || sc.Skip.Env != "CI" {
		t.Errorf("skip = %+v", sc.Skip)
	}
	if len(sc.Teardown) != 1 || sc.Teardown[0].Command != "rm ${tmp}" {
		t.Errorf("teardown = %+v", sc.Teardown)
	}
	// Teardown variable references count toward the scenario's variable set,
	// alongside the main steps'.
	if strings.Join(sortStrings(sc.Variables), ",") != "tmp,who" {
		t.Errorf("variables = %v, want [tmp who]", sc.Variables)
	}
}

// TestBuildStep_AssertVarsCollected is the manifest-side regression for the
// assert-variable under-reporting bug: a variable referenced ONLY inside an
// assert step must appear in the scenario's manifest variable list.
func TestBuildStep_AssertVarsCollected(t *testing.T) {
	t.Parallel()
	s := &spec.Spec{
		Suite: spec.Suite{Name: "s"},
		Scenarios: []spec.Scenario{{
			Name: "sc",
			Steps: []spec.Step{
				{Run: &spec.Run{Command: "run-cli"}}, // no vars here
				{Assert: &spec.Assert{Stdout: &spec.StreamAssert{Equals: strPtr("${expected}")}}},
			},
		}},
	}
	sc := Build([]Input{{Spec: s, Path: "s.atago.yaml"}}).Specs[0].Scenarios[0]
	found := false
	for _, v := range sc.Variables {
		if v == "expected" {
			found = true
		}
	}
	if !found {
		t.Errorf("assert-only variable ${expected} missing from manifest variables: %v", sc.Variables)
	}
}

func strPtr(s string) *string { return &s }

func keysOf(m map[string]Step) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return sortStrings(out)
}

func sortStrings(in []string) []string {
	out := append([]string(nil), in...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j-1] > out[j]; j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}
