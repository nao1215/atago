package report

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/atago/internal/engine"
)

// A JUnit XML report (spec.md §26), consumable by CI systems and test-result
// viewers. Rendered by Render (FormatJUnit) via buildJUnit/writeJUnit.
type junitTestsuites struct {
	XMLName  xml.Name         `xml:"testsuites"`
	Tests    int              `xml:"tests,attr"`
	Failures int              `xml:"failures,attr"`
	Errors   int              `xml:"errors,attr"`
	Skipped  int              `xml:"skipped,attr"`
	Time     float64          `xml:"time,attr"`
	Suites   []junitTestsuite `xml:"testsuite"`
}

type junitTestsuite struct {
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Errors    int             `xml:"errors,attr"`
	Skipped   int             `xml:"skipped,attr"`
	Time      float64         `xml:"time,attr"`
	Testcases []junitTestcase `xml:"testcase"`
}

type junitTestcase struct {
	Name    string        `xml:"name,attr"`
	Time    float64       `xml:"time,attr"`
	Failure *junitMessage `xml:"failure,omitempty"`
	Error   *junitMessage `xml:"error,omitempty"`
	Skipped *junitSkipped `xml:"skipped,omitempty"`
}

type junitMessage struct {
	Message string `xml:"message,attr"`
	Body    string `xml:",chardata"`
}

type junitSkipped struct {
	Message string `xml:"message,attr"`
}

func buildJUnit(results []*engine.SuiteResult) junitTestsuites {
	root := junitTestsuites{}
	for _, res := range results {
		ts := junitTestsuite{Name: res.Suite, Time: res.Duration.Seconds()}
		for i := range res.Scenarios {
			sc := &res.Scenarios[i]
			tc := junitTestcase{Name: sc.Name, Time: sc.Duration.Seconds()}
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
			}
			ts.Testcases = append(ts.Testcases, tc)
			ts.Tests++
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

func firstErrorMessage(sc *engine.ScenarioResult) string {
	for _, step := range sc.Steps {
		if step.ErrMsg != "" {
			return step.ErrMsg
		}
	}
	return "execution error"
}

// detailText renders the human failure block as plain text for the XML body.
func detailText(sc *engine.ScenarioResult) string {
	var b strings.Builder
	for _, step := range sc.Steps {
		for _, ck := range step.Checks {
			if ck == nil || ck.OK {
				continue
			}
			fmt.Fprintf(&b, "Step: %s\n", ck.Desc)
			if ck.Expected != "" {
				fmt.Fprintf(&b, "Expected: %s\n", ck.Expected)
			}
			if ck.Actual != "" {
				fmt.Fprintf(&b, "Actual: %s\n", ck.Actual)
			}
			if ck.Hint != "" {
				fmt.Fprintf(&b, "Hint: %s\n", ck.Hint)
			}
		}
		if step.ErrMsg != "" {
			fmt.Fprintf(&b, "Error %s: %s\n", stepErrorContext(step), step.ErrMsg)
		}
	}
	return strings.TrimRight(b.String(), "\n")
}
