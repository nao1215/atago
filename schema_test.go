package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/nao1215/atago/internal/engine"
	"github.com/nao1215/atago/internal/loader"
	"github.com/nao1215/atago/internal/manifest"
	"github.com/nao1215/atago/internal/report"
	"github.com/nao1215/atago/internal/spec"
)

// loadSchema compiles the published JSON Schema.
func loadSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	f, err := os.Open("schema/atago.schema.json")
	if err != nil {
		t.Fatalf("open schema: %v", err)
	}
	defer f.Close()

	doc, err := jsonschema.UnmarshalJSON(f)
	if err != nil {
		t.Fatalf("parse schema: %v", err)
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource("atago.schema.json", doc); err != nil {
		t.Fatalf("add resource: %v", err)
	}
	s, err := c.Compile("atago.schema.json")
	if err != nil {
		t.Fatalf("compile schema: %v", err)
	}
	return s
}

// yamlToAny decodes YAML into the generic types the validator expects.
func yamlToAny(t *testing.T, data []byte) any {
	t.Helper()
	var v any
	if err := yaml.Unmarshal(data, &v); err != nil {
		t.Fatalf("yaml unmarshal: %v", err)
	}
	return v
}

// shippedSpecPaths returns every *.atago.yaml this repository ships, walked
// from the module root rather than from a list of directories: a suite added
// under a new path is then covered by the guards below without anyone
// remembering they exist, which is exactly how fourteen specs drifted out of
// schema conformance while a two-file guard stayed green. Dot directories (the
// git metadata, agent worktrees) and dist/ (release output) carry no shipped
// spec, so the walk skips them wholesale.
func shippedSpecPaths(t *testing.T) []string {
	t.Helper()
	var specs []string
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if name := d.Name(); path != "." && (strings.HasPrefix(name, ".") || name == "dist") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".atago.yaml") || strings.HasSuffix(path, ".atago.yml") {
			specs = append(specs, filepath.ToSlash(filepath.Clean(path)))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk repository: %v", err)
	}
	// A walk that silently matches nothing turns every corpus guard into a
	// no-op, so assert the two trees that must always contribute.
	var examples, e2e int
	for _, p := range specs {
		switch {
		case strings.HasPrefix(p, "examples/"):
			examples++
		case strings.HasPrefix(p, "test/e2e/"):
			e2e++
		}
	}
	if examples == 0 || e2e == 0 {
		t.Fatalf("spec walk found %d specs but %d under examples/ and %d under test/e2e/; the corpus guards would pass on nothing", len(specs), examples, e2e)
	}
	return specs
}

// TestSchema_EveryShippedSpecConforms is the drift guard between the specs this
// repository ships and the schema editors validate them against. It used to
// check two demo files, which is how a schema that rejects a composable file
// assert and every store json-path capture shipped while the guard stayed
// green: a user copying examples/http.atago.yaml saw their editor reject it.
// Every spec in the tree is validated here, so a schema change that narrows the
// accepted vocabulary fails the build instead of a user's editor.
func TestSchema_EveryShippedSpecConforms(t *testing.T) {
	s := loadSchema(t)
	for _, p := range shippedSpecPaths(t) {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		if err := s.Validate(yamlToAny(t, data)); err != nil {
			t.Errorf("%s does not conform to schema/atago.schema.json:\n%v", p, err)
		}
	}
}

// TestSchema_AcceptsRunnerShapes confirms each discriminated runner shape (#44)
// accepts its own fields.
func TestSchema_AcceptsRunnerShapes(t *testing.T) {
	s := loadSchema(t)
	good := map[string]string{
		"cmd": `version: "1"
suite: {name: x}
runners:
  local: {type: cmd, cwd: ./sub, timeout: 5s}
scenarios:
  - name: a
    steps: [{run: {command: echo}}]`,
		"http": `version: "1"
suite: {name: x}
runners:
  api: {type: http, base_url: "http://localhost:8080"}
scenarios:
  - name: a
    steps: [{run: {command: echo}}]`,
		"db": `version: "1"
suite: {name: x}
runners:
  store: {type: db, dsn: "sqlite:./x.db", driver: sqlite}
scenarios:
  - name: a
    steps: [{run: {command: echo}}]`,
		"ssh": `version: "1"
suite: {name: x}
runners:
  box: {type: ssh, host: example.com, user: root, key_file: ./id, insecure_host_key: true}
scenarios:
  - name: a
    steps: [{run: {command: echo}}]`,
		"grpc": `version: "1"
suite: {name: x}
runners:
  svc: {type: grpc, target: "localhost:50051", tls: true}
scenarios:
  - name: a
    steps: [{run: {command: echo}}]`,
		"browser": `version: "1"
suite: {name: x}
runners:
  web: {type: browser, timeout: 30s}
scenarios:
  - name: a
    steps: [{run: {command: echo}}]`,
	}
	for name, src := range good {
		t.Run(name, func(t *testing.T) {
			if err := s.Validate(yamlToAny(t, []byte(src))); err != nil {
				t.Errorf("schema rejected valid %s runner:\n%v", name, err)
			}
		})
	}
}

// TestSchema_AcceptsComposableFileMatchers pins the file assert shapes the
// loader documents as composable: a size bound stands alone or joins a content
// matcher, and min_size with max_size bounds a range. The schema used to model
// the matcher list with oneOf, which reads as "exactly one" and turned every one
// of these into an editor error on a spec atago runs — including a shipped
// example. Each case is loaded as well as validated, so the test fails if the
// two ever disagree again in either direction.
func TestSchema_AcceptsComposableFileMatchers(t *testing.T) {
	s := loadSchema(t)
	cases := map[string]string{
		"exists with size":     `assert: {file: {path: out.txt, exists: true, size: 0}}`,
		"size alone":           `assert: {file: {path: out.txt, size: 12}}`,
		"size range":           `assert: {file: {path: out.txt, min_size: 1, max_size: 4096}}`,
		"contains with count":  `assert: {file: {path: out.txt, contains: warn, count: 2}}`,
		"contains with a size": `assert: {file: {path: out.txt, contains: warn, max_size: 4096}}`,
	}
	for name, step := range cases {
		t.Run(name, func(t *testing.T) {
			src := "version: \"1\"\nsuite: {name: x}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - " + step + "\n"
			if _, err := loader.LoadBytes("t.atago.yaml", []byte(src)); err != nil {
				t.Fatalf("loader rejected %s: %v", name, err)
			}
			if err := s.Validate(yamlToAny(t, []byte(src))); err != nil {
				t.Errorf("schema rejected %s, which the loader accepts:\n%v", name, err)
			}
		})
	}
}

// TestSchema_AcceptsStoreCaptures pins the store shapes against the schema. A
// capture selects a value rather than judging one, so its json node carries a
// path and no matcher — the only spelling the loader accepts there. The schema
// reached that node through the assert definition, whose oneOf demanded a
// matcher, so every store-by-json-path spec in this repository (three shipped
// examples among them) failed validation while running green.
func TestSchema_AcceptsStoreCaptures(t *testing.T) {
	s := loadSchema(t)
	cases := map[string]string{
		"stdout json path": `store: {name: id, from: {stdout: {json: {path: "$.id"}}}}`,
		"stdout matches":   `store: {name: id, from: {stdout: {matches: "id=(\\d+)"}}}`,
		"stdout trim":      `store: {name: all, from: {stdout: {trim: true}}}`,
		"file json path":   `store: {name: v, from: {file: {path: out.json, json: {path: "$.v"}}}}`,
		"file text":        `store: {name: body, from: {file: {path: out.txt, text: true}}}`,
	}
	for name, step := range cases {
		t.Run(name, func(t *testing.T) {
			src := "version: \"1\"\nsuite: {name: x}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - " + step + "\n"
			if _, err := loader.LoadBytes("t.atago.yaml", []byte(src)); err != nil {
				t.Fatalf("loader rejected %s: %v", name, err)
			}
			if err := s.Validate(yamlToAny(t, []byte(src))); err != nil {
				t.Errorf("schema rejected %s, which the loader accepts:\n%v", name, err)
			}
		})
	}
}

// TestSchema_AcceptsHermeticEnv confirms clear_env/pass_env (#16) are accepted
// on run, pty, service, and the defaults blocks.
func TestSchema_AcceptsHermeticEnv(t *testing.T) {
	s := loadSchema(t)
	src := `version: "1"
suite:
  name: x
  setup:
    - service: {name: shared, command: ./srv, clear_env: true, pass_env: [PATH]}
defaults:
  run: {clear_env: true, pass_env: [PATH, HOME]}
  service: {clear_env: true, pass_env: [PATH]}
scenarios:
  - name: a
    services:
      - {name: mock, command: ./mock, clear_env: true, pass_env: [PATH]}
    steps:
      - run: {command: env, clear_env: true, pass_env: [PATH, HOME], env: {A: b}}
      - pty: {command: cat, clear_env: true, pass_env: [TERM], session: [{send: ""}]}`
	if err := s.Validate(yamlToAny(t, []byte(src))); err != nil {
		t.Errorf("schema rejected valid hermetic-env spec:\n%v", err)
	}
}

// TestSchema_AcceptsSuiteTimeout confirms suite.timeout (#17) is accepted.
func TestSchema_AcceptsSuiteTimeout(t *testing.T) {
	s := loadSchema(t)
	src := `version: "1"
suite: {name: x, timeout: 2m}
scenarios:
  - name: a
    steps: [{run: {command: echo, timeout: "0"}}]`
	if err := s.Validate(yamlToAny(t, []byte(src))); err != nil {
		t.Errorf("schema rejected valid suite.timeout spec:\n%v", err)
	}
}

// TestSchema_AcceptsDescriptions confirms the published schema accepts the
// optional suite/scenario `description:` that `atago doc` renders, so an
// editor validating against the schema does not flag a spec the loader runs.
func TestSchema_AcceptsDescriptions(t *testing.T) {
	s := loadSchema(t)
	src := `version: "1"
suite:
  name: x
  description: |
    What this suite guarantees.

    Second paragraph.
scenarios:
  - name: a
    description: Why this scenario exists.
    steps: [{run: {command: echo}}]`
	if err := s.Validate(yamlToAny(t, []byte(src))); err != nil {
		t.Errorf("schema rejected valid suite/scenario description spec:\n%v", err)
	}
}

// TestSchema_AcceptsStdinSources confirms the three stdin shapes (#18) are
// accepted.
func TestSchema_AcceptsStdinSources(t *testing.T) {
	s := loadSchema(t)
	src := `version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps:
      - run: {command: cat, stdin: "inline"}
      - run: {command: cat, stdin: {file: in.txt}}
      - run: {command: cat, stdin: {base64: AAEC/w==}}`
	if err := s.Validate(yamlToAny(t, []byte(src))); err != nil {
		t.Errorf("schema rejected valid stdin sources:\n%v", err)
	}
}

// TestSchema_AcceptsExitCodeIn confirms exit_code: {in: [...]} (#19) is
// accepted.
func TestSchema_AcceptsExitCodeIn(t *testing.T) {
	s := loadSchema(t)
	src := `version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps:
      - run: {command: echo}
      - assert: {exit_code: {in: [0, 2]}}`
	if err := s.Validate(yamlToAny(t, []byte(src))); err != nil {
		t.Errorf("schema rejected valid exit_code in-set:\n%v", err)
	}
}

// TestSchema_AcceptsDuration confirms the duration assertion (#31) is
// accepted and a conflicting bound pair is rejected.
func TestSchema_AcceptsDuration(t *testing.T) {
	s := loadSchema(t)
	good := `version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps:
      - run: {command: echo}
      - assert:
          duration: {gte: 100ms, lt: 60s}`
	if err := s.Validate(yamlToAny(t, []byte(good))); err != nil {
		t.Errorf("schema rejected valid duration assert:\n%v", err)
	}
}

// TestSchema_AcceptsSignalStep confirms the signal step (#23) is accepted.
func TestSchema_AcceptsSignalStep(t *testing.T) {
	s := loadSchema(t)
	src := `version: "1"
suite: {name: x}
scenarios:
  - name: a
    services:
      - {name: srv, command: ./srv}
    steps:
      - signal: {service: srv, signal: TERM, wait: {timeout: 5s}}
      - signal: {service: srv, signal: KILL}`
	if err := s.Validate(yamlToAny(t, []byte(src))); err != nil {
		t.Errorf("schema rejected valid signal steps:\n%v", err)
	}
}

// TestSchema_AcceptsMockServers confirms mock_servers, the suite mock_server
// step, and the mock assert target (#24) are accepted.
func TestSchema_AcceptsMockServers(t *testing.T) {
	s := loadSchema(t)
	src := `version: "1"
suite:
  name: x
  setup:
    - mock_server:
        name: shared
        routes: [{method: GET, path: /ping, body: pong}]
scenarios:
  - name: a
    mock_servers:
      - name: api
        routes:
          - method: POST
            path: /v1/reports
            status: 201
            json: { id: "r-1" }
          - method: GET
            path: /file
            body_file: canned.json
            delay: 200ms
    steps:
      - run: {command: echo hi}
      - assert:
          mock:
            name: api
            path: /v1/reports
            method: POST
            count: 1
            header: { name: Authorization, matches: "^Bearer " }
            body: { contains: report }`
	if err := s.Validate(yamlToAny(t, []byte(src))); err != nil {
		t.Errorf("schema rejected valid mock-server spec:\n%v", err)
	}
}

// TestSchema_RejectsInvalid confirms the schema actually catches bad specs.
func TestSchema_RejectsInvalid(t *testing.T) {
	s := loadSchema(t)
	bad := map[string]string{
		"wrong version": `version: "2"
suite: {name: x}
scenarios:
  - name: a
    steps: [{run: {command: echo}}]`,
		"step with two actions": `version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps:
      - run: {command: echo}
        assert: {exit_code: 0}`,
		"missing suite name": `version: "1"
suite: {}
scenarios:
  - name: a
    steps: [{run: {command: echo}}]`,
		// Issue #22: headerMatch with only `name` must be rejected (one of
		// contains/equals is required) so the schema matches loader semantics.
		"header match without contains/equals": `version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps:
      - http: {method: GET, path: /}
      - assert: {header: {name: Content-Type}}`,
		// Issue #22: store.from with no source must be rejected (exactly one of
		// the seven sources is required).
		"store from with no source": `version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps:
      - run: {command: echo}
      - store: {name: x, from: {}}`,
		// Issue #44: an http runner must not carry SSH-only fields.
		"http runner with ssh field": `version: "1"
suite: {name: x}
runners:
  api: {type: http, base_url: "http://localhost", host: example.com}
scenarios:
  - name: a
    steps: [{run: {command: echo}}]`,
		// Issue #44: a grpc runner must not carry db-only fields.
		"grpc runner with db field": `version: "1"
suite: {name: x}
runners:
  svc: {type: grpc, target: "localhost:50051", dsn: "sqlite:./x.db"}
scenarios:
  - name: a
    steps: [{run: {command: echo}}]`,
		// Issue #44: a db runner must not carry the http base_url field.
		"db runner with base_url": `version: "1"
suite: {name: x}
runners:
  store: {type: db, dsn: "sqlite:./x.db", base_url: "http://x"}
scenarios:
  - name: a
    steps: [{run: {command: echo}}]`,
		// Issue #44: schema now requires the per-type mandatory field (db.dsn).
		"db runner missing dsn": `version: "1"
suite: {name: x}
runners:
  store: {type: db}
scenarios:
  - name: a
    steps: [{run: {command: echo}}]`,
		// Issue #16: pass_env entries are variable names, not key=value pairs.
		"pass_env with non-string entry": `version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps: [{run: {command: env, clear_env: true, pass_env: [{PATH: yes}]}}]`,
		// Issue #16: clear_env is a boolean, not a string.
		"clear_env with string value": `version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps: [{run: {command: env, clear_env: "yes"}}]`,
		// Issue #17: suite.timeout is a Go duration string, not a number.
		"suite timeout as number": `version: "1"
suite: {name: x, timeout: 30}
scenarios:
  - name: a
    steps: [{run: {command: echo}}]`,
		// Issue #18: the stdin mapping form sets exactly one of file/base64.
		"stdin with both file and base64": `version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps: [{run: {command: cat, stdin: {file: in.txt, base64: AAEC}}}]`,
		// Issue #18: unknown stdin keys are rejected.
		"stdin with unknown key": `version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps: [{run: {command: cat, stdin: {fil: in.txt}}}]`,
		// Issue #18: defaults.run.stdin is per-step input data.
		"defaults run stdin": `version: "1"
suite: {name: x}
defaults:
  run: {stdin: shared}
scenarios:
  - name: a
    steps: [{run: {command: cat}}]`,
		// Issue #19: an empty in list is rejected (minItems 1).
		"exit_code in empty list": `version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps:
      - run: {command: echo}
      - assert: {exit_code: {in: []}}`,
		// Issue #19: not and in cannot be combined (oneOf shapes).
		"exit_code not and in mixed": `version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps:
      - run: {command: echo}
      - assert: {exit_code: {not: 1, in: [0]}}`,
	}
	for name, src := range bad {
		t.Run(name, func(t *testing.T) {
			if err := s.Validate(yamlToAny(t, []byte(src))); err == nil {
				t.Errorf("schema accepted invalid spec %q", name)
			}
		})
	}
}

// fixtureSpecPath is the committed, fully-passing spec the golden examples under
// schema/examples/ are generated from. It must match the spec_path recorded in
// those examples so the golden byte-equality guard holds.
const fixtureSpecPath = "test/e2e/atago/version.atago.yaml"

// compileSchema compiles a published JSON Schema by path, failing the test if it
// is not valid JSON or not a valid JSON Schema (draft 2020-12).
func compileSchema(t *testing.T, path string) *jsonschema.Schema {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open schema %s: %v", path, err)
	}
	defer f.Close()

	doc, err := jsonschema.UnmarshalJSON(f)
	if err != nil {
		t.Fatalf("parse schema %s: %v", path, err)
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource(path, doc); err != nil {
		t.Fatalf("add resource %s: %v", path, err)
	}
	s, err := c.Compile(path)
	if err != nil {
		t.Fatalf("compile schema %s: %v", path, err)
	}
	return s
}

// readJSONAny reads a committed JSON document into the generic types the
// validator expects.
func readJSONAny(t *testing.T, path string) any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	return v
}

// TestOutputSchemas_Compile guards that both published output schemas parse and
// compile as valid JSON Schema.
func TestOutputSchemas_Compile(t *testing.T) {
	for _, path := range []string{
		"schema/manifest.schema.json",
		"schema/report.schema.json",
	} {
		if s := compileSchema(t, path); s == nil {
			t.Errorf("compileSchema(%s) returned nil", path)
		}
	}
}

// TestManifestExample_GoldenDrift regenerates the manifest example in-process
// from the committed fixture spec and asserts it byte-equals the committed
// golden. The manifest is fully deterministic, so any drift between the builder
// and the published example fails here.
func TestManifestExample_GoldenDrift(t *testing.T) {
	s, err := loader.Load(fixtureSpecPath)
	if err != nil {
		t.Fatalf("load %s: %v", fixtureSpecPath, err)
	}
	doc := manifest.Build([]manifest.Input{{Spec: s, Path: fixtureSpecPath}})
	got, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	got = append(got, '\n')

	want, err := os.ReadFile("schema/examples/manifest.example.json")
	if err != nil {
		t.Fatalf("read manifest example: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("manifest example is stale; regenerate schema/examples/manifest.example.json\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// TestManifestExample_Conforms validates the committed manifest example against
// the manifest schema.
func TestManifestExample_Conforms(t *testing.T) {
	s := compileSchema(t, "schema/manifest.schema.json")
	if err := s.Validate(readJSONAny(t, "schema/examples/manifest.example.json")); err != nil {
		t.Errorf("manifest example does not conform to schema:\n%v", err)
	}
}

// TestManifest_SuiteLifecycleConforms builds a manifest for a spec that exercises
// the suite lifecycle — env, setup (with a service), teardown, and the derived
// suite_variables — and validates the output against the manifest schema. The
// schema previously omitted every suite_* field except suite_timeout, so a
// suite-bearing manifest would have failed its own published schema (#244).
func TestManifest_SuiteLifecycleConforms(t *testing.T) {
	src := `
version: "1"
suite:
  name: life
  env:
    SHARED: shared-value
  setup:
    - run: {shell: true, command: "echo build ${srcdir}"}
    - service:
        name: db
        command: ./db
        env:
          DSN: "${dsn_ref}"
        ready:
          file: "${suitedir}/ready"
  teardown:
    - run: {shell: true, command: cleanup}
scenarios:
  - name: sc
    steps:
      - run: {shell: true, command: echo hi}
`
	s, err := loader.LoadBytes("life.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	doc := manifest.Build([]manifest.Input{{Spec: s, Path: "life.atago.yaml"}})
	blob, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	var v any
	if err := json.Unmarshal(blob, &v); err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}
	schema := compileSchema(t, "schema/manifest.schema.json")
	if err := schema.Validate(v); err != nil {
		t.Errorf("suite-lifecycle manifest does not conform to schema:\n%v", err)
	}
	// The suite service's env value and ready-probe references must surface.
	sv := doc.Specs[0].SuiteVariables
	for _, want := range []string{"srcdir", "dsn_ref", "suitedir"} {
		found := false
		for _, got := range sv {
			if got == want {
				found = true
			}
		}
		if !found {
			t.Errorf("suite_variables = %v, want it to include %q", sv, want)
		}
	}
	// The shell-enabled setup run must surface in the suite security notes,
	// proving the field round-trips through the schema above.
	sec := strings.Join(doc.Specs[0].SuiteSecurity, "\n")
	if !strings.Contains(sec, "shell execution enabled: echo build ${srcdir}") {
		t.Errorf("suite_security = %v, want the shell-enabled setup run flagged", doc.Specs[0].SuiteSecurity)
	}
}

// TestReportExample_Conforms validates the committed report example against the
// report schema. The report embeds wall-clock duration_ms fields, so the
// committed example zeroes them and is guarded by schema conformance rather than
// in-process byte-equality.
func TestReportExample_Conforms(t *testing.T) {
	s := compileSchema(t, "schema/report.schema.json")
	if err := s.Validate(readJSONAny(t, "schema/examples/report.example.json")); err != nil {
		t.Errorf("report example does not conform to schema:\n%v", err)
	}
}

// TestManifest_ExpectFailAndDeterministicConform builds a manifest for a spec
// carrying the two features the committed example never exercises — a
// scenario's `expect_fail:` and a step's `deterministic:` — and validates the
// real output against the manifest schema. Both shipped in the writer without a
// schema property, so a manifest that used them failed the schema atago
// publishes for it (#496).
func TestManifest_ExpectFailAndDeterministicConform(t *testing.T) {
	src := `
version: "1"
suite:
  name: known-bugs
scenarios:
  - name: a known bug
    expect_fail:
      reason: still broken
      issue: "https://github.com/nao1215/atago/issues/496"
    steps:
      - run:
          shell: true
          command: "exit 1"
          deterministic:
            runs: 3
            compare: [stdout, exit_code]
      - assert: {exit_code: 0}
`
	s, err := loader.LoadBytes("xf.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	doc := manifest.Build([]manifest.Input{{Spec: s, Path: "xf.atago.yaml"}})
	blob, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	var v any
	if err := json.Unmarshal(blob, &v); err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}
	// The manifest must actually carry both fields, or the conformance check
	// below would pass over a document that never exercised them.
	sc := doc.Specs[0].Scenarios[0]
	if sc.ExpectFail == nil {
		t.Fatal("manifest scenario carries no expect_fail")
	}
	if sc.Steps[0].Deterministic == nil {
		t.Fatal("manifest step carries no deterministic")
	}
	if err := compileSchema(t, "schema/manifest.schema.json").Validate(v); err != nil {
		t.Errorf("expect_fail/deterministic manifest does not conform to schema:\n%v", err)
	}
}

// TestReport_FeatureFieldsConform renders a real `--report json` document for
// the three features the committed example never exercises — a spec that failed
// to load, a suite whose setup failed, and a scenario declaring `expect_fail:` —
// and validates it against the report schema. Each shipped in the writer without
// a schema property, so any report using them failed the schema atago publishes
// for editors and CI consumers (#496).
func TestReport_FeatureFieldsConform(t *testing.T) {
	specs := map[string]string{
		"xf.atago.yaml": `
version: "1"
suite:
  name: known-bug
scenarios:
  - name: a known bug
    expect_fail:
      reason: still broken
      issue: "https://github.com/nao1215/atago/issues/496"
    steps:
      - run: {shell: true, command: "exit 1"}
      - assert: {exit_code: 0}
`,
		"setup.atago.yaml": `
version: "1"
suite:
  name: broken-setup
  setup:
    - run: {shell: true, command: "exit 1"}
    - assert: {exit_code: 0}
scenarios:
  - name: never runs
    steps:
      - run: {shell: true, command: "echo hi"}
`,
	}
	var results []*engine.SuiteResult
	for _, path := range []string{"xf.atago.yaml", "setup.atago.yaml"} {
		sp, err := loader.LoadBytes(path, []byte(specs[path]))
		if err != nil {
			t.Fatalf("load %s: %v", path, err)
		}
		results = append(results, engine.New().Run(context.Background(), sp, path))
	}
	var buf bytes.Buffer
	if err := report.Render(&buf, report.FormatJSON, results,
		report.WithLoadFailures(report.LoadFailure{SpecPath: "broken.atago.yaml", Message: "yaml: line 3: mapping values are not allowed"})); err != nil {
		t.Fatalf("render json report: %v", err)
	}
	// Every field under test must really be in the document, or conformance
	// would be asserted over a report that exercised none of them.
	for _, want := range []string{`"load_failures"`, `"setup_failures"`, `"expect_fail"`} {
		if !strings.Contains(buf.String(), want) {
			t.Fatalf("rendered report carries no %s:\n%s", want, buf.String())
		}
	}
	var v any
	if err := json.Unmarshal(buf.Bytes(), &v); err != nil {
		t.Fatalf("unmarshal report: %v", err)
	}
	if err := compileSchema(t, "schema/report.schema.json").Validate(v); err != nil {
		t.Errorf("report does not conform to schema:\n%v", err)
	}
}

// TestSchema_AssertTargetsAreAllDeclared is the last of the assert-target
// coverage guards: the published schema is what an editor validates a spec
// against, so a target the runtime accepts but the schema omits makes a correct
// spec light up red in the author's editor — and one listed in `properties` but
// missing from `anyOf` would be accepted only in combination with another target.
// Both halves are derived from spec.AllAssertTargets rather than restated here.
func TestSchema_AssertTargetsAreAllDeclared(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("schema/atago.schema.json")
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	// Navigated as a generic document rather than decoded into structs: the keys
	// are JSON Schema's own ($defs, anyOf), not names atago chose.
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode schema: %v", err)
	}
	defs, ok := doc["$defs"].(map[string]any)
	if !ok {
		t.Fatal("schema has no $defs object")
	}
	assertDef, ok := defs["assert"].(map[string]any)
	if !ok {
		t.Fatal("schema has no $defs/assert object")
	}
	properties, ok := assertDef["properties"].(map[string]any)
	if !ok {
		t.Fatal("$defs/assert has no properties object")
	}
	alternatives, ok := assertDef["anyOf"].([]any)
	if !ok {
		t.Fatal("$defs/assert has no anyOf list")
	}
	required := map[string]bool{}
	for _, alt := range alternatives {
		m, ok := alt.(map[string]any)
		if !ok {
			continue
		}
		names, ok := m["required"].([]any)
		if !ok {
			continue
		}
		for _, n := range names {
			if s, ok := n.(string); ok {
				required[s] = true
			}
		}
	}
	for _, target := range spec.AllAssertTargets() {
		if _, ok := properties[string(target)]; !ok {
			t.Errorf("assert target %q has no property in $defs/assert; an editor would reject a valid spec", target)
		}
		if !required[string(target)] {
			t.Errorf("assert target %q is not one of the anyOf alternatives; a spec setting only it would fail validation", target)
		}
	}
	for key := range properties {
		if !required[key] {
			t.Errorf("$defs/assert declares %q but no anyOf alternative requires it", key)
		}
	}
}

// TestSchema_PTYSessionActionsAreExclusive proves the schema enforces the same
// one-of rule the loader does (#379). The description has always said a session
// entry carries exactly one action; without the constraint an editor validating
// against the schema would accept a spec atago then rejects at load, which is
// the worst order to learn it in.
func TestSchema_PTYSessionActionsAreExclusive(t *testing.T) {
	s := loadSchema(t)
	valid := `version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps:
      - pty:
          command: mytool
          session:
            - expect: "ready"
            - send: {key: enter}
            - resize: {rows: 40, cols: 120}
            - expect_screen: {contains: "done"}`
	if err := s.Validate(yamlToAny(t, []byte(valid))); err != nil {
		t.Errorf("schema rejected a valid pty session:\n%v", err)
	}

	for name, src := range map[string]string{
		"expect and send": `version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps:
      - pty:
          command: mytool
          session:
            - {expect: "ready", send: "yo"}`,
		"resize and send": `version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps:
      - pty:
          command: mytool
          session:
            - {resize: {rows: 4, cols: 8}, send: "yo"}`,
		"no action at all": `version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps:
      - pty:
          command: mytool
          session:
            - {}`,
	} {
		if err := s.Validate(yamlToAny(t, []byte(src))); err == nil {
			t.Errorf("schema accepted a session entry with %s; it must carry exactly one action", name)
		}
	}
}
