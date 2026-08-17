package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/atago/internal/diag"
	"github.com/nao1215/atago/internal/engine"
	"github.com/nao1215/atago/internal/plural"
)

// writeGHA emits GitHub Actions workflow-command annotations, so
// failures surface inline in the Actions UI. One `::error::` line per failed or
// errored scenario, plus a final `::notice::` summary. Rendered by Render
// (FormatGHA).
func writeGHA(w io.Writer, results []*engine.SuiteResult, allowXPass bool, loadFailures []LoadFailure, snapsUpdated int) error {
	var b strings.Builder
	var agg engine.Counts
	var total int
	// A spec that never parsed produced no scenario to annotate, so annotate the
	// file itself, as an error matching the non-zero exit.
	for _, lf := range loadFailures {
		// The one annotation that has a file to point at: `file=` puts it on the
		// spec in the Actions UI, where a scenario failure has only a title.
		fmt.Fprintf(&b, "::error file=%s,title=%s::%s\n",
			ghaEscapeProp(lf.SpecPath), ghaEscapeProp(lf.SpecPath),
			ghaEscapeData("spec failed to load: "+oneLine(lf.Message)))
	}
	for _, res := range results {
		for i := range res.Scenarios {
			sc := &res.Scenarios[i]
			switch sc.Status {
			case engine.StatusFailed:
				fmt.Fprintf(&b, "::error title=%s::%s\n",
					ghaEscapeProp(res.Suite+" / "+sc.Name), ghaEscapeData(firstFailureMessage(sc)+" — "+oneLine(detailText(sc))))
			case engine.StatusError:
				// The code goes in the title, which is what GitHub renders in the
				// annotations list; buried in the body it would only be visible to
				// someone who already opened the annotation.
				msg := firstErrorMessage(sc)
				title := res.Suite + " / " + sc.Name
				if codes := diag.Codes(msg); len(codes) > 0 {
					title = codes[0].String() + " " + title
				}
				fmt.Fprintf(&b, "::error title=%s::%s\n",
					ghaEscapeProp(title), ghaEscapeData(msg))
			case engine.StatusFlaky:
				// Green for the job, loud in the annotations (#29, #138).
				fmt.Fprintf(&b, "::warning title=%s::%s\n",
					ghaEscapeProp(res.Suite+" / "+sc.Name), ghaEscapeData(flakyMessage(sc)))
			case engine.StatusXPass:
				// An error annotation, matching the exit code: the fix landed and
				// the scenario has to be promoted, which nobody does for a notice.
				// Under --allow-xpass the exit code is 0, so the annotation drops
				// to a warning rather than failing a job that passed.
				kind := "error"
				if allowXPass {
					kind = "warning"
				}
				fmt.Fprintf(&b, "::%s title=%s::%s\n",
					kind, ghaEscapeProp(res.Suite+" / "+sc.Name), ghaEscapeData(xpassMessage(sc)))
			case engine.StatusXFail:
				// A notice, not a warning: a known bug that is still broken is the
				// expected outcome, and a warning per known bug would train
				// reviewers to ignore the annotation channel.
				fmt.Fprintf(&b, "::notice title=%s::%s\n",
					ghaEscapeProp(res.Suite+" / "+sc.Name), ghaEscapeData("xfail: "+expectFailSummary(sc)))
			}
			// A failed teardown never changes the verdict or the exit code, so
			// it is a warning — the flaky pattern: green for the job, loud in
			// the annotations. Silence here meant a green Actions run with no
			// trace that cleanup of external resources failed.
			if msg := firstStepFailureMessage(sc.Teardown); msg != "" {
				fmt.Fprintf(&b, "::warning title=%s::%s\n",
					ghaEscapeProp(res.Suite+" / "+sc.Name),
					ghaEscapeData("teardown failed: "+msg+" — "+oneLine(stepsDetailText(sc.Teardown))))
			}
		}
		if msg := firstStepFailureMessage(res.Teardown); msg != "" {
			fmt.Fprintf(&b, "::warning title=%s::%s\n",
				ghaEscapeProp(res.Suite),
				ghaEscapeData("suite teardown failed: "+msg+" — "+oneLine(stepsDetailText(res.Teardown))))
		}
		// A suite that errored before any scenario ran (#7) surfaces its cause as
		// an error annotation, so the Actions UI is never silent for a non-zero
		// exit that produced no scenario rows.
		if suiteErroredWithoutScenarios(res) {
			pts := suiteFailurePoints(res)
			for _, p := range pts {
				fmt.Fprintf(&b, "::error title=%s::%s\n",
					ghaEscapeProp(res.Suite+" / "+p.name), ghaEscapeData(p.message))
			}
			// res.Counts() and len(res.Scenarios) are both zero here (no scenario
			// rows), so the synthesized points are the suite's only failure rows.
			// Count them as errored — matching the testcases junit emits and the
			// not-ok points tap emits for the same run — so the ::notice:: summary
			// agrees with the ::error:: annotations just written above instead of
			// reading an all-zero, all-green line.
			agg = agg.Add(engine.Counts{Errored: len(pts)})
			total += len(pts)
		}
		agg = agg.Add(res.Counts())
		total += len(res.Scenarios)
	}
	// A snapshot rewrite is loud in the annotations and green for the job, the
	// flaky pattern: --update-snapshots replaced committed expected results, and
	// a job that did so must not read like an ordinary green run.
	if snapsUpdated > 0 {
		fmt.Fprintf(&b, "::warning title=atago::%s\n", ghaEscapeData(fmt.Sprintf(
			"%s updated by --update-snapshots; the committed expected results were rewritten",
			plural.Count(snapsUpdated, "snapshot", "snapshots"))))
	}
	fmt.Fprintf(&b, "::notice title=atago::%s\n", ghaEscapeData(fmt.Sprintf(
		"%d scenarios: %d passed, %d failed, %d errored, %d skipped%s%s%s",
		total, agg.Passed, agg.Failed, agg.Errored, agg.Skipped,
		flakySuffix(agg)+expectFailSuffix(agg), loadFailureSuffix(len(loadFailures)),
		snapshotSuffix(snapsUpdated))))
	_, err := io.WriteString(w, b.String())
	return err
}

func oneLine(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(s), "\n", "; ")
}

// ghaEscapeData escapes a workflow-command message body per the GitHub spec.
// Beyond the required %/CR/LF encoding it folds any other raw control byte from
// captured output (an ANSI escape, a bell) to U+FFFD, so an annotation never
// carries a byte the Actions UI would mangle.
func ghaEscapeData(s string) string {
	s = strings.ReplaceAll(s, "%", "%25")
	s = strings.ReplaceAll(s, "\r", "%0D")
	s = strings.ReplaceAll(s, "\n", "%0A")
	return sanitizeControlBytes(s)
}

// ghaEscapeProp escapes a workflow-command property value (stricter than data).
func ghaEscapeProp(s string) string {
	s = ghaEscapeData(s)
	s = strings.ReplaceAll(s, ":", "%3A")
	s = strings.ReplaceAll(s, ",", "%2C")
	return s
}
