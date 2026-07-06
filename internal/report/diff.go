package report

import (
	"fmt"
	"strings"

	"github.com/nao1215/atago/internal/assert"
)

// diffMaxInputLines caps each side fed to the LCS table so a pathological
// payload cannot blow up memory; anything beyond is cut with a note (the
// full payloads live in --artifacts-dir sidecars).
const diffMaxInputLines = 400

// diffMaxOutputLines caps the rendered diff, mirroring the excerpt policy on
// Expected/Actual.
const diffMaxOutputLines = 120

// noNewlineMarker is the classic diff annotation for a side that does not end
// in a newline, so trailing-newline differences are visible instead of
// invisible.
const noNewlineMarker = `\ No newline at end of file`

// checkDiff returns the unified diff for a failed check when a diff is the
// clearer rendering: both full payloads present (equals/snapshot failures
// carry them) and at least one side multi-line. Empty string means "keep the
// compact Expected/Actual form". Payloads are masked by the engine before
// they reach the report, so secrets cannot leak through context lines.
func checkDiff(ck *assert.CheckResult) string {
	if ck == nil || ck.OK || len(ck.ArtifactExpected) == 0 || len(ck.ArtifactActual) == 0 {
		return ""
	}
	// Tree-manifest failures carry their own added/removed/changed summary
	// (#25) — a better rendering for trees than a generic unified diff.
	if ck.ArtifactKind == "tree" {
		return ""
	}
	expected, actual := string(ck.ArtifactExpected), string(ck.ArtifactActual)
	if !strings.Contains(strings.TrimRight(expected, "\n"), "\n") && !strings.Contains(strings.TrimRight(actual, "\n"), "\n") {
		return ""
	}
	expectedLabel, actualLabel := "expected", "actual"
	if ck.ArtifactKind == "snapshot" {
		expectedLabel, actualLabel = "snapshot (golden)", "actual"
	}
	return unifiedDiff(expected, actual, expectedLabel, actualLabel)
}

// unifiedDiff renders a classic unified diff (3 context lines) between two
// texts. It returns "" when the texts are equal.
func unifiedDiff(expected, actual, expectedLabel, actualLabel string) string {
	if expected == actual {
		return ""
	}
	a, aTrunc := splitDiffLines(expected)
	b, bTrunc := splitDiffLines(actual)

	ops := diffOps(a, b)

	var out []string
	out = append(out, "--- "+expectedLabel, "+++ "+actualLabel)
	hunks := groupHunks(ops, 3)
	for _, h := range hunks {
		out = append(out, fmt.Sprintf("@@ -%s +%s @@", hunkRange(h.aStart, h.aCount), hunkRange(h.bStart, h.bCount)))
		for _, op := range h.ops {
			switch op.kind {
			case ' ':
				out = append(out, " "+a[op.aIdx])
			case '-':
				out = append(out, "-"+a[op.aIdx])
			case '+':
				out = append(out, "+"+b[op.bIdx])
			}
		}
	}
	// Cap the rendered hunks first so the annotations below always survive.
	if len(out) > diffMaxOutputLines {
		out = append(out[:diffMaxOutputLines], "... (diff truncated)")
	}
	if aTrunc || bTrunc {
		out = append(out, fmt.Sprintf("... (inputs truncated at %d lines; full payloads in artifacts)", diffMaxInputLines))
	}
	// The marker describes a side whose last line lacks a trailing newline; an
	// empty side has no last line, so it never carries the marker.
	if expected != "" && !strings.HasSuffix(expected, "\n") && strings.HasSuffix(actual, "\n") {
		out = append(out, noNewlineMarker+" (expected)")
	}
	if actual != "" && strings.HasSuffix(expected, "\n") && !strings.HasSuffix(actual, "\n") {
		out = append(out, noNewlineMarker+" (actual)")
	}
	return strings.Join(out, "\n")
}

// hunkRange formats one side of a unified-diff `@@` header. A side that
// contributes zero lines (a pure insertion or deletion) is numbered 0, the GNU
// convention patch(1) and strict diff parsers rely on; otherwise the 1-based
// start line and its count.
func hunkRange(start, count int) string {
	if count == 0 {
		return fmt.Sprintf("%d,0", start)
	}
	return fmt.Sprintf("%d,%d", start+1, count)
}

// splitDiffLines splits text into lines for diffing (CRLF folded so an OS
// line-ending artifact never renders a whole-file diff) and caps the count.
func splitDiffLines(text string) (lines []string, truncated bool) {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.TrimSuffix(text, "\n")
	if text == "" {
		return nil, false
	}
	lines = strings.Split(text, "\n")
	if len(lines) > diffMaxInputLines {
		return lines[:diffMaxInputLines], true
	}
	return lines, false
}

// diffOp is one line of the diff script: ' ' common, '-' only in a, '+' only
// in b.
type diffOp struct {
	kind       byte
	aIdx, bIdx int
}

// diffOps computes a line-level diff via a classic LCS table — fine at the
// capped input sizes and dependency-free, matching the repo's small-helper
// style (color.go, excerpt).
func diffOps(a, b []string) []diffOp {
	n, m := len(a), len(b)
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i][j] = lcs[i+1][j]
			} else {
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}
	var ops []diffOp
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case a[i] == b[j]:
			ops = append(ops, diffOp{' ', i, j})
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			ops = append(ops, diffOp{'-', i, -1})
			i++
		default:
			ops = append(ops, diffOp{'+', -1, j})
			j++
		}
	}
	for ; i < n; i++ {
		ops = append(ops, diffOp{'-', i, -1})
	}
	for ; j < m; j++ {
		ops = append(ops, diffOp{'+', -1, j})
	}
	return ops
}

// hunk is one @@-block: a run of changes with up to `context` common lines on
// each side.
type hunk struct {
	aStart, aCount int
	bStart, bCount int
	ops            []diffOp
}

// groupHunks folds the op script into unified-diff hunks with the given
// context width, merging hunks whose context would overlap.
func groupHunks(ops []diffOp, context int) []hunk {
	// Indices of change ops.
	var changes []int
	for i, op := range ops {
		if op.kind != ' ' {
			changes = append(changes, i)
		}
	}
	if len(changes) == 0 {
		return nil
	}
	var hunks []hunk
	start := max(0, changes[0]-context)
	end := min(len(ops), changes[0]+context+1)
	for _, c := range changes[1:] {
		if c-context <= end {
			end = min(len(ops), c+context+1)
			continue
		}
		hunks = append(hunks, buildHunk(ops[start:end]))
		start = max(0, c-context)
		end = min(len(ops), c+context+1)
	}
	hunks = append(hunks, buildHunk(ops[start:end]))
	return hunks
}

func buildHunk(ops []diffOp) hunk {
	h := hunk{ops: ops, aStart: -1, bStart: -1}
	for _, op := range ops {
		if op.aIdx >= 0 {
			if h.aStart < 0 {
				h.aStart = op.aIdx
			}
			h.aCount++
		}
		if op.bIdx >= 0 {
			if h.bStart < 0 {
				h.bStart = op.bIdx
			}
			h.bCount++
		}
	}
	if h.aStart < 0 {
		h.aStart = 0
	}
	if h.bStart < 0 {
		h.bStart = 0
	}
	return h
}

// colorizeDiff applies the console palette to a rendered diff: removals red,
// additions green, hunk headers and file labels dim. A no-op when color is
// off (--ci / NO_COLOR / non-TTY).
func colorizeDiff(on bool, diff string) string {
	if !on || diff == "" {
		return diff
	}
	lines := strings.Split(diff, "\n")
	for i, l := range lines {
		switch {
		// The trailing space disambiguates structural lines from content
		// lines that happen to start with --/++ (YAML document separators).
		case strings.HasPrefix(l, "--- ") || strings.HasPrefix(l, "+++ ") || strings.HasPrefix(l, "@@ "):
			lines[i] = colorize(true, cDim, l)
		case strings.HasPrefix(l, "-"):
			lines[i] = colorize(true, cRed, l)
		case strings.HasPrefix(l, "+"):
			lines[i] = colorize(true, cGreen, l)
		}
	}
	return strings.Join(lines, "\n")
}
