package report

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"pgregory.net/rapid"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/engine"
	"github.com/nao1215/atago/internal/plural"
	"github.com/nao1215/atago/internal/spec"
)

// The fixed fixtures in format_parity_test.go pin one scenario per status.
// The properties below generate whole runs instead: several suites, any mix of
// the seven terminal statuses, and names built from the characters each format
// has to escape (XML metacharacters, TAP's '#' and backslash, newlines,
// control bytes, wide runes). For every generated run, each machine-readable
// format has to agree with the run it was given, so a formatter that drops,
// double-counts, or lets a name forge structure is caught on inputs nobody
// thought to write by hand.

var allStatuses = []engine.Status{
	engine.StatusPassed, engine.StatusFailed, engine.StatusError, engine.StatusSkipped,
	engine.StatusFlaky, engine.StatusXFail, engine.StatusXPass,
}

// nameFragments are the pieces names are assembled from. Each one is there
// because some format treats it specially.
var nameFragments = []string{
	"a", "build", " ", "  ", "/", "-", "#", "\\", "\\#", " # SKIP", " # TODO", "ok 1 - x",
	"not ok", "\n", "\r\n", "\t", "<", ">", "&", "\"", "'", "]]>", "%", "::", "\x1b[31m",
	"\x07", "\x7f", "é", "日本", "👍", "\u200d", "1..3",
}

func genName(t *rapid.T, label string) string {
	parts := rapid.SliceOfN(rapid.SampledFrom(nameFragments), 1, 6).Draw(t, label)
	return strings.Join(parts, "")
}

func genScenario(t *rapid.T, label string) engine.ScenarioResult {
	st := rapid.SampledFrom(allStatuses).Draw(t, label+"-status")
	sc := engine.ScenarioResult{Name: genName(t, label+"-name"), Status: st, Duration: time.Millisecond}
	pass := []engine.StepResult{{Kind: "assert", Checks: []*assert.CheckResult{{OK: true}}}}
	fail := []engine.StepResult{{Kind: "assert", Checks: []*assert.CheckResult{{
		OK: false, Desc: "assert stdout contains x", Expected: "x", Actual: "y", Hint: "differs"}}}}
	switch st {
	case engine.StatusPassed:
		sc.Steps = pass
	case engine.StatusFailed:
		sc.Steps = fail
	case engine.StatusError:
		sc.Steps = []engine.StepResult{{Kind: "run", ErrMsg: "command not found"}}
	case engine.StatusSkipped:
		sc.SkipReason = genName(t, label+"-skip")
	case engine.StatusFlaky:
		// Two shapes reach the report: a retry that recovered keeps the
		// passing attempt, and a --repeat fold keeps the failing iteration's
		// steps, whose failures JSON still lists.
		sc.Attempts = 2
		sc.Steps = pass
		if rapid.Bool().Draw(t, label+"-flaky-repeat") {
			sc.Steps = fail
		}
	case engine.StatusXFail:
		sc.ExpectFail = &spec.ExpectFail{Reason: genName(t, label+"-reason")}
		sc.Steps = fail
	case engine.StatusXPass:
		sc.ExpectFail = &spec.ExpectFail{Reason: genName(t, label+"-reason")}
		sc.Steps = pass
	}
	return sc
}

// genRun draws one to three suites of one to eight scenarios each. Suites with
// no scenario take a separate synthesized path (they error before running) and
// are covered by TestRender_CrossFormatCountParity_SetupErrored.
func genRun(t *rapid.T) []*engine.SuiteResult {
	n := rapid.IntRange(1, 3).Draw(t, "suites")
	out := make([]*engine.SuiteResult, 0, n)
	for i := range n {
		m := rapid.IntRange(1, 8).Draw(t, fmt.Sprintf("scenarios-%d", i))
		res := &engine.SuiteResult{
			Suite:    genName(t, fmt.Sprintf("suite-%d", i)),
			SpecPath: fmt.Sprintf("s%d.atago.yaml", i),
			Status:   engine.StatusPassed,
			Duration: time.Millisecond,
		}
		for j := range m {
			sc := genScenario(t, fmt.Sprintf("sc-%d-%d", i, j))
			if sc.Status == engine.StatusFailed || sc.Status == engine.StatusError || sc.Status == engine.StatusXPass {
				res.Status = engine.StatusFailed
			}
			res.Scenarios = append(res.Scenarios, sc)
		}
		if rapid.IntRange(0, 4).Draw(t, fmt.Sprintf("security-%d", i)) == 0 {
			res.SecurityViolation = true
			res.Status = engine.StatusFailed
		}
		out = append(out, res)
	}
	return out
}

// genLoadFailures draws zero to two spec files that never loaded. Each one is
// a failing point of its own in every format.
func genLoadFailures(t *rapid.T) []LoadFailure {
	n := rapid.IntRange(0, 2).Draw(t, "load-failures")
	out := make([]LoadFailure, 0, n)
	for i := range n {
		out = append(out, LoadFailure{
			SpecPath: genName(t, fmt.Sprintf("lf-path-%d", i)),
			Message:  genName(t, fmt.Sprintf("lf-msg-%d", i)),
		})
	}
	return out
}

func securityViolations(results []*engine.SuiteResult) int {
	n := 0
	for _, r := range results {
		if r.SecurityViolation {
			n++
		}
	}
	return n
}

func tally(results []*engine.SuiteResult) map[engine.Status]int {
	c := map[engine.Status]int{}
	for _, r := range results {
		for _, sc := range r.Scenarios {
			c[sc.Status]++
		}
	}
	return c
}

// failedChecks counts the failing checks in a scenario's steps.
func failedChecks(sc engine.ScenarioResult) int {
	n := 0
	for _, st := range sc.Steps {
		for _, c := range st.Checks {
			if !c.OK {
				n++
			}
		}
	}
	return n
}

func renderAll(t *rapid.T, f Format, results []*engine.SuiteResult, opts ...Option) string {
	var b bytes.Buffer
	if err := Render(&b, f, results, opts...); err != nil {
		t.Fatalf("render %s: %v", f, err)
	}
	return b.String()
}

func total(c map[engine.Status]int) int {
	n := 0
	for _, v := range c {
		n += v
	}
	return n
}

// TestProperty_JSONRoundTripsEveryScenario: the JSON document lists every
// scenario of every suite, in order, with its exact name and status, and its
// failures[] holds exactly the failed, errored and xpass scenarios, the failing
// checks a flaky --repeat fold kept, and one entry for a security violation.
func TestProperty_JSONRoundTripsEveryScenario(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		results := genRun(t)
		var doc jsonDocument
		if err := json.Unmarshal([]byte(renderAll(t, FormatJSON, results)), &doc); err != nil {
			t.Fatalf("json invalid: %v", err)
		}
		if len(doc.Suites) != len(results) {
			t.Fatalf("suites = %d, want %d", len(doc.Suites), len(results))
		}
		for i, res := range results {
			got := doc.Suites[i]
			if got.Suite != res.Suite {
				t.Fatalf("suite %d name = %q, want %q", i, got.Suite, res.Suite)
			}
			if len(got.Scenarios) != len(res.Scenarios) {
				t.Fatalf("suite %d scenarios = %d, want %d", i, len(got.Scenarios), len(res.Scenarios))
			}
			failures := 0
			for j, sc := range res.Scenarios {
				if got.Scenarios[j].Name != sc.Name || got.Scenarios[j].Status != string(sc.Status) {
					t.Fatalf("suite %d scenario %d = %q/%s, want %q/%s",
						i, j, got.Scenarios[j].Name, got.Scenarios[j].Status, sc.Name, sc.Status)
				}
				switch sc.Status {
				case engine.StatusFailed, engine.StatusError, engine.StatusXPass:
					failures++
				case engine.StatusFlaky:
					failures += failedChecks(sc)
				case engine.StatusPassed, engine.StatusSkipped, engine.StatusXFail:
				}
			}
			if got.SecurityViolation != res.SecurityViolation {
				t.Fatalf("suite %d security_violation = %v, want %v", i, got.SecurityViolation, res.SecurityViolation)
			}
			if res.SecurityViolation {
				failures++
			}
			if len(got.Failures) != failures {
				t.Fatalf("suite %d failures[] = %d, want %d", i, len(got.Failures), failures)
			}
		}
	})
}

// xmlChar reports whether r may appear in an XML 1.0 document.
func xmlChar(r rune) bool {
	return r == '\t' || r == '\n' || r == '\r' ||
		(r >= 0x20 && r <= 0xD7FF) || (r >= 0xE000 && r <= 0xFFFD) || (r >= 0x10000 && r <= 0x10FFFF)
}

// xmlSafe reports whether every rune of s can be written into an XML 1.0
// document as itself. Other runes are replaced on the way out, so only these
// names are expected to round-trip exactly.
func xmlSafe(s string) bool {
	for _, r := range s {
		if !xmlChar(r) || r == utf8.RuneError {
			return false
		}
	}
	return true
}

// TestProperty_JUnitCountsMatchTheRun: the JUnit document parses, its root
// counters agree with the run (xfail is a skip, xpass a failure, flaky a pass
// carrying flakyFailure, a security violation or an unloadable spec one more
// errored testcase), the per-suite counters add up to the root, and a testcase
// name made of XML-legal characters comes back unchanged.
func TestProperty_JUnitCountsMatchTheRun(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		results := genRun(t)
		lfs := genLoadFailures(t)
		c := tally(results)
		extra := securityViolations(results) + len(lfs)
		var root junitTestsuites
		if err := xml.Unmarshal([]byte(renderAll(t, FormatJUnit, results, WithLoadFailures(lfs...))), &root); err != nil {
			t.Fatalf("junit invalid: %v", err)
		}
		if root.Tests != total(c)+extra {
			t.Fatalf("tests = %d, want %d", root.Tests, total(c)+extra)
		}
		if want := c[engine.StatusFailed] + c[engine.StatusXPass]; root.Failures != want {
			t.Fatalf("failures = %d, want %d", root.Failures, want)
		}
		if want := c[engine.StatusError] + extra; root.Errors != want {
			t.Fatalf("errors = %d, want %d", root.Errors, want)
		}
		if want := c[engine.StatusSkipped] + c[engine.StatusXFail]; root.Skipped != want {
			t.Fatalf("skipped = %d, want %d", root.Skipped, want)
		}
		if len(root.Suites) != len(results)+len(lfs) {
			t.Fatalf("testsuite elements = %d, want %d", len(root.Suites), len(results)+len(lfs))
		}
		var tests, flaky int
		for _, ts := range root.Suites {
			tests += ts.Tests
			if ts.Tests != len(ts.Testcases) {
				t.Fatalf("testsuite %q says tests=%d but holds %d testcases", ts.Name, ts.Tests, len(ts.Testcases))
			}
		}
		// The unloadable specs lead the document; the run's suites follow in order.
		for i, ts := range root.Suites[len(lfs):] {
			wantCases := len(results[i].Scenarios)
			if results[i].SecurityViolation {
				wantCases++
			}
			if len(ts.Testcases) != wantCases {
				t.Fatalf("suite %d testcases = %d, want %d", i, len(ts.Testcases), wantCases)
			}
			for j, tc := range ts.Testcases[:len(results[i].Scenarios)] {
				if tc.FlakyFailure != nil {
					flaky++
				}
				if want := results[i].Scenarios[j].Name; xmlSafe(want) && tc.Name != want {
					t.Fatalf("suite %d testcase %d name = %q, want %q", i, j, tc.Name, want)
				}
			}
		}
		if tests != root.Tests {
			t.Fatalf("sum of testsuite tests = %d, root says %d", tests, root.Tests)
		}
		if flaky != c[engine.StatusFlaky] {
			t.Fatalf("flakyFailure elements = %d, want %d", flaky, c[engine.StatusFlaky])
		}
	})
}

// tapPoint is one ok / not ok line read back the way a TAP 14 consumer reads
// it: the description is unescaped (`\\` is a backslash, `\#` a hash) and the
// first unescaped `#` starts the directive.
type tapPoint struct {
	ok          bool
	num         int
	description string
	directive   string
}

func parseTAPPoint(line string) (tapPoint, bool) {
	var p tapPoint
	rest, found := strings.CutPrefix(line, "ok ")
	if found {
		p.ok = true
	} else if rest, found = strings.CutPrefix(line, "not ok "); !found {
		return p, false
	}
	numText, rest, _ := strings.Cut(rest, " ")
	n, err := strconv.Atoi(numText)
	if err != nil {
		return p, false
	}
	p.num = n
	rest = strings.TrimPrefix(rest, "- ")
	var desc strings.Builder
	for i := 0; i < len(rest); i++ {
		ch := rest[i]
		if ch == '\\' && i+1 < len(rest) && (rest[i+1] == '\\' || rest[i+1] == '#') {
			desc.WriteByte(rest[i+1])
			i++
			continue
		}
		if ch == '#' {
			p.directive = strings.TrimSpace(rest[i+1:])
			break
		}
		desc.WriteByte(ch)
	}
	p.description = strings.TrimSpace(desc.String())
	return p, true
}

// TestProperty_TAPPointsMatchTheRun: the plan equals the number of points,
// there is exactly one point per unloadable spec, scenario and security
// violation, numbered 1..N in order (a name can never forge an extra `ok` line), ok / not ok follows the verdict, the
// directive a consumer reads is SKIP for a skip, TODO for xfail and xpass, and
// nothing otherwise, and the description it reads back is the suite and
// scenario name flattened onto one line.
func TestProperty_TAPPointsMatchTheRun(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		results := genRun(t)
		lfs := genLoadFailures(t)
		// Each expected point: a load failure, a scenario, or a suite's
		// security violation, in the order the stream promises.
		type point struct {
			sc   engine.ScenarioResult
			desc string
		}
		var want []point
		for _, lf := range lfs {
			want = append(want, point{sc: engine.ScenarioResult{Status: engine.StatusError}, desc: lf.SpecPath})
		}
		for _, r := range results {
			for _, sc := range r.Scenarios {
				want = append(want, point{sc: sc, desc: r.Suite + " / " + sc.Name})
			}
			if r.SecurityViolation {
				want = append(want, point{sc: engine.ScenarioResult{Status: engine.StatusError}, desc: r.Suite + " / security policy"})
			}
		}
		out := renderAll(t, FormatTAP, results, WithLoadFailures(lfs...))
		// Checked byte by byte rather than against tapFlatten, which the
		// renderer shares: a raw control byte from a generated name must never
		// reach the stream, wherever it would land. Tab is whitespace TAP
		// allows in a description and is kept on purpose.
		for i := 0; i < len(out); i++ {
			if c := out[i]; (c < 0x20 && c != '\n' && c != '\t') || c == 0x7f {
				t.Fatalf("raw control byte %#x at offset %d:\n%q", c, i, out)
			}
		}
		if !strings.Contains(out, "\n1.."+strconv.Itoa(len(want))+"\n") {
			t.Fatalf("plan is not 1..%d:\n%s", len(want), out)
		}
		var points []tapPoint
		for _, line := range strings.Split(out, "\n") {
			if p, ok := parseTAPPoint(line); ok {
				points = append(points, p)
			}
		}
		if len(points) != len(want) {
			t.Fatalf("points = %d, want %d:\n%s", len(points), len(want), out)
		}
		for i, p := range points {
			sc := want[i].sc
			if p.num != i+1 {
				t.Fatalf("point %d is numbered %d:\n%s", i+1, p.num, out)
			}
			wantOK := sc.Status == engine.StatusPassed || sc.Status == engine.StatusSkipped ||
				sc.Status == engine.StatusFlaky || sc.Status == engine.StatusXPass
			if p.ok != wantOK {
				t.Fatalf("point %d (%s) ok = %v, want %v:\n%s", i+1, sc.Status, p.ok, wantOK, out)
			}
			var wantDirective string
			switch sc.Status {
			case engine.StatusSkipped:
				wantDirective = "SKIP"
			case engine.StatusXFail, engine.StatusXPass:
				wantDirective = "TODO"
			case engine.StatusPassed, engine.StatusFailed, engine.StatusError, engine.StatusFlaky:
				wantDirective = ""
			}
			gotDirective, _, _ := strings.Cut(p.directive, " ")
			if p.directive == "" {
				gotDirective = ""
			}
			if gotDirective != wantDirective {
				t.Fatalf("point %d (%s) directive = %q, want %q:\n%s", i+1, sc.Status, p.directive, wantDirective, out)
			}
			if wantDesc := strings.TrimSpace(tapFlatten(want[i].desc)); p.description != wantDesc {
				t.Fatalf("point %d description reads back as %q, want %q:\n%s", i+1, p.description, wantDesc, out)
			}
		}
	})
}

// TestProperty_GHAAndConsoleAgreeOnTheTally: the gha notice and the console
// summary carry the same tally line, worded the same way ("1 scenario",
// "2 scenarios"), and its numbers are the run's.
func TestProperty_GHAAndConsoleAgreeOnTheTally(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		results := genRun(t)
		// The gha notice counts a security violation as one more errored point
		// (as TAP and JUnit do) while the console tally counts scenarios only and
		// reports the violation in its own block, so the two lines are compared on
		// runs without one (TestRender_SecurityViolationInTeardownIsNotGreen pins
		// both shapes).
		for _, r := range results {
			r.SecurityViolation = false
		}
		c := tally(results)
		head := fmt.Sprintf("%s: %d passed, %d failed, %d errored, %d skipped",
			plural.Count(total(c), "scenario", "scenarios"), c[engine.StatusPassed], c[engine.StatusFailed], c[engine.StatusError], c[engine.StatusSkipped])
		for _, f := range []Format{FormatGHA, FormatConsole} {
			if out := renderAll(t, f, results); !strings.Contains(out, head) {
				t.Fatalf("%s output lacks %q:\n%s", f, head, out)
			}
		}
	})
}
