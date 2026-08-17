package report

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/engine"
	"github.com/nao1215/atago/internal/spec"
)

// This file is the run-level twin of the spec package's Decides tests: a
// catalog of every fact a run can carry — a verdict class, a teardown failure,
// a rewritten golden, a load failure, the evidence paths — with an explicit
// decision for every report format: either a substring proving the format
// carries it, or the recorded reason it deliberately does not. The recurring
// bug this closes off is the partially wired fact: a load failure reached only
// the console (#120/#479), a failed teardown only console and json (#511), the
// artifact and service-log paths the same pair (#528), and the snapshot
// rewrite count left no trace anywhere (#532/#542) — four rounds of the same
// defect, each fixed by hand-wiring one fact into the formats it missed. A
// fact added to this catalog is red for every format nobody decided about.
//
// Two reflection guards anchor the catalog to real structs so it cannot rot:
// every renderOptions field and every engine.Counts field must be claimed by a
// catalog entry (or exempted with a reason), so a new run-level option or a
// new verdict class fails these tests until someone decides how each format
// reports it.

// factProof is one format's decision about one fact.
type factProof struct {
	// contains proves the fact is present in the rendered output.
	contains string
	// absent, when set alongside contains, additionally proves an output the
	// fact must NOT produce (an allow-flag removing a synthesized entry).
	absent string
	// exempt records why this format deliberately omits the fact. Exactly one
	// of contains/exempt must be set.
	exempt string
}

// reportFact is one run-level fact: the fixture that carries it and the
// decision per format.
type reportFact struct {
	results []*engine.SuiteResult
	opts    []Option
	proofs  map[Format]factProof
	// countAnchor / optionAnchor tie the fact to the struct field whose
	// existence forces it to be cataloged (see the reflection guards).
	countAnchor  string
	optionAnchor []string
}

// suiteOf wraps scenarios into the single-suite fixture the matrix renders.
func suiteOf(scs ...engine.ScenarioResult) []*engine.SuiteResult {
	return []*engine.SuiteResult{{
		Suite:     "matrix",
		SpecPath:  "matrix.atago.yaml",
		Status:    engine.StatusPassed,
		Scenarios: scs,
	}}
}

// failingCheck builds the failed-assertion evidence the failure facts share.
func failingCheck() *assert.CheckResult {
	return &assert.CheckResult{
		Desc:          "assert stdout contains \"never\"",
		Expected:      "EXPECTED-MARKER",
		Actual:        "ACTUAL-MARKER",
		Hint:          "HINT-MARKER",
		ArtifactFiles: []assert.ArtifactFile{{Role: "actual", Path: "art/PATH-MARKER.txt"}},
	}
}

func factCatalog() map[string]reportFact {
	passed := suiteOf(engine.ScenarioResult{Name: "greets", Status: engine.StatusPassed,
		Steps: []engine.StepResult{{Kind: spec.StepAssert, Checks: []*assert.CheckResult{{OK: true, Desc: "assert exit_code is 0"}}}}})
	failed := suiteOf(engine.ScenarioResult{Name: "greets", Status: engine.StatusFailed,
		ServiceLogs: []engine.ServiceLog{{Name: "svc", Path: "logs/SVCLOG-MARKER.log"}},
		Steps:       []engine.StepResult{{Kind: spec.StepAssert, Checks: []*assert.CheckResult{failingCheck()}}}})
	failed[0].Status = engine.StatusFailed
	errored := suiteOf(engine.ScenarioResult{Name: "greets", Status: engine.StatusError,
		Steps: []engine.StepResult{{Kind: spec.StepRun, ErrMsg: "ERRMSG-MARKER"}}})
	errored[0].Status = engine.StatusError
	skipped := suiteOf(engine.ScenarioResult{Name: "greets", Status: engine.StatusSkipped, SkipReason: "SKIPREASON-MARKER"})
	flaky := suiteOf(engine.ScenarioResult{Name: "greets", Status: engine.StatusFlaky, Attempts: 2,
		Steps: []engine.StepResult{{Kind: spec.StepAssert, Checks: []*assert.CheckResult{{OK: true, Desc: "assert exit_code is 0"}}}}})
	xfail := suiteOf(engine.ScenarioResult{Name: "greets", Status: engine.StatusXFail,
		ExpectFail: &spec.ExpectFail{Reason: "XFREASON-MARKER"}})
	xpass := suiteOf(engine.ScenarioResult{Name: "greets", Status: engine.StatusXPass,
		ExpectFail: &spec.ExpectFail{Reason: "XFREASON-MARKER"}})
	scenarioTeardown := suiteOf(engine.ScenarioResult{Name: "greets", Status: engine.StatusPassed,
		Teardown: []engine.StepResult{{Kind: spec.StepAssert, Checks: []*assert.CheckResult{{Desc: "TDDESC-MARKER"}}}}})
	suiteTeardown := suiteOf(engine.ScenarioResult{Name: "greets", Status: engine.StatusPassed})
	suiteTeardown[0].Teardown = []engine.StepResult{{Kind: spec.StepAssert, Checks: []*assert.CheckResult{{Desc: "SUITETD-MARKER"}}}}

	return map[string]reportFact{
		"passed scenario": {
			results: passed, countAnchor: "Passed",
			proofs: map[Format]factProof{
				FormatConsole: {contains: "1 passed"},
				FormatJSON:    {contains: `"status": "passed"`},
				FormatJUnit:   {contains: `failures="0"`},
				FormatTAP:     {contains: "ok 1 - "},
				FormatGHA:     {contains: "1 passed"},
			},
		},
		"failed scenario": {
			results: failed, countAnchor: "Failed",
			proofs: map[Format]factProof{
				FormatConsole: {contains: "FAILED:"},
				FormatJSON:    {contains: `"status": "failed"`},
				FormatJUnit:   {contains: "<failure"},
				FormatTAP:     {contains: "not ok 1 - "},
				FormatGHA:     {contains: "::error"},
			},
		},
		"errored scenario": {
			results: errored, countAnchor: "Errored",
			proofs: map[Format]factProof{
				FormatConsole: {contains: "ERRMSG-MARKER"},
				FormatJSON:    {contains: `"status": "error"`},
				FormatJUnit:   {contains: "<error"},
				FormatTAP:     {contains: "not ok 1 - "},
				FormatGHA:     {contains: "::error"},
			},
		},
		"skipped scenario with its reason": {
			results: skipped, countAnchor: "Skipped",
			proofs: map[Format]factProof{
				FormatConsole: {contains: "1 skipped"},
				FormatJSON:    {contains: `"skip_reason": "SKIPREASON-MARKER"`},
				FormatJUnit:   {contains: "<skipped"},
				FormatTAP:     {contains: "# SKIP SKIPREASON-MARKER"},
				FormatGHA:     {contains: "1 skipped"},
			},
		},
		"flaky scenario with its instability": {
			results: flaky, countAnchor: "Flaky",
			proofs: map[Format]factProof{
				FormatConsole: {contains: "1 flaky"},
				FormatJSON:    {contains: `"status": "flaky"`},
				FormatJUnit:   {contains: "flakyFailure"},
				FormatTAP:     {contains: "flaky: passed after 2 attempts"},
				FormatGHA:     {contains: "flaky: passed after 2 attempts"},
			},
		},
		"expected failure with its reason": {
			results: xfail, countAnchor: "XFail",
			proofs: map[Format]factProof{
				FormatConsole: {contains: "XFAIL:"},
				FormatJSON:    {contains: `"status": "xfail"`},
				FormatJUnit:   {contains: "xfail: XFREASON-MARKER"},
				FormatTAP:     {contains: "# TODO XFREASON-MARKER"},
				FormatGHA:     {contains: "xfail: XFREASON-MARKER"},
			},
		},
		"unexpectedly passing known bug": {
			results: xpass, countAnchor: "XPass",
			proofs: map[Format]factProof{
				FormatConsole: {contains: "XPASS:"},
				FormatJSON:    {contains: `"status": "xpass"`},
				FormatJUnit:   {contains: "xpass: the known bug is fixed"},
				FormatTAP:     {contains: "xpass: the known bug is fixed"},
				FormatGHA:     {contains: "xpass: the known bug is fixed"},
			},
		},
		"failure evidence (hint, expected/actual)": {
			results: failed,
			proofs: map[Format]factProof{
				FormatConsole: {contains: "HINT-MARKER"},
				FormatJSON:    {contains: `"hint": "HINT-MARKER"`},
				FormatJUnit:   {contains: "Hint: HINT-MARKER"},
				FormatTAP:     {contains: "HINT-MARKER"},
				FormatGHA:     {contains: "HINT-MARKER"},
			},
		},
		"artifact sidecar paths": {
			results: failed,
			proofs: map[Format]factProof{
				FormatConsole: {contains: "art/PATH-MARKER.txt"},
				FormatJSON:    {contains: `"artifacts"`},
				FormatJUnit:   {contains: "Artifact (actual): art/PATH-MARKER.txt"},
				FormatTAP:     {contains: "Artifact (actual): art/PATH-MARKER.txt"},
				FormatGHA:     {contains: "Artifact (actual): art/PATH-MARKER.txt"},
			},
		},
		"preserved service log paths": {
			results: failed,
			proofs: map[Format]factProof{
				FormatConsole: {contains: "logs/SVCLOG-MARKER.log"},
				FormatJSON:    {contains: `"service_logs"`},
				FormatJUnit:   {contains: "Service log (svc): logs/SVCLOG-MARKER.log"},
				FormatTAP:     {contains: "Service log (svc): logs/SVCLOG-MARKER.log"},
				FormatGHA:     {contains: "Service log (svc): logs/SVCLOG-MARKER.log"},
			},
		},
		"scenario teardown failure": {
			results: scenarioTeardown,
			proofs: map[Format]factProof{
				FormatConsole: {contains: "TEARDOWN FAILED:"},
				FormatJSON:    {contains: `"teardown_failures"`},
				FormatJUnit:   {contains: "teardown failed (the verdict is decided by the steps alone):"},
				FormatTAP:     {contains: "# teardown failed: TDDESC-MARKER"},
				FormatGHA:     {contains: "teardown failed: TDDESC-MARKER"},
			},
		},
		"suite teardown failure": {
			results: suiteTeardown,
			proofs: map[Format]factProof{
				FormatConsole: {contains: "SUITE TEARDOWN FAILED:"},
				FormatJSON:    {contains: "SUITETD-MARKER"},
				FormatJUnit:   {contains: "suite teardown failed (teardown outcomes never change the suite status):"},
				FormatTAP:     {contains: "# suite teardown failed: SUITETD-MARKER"},
				FormatGHA:     {contains: "suite teardown failed: SUITETD-MARKER"},
			},
		},
		"spec that failed to load": {
			results: passed, optionAnchor: []string{"loadFailures"},
			opts: []Option{WithLoadFailures(LoadFailure{SpecPath: "bad.atago.yaml", Message: "LOADMSG-MARKER"})},
			proofs: map[Format]factProof{
				FormatConsole: {contains: "failed to load"},
				FormatJSON:    {contains: `"load_failures"`},
				FormatJUnit:   {contains: "spec failed to load"},
				FormatTAP:     {contains: "spec failed to load"},
				FormatGHA:     {contains: "spec failed to load"},
			},
		},
		"rewritten snapshot goldens": {
			results: passed, optionAnchor: []string{"snapshotsUpdated"},
			opts: []Option{WithSnapshotsUpdated(3)},
			proofs: map[Format]factProof{
				FormatConsole: {contains: "3 snapshots updated"},
				FormatJSON:    {contains: `"snapshots_updated": 3`},
				FormatJUnit:   {exempt: "deliberate (#542): surefire's schema has no run-level annotation slot, and hanging the count on an arbitrary testsuite would misattribute it"},
				FormatTAP:     {contains: "# 3 snapshots updated"},
				FormatGHA:     {contains: "3 snapshots updated"},
			},
		},
		"run wall-clock time": {
			results: passed, optionAnchor: []string{"elapsed", "hasElapsed"},
			opts: []Option{WithElapsed(2 * time.Second)},
			proofs: map[Format]factProof{
				FormatConsole: {contains: "(2s)"},
				FormatJSON:    {exempt: "the machine formats carry per-suite durations from the results; the run's wall clock is the console summary's nicety"},
				FormatJUnit:   {exempt: "same: testsuite time attributes come from the per-suite durations"},
				FormatTAP:     {exempt: "TAP carries points and diagnostics, not a run duration"},
				FormatGHA:     {exempt: "the Actions UI shows the job's own duration; the summary line repeats the counts only"},
			},
		},
		"accepted flakiness (--allow-flaky)": {
			results: flaky, optionAnchor: []string{"allowFlaky"},
			opts: []Option{WithAllowFlaky(true)},
			proofs: map[Format]factProof{
				FormatConsole: {contains: "PASSED"},
				FormatJSON:    {exempt: "the per-scenario classification is identical either way; the flag moves only the exit code and the console verdict word"},
				FormatJUnit:   {exempt: "same: a flaky scenario is a flakyFailure with or without the flag"},
				FormatTAP:     {exempt: "same: the point is ok with the instability in its diagnostic either way"},
				FormatGHA:     {exempt: "same: the annotation stays a warning either way"},
			},
		},
		"accepted xpass (--allow-xpass)": {
			results: xpass, optionAnchor: []string{"allowXPass"},
			opts: []Option{WithAllowXPass(true)},
			proofs: map[Format]factProof{
				FormatConsole: {contains: "PASSED"},
				FormatJSON:    {contains: `"failures": []`},
				FormatJUnit:   {contains: "flakyFailure", absent: "<failure"},
				FormatTAP:     {exempt: "an ok TODO point already encodes unexpected-success; TAP has no verdict the flag could soften"},
				FormatGHA:     {contains: "::warning", absent: "::error"},
			},
		},
	}
}

// TestReportFacts_DecideEveryFormat renders each cataloged fact in every
// format and holds the decision: the proving substring is present (and the
// forbidden one absent), or the exemption carries a reason. A fact with no
// decision for some format — what every historical instance of this bug family
// was — fails with the question to answer.
func TestReportFacts_DecideEveryFormat(t *testing.T) {
	t.Parallel()
	for name, fact := range factCatalog() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			for _, f := range AllFormats() {
				proof, ok := fact.proofs[f]
				if !ok {
					t.Errorf("format %q: no decision for fact %q; add a proving substring or an exemption with its reason", f, name)
					continue
				}
				if (proof.contains == "") == (proof.exempt == "") {
					t.Errorf("format %q: fact %q must set exactly one of contains/exempt", f, name)
					continue
				}
				if proof.exempt != "" {
					continue
				}
				var b bytes.Buffer
				if err := Render(&b, f, fact.results, fact.opts...); err != nil {
					t.Fatalf("format %q: render: %v", f, err)
				}
				out := b.String()
				if !strings.Contains(out, proof.contains) {
					t.Errorf("format %q does not carry fact %q: want substring %q in:\n%s", f, name, proof.contains, out)
				}
				if proof.absent != "" && strings.Contains(out, proof.absent) {
					t.Errorf("format %q: fact %q forbids substring %q, present in:\n%s", f, name, proof.absent, out)
				}
			}
		})
	}
}

// TestReportFacts_CoverEveryCount anchors the catalog to engine.Counts: every
// verdict class the run tallies must be a cataloged fact, so a new status —
// the next StatusFlaky — cannot ship without deciding how each format reports
// it.
func TestReportFacts_CoverEveryCount(t *testing.T) {
	t.Parallel()
	anchored := map[string]bool{}
	for _, fact := range factCatalog() {
		if fact.countAnchor != "" {
			anchored[fact.countAnchor] = true
		}
	}
	rt := reflect.TypeOf(engine.Counts{})
	for i := range rt.NumField() {
		name := rt.Field(i).Name
		if !anchored[name] {
			t.Errorf("engine.Counts.%s has no cataloged fact; add one so every format decides how to report that verdict class", name)
		}
	}
}

// TestReportFacts_CoverEveryRenderOption anchors the catalog to renderOptions:
// every run-level input Render accepts must be claimed by a cataloged fact, so
// a new WithX option — the next snapshotsUpdated — cannot ship without
// deciding how each format reports what it carries.
func TestReportFacts_CoverEveryRenderOption(t *testing.T) {
	t.Parallel()
	anchored := map[string]bool{}
	for _, fact := range factCatalog() {
		for _, f := range fact.optionAnchor {
			anchored[f] = true
		}
	}
	rt := reflect.TypeOf(renderOptions{})
	for i := range rt.NumField() {
		name := rt.Field(i).Name
		if !anchored[name] {
			t.Errorf("renderOptions.%s has no cataloged fact; add one (or extend an existing fact's optionAnchor) so every format decides how to report it", name)
		}
	}
}

// TestAllFormats_RenderKnowsEveryFormat pins the format list to the Render
// dispatch from both sides: every listed format renders, and a format outside
// the list is refused — the list and the switch used to be two hand copies.
func TestAllFormats_RenderKnowsEveryFormat(t *testing.T) {
	t.Parallel()
	for _, f := range AllFormats() {
		var b bytes.Buffer
		if err := Render(&b, f, nil); err != nil {
			t.Errorf("format %q is listed but Render refuses it: %v", f, err)
		}
		if !f.Valid() {
			t.Errorf("format %q is listed but Valid() rejects it", f)
		}
	}
	if err := Render(&bytes.Buffer{}, Format("nope"), nil); err == nil {
		t.Error("an unlisted format was rendered")
	}
	if Format("nope").Valid() {
		t.Error("an unlisted format is Valid()")
	}
}
