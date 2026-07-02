package manifest

import (
	"encoding/json"
	"testing"

	"github.com/nao1215/atago/internal/loader"
)

// TestBuild_Golden pins the exact serialized shape and field ordering of a small
// spec's manifest, so any drift in the document format is caught (#49).
func TestBuild_Golden(t *testing.T) {
	t.Parallel()
	const src = `
version: "1"
suite:
  name: greet
scenarios:
  - name: prints a greeting
    steps:
      - run: {command: echo hi}
      - assert: {stdout: {contains: hi}}
`
	s, err := loader.LoadBytes("greet.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	got, err := json.MarshalIndent(Build([]Input{{Spec: s, Path: "greet.atago.yaml"}}), "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	const want = `{
  "schema_version": "1",
  "specs": [
    {
      "spec_path": "greet.atago.yaml",
      "suite": "greet",
      "network": {
        "policy": "unrestricted"
      },
      "scenarios": [
        {
          "name": "prints a greeting",
          "steps": [
            {
              "index": 0,
              "kind": "run",
              "action": "run echo hi",
              "command": "echo hi"
            },
            {
              "index": 1,
              "kind": "assert",
              "action": "assert stdout",
              "target": "stdout"
            }
          ]
        }
      ]
    }
  ]
}`
	if string(got) != want {
		t.Errorf("manifest golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
