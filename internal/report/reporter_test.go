package report

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/engine"
)

// erroredResults builds a suite carrying an errored scenario so the error-path
// branches of every reporter are exercised.
func erroredResults() *engine.SuiteResult {
	return &engine.SuiteResult{
		Suite:    "e1",
		SpecPath: "e1.atago.yaml",
		Status:   engine.StatusError,
		Duration: 2 * time.Millisecond,
		Scenarios: []engine.ScenarioResult{
			{Name: "boom", Status: engine.StatusError, Duration: time.Millisecond, Steps: []engine.StepResult{
				{Index: 0, Kind: "run", ErrMsg: "command not found"},
			}},
		},
	}
}

func TestFormat_Valid(t *testing.T) {
	t.Parallel()
	for _, f := range []Format{FormatConsole, FormatJSON, FormatJUnit, FormatGHA, FormatTAP} {
		if !f.Valid() {
			t.Errorf("%q should be valid", f)
		}
	}
	if Format("bogus").Valid() {
		t.Error("bogus format should be invalid")
	}
}

// render is a single-suite helper around Render (the sole rendering entry point
// after the dead Reporter/For path was removed, issue #25).
func render(t *testing.T, f Format, res *engine.SuiteResult) string {
	t.Helper()
	var b bytes.Buffer
	if err := Render(&b, f, []*engine.SuiteResult{res}); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

// TestRender_PerFormat exercises each format's rendering via the
// single Render entry point, on both a failing and an errored suite.
func TestRender_PerFormat(t *testing.T) {
	t.Parallel()
	failing := sampleResults()[0]
	errored := erroredResults()

	t.Run("console failing", func(t *testing.T) {
		t.Parallel()
		if out := render(t, FormatConsole, failing); !strings.Contains(out, "FAILED:") {
			t.Errorf("console missing FAILED:\n%s", out)
		}
	})

	t.Run("console errored", func(t *testing.T) {
		t.Parallel()
		out := render(t, FormatConsole, errored)
		if !strings.Contains(out, "ERROR:") || !strings.Contains(out, "command not found") {
			t.Errorf("console missing error block:\n%s", out)
		}
	})

	t.Run("json failing", func(t *testing.T) {
		t.Parallel()
		var doc jsonDocument
		out := render(t, FormatJSON, failing)
		if err := json.Unmarshal([]byte(out), &doc); err != nil {
			t.Fatalf("JSON invalid: %v\n%s", err, out)
		}
		if doc.SchemaVersion != jsonSchemaVersion {
			t.Errorf("schema_version = %q, want %q", doc.SchemaVersion, jsonSchemaVersion)
		}
		if len(doc.Suites) != 1 {
			t.Fatalf("suites = %d, want 1", len(doc.Suites))
		}
		rep := doc.Suites[0]
		if rep.Suite != "s1" || len(rep.Scenarios) != 3 || len(rep.Failures) != 1 {
			t.Errorf("json shape = suite %q scenarios %d failures %d", rep.Suite, len(rep.Scenarios), len(rep.Failures))
		}
	})

	t.Run("json errored carries error field", func(t *testing.T) {
		t.Parallel()
		var doc jsonDocument
		if err := json.Unmarshal([]byte(render(t, FormatJSON, errored)), &doc); err != nil {
			t.Fatal(err)
		}
		if len(doc.Suites) != 1 {
			t.Fatalf("suites = %d, want 1", len(doc.Suites))
		}
		rep := doc.Suites[0]
		if len(rep.Failures) != 1 || rep.Failures[0].Error != "command not found" {
			t.Errorf("json error failure = %+v", rep.Failures)
		}
	})

	t.Run("junit errored", func(t *testing.T) {
		t.Parallel()
		var root junitTestsuites
		if err := xml.Unmarshal([]byte(render(t, FormatJUnit, errored)), &root); err != nil {
			t.Fatalf("JUnit invalid: %v", err)
		}
		if root.Errors != 1 {
			t.Errorf("junit errors = %d, want 1", root.Errors)
		}
	})

	t.Run("gha errored", func(t *testing.T) {
		t.Parallel()
		if out := render(t, FormatGHA, errored); !strings.Contains(out, "::error title=e1 / boom::") {
			t.Errorf("gha missing error annotation:\n%s", out)
		}
	})

	t.Run("tap errored", func(t *testing.T) {
		t.Parallel()
		out := render(t, FormatTAP, errored)
		if !strings.Contains(out, "not ok 1 - e1 / boom") || !strings.Contains(out, "command not found") {
			t.Errorf("tap missing errored point:\n%s", out)
		}
	})
}

// TestRender_JSONStableShape verifies the top-level shape is identical whether
// one suite or many are rendered: always {"schema_version","suites":[...]} with
// suites as an array (#43). This is the contract machine consumers rely on.
func TestRender_JSONStableShape(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		results   []*engine.SuiteResult
		wantCount int
	}{
		"single suite": {sampleResults(), 1},
		"multi suite":  {append(sampleResults(), erroredResults()), 2},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var b bytes.Buffer
			if err := Render(&b, FormatJSON, tc.results); err != nil {
				t.Fatal(err)
			}
			var doc jsonDocument
			if err := json.Unmarshal(b.Bytes(), &doc); err != nil {
				t.Fatalf("top-level JSON must be the wrapper object: %v\n%s", err, b.String())
			}
			if doc.SchemaVersion != jsonSchemaVersion {
				t.Errorf("schema_version = %q, want %q", doc.SchemaVersion, jsonSchemaVersion)
			}
			if len(doc.Suites) != tc.wantCount {
				t.Errorf("got %d suites, want %d", len(doc.Suites), tc.wantCount)
			}
			// A bare array or bare object must NOT parse at the top level.
			var arr []jsonReport
			if json.Unmarshal(b.Bytes(), &arr) == nil {
				t.Errorf("top-level should not be a bare array:\n%s", b.String())
			}
		})
	}
}

func TestRender_UnknownFormat(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	if err := Render(&b, Format("nope"), sampleResults()); err == nil {
		t.Error("expected error for unknown format")
	}
}
