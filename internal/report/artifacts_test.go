package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/engine"
)

// artifactResults builds a suite with a single failed assertion that carries two
// sidecar artifact references, as the engine would after writing them (#48).
func artifactResults() []*engine.SuiteResult {
	return []*engine.SuiteResult{{
		Suite:    "s1",
		SpecPath: "dir/s1.atago.yaml",
		Status:   engine.StatusFailed,
		Scenarios: []engine.ScenarioResult{{
			Name:   "f",
			Status: engine.StatusFailed,
			Steps: []engine.StepResult{{Kind: "assert", Checks: []*assert.CheckResult{{
				OK: false, Desc: "assert stdout contains \"x\"", Expected: "x", Actual: "y", Hint: "missing x",
				ArtifactFiles: []assert.ArtifactFile{
					{Role: "actual", Path: "s1-abcd1234/f-0/step-01-stdout.actual.txt"},
					{Role: "expected", Path: "s1-abcd1234/f-0/step-01-stdout.expected.txt"},
				},
			}}}},
		}},
	}}
}

func TestRender_JSON_IncludesArtifactPaths(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	if err := Render(&b, FormatJSON, artifactResults()); err != nil {
		t.Fatalf("render: %v", err)
	}
	var doc struct {
		Suites []struct {
			Failures []struct {
				Artifacts []struct {
					Role string `json:"role"`
					Path string `json:"path"`
				} `json:"artifacts"`
			} `json:"failures"`
		} `json:"suites"`
	}
	if err := json.Unmarshal(b.Bytes(), &doc); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, b.String())
	}
	arts := doc.Suites[0].Failures[0].Artifacts
	if len(arts) != 2 {
		t.Fatalf("artifacts = %+v, want 2", arts)
	}
	if arts[0].Role != "actual" || !strings.HasSuffix(arts[0].Path, "stdout.actual.txt") {
		t.Errorf("actual artifact = %+v", arts[0])
	}
	if arts[1].Role != "expected" {
		t.Errorf("expected artifact role = %q", arts[1].Role)
	}
}

func TestRender_JSON_OmitsArtifactsWhenNone(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	if err := Render(&b, FormatJSON, sampleResults()); err != nil {
		t.Fatalf("render: %v", err)
	}
	if strings.Contains(b.String(), "\"artifacts\"") {
		t.Errorf("artifacts field should be omitted when empty:\n%s", b.String())
	}
}

func TestRender_JSON_IncludesServiceLogs(t *testing.T) {
	t.Parallel()
	results := []*engine.SuiteResult{{
		Suite:    "s1",
		SpecPath: "svc.atago.yaml",
		Status:   engine.StatusError,
		Scenarios: []engine.ScenarioResult{{
			Name:   "svc failed",
			Status: engine.StatusError,
			Steps:  []engine.StepResult{{Kind: "", ErrMsg: "service not ready"}},
			ServiceLogs: []engine.ServiceLog{
				{Name: "peer", Path: "s1-abcd1234/svc-failed-0/service-peer.log"},
			},
		}},
	}}
	var b bytes.Buffer
	if err := Render(&b, FormatJSON, results); err != nil {
		t.Fatalf("render: %v", err)
	}
	var doc struct {
		Suites []struct {
			Scenarios []struct {
				ServiceLogs []struct {
					Name string `json:"name"`
					Path string `json:"path"`
				} `json:"service_logs"`
			} `json:"scenarios"`
		} `json:"suites"`
	}
	if err := json.Unmarshal(b.Bytes(), &doc); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, b.String())
	}
	logs := doc.Suites[0].Scenarios[0].ServiceLogs
	if len(logs) != 1 || logs[0].Name != "peer" || !strings.HasSuffix(logs[0].Path, "service-peer.log") {
		t.Errorf("service_logs = %+v", logs)
	}
}

func TestRender_JSON_OmitsServiceLogsWhenNone(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	if err := Render(&b, FormatJSON, sampleResults()); err != nil {
		t.Fatalf("render: %v", err)
	}
	if strings.Contains(b.String(), "service_logs") {
		t.Errorf("service_logs should be omitted when empty:\n%s", b.String())
	}
}

func TestRender_Console_ShowsArtifactHints(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	if err := Render(&b, FormatConsole, artifactResults()); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := b.String()
	if !strings.Contains(out, "Artifacts:") {
		t.Errorf("console output missing Artifacts section:\n%s", out)
	}
	if !strings.Contains(out, "step-01-stdout.actual.txt") {
		t.Errorf("console output missing artifact path:\n%s", out)
	}
	// The console must stay compact: it references paths, not payload dumps.
	if strings.Count(out, "actual.txt") == 0 {
		t.Errorf("expected an artifact path reference")
	}
}
