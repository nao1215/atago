package report

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/assert"
)

// TestDiffOps_Reconstruct is the fundamental diff invariant: the edit script
// diffOps(a, b) must, in order, reconstruct a from its keep+remove ops and b
// from its keep+add ops. A script that violates this renders a misleading
// failure diff to the user. The inputs draw from a tiny alphabet with repeats so
// the LCS has real common subsequences to find, not just wholesale replacement.
func TestDiffOps_Reconstruct(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewSource(1))
	alphabet := []string{"", "a", "b", "c", "d", "a", "b"}
	gen := func() []string {
		out := make([]string, rng.Intn(9))
		for i := range out {
			out[i] = alphabet[rng.Intn(len(alphabet))]
		}
		return out
	}
	for iter := range 5000 {
		a, b := gen(), gen()
		var gotA, gotB []string
		for _, op := range diffOps(a, b) {
			switch op.kind {
			case ' ':
				gotA = append(gotA, a[op.aIdx])
				gotB = append(gotB, b[op.bIdx])
			case '-':
				gotA = append(gotA, a[op.aIdx])
			case '+':
				gotB = append(gotB, b[op.bIdx])
			}
		}
		if !slices.Equal(gotA, a) {
			t.Fatalf("iter %d: keep+remove ops do not reconstruct a\n a=%q got=%q b=%q", iter, a, gotA, b)
		}
		if !slices.Equal(gotB, b) {
			t.Fatalf("iter %d: keep+add ops do not reconstruct b\n a=%q b=%q got=%q", iter, a, b, gotB)
		}
	}
}

// TestUnifiedDiff_HunkPlacement pins hunk generation for a change at the
// start, middle, and end of a document (#28).
func TestUnifiedDiff_HunkPlacement(t *testing.T) {
	t.Parallel()
	base := "a\nb\nc\nd\ne\nf\ng\nh\ni\nj\n"

	middle := unifiedDiff(base, strings.Replace(base, "e\n", "E\n", 1), "expected", "actual")
	for _, want := range []string{"--- expected", "+++ actual", "@@ -2,7 +2,7 @@", "-e", "+E", " d", " f"} {
		if !strings.Contains(middle, want) {
			t.Errorf("middle diff missing %q:\n%s", want, middle)
		}
	}
	if strings.Contains(middle, " a\n") || strings.Contains(middle, " j") {
		t.Errorf("middle diff leaked far context:\n%s", middle)
	}

	start := unifiedDiff(base, strings.Replace(base, "a\n", "A\n", 1), "expected", "actual")
	if !strings.Contains(start, "@@ -1,4 +1,4 @@") || !strings.Contains(start, "-a") || !strings.Contains(start, "+A") {
		t.Errorf("start diff wrong:\n%s", start)
	}

	end := unifiedDiff(base, strings.Replace(base, "j\n", "J\n", 1), "expected", "actual")
	if !strings.Contains(end, "-j") || !strings.Contains(end, "+J") {
		t.Errorf("end diff wrong:\n%s", end)
	}
}

// TestUnifiedDiff_CRLFAndTrailingNewline proves CRLF inputs fold (an OS
// artifact must not diff every line) while both a trailing-newline difference
// and a line-ending-only difference are rendered explicitly. A byte-exact
// failure caused only by CRLF-vs-LF must never render a blank diff: the folded
// lines are equal, so there are no content hunks, but the assertion still
// failed and the reason must be surfaced.
func TestUnifiedDiff_CRLFAndTrailingNewline(t *testing.T) {
	t.Parallel()
	// CRLF is the only difference: the folded lines are identical, so no
	// content hunk is produced — but the diff must not be blank and must point
	// at the line-ending difference.
	got := unifiedDiff("alpha\nbeta\n", "alpha\r\nbeta\r\n", "expected", "actual")
	if strings.TrimSpace(got) == "" {
		t.Fatalf("CRLF-only difference rendered a blank diff")
	}
	if !strings.Contains(got, "CRLF") && !strings.Contains(strings.ToLower(got), "line ending") {
		t.Errorf("CRLF-only difference should point at the line-ending difference:\n%s", got)
	}
	// Folding still holds: the content lines must not each diff as changed.
	for _, l := range strings.Split(got, "\n") {
		if (strings.HasPrefix(l, "-") && !strings.HasPrefix(l, "---")) ||
			(strings.HasPrefix(l, "+") && !strings.HasPrefix(l, "+++")) {
			t.Errorf("CRLF-only difference produced content hunks:\n%s", got)
		}
	}

	got = unifiedDiff("a\nb\n", "a\nb", "expected", "actual")
	if !strings.Contains(got, noNewlineMarker+" (actual)") {
		t.Errorf("trailing-newline difference not annotated:\n%s", got)
	}
}

// TestUnifiedDiff_Truncation proves oversized inputs and long diffs are cut
// with explicit notes.
func TestUnifiedDiff_Truncation(t *testing.T) {
	t.Parallel()
	var a, b strings.Builder
	for i := 0; i < diffMaxInputLines+50; i++ {
		a.WriteString("same\n")
		b.WriteString("diff\n")
	}
	got := unifiedDiff(a.String(), b.String(), "expected", "actual")
	if !strings.Contains(got, "(diff truncated)") {
		t.Errorf("long diff not truncated:\n...%s", got[len(got)-200:])
	}
	if !strings.Contains(got, "inputs truncated") {
		t.Errorf("oversized inputs not annotated")
	}
}

// TestCheckDiff_Eligibility proves the diff renders only for multi-line
// two-sided failures; single-line failures keep the compact form.
func TestCheckDiff_Eligibility(t *testing.T) {
	t.Parallel()
	multi := &assert.CheckResult{
		ArtifactExpected: []byte("one\ntwo\nthree\n"),
		ArtifactActual:   []byte("one\nTWO\nthree\n"),
	}
	if got := checkDiff(multi); !strings.Contains(got, "-two") || !strings.Contains(got, "+TWO") {
		t.Errorf("multi-line diff = %q", got)
	}
	single := &assert.CheckResult{ArtifactExpected: []byte("x\n"), ArtifactActual: []byte("y\n")}
	if got := checkDiff(single); got != "" {
		t.Errorf("single-line failure should keep the compact form, got %q", got)
	}
	oneSided := &assert.CheckResult{ArtifactActual: []byte("a\nb\n")}
	if got := checkDiff(oneSided); got != "" {
		t.Errorf("one-sided failure should not diff, got %q", got)
	}
	snap := &assert.CheckResult{
		ArtifactKind:     "snapshot",
		ArtifactExpected: []byte("a\nb\n"),
		ArtifactActual:   []byte("a\nc\n"),
	}
	if got := checkDiff(snap); !strings.Contains(got, "--- snapshot (golden)") {
		t.Errorf("snapshot diff should label the golden side, got %q", got)
	}
}

// TestColorizeDiff proves red/green/dim application and the NO_COLOR-off
// path's byte-stability.
func TestColorizeDiff(t *testing.T) {
	t.Parallel()
	diff := "--- expected\n+++ actual\n@@ -1,1 +1,1 @@\n-old\n+new\n context"
	if got := colorizeDiff(false, diff); got != diff {
		t.Errorf("color off must be byte-stable")
	}
	got := colorizeDiff(true, diff)
	if !strings.Contains(got, cRed+"-old"+cReset) || !strings.Contains(got, cGreen+"+new"+cReset) {
		t.Errorf("colorized = %q", got)
	}
	if !strings.Contains(got, cDim+"@@ -1,1 +1,1 @@"+cReset) {
		t.Errorf("hunk header not dimmed: %q", got)
	}
	if strings.Contains(got, cRed+"--- expected") {
		t.Errorf("file label wrongly red: %q", got)
	}
	// A removed content line starting with "--" (a YAML document separator
	// becomes "---" after the diff marker) must stay red, not dim.
	yaml := colorizeDiff(true, "-"+"--\n+"+"++")
	if !strings.Contains(yaml, cRed+"---"+cReset) {
		t.Errorf("removed YAML separator not red: %q", yaml)
	}
	if !strings.Contains(yaml, cGreen+"+++"+cReset) {
		t.Errorf("added ++ content not green: %q", yaml)
	}
}

// TestUnifiedDiff_OutputTruncationAndMultiHunk covers the rendered-output cap
// (a diff longer than diffMaxOutputLines is truncated with a marker) and
// multi-hunk grouping (two changes separated by more than 2*context common lines
// produce two @@ headers), plus the actual-side no-newline marker.
func TestUnifiedDiff_OutputTruncationAndMultiHunk(t *testing.T) {
	t.Parallel()

	// Output truncation: 130 wholly-different lines (< diffMaxInputLines, so no
	// input truncation) diff to 260 hunk lines, exceeding diffMaxOutputLines.
	var exp, act strings.Builder
	for i := 0; i < 130; i++ {
		fmt.Fprintf(&exp, "old-%d\n", i)
		fmt.Fprintf(&act, "new-%d\n", i)
	}
	got := unifiedDiff(exp.String(), act.String(), "expected", "actual")
	if !strings.Contains(got, "... (diff truncated)") {
		t.Errorf("expected output-truncation marker, got %d bytes", len(got))
	}
	if strings.Contains(got, "inputs truncated") {
		t.Errorf("130 lines should not trip input truncation:\n%s", got)
	}

	// Multi-hunk: a long shared body with an isolated change at each end.
	base := make([]string, 0, 40)
	for i := 0; i < 40; i++ {
		base = append(base, fmt.Sprintf("line-%d", i))
	}
	a := append([]string{"HEAD-A"}, base...)
	a = append(a, "TAIL-A")
	b := append([]string{"HEAD-B"}, base...)
	b = append(b, "TAIL-B")
	md := unifiedDiff(strings.Join(a, "\n"), strings.Join(b, "\n"), "expected", "actual")
	if n := strings.Count(md, "@@ -"); n < 2 {
		t.Errorf("expected >= 2 hunks for two far-apart changes, got %d:\n%s", n, md)
	}

	// The actual-side no-newline marker fires when expected ends in \n but actual
	// does not (the mirror of the existing expected-side test).
	nl := unifiedDiff("a\n", "a", "expected", "actual")
	if !strings.Contains(nl, noNewlineMarker+" (actual)") {
		t.Errorf("actual-side no-newline marker missing:\n%s", nl)
	}
}

// TestUnifiedDiff_ZeroCountHunkHeader is a regression: a hunk whose one side
// contributes zero lines (a pure insertion or deletion) must number that side
// 0, per the GNU unified-diff convention that patch(1) and strict parsers rely
// on — not 1. The empty side also has no trailing-newline state to report.
func TestUnifiedDiff_ZeroCountHunkHeader(t *testing.T) {
	t.Parallel()
	// Pure insertion: the expected side is empty (0 lines).
	got := unifiedDiff("", "line1\nline2\n", "expected", "actual")
	if !strings.Contains(got, "@@ -0,0 +1,2 @@") {
		t.Errorf("insertion hunk header wrong:\n%s\nwant a \"@@ -0,0 +1,2 @@\" header", got)
	}
	if strings.Contains(got, "No newline at end of file (expected)") {
		t.Errorf("spurious no-newline marker for an empty expected side:\n%s", got)
	}
	// Pure deletion: the actual side is empty (0 lines).
	got = unifiedDiff("line1\nline2\n", "", "expected", "actual")
	if !strings.Contains(got, "@@ -1,2 +0,0 @@") {
		t.Errorf("deletion hunk header wrong:\n%s\nwant a \"@@ -1,2 +0,0 @@\" header", got)
	}
	if strings.Contains(got, "No newline at end of file (actual)") {
		t.Errorf("spurious no-newline marker for an empty actual side:\n%s", got)
	}
	// When BOTH non-empty sides lack a trailing newline, both are annotated —
	// the marker is judged per side, not by comparing the two sides.
	got = unifiedDiff("a", "b", "expected", "actual")
	if !strings.Contains(got, noNewlineMarker+" (expected)") || !strings.Contains(got, noNewlineMarker+" (actual)") {
		t.Errorf("both-no-newline sides not both annotated:\n%s", got)
	}
}
