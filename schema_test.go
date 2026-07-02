package main

import (
	"os"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/santhosh-tekuri/jsonschema/v6"
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
	}
	for name, src := range bad {
		t.Run(name, func(t *testing.T) {
			if err := s.Validate(yamlToAny(t, []byte(src))); err == nil {
				t.Errorf("schema accepted invalid spec %q", name)
			}
		})
	}
}
