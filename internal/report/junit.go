package report

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/nao1215/atago/internal/engine"
)

// junitSeconds is a duration in seconds rendered as a plain fixed-point decimal.
// encoding/xml marshals a float64 with %g, so a sub-millisecond time would come
// out in scientific notation (e.g. "1.5e-06"), which the JUnit/Surefire schema
// and strict parsers reject. Six decimals keep microsecond resolution.
type junitSeconds float64

func (s junitSeconds) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: name, Value: strconv.FormatFloat(float64(s), 'f', 6, 64)}, nil
}

// A JUnit XML report, consumable by CI systems and test-result
// viewers. Rendered by Render (FormatJUnit) via buildJUnit/writeJUnit.
type junitTestsuites struct {
	XMLName  xml.Name         `xml:"testsuites"`
	Tests    int              `xml:"tests,attr"`
	Failures int              `xml:"failures,attr"`
	Errors   int              `xml:"errors,attr"`
	Skipped  int              `xml:"skipped,attr"`
	Time     junitSeconds     `xml:"time,attr"`
	Suites   []junitTestsuite `xml:"testsuite"`
}

type junitTestsuite struct {
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Errors    int             `xml:"errors,attr"`
	Skipped   int             `xml:"skipped,attr"`
	Time      junitSeconds    `xml:"time,attr"`
	Testcases []junitTestcase `xml:"testcase"`
	// SystemErr carries a failed suite.teardown. It never changes the counts —
	// teardown outcomes never change a verdict — but a junit consumer used to
	// see a clean document with zero trace that cleanup failed.
	SystemErr string `xml:"system-err,omitempty"`
}

type junitTestcase struct {
	Name    string        `xml:"name,attr"`
	Time    junitSeconds  `xml:"time,attr"`
	Failure *junitMessage `xml:"failure,omitempty"`
	Error   *junitMessage `xml:"error,omitempty"`
	Skipped *junitSkipped `xml:"skipped,omitempty"`
	// FlakyFailure is the Maven-surefire convention for a test that failed
	// and then passed on retry (#29): the testcase itself counts as passed,
	// the element preserves the evidence.
	FlakyFailure *junitMessage `xml:"flakyFailure,omitempty"`
	// SystemErr carries the scenario's failed teardown, junit's slot for
	// output that accompanies a result without deciding it (the flakyFailure
	// pattern one severity down): the verdict is decided by the steps alone,
	// so a failed cleanup must not add a <failure>, but staying silent hid it
	// from every junit consumer.
	SystemErr string `xml:"system-err,omitempty"`
}

type junitMessage struct {
	Message string `xml:"message,attr"`
	Body    string `xml:",chardata"`
}

type junitSkipped struct {
	Message string `xml:"message,attr"`
}

func buildJUnit(results []*engine.SuiteResult, allowXPass bool, loadFailures []LoadFailure) junitTestsuites {
	root := junitTestsuites{}
	// A spec that never parsed belongs to no suite, so it gets one of its own
	// carrying a single errored testcase — the shape a collection error takes in
	// every other tool that produces JUnit XML.
	for _, lf := range loadFailures {
		root.Suites = append(root.Suites, junitTestsuite{
			Name: lf.SpecPath, Tests: 1, Errors: 1,
			Testcases: []junitTestcase{{
				Name:  "load",
				Error: &junitMessage{Message: "spec failed to load", Body: lf.Message},
			}},
		})
		root.Tests++
		root.Errors++
	}
	for _, res := range results {
		ts := junitTestsuite{Name: res.Suite, Time: junitSeconds(res.Duration.Seconds())}
		for i := range res.Scenarios {
			sc := &res.Scenarios[i]
			tc := junitTestcase{Name: sc.Name, Time: junitSeconds(sc.Duration.Seconds())}
			switch sc.Status {
			case engine.StatusFailed:
				tc.Failure = &junitMessage{Message: firstFailureMessage(sc), Body: detailText(sc)}
				ts.Failures++
			case engine.StatusError:
				tc.Error = &junitMessage{Message: firstErrorMessage(sc), Body: detailText(sc)}
				ts.Errors++
			case engine.StatusSkipped:
				tc.Skipped = &junitSkipped{Message: sc.SkipReason}
				ts.Skipped++
			case engine.StatusFlaky:
				tc.FlakyFailure = &junitMessage{
					Message: flakyMessage(sc),
					Body:    detailText(sc),
				}
			case engine.StatusXFail:
				// pytest's convention, and the only one JUnit XML can express: an
				// expected failure is a skipped testcase carrying its reason. A
				// <failure> would paint the build red for a known bug, which is
				// the whole thing expect_fail exists to avoid.
				tc.Skipped = &junitSkipped{Message: "xfail: " + expectFailSummary(sc)}
				ts.Skipped++
			case engine.StatusXPass:
				if allowXPass {
					// Green for the exit code, so green here too — a <failure>
					// in a passing build is exactly the contradiction the flaky
					// verdict already avoids. flakyFailure is JUnit's slot for
					// "worth surfacing, not a failure".
					tc.FlakyFailure = &junitMessage{Message: xpassMessage(sc), Body: detailText(sc)}
					break
				}
				tc.Failure = &junitMessage{Message: xpassMessage(sc), Body: detailText(sc)}
				ts.Failures++
			case engine.StatusPassed:
				// A bare <testcase> with no child element: JUnit's way of saying
				// it passed.
			}
			if td := stepsDetailText(sc.Teardown); td != "" {
				tc.SystemErr = "teardown failed (the verdict is decided by the steps alone):\n" + td
			}
			ts.Testcases = append(ts.Testcases, tc)
			ts.Tests++
		}
		// A suite that errored before producing any scenario row (a suite.setup
		// failure with nothing selected, #7) has no testcase to carry the error;
		// synthesize one so the suite is never a green empty <testsuite> while the
		// process exits non-zero.
		if suiteErroredWithoutScenarios(res) {
			for _, p := range suiteFailurePoints(res) {
				ts.Testcases = append(ts.Testcases, junitTestcase{
					Name:  p.name,
					Error: &junitMessage{Message: p.message, Body: p.body},
				})
				ts.Errors++
				ts.Tests++
			}
		}
		if td := stepsDetailText(res.Teardown); td != "" {
			ts.SystemErr = "suite teardown failed (teardown outcomes never change the suite status):\n" + td
		}
		root.Suites = append(root.Suites, ts)
		root.Tests += ts.Tests
		root.Failures += ts.Failures
		root.Errors += ts.Errors
		root.Skipped += ts.Skipped
		root.Time += ts.Time
	}
	return root
}

func writeJUnit(w io.Writer, root junitTestsuites) error {
	if _, err := io.WriteString(w, xml.Header); err != nil {
		return err
	}
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	if err := enc.Encode(root); err != nil {
		return err
	}
	_, err := io.WriteString(w, "\n")
	return err
}

func firstFailureMessage(sc *engine.ScenarioResult) string {
	for _, step := range sc.Steps {
		for _, ck := range step.Checks {
			if ck != nil && !ck.OK {
				return ck.Desc
			}
		}
	}
	return "assertion failed"
}

// firstStepFailureMessage returns the first failing check's description or the
// first errored step's message across a step list, or "" when it is clean. It
// names a teardown failure in the one-line slots (a tap comment, a gha warning
// title) the way firstFailureMessage names a scenario failure.
func firstStepFailureMessage(steps []engine.StepResult) string {
	for _, step := range steps {
		for _, ck := range step.Checks {
			if ck != nil && !ck.OK {
				return ck.Desc
			}
		}
		if step.ErrMsg != "" {
			return step.ErrMsg
		}
	}
	return ""
}

func firstErrorMessage(sc *engine.ScenarioResult) string {
	for _, step := range sc.Steps {
		if step.ErrMsg != "" {
			return step.ErrMsg
		}
	}
	return "execution error"
}

// detailText renders the human failure block as plain text for the XML body,
// closing with any preserved background-service logs (#51) — evidence the run
// wrote for whoever reads the failure, which only the console and the json
// report used to reference.
func detailText(sc *engine.ScenarioResult) string {
	body := stepsDetailText(sc.Steps)
	if len(sc.ServiceLogs) == 0 {
		return body
	}
	var b strings.Builder
	b.WriteString(body)
	if body != "" {
		b.WriteByte('\n')
	}
	for _, sl := range sc.ServiceLogs {
		fmt.Fprintf(&b, "Service log (%s): %s\n", sl.Name, sl.Path)
	}
	return strings.TrimRight(b.String(), "\n")
}

// stepsDetailText renders the failed checks and errored steps of any step list
// — a scenario's steps, its teardown, or the suite lifecycle — as plain text,
// or "" when the list is clean.
func stepsDetailText(steps []engine.StepResult) string {
	var b strings.Builder
	for _, step := range steps {
		for _, ck := range step.Checks {
			if ck == nil || ck.OK {
				continue
			}
			fmt.Fprintf(&b, "Step: %s\n", ck.Desc)
			// Multi-line equals/snapshot failures embed the uncolored unified
			// diff (#28); the compact form covers everything else. These
			// bodies feed junit/tap/gha, which must stay ANSI-free.
			if diff := checkDiff(ck); diff != "" {
				fmt.Fprintf(&b, "Diff (-expected +actual):\n%s\n", diff)
			} else {
				expected, actual := visiblePair(ck.Expected, ck.Actual)
				if expected != "" {
					fmt.Fprintf(&b, "Expected: %s\n", expected)
				}
				if actual != "" {
					fmt.Fprintf(&b, "Actual: %s\n", actual)
				}
			}
			if ck.Hint != "" {
				fmt.Fprintf(&b, "Hint: %s\n", ck.Hint)
			}
			// Reference the durable sidecars by path (#48). The console and the
			// json report carried them and these formats did not, so the CI
			// systems they exist for could not reach the files the run wrote.
			for _, a := range ck.ArtifactFiles {
				fmt.Fprintf(&b, "Artifact (%s): %s\n", a.Role, a.Path)
			}
		}
		if step.ErrMsg != "" {
			fmt.Fprintf(&b, "Error %s: %s\n", stepErrorContext(step), step.ErrMsg)
		}
	}
	return strings.TrimRight(b.String(), "\n")
}
