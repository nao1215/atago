package loader

import (
	"os"
	"path/filepath"
	"testing"
)

const sourceSpec = `version: "1"
suite:
  name: located
runners:
  api:
    type: http
    base_url: https://example.com
scenarios:
  - name: first
    steps:
      - run: {command: "true"}
      - assert: {exit_code: 0}
  - name: second
    matrix:
      - n: "1"
      - n: "2"
    steps:
      - run: {command: "true"}
`

func writeTemp(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "s.atago.yaml")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadWithSource_Positions(t *testing.T) {
	p := writeTemp(t, sourceSpec)
	s, src, err := LoadWithSource(p)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if src == nil {
		t.Fatal("nil source locator")
	}

	if line, _ := src.SuitePos(); line != 3 {
		t.Errorf("suite line = %d, want 3", line)
	}
	// A runner path resolves to its value node (the first field of the mapping),
	// i.e. the `type:` line just below the `api:` key.
	if line, _ := src.RunnerPos("api"); line != 6 {
		t.Errorf("runner api line = %d, want 6", line)
	}
	// Authored scenario 0 ("first") starts on line 9 (its name).
	if line, _ := src.ScenarioPos(0); line != 9 {
		t.Errorf("scenario[0] line = %d, want 9", line)
	}
	// Authored scenario 1 ("second") name is on line 13.
	if line, _ := src.ScenarioPos(1); line != 13 {
		t.Errorf("scenario[1] line = %d, want 13", line)
	}
	// Step 0 of scenario 0 is the run step on line 11.
	if line, _ := src.StepPos(0, 0); line != 11 {
		t.Errorf("scenario[0].steps[0] line = %d, want 11", line)
	}

	// Every matrix instance of "second" retains its authored SourceIndex (1), so
	// they all resolve to the same authored line.
	var matrixInstances int
	for i := range s.Scenarios {
		if s.Scenarios[i].SourceIndex == 1 {
			matrixInstances++
			if line, _ := src.ScenarioPos(s.Scenarios[i].SourceIndex); line != 13 {
				t.Errorf("matrix instance line = %d, want 13", line)
			}
		}
	}
	if matrixInstances != 2 {
		t.Errorf("expected 2 matrix instances sharing SourceIndex 1, got %d", matrixInstances)
	}
}

func TestLoadWithSource_UnknownPositionsAreZero(t *testing.T) {
	p := writeTemp(t, sourceSpec)
	_, src, err := LoadWithSource(p)
	if err != nil {
		t.Fatal(err)
	}
	// A runner that does not exist resolves to an unknown (zero) position.
	if line, col := src.RunnerPos("nope"); line != 0 || col != 0 {
		t.Errorf("unknown runner pos = (%d,%d), want (0,0)", line, col)
	}
	// A scenario index past the end resolves to unknown.
	if line, _ := src.ScenarioPos(99); line != 0 {
		t.Errorf("out-of-range scenario line = %d, want 0", line)
	}
}
