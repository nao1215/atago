package docgen

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/loader"
)

func load(t *testing.T, name, body string) *Source {
	t.Helper()
	s, err := loader.LoadBytes(name, []byte(body))
	if err != nil {
		t.Fatalf("load %s: %v", name, err)
	}
	return &Source{Path: name, Spec: s}
}

func gen(t *testing.T, sources ...Source) string {
	t.Helper()
	var b bytes.Buffer
	if err := Generate(&b, sources); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

// --- #66 summary + TOC -----------------------------------------------------

func TestGenerate_SummaryAndTOC(t *testing.T) {
	a := load(t, "a.atago.yaml", `version: "1"
suite: {name: alpha}
scenarios:
  - name: first
    tags: [smoke, fast]
    steps: [{run: {command: "true"}}, {assert: {exit_code: 0}}]
  - name: second
    tags: [smoke]
    steps: [{run: {command: "true"}}, {assert: {exit_code: 0}}]
`)
	b := load(t, "b.atago.yaml", `version: "1"
suite: {name: beta}
scenarios:
  - name: third
    steps: [{run: {command: "true"}}, {assert: {exit_code: 0}}]
`)
	out := gen(t, *a, *b)

	for _, want := range []string{
		"## Summary",
		"2 suites · 3 scenarios",
		"`fast` (1), `smoke` (2)",
		"## Contents",
		"[alpha](#alpha)",
		"[first](#scenario-first)",
		"[third](#scenario-third)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

// TestGenerate_TOCAnchorsAreUnique verifies duplicate scenario names get the
// GitHub-style numeric suffix so TOC links stay distinct.
func TestGenerate_TOCAnchorsAreUnique(t *testing.T) {
	a := load(t, "a.atago.yaml", `version: "1"
suite: {name: dup}
scenarios:
  - name: same
    steps: [{run: {command: "true"}}, {assert: {exit_code: 0}}]
`)
	b := load(t, "b.atago.yaml", `version: "1"
suite: {name: dup2}
scenarios:
  - name: same
    steps: [{run: {command: "true"}}, {assert: {exit_code: 0}}]
`)
	out := gen(t, *a, *b)
	if !strings.Contains(out, "(#scenario-same)") {
		t.Errorf("missing first anchor:\n%s", out)
	}
	if !strings.Contains(out, "(#scenario-same-1)") {
		t.Errorf("duplicate scenario name did not get a -1 suffix:\n%s", out)
	}
}

// TestAnchors_MatchesGitHubRules pins the exported slugger drift guards rely
// on: punctuation drops, spaces hyphenate, underscores survive, and a repeated
// heading gets GitHub's incrementing suffix.
func TestAnchors_MatchesGitHubRules(t *testing.T) {
	got := Anchors([]string{
		"Test error handling: exit codes and stderr",
		"Pin a generator's whole output tree",
		"snake_case heading",
		"Repeat",
		"Repeat",
	})
	want := []string{
		"test-error-handling-exit-codes-and-stderr",
		"pin-a-generators-whole-output-tree",
		"snake_case-heading",
		"repeat",
		"repeat-1",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Anchors[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// --- #67 previews ----------------------------------------------------------

func TestGenerate_InputPreviews(t *testing.T) {
	s := load(t, "s.atago.yaml", `version: "1"
suite: {name: prev}
scenarios:
  - name: with inputs
    steps:
      - fixture: {file: data.csv, content: "id,name\n1,ada\n"}
      - run: {command: "sort", stdin: "banana\napple\n"}
      - assert: {exit_code: 0}
`)
	out := gen(t, *s)
	for _, want := range []string{
		"#### Inputs",
		"Fixture `data.csv`",
		"id,name",
		"stdin for `sort`",
		"banana",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing input preview %q:\n%s", want, out)
		}
	}
}

func TestGenerate_ExactOutputPreview(t *testing.T) {
	s := load(t, "s.atago.yaml", `version: "1"
suite: {name: exact}
scenarios:
  - name: multiline exact
    steps:
      - run: {command: "printf"}
      - assert:
          stdout:
            equals: "line one\nline two\nline three\n"
`)
	out := gen(t, *s)
	if !strings.Contains(out, "#### Expected output") {
		t.Errorf("missing expected-output section:\n%s", out)
	}
	if !strings.Contains(out, "line one") {
		t.Errorf("missing exact preview body:\n%s", out)
	}
}

func TestTruncatePreview_Marks(t *testing.T) {
	var lines []string
	for range 30 {
		lines = append(lines, "line")
	}
	got := truncatePreview(strings.Join(lines, "\n"))
	if !strings.Contains(got, "… (truncated, 10 more lines)") {
		t.Errorf("truncation marker missing/incorrect:\n%s", got)
	}
	if strings.Count(got, "\n") > previewMaxLines+1 {
		t.Errorf("preview exceeded line budget:\n%s", got)
	}
}

// --- #68 split-by-spec -----------------------------------------------------

func TestGenerateSplit(t *testing.T) {
	a := load(t, "dir/one.atago.yaml", `version: "1"
suite: {name: one}
scenarios:
  - name: s1
    steps: [{run: {command: "true"}}, {assert: {exit_code: 0}}]
`)
	b := load(t, "dir/two.atago.yaml", `version: "1"
suite: {name: two}
scenarios:
  - name: s2
    steps: [{run: {command: "true"}}, {assert: {exit_code: 0}}]
`)
	index, docs, err := GenerateSplit([]Source{*a, *b}, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 2 {
		t.Fatalf("docs = %d, want 2", len(docs))
	}
	if docs[0].Name != "one.md" || docs[1].Name != "two.md" {
		t.Errorf("names = %q, %q", docs[0].Name, docs[1].Name)
	}
	idx := string(index)
	for _, want := range []string{"# atago Behavior Specs — Index", "[one](one.md)", "[two](two.md)", "2 suites · 2 scenarios"} {
		if !strings.Contains(idx, want) {
			t.Errorf("index missing %q:\n%s", want, idx)
		}
	}
	if !bytes.Contains(docs[0].Content, []byte("Scenario: s1")) {
		t.Errorf("per-spec doc missing its scenario:\n%s", docs[0].Content)
	}
}

func TestSplitFilenames_Collisions(t *testing.T) {
	a := load(t, "x/dup.atago.yaml", `version: "1"
suite: {name: dupa}
scenarios: [{name: a, steps: [{run: {command: "true"}}, {assert: {exit_code: 0}}]}]
`)
	b := load(t, "y/dup.atago.yaml", `version: "1"
suite: {name: dupb}
scenarios: [{name: b, steps: [{run: {command: "true"}}, {assert: {exit_code: 0}}]}]
`)
	names := splitFilenames([]Source{*a, *b})
	if names[0] != "dup.md" || names[1] != "dup-1.md" {
		t.Errorf("collision names = %v, want [dup.md dup-1.md]", names)
	}
}

// TestGenerate_ThenGroupedByCommand: with several action steps, each Then
// bullet group opens with the command it checks, so "exit code is 0" and
// "exit code is not 0" cannot be confused across commands.
func TestGenerate_ThenGroupedByCommand(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
scenarios:
  - name: two commands, two outcomes
    steps:
      - run: {command: tool init}
      - assert: {exit_code: 0}
      - run: {command: tool status}
      - assert:
          exit_code:
            not: 0
`
	out := renderDoc(t, src)
	for _, w := range []string{
		"- after `tool init`:\n  - exit code is `0`",
		"- after `tool status`:\n  - exit code is not `0`",
	} {
		if !strings.Contains(out, w) {
			t.Errorf("doc output missing %q\n--- got ---\n%s", w, out)
		}
	}
}

// TestGenerate_SingleCommandThenStaysFlat: one action keeps the flat list — no
// "after ..." noise when there is nothing to disambiguate.
func TestGenerate_SingleCommandThenStaysFlat(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
scenarios:
  - name: one command
    steps:
      - run: {command: tool run}
      - assert: {exit_code: 0}
`
	out := renderDoc(t, src)
	if strings.Contains(out, "- after ") {
		t.Errorf("single-action scenario should keep a flat Then list\n--- got ---\n%s", out)
	}
	if !strings.Contains(out, "- exit code is `0`") {
		t.Errorf("flat Then bullet missing\n--- got ---\n%s", out)
	}
}

// TestGenerate_MatrixVarsExpandedInBody: a matrix instance renders its concrete
// row values in commands and assertions, matching its already-expanded name.
func TestGenerate_MatrixVarsExpandedInBody(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
scenarios:
  - name: "greets ${who}"
    matrix:
      - { who: Alice }
    steps:
      - run: {command: "greet ${who}"}
      - assert:
          stdout:
            contains: ${who}
`
	out := renderDoc(t, src)
	for _, w := range []string{"greet Alice", "stdout contains `Alice`"} {
		if !strings.Contains(out, w) {
			t.Errorf("doc output missing %q\n--- got ---\n%s", w, out)
		}
	}
	if strings.Contains(out, "${who}") {
		t.Errorf("matrix variable left unexpanded in doc body\n--- got ---\n%s", out)
	}
}

// TestGenerate_StoreStepDocumented: a store step appears in When as a comment,
// so a later ${name} reference is explained where the value is captured.
func TestGenerate_StoreStepDocumented(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
scenarios:
  - name: capture and reuse
    steps:
      - run: {command: tool id}
      - store:
          name: id
          from:
            stdout:
              json: {path: "$.id"}
      - run: {command: "tool get ${id}"}
      - assert: {exit_code: 0}
`
	out := renderDoc(t, src)
	if !strings.Contains(out, "# capture ${id} from stdout") {
		t.Errorf("store step not documented in When\n--- got ---\n%s", out)
	}
}

// renderDoc loads src and renders it as a doc, returning the Markdown.
func renderDoc(t *testing.T, src string) string {
	t.Helper()
	s, err := loader.LoadBytes("s.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	var b bytes.Buffer
	if err := Generate(&b, []Source{{Path: "s.atago.yaml", Spec: s}}); err != nil {
		t.Fatal(err)
	}
	return b.String()
}
