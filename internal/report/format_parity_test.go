package report

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/engine"
)

// mixedResults builds one suite carrying every terminal scenario status —
// passed, failed, errored, skipped, and flaky — so a single fixture exercises
// each format's classification of all five. The flaky scenario mirrors what the
// engine actually returns from --retry-failed: the recovering (passing)
// attempt's steps, with StatusFlaky and Attempts > 1 (see engine attempts.go).
// So its retained steps hold no failing check — the report layer must still keep
// it out of the failure bucket in every format.
func mixedResults() *engine.SuiteResult {
	return &engine.SuiteResult{
		Suite:    "mix",
		SpecPath: "mix.atago.yaml",
		Status:   engine.StatusFailed,
		Duration: 5 * time.Millisecond,
		Scenarios: []engine.ScenarioResult{
			{Name: "p", Status: engine.StatusPassed, Duration: time.Millisecond,
				Steps: []engine.StepResult{{Kind: "assert", Checks: []*assert.CheckResult{{OK: true}}}}},
			{Name: "f", Status: engine.StatusFailed, Duration: time.Millisecond, Steps: []engine.StepResult{
				{Kind: "assert", Checks: []*assert.CheckResult{{
					OK: false, Desc: "assert exit_code is 0", Expected: "exit code 0", Actual: "exit code 3", Hint: "differs"}}},
			}},
			{Name: "e", Status: engine.StatusError, Duration: time.Millisecond, Steps: []engine.StepResult{
				{Index: 0, Kind: "run", ErrMsg: "command not found"},
			}},
			{Name: "s", Status: engine.StatusSkipped, SkipReason: "only on os=plan9"},
			{Name: "k", Status: engine.StatusFlaky, Attempts: 2, Duration: time.Millisecond,
				Steps: []engine.StepResult{{Kind: "assert", Checks: []*assert.CheckResult{{OK: true}}}}},
		},
	}
}

// TestRender_CrossFormatCountParity pins the format-cross invariant: the same
// run result, rendered to every format, must agree on how many scenarios passed,
// failed, errored, skipped, and were flaky. A flaky (failed-then-recovered)
// scenario is green for the verdict, so it must land in the failure bucket of no
// format — junit routes it to flakyFailure, tap emits `ok`, gha a warning, and
// json keeps it out of failures[]. This is the differential oracle a future
// change to any one formatter would trip.
func TestRender_CrossFormatCountParity(t *testing.T) {
	t.Parallel()
	res := mixedResults()
	const (
		wantTotal   = 5
		wantPassed  = 1
		wantFailed  = 1
		wantErrored = 1
		wantSkipped = 1
		wantFlaky   = 1
	)
	// The failure bucket is what a machine consumer treats as "acted-on
	// failures": hard failures plus errors, never flaky recoveries.
	const wantFailureBucket = wantFailed + wantErrored

	t.Run("junit", func(t *testing.T) {
		t.Parallel()
		var root junitTestsuites
		if err := xml.Unmarshal([]byte(render(t, FormatJUnit, res)), &root); err != nil {
			t.Fatalf("junit invalid: %v", err)
		}
		if root.Tests != wantTotal {
			t.Errorf("junit tests = %d, want %d", root.Tests, wantTotal)
		}
		if root.Failures != wantFailed {
			t.Errorf("junit failures = %d, want %d", root.Failures, wantFailed)
		}
		if root.Errors != wantErrored {
			t.Errorf("junit errors = %d, want %d", root.Errors, wantErrored)
		}
		if root.Skipped != wantSkipped {
			t.Errorf("junit skipped = %d, want %d", root.Skipped, wantSkipped)
		}
		// The flaky scenario is present as a testcase but counted in none of the
		// failure/error/skip buckets: it is a green test carrying a flakyFailure.
		var flaky int
		for _, ts := range root.Suites {
			for _, tc := range ts.Testcases {
				if tc.FlakyFailure != nil {
					flaky++
				}
			}
		}
		if flaky != wantFlaky {
			t.Errorf("junit flakyFailure elements = %d, want %d", flaky, wantFlaky)
		}
	})

	t.Run("json", func(t *testing.T) {
		t.Parallel()
		var doc jsonDocument
		if err := json.Unmarshal([]byte(render(t, FormatJSON, res)), &doc); err != nil {
			t.Fatalf("json invalid: %v", err)
		}
		if len(doc.Suites) != 1 {
			t.Fatalf("suites = %d, want 1", len(doc.Suites))
		}
		rep := doc.Suites[0]
		byStatus := map[string]int{}
		for _, sc := range rep.Scenarios {
			byStatus[sc.Status]++
		}
		if len(rep.Scenarios) != wantTotal {
			t.Errorf("json scenarios = %d, want %d", len(rep.Scenarios), wantTotal)
		}
		if byStatus["passed"] != wantPassed || byStatus["failed"] != wantFailed ||
			byStatus["error"] != wantErrored || byStatus["skipped"] != wantSkipped ||
			byStatus["flaky"] != wantFlaky {
			t.Errorf("json status tally = %v", byStatus)
		}
		// failures[] carries hard failures and errors only; the flaky recovery
		// must never appear there, or a consumer counts it as a real failure.
		if len(rep.Failures) != wantFailureBucket {
			t.Errorf("json failures = %d, want %d: %+v", len(rep.Failures), wantFailureBucket, rep.Failures)
		}
		for _, f := range rep.Failures {
			if f.Scenario == "k" {
				t.Errorf("json failures[] must not include the flaky scenario: %+v", f)
			}
		}
	})

	t.Run("tap", func(t *testing.T) {
		t.Parallel()
		out := render(t, FormatTAP, res)
		if !strings.Contains(out, "1.."+itoa(wantTotal)+"\n") {
			t.Errorf("tap plan is not 1..%d:\n%s", wantTotal, out)
		}
		var okN, notOkN, skipN int
		for _, line := range strings.Split(out, "\n") {
			switch {
			case strings.HasPrefix(line, "not ok "):
				notOkN++
			case strings.HasPrefix(line, "ok "):
				okN++
				if strings.Contains(line, "# SKIP") {
					skipN++
				}
			}
		}
		// ok points: passed + skipped + flaky; not ok: failed + errored.
		if okN != wantPassed+wantSkipped+wantFlaky {
			t.Errorf("tap ok lines = %d, want %d\n%s", okN, wantPassed+wantSkipped+wantFlaky, out)
		}
		if notOkN != wantFailureBucket {
			t.Errorf("tap not-ok lines = %d, want %d\n%s", notOkN, wantFailureBucket, out)
		}
		if skipN != wantSkipped {
			t.Errorf("tap # SKIP lines = %d, want %d\n%s", skipN, wantSkipped, out)
		}
	})

	t.Run("gha", func(t *testing.T) {
		t.Parallel()
		out := render(t, FormatGHA, res)
		wantNotice := "5 scenarios: 1 passed, 1 failed, 1 errored, 1 skipped, 1 flaky"
		if !strings.Contains(out, wantNotice) {
			t.Errorf("gha notice summary missing %q:\n%s", wantNotice, out)
		}
		// The flaky scenario is a warning, not an error annotation.
		if strings.Contains(out, "::error title=mix / k::") {
			t.Errorf("gha emitted an error annotation for the flaky scenario:\n%s", out)
		}
		if !strings.Contains(out, "::warning title=mix / k::") {
			t.Errorf("gha missing warning annotation for the flaky scenario:\n%s", out)
		}
	})
}

// TestRender_CrossFormatCountParity_SetupErrored pins the format-cross count
// invariant for a suite that errored before any scenario ran (all scenarios
// filtered out, exit 4): junit, tap, and gha must agree on the count. junit
// synthesizes tests=1/errors=1 and tap a 1..1 plan with one not-ok, so the gha
// ::notice:: summary must likewise count the synthesized failure (1 errored),
// not contradict its own ::error:: annotation with an all-zero line.
func TestRender_CrossFormatCountParity_SetupErrored(t *testing.T) {
	t.Parallel()
	res := suiteSetupErrorEmpty()[0]

	t.Run("junit", func(t *testing.T) {
		t.Parallel()
		var root junitTestsuites
		if err := xml.Unmarshal([]byte(render(t, FormatJUnit, res)), &root); err != nil {
			t.Fatalf("junit invalid: %v", err)
		}
		if root.Tests != 1 || root.Errors != 1 {
			t.Errorf("junit tests=%d errors=%d, want 1/1", root.Tests, root.Errors)
		}
	})

	t.Run("tap", func(t *testing.T) {
		t.Parallel()
		out := render(t, FormatTAP, res)
		if !strings.Contains(out, "1..1\n") {
			t.Errorf("tap plan is not 1..1:\n%s", out)
		}
		if n := strings.Count(out, "\nnot ok "); n != 1 {
			t.Errorf("tap not-ok lines = %d, want 1:\n%s", n, out)
		}
	})

	t.Run("gha notice reflects the error, not all-zero", func(t *testing.T) {
		t.Parallel()
		out := render(t, FormatGHA, res)
		if !strings.Contains(out, "::error") {
			t.Fatalf("gha emitted no ::error:: annotation:\n%s", out)
		}
		if strings.Contains(out, "0 scenarios: 0 passed, 0 failed, 0 errored, 0 skipped") {
			t.Errorf("gha notice is all-zero, contradicting its own ::error:: and junit/tap:\n%s", out)
		}
		if !strings.Contains(out, "1 errored") {
			t.Errorf("gha notice should count the synthesized failure as 1 errored:\n%s", out)
		}
	})
}

// TestRender_HostileCharsStayWellFormed feeds XML/JSON/TAP-hostile bytes through
// the scenario name and failure detail — angle brackets, ampersands, a CDATA
// terminator, quotes, an embedded newline, a C0 control byte, and a multibyte
// rune — and asserts the machine formats stay parseable. junit must remain
// well-formed XML and json must remain valid JSON; a formatter that forgot to
// escape one of these would produce output no CI consumer can read.
func TestRender_HostileCharsStayWellFormed(t *testing.T) {
	t.Parallel()
	const hostile = "a<b>&\"'x]]>\x01\nz \xf0\x9f\x92\xa5 end"
	res := &engine.SuiteResult{
		Suite:    "h<&>",
		SpecPath: "h.atago.yaml",
		Status:   engine.StatusFailed,
		Duration: time.Millisecond,
		Scenarios: []engine.ScenarioResult{
			{Name: hostile, Status: engine.StatusFailed, Duration: time.Millisecond, Steps: []engine.StepResult{
				{Kind: "assert", Checks: []*assert.CheckResult{{
					OK: false, Desc: hostile, Expected: hostile, Actual: hostile, Hint: hostile}}},
			}},
		},
	}

	t.Run("junit stays well-formed", func(t *testing.T) {
		t.Parallel()
		var root junitTestsuites
		if err := xml.Unmarshal([]byte(render(t, FormatJUnit, res)), &root); err != nil {
			t.Fatalf("junit XML not well-formed with hostile chars: %v", err)
		}
		if root.Tests != 1 || root.Failures != 1 {
			t.Errorf("junit counts = tests %d failures %d, want 1/1", root.Tests, root.Failures)
		}
	})

	t.Run("json stays valid", func(t *testing.T) {
		t.Parallel()
		var doc jsonDocument
		if err := json.Unmarshal([]byte(render(t, FormatJSON, res)), &doc); err != nil {
			t.Fatalf("json not valid with hostile chars: %v", err)
		}
		if len(doc.Suites) != 1 || len(doc.Suites[0].Failures) != 1 {
			t.Fatalf("json shape unexpected: %+v", doc.Suites)
		}
	})

	t.Run("tap and gha do not panic", func(t *testing.T) {
		t.Parallel()
		// A newline in the description must be flattened so a point stays on one
		// line, and no raw CR may leak into a rendered point.
		tap := render(t, FormatTAP, res)
		for _, line := range strings.Split(tap, "\n") {
			if strings.HasPrefix(line, "not ok ") && strings.Contains(line, "\r") {
				t.Errorf("tap point line carries a raw CR:\n%s", line)
			}
		}
		_ = render(t, FormatGHA, res)
	})

	// A raw control byte (here \x01, but a captured ANSI escape is the common
	// case) must not survive into a TAP YAML diagnostic or a GHA annotation, or
	// the surrounding document is malformed. Tab and newline are structural.
	t.Run("tap and gha carry no raw control bytes", func(t *testing.T) {
		t.Parallel()
		for _, f := range []Format{FormatTAP, FormatGHA} {
			out := render(t, f, res)
			if i := strings.IndexFunc(out, rejectedControlRune); i >= 0 {
				t.Errorf("%v output carries a raw control byte at offset %d:\n%q", f, i, out)
			}
		}
	})
}

// rejectedControlRune reports a control character no machine-readable report may
// carry verbatim: any C0 byte other than tab/newline, DEL, or a C1 control.
func rejectedControlRune(r rune) bool {
	if r == '\t' || r == '\n' {
		return false
	}
	return r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f)
}

// TestTAP_DiagnosticMessageNotDoubleEscaped is a regression: a `#` inside a TAP
// diagnostic message is not special (it is a plain char in a double-quoted YAML
// scalar), so it must round-trip verbatim. The old code ran the message through
// tapInline — which escapes `#`→`\#` for the bare ok/not-ok line — then %q,
// injecting a spurious backslash a TAP consumer would decode.
func TestTAP_DiagnosticMessageNotDoubleEscaped(t *testing.T) {
	t.Parallel()
	res := &engine.SuiteResult{
		Suite: "s", SpecPath: "s.atago.yaml", Status: engine.StatusFailed, Duration: time.Millisecond,
		Scenarios: []engine.ScenarioResult{
			{Name: "sc", Status: engine.StatusFailed, Duration: time.Millisecond, Steps: []engine.StepResult{
				{Kind: "assert", Checks: []*assert.CheckResult{{OK: false, Desc: "issue #42 body contains foo"}}},
			}},
		},
	}
	tap := render(t, FormatTAP, res)
	if !strings.Contains(tap, `message: "issue #42 body contains foo"`) {
		t.Errorf("tap diagnostic message double-escaped or altered:\n%s\nwant a plain `message: \"issue #42 body contains foo\"`", tap)
	}
	if strings.Contains(tap, `\#`) {
		t.Errorf("tap diagnostic message injected a spurious backslash before #:\n%s", tap)
	}
}

// TestJUnit_TimeAttrIsPlainDecimal is a regression: a sub-millisecond scenario
// duration must render as a plain decimal seconds value, not Go's default float
// %g scientific notation (e.g. "1.5e-06"), which the JUnit/Surefire schema and
// strict parsers reject.
func TestJUnit_TimeAttrIsPlainDecimal(t *testing.T) {
	t.Parallel()
	res := &engine.SuiteResult{
		Suite: "s", SpecPath: "s.atago.yaml", Status: engine.StatusPassed, Duration: 1500,
		Scenarios: []engine.ScenarioResult{
			{Name: "quick", Status: engine.StatusPassed, Duration: 1500, // 1500ns
				Steps: []engine.StepResult{{Kind: "assert", Checks: []*assert.CheckResult{{OK: true}}}}},
		},
	}
	out := render(t, FormatJUnit, res)
	if strings.Contains(out, "e-") || strings.Contains(out, "e+") {
		t.Errorf("junit time attribute uses scientific notation:\n%s", out)
	}
	// It must still parse as XML with the documented shape.
	var root junitTestsuites
	if err := xml.Unmarshal([]byte(out), &root); err != nil {
		t.Fatalf("junit XML not well-formed: %v", err)
	}
}

// TestRender_SummaryUsesElapsedNotSuiteSum is a regression: the console summary
// headline duration must be the run's real wall-clock time, not the sum of
// per-suite durations, which overcounts when --parallel runs suites concurrently
// (two one-second suites in parallel finish in ~1s, not 2s).
func TestRender_SummaryUsesElapsedNotSuiteSum(t *testing.T) {
	t.Parallel()
	suite := func(name string) *engine.SuiteResult {
		return &engine.SuiteResult{
			Suite: name, SpecPath: name + ".atago.yaml", Status: engine.StatusPassed, Duration: time.Second,
			Scenarios: []engine.ScenarioResult{
				{Name: "sc", Status: engine.StatusPassed, Duration: time.Second,
					Steps: []engine.StepResult{{Kind: "assert", Checks: []*assert.CheckResult{{OK: true}}}}},
			},
		}
	}
	var b strings.Builder
	if err := Render(&b, FormatConsole, []*engine.SuiteResult{suite("a"), suite("b")}, WithElapsed(time.Second)); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	if !strings.Contains(out, "(1s)") {
		t.Errorf("summary did not report the elapsed wall-clock (1s):\n%s", out)
	}
	if strings.Contains(out, "(2s)") {
		t.Errorf("summary summed per-suite durations (2s) instead of using elapsed:\n%s", out)
	}
}
