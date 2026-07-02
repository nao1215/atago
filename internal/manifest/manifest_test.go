package manifest

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/loader"
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
