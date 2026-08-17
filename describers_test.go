package main

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/assertdesc"
	"github.com/nao1215/atago/internal/docgen"
	"github.com/nao1215/atago/internal/explain"
	"github.com/nao1215/atago/internal/loader"
	"github.com/nao1215/atago/internal/manifest"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/spectest"
)

// describerOutputs renders one spec through every describer atago ships, so a
// test can ask the same question of all of them.
type describerOutputs struct {
	explain  string
	doc      string
	manifest string
}

func describeAllWays(t *testing.T, path, src string) describerOutputs {
	t.Helper()
	s, err := loader.LoadBytes(path, []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	var ex strings.Builder
	if err := explain.Explain(&ex, s, path); err != nil {
		t.Fatalf("explain: %v", err)
	}
	var doc strings.Builder
	if err := docgen.Generate(&doc, []docgen.Source{{Path: path, Spec: s}}); err != nil {
		t.Fatalf("doc: %v", err)
	}
	// HTML escaping off so a redirect in a command reads as ">" rather than as
	// \u003e; the question here is which text each describer reports, not how
	// one of them encodes it.
	var m strings.Builder
	enc := json.NewEncoder(&m)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(manifest.Build([]manifest.Input{{Spec: s, Path: path}})); err != nil {
		t.Fatalf("manifest: %v", err)
	}
	return describerOutputs{explain: ex.String(), doc: doc.String(), manifest: m.String()}
}

func (d describerOutputs) each(fn func(name, out string)) {
	fn("explain", d.explain)
	fn("doc", d.doc)
	fn("manifest", d.manifest)
}

// TestDescribers_AgreeOnAnExpandedMatrixRow pins the answer every describer
// gives to one question: what does this expanded matrix row run? The row is
// bound at load time and the scenario NAME already carries its values, so a
// describer that leaves `${who}` in the step text is describing a command
// nobody runs — and two rows of one matrix then read identically under two
// different headings.
//
// Variables the row does not bind stay literal everywhere: `${env:...}` and a
// value a store captures are unknown before the run, and printing host state
// into a generated page is how a secret ends up committed.
func TestDescribers_AgreeOnAnExpandedMatrixRow(t *testing.T) {
	t.Parallel()
	const src = `
version: "1"
suite:
  name: matrix describers
scenarios:
  - name: greets ${who}
    matrix:
      - {who: alice}
      - {who: bob}
    steps:
      - run:
          shell: true
          command: "echo ${who} > ${who}.txt"
      - assert:
          exit_code: 0
          file: {path: "${who}.txt", contains: "${who}"}
  - name: keeps a runtime reference literal
    steps:
      - run: {command: "echo ${env:HOME}"}
      - assert: {exit_code: 0}
`
	out := describeAllWays(t, "m.atago.yaml", src)

	out.each(func(name, text string) {
		for _, row := range []string{"alice", "bob"} {
			if !strings.Contains(text, "echo "+row+" > "+row+".txt") {
				t.Errorf("%s does not render the %s row's command with its bound value:\n%s", name, row, text)
			}
		}
		if strings.Contains(text, "${who}") {
			t.Errorf("%s leaves a row-bound ${who} unexpanded:\n%s", name, text)
		}
		if !strings.Contains(text, "${env:HOME}") {
			t.Errorf("%s resolved a reference only the run can resolve; ${env:HOME} must stay literal:\n%s", name, text)
		}
	})
}

// unexpandedRef finds a live ${name} reference left in rendered output.
var unexpandedRef = regexp.MustCompile(`\$\{[a-zA-Z_][a-zA-Z0-9_]*\}`)

// TestDescribers_DistinguishAStrongAssertionFromAWeakOne is the property the
// manifest failed: a suite whose assertions were gutted produced a
// byte-identical document, because every assert was reduced to its target name.
// Two assertions on the same target must read differently in every describer,
// or none of them can be used to review what a suite actually checks.
func TestDescribers_DistinguishAStrongAssertionFromAWeakOne(t *testing.T) {
	t.Parallel()
	const strong = `
version: "1"
suite:
  name: same
scenarios:
  - name: checks
    steps:
      - run: {command: echo hi}
      - assert:
          exit_code: 0
          stdout: {equals: "hi\n"}
`
	const weak = `
version: "1"
suite:
  name: same
scenarios:
  - name: checks
    steps:
      - run: {command: echo hi}
      - assert:
          exit_code: {in: [0, 1, 2]}
          stdout: {not_contains: "zzz"}
`
	s := describeAllWays(t, "s.atago.yaml", strong)
	w := describeAllWays(t, "s.atago.yaml", weak)

	for _, c := range []struct{ name, strong, weak string }{
		{"explain", s.explain, w.explain},
		{"doc", s.doc, w.doc},
		{"manifest", s.manifest, w.manifest},
	} {
		if c.strong == c.weak {
			t.Errorf("%s renders a gutted suite identically to the strong one, so it cannot say what a suite checks:\n%s", c.name, c.strong)
		}
	}
}

// manifestAssertPhrases renders one target's specimen assertion the way the
// manifest records it.
func manifestAssertPhrases(t *testing.T, target spec.AssertTarget) []string {
	t.Helper()
	return assertdesc.Describe(spectest.AssertForTarget(target))
}

// TestDescribers_CoverEveryAssertTargetWithoutLeavingARawReference walks every
// assert target and proves each describer says something about it — the
// manifest used to emit only the target's name, which reads as a category
// rather than as a claim.
func TestDescribers_CoverEveryAssertTargetWithoutLeavingARawReference(t *testing.T) {
	t.Parallel()
	for _, target := range spec.AllAssertTargets() {
		got := manifestAssertPhrases(t, target)
		if len(got) == 0 {
			t.Errorf("target %q produces no manifest description", target)
			continue
		}
		for _, phrase := range got {
			if phrase == string(target) {
				t.Errorf("target %q is described by its own name, which says nothing about the assertion", target)
			}
			if unexpandedRef.MatchString(phrase) {
				t.Errorf("target %q renders an unexpanded reference: %q", target, phrase)
			}
		}
	}
}
