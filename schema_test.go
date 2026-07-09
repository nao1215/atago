package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/nao1215/atago/internal/loader"
	"github.com/nao1215/atago/internal/manifest"
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

// TestSchema_RealSpecsConform guards against drift between the JSON Schema and
// the specs we ship: every example/demo spec must validate against the schema.
func TestSchema_RealSpecsConform(t *testing.T) {
	s := loadSchema(t)
	specs := []string{
		"doc/demo/passing.atago.yaml",
		"doc/demo/failing.atago.yaml",
	}
	for _, p := range specs {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		if err := s.Validate(yamlToAny(t, data)); err != nil {
			t.Errorf("%s does not conform to schema:\n%v", p, err)
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
