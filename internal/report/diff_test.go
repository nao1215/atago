package report

import (
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/assert"
)

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
// artifact must not diff every line) while a trailing-newline difference is
// rendered explicitly.
func TestUnifiedDiff_CRLFAndTrailingNewline(t *testing.T) {
	t.Parallel()
	if got := unifiedDiff("a\r\nb\r\n", "a\nb\n", "expected", "actual"); !strings.Contains(got, noNewlineMarker) && got != "" {
		// CRLF folding makes the contents equal; only the byte-level
		// difference remains, which the caller treats as equal lines. The
		// diff is header-only in that case — assert no +/- content lines.
		for _, l := range strings.Split(got, "\n") {
			if strings.HasPrefix(l, "-") && !strings.HasPrefix(l, "---") {
				t.Errorf("CRLF-only difference produced content hunks:\n%s", got)
			}
		}
	}
	got := unifiedDiff("a\nb\n", "a\nb", "expected", "actual")
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
}
