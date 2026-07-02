package manifest

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/spec"
)

// fakeLocator is a deterministic SourceLocator for tests: it returns predictable
// positions so the manifest's source wiring can be asserted without parsing YAML.
type fakeLocator struct{}

func (fakeLocator) SuitePos() (int, int)             { return 2, 3 }
func (fakeLocator) RunnerPos(name string) (int, int) { return 10, 3 }
func (fakeLocator) ScenarioPos(i int) (int, int)     { return 20 + i, 5 }
func (fakeLocator) StepPos(scIdx, stepIdx int) (int, int) {
	return 30 + scIdx*10 + stepIdx, 7
}

func sampleSpec() *spec.Spec {
	return &spec.Spec{
		Version: "1",
		Suite:   spec.Suite{Name: "s"},
		Runners: map[string]spec.Runner{
			"api": {Type: "http", BaseURL: "https://example.com"},
		},
		Scenarios: []spec.Scenario{
			{
				Name:        "first",
				SourceIndex: 0,
				Steps: []spec.Step{
					{Run: &spec.Run{Command: "true"}},
				},
			},
		},
	}
}

func TestBuild_WithSourceLocator(t *testing.T) {
	doc := Build([]Input{{Spec: sampleSpec(), Path: "s.atago.yaml", Source: fakeLocator{}}})
	sp := doc.Specs[0]
	if sp.Source == nil || sp.Source.Line != 2 || sp.Source.Column != 3 {
		t.Fatalf("suite source = %+v, want line 2 col 3", sp.Source)
	}
	if len(sp.Runners) != 1 || sp.Runners[0].Source == nil || sp.Runners[0].Source.Line != 10 {
		t.Fatalf("runner source = %+v, want line 10", sp.Runners)
	}
	sc := sp.Scenarios[0]
	if sc.Source == nil || sc.Source.Line != 20 {
		t.Fatalf("scenario source = %+v, want line 20", sc.Source)
	}
	if sc.Steps[0].Source == nil || sc.Steps[0].Source.Line != 30 {
		t.Fatalf("step source = %+v, want line 30", sc.Steps[0].Source)
	}
}

// TestBuild_WithoutSourceLocator_OmitsSource is the compatibility guard: with no
// locator (as `atago list` uses), no `source` field is emitted anywhere.
func TestBuild_WithoutSourceLocator_OmitsSource(t *testing.T) {
	doc := Build([]Input{{Spec: sampleSpec(), Path: "s.atago.yaml"}})
	payload, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), "\"source\"") {
		t.Errorf("manifest emitted a source field without a locator:\n%s", payload)
	}
}

// TestBuild_SourceDeterministic guards that source-annotated output is stable
// across repeated builds.
func TestBuild_SourceDeterministic(t *testing.T) {
	build := func() string {
		doc := Build([]Input{{Spec: sampleSpec(), Path: "s.atago.yaml", Source: fakeLocator{}}})
		b, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	first := build()
	second := build()
	if first != second {
		t.Error("source-annotated manifest is not deterministic")
	}
}
