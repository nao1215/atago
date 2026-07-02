package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/nao1215/atago/internal/loader"
	"github.com/nao1215/atago/internal/manifest"
)

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
