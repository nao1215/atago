package assert

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nao1215/atago/internal/plural"
	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
)

// countBounds are the occurrence bounds an assertion asked for (#396). Count is
// exact and mutually exclusive with the Min/Max pair; the loader enforces that,
// so at most one shape reaches here.
type countBounds struct {
	Count *int
	Min   *int
	Max   *int
}

// occurrenceTarget is the single countable matcher a count bound applies to:
// either a literal substring or a regexp source. Exactly one is set.
type occurrenceTarget struct {
	Literal string
	Pattern string
}

// countTarget picks the countable matcher off a stream assert. The loader has
// already rejected the ambiguous shapes (both matchers set, a multi-element
// contains), so this reads the one that survived.
func countTarget(s *spec.StreamAssert) occurrenceTarget {
	if s.Matches != nil {
		return occurrenceTarget{Pattern: *s.Matches}
	}
	if len(s.Contains) > 0 {
		return occurrenceTarget{Literal: foldCRLF(s.Contains[0])}
	}
	return occurrenceTarget{}
}

// checkOccurrences counts how many times the target occurs in got and compares
// that against the bounds.
//
// Both matchers count NON-OVERLAPPING occurrences, which is what makes the two
// answer the same question: "aa" occurs once in "aaa", not twice, and a regexp
// counts its own non-overlapping matches (regexp.FindAllString's rule). An
// overlapping count would make `count: 1` mean something different depending on
// which matcher spelled it, and the number a reader wants is almost always "how
// many times did this line get printed".
func checkOccurrences(name, got string, target occurrenceTarget, bounds countBounds) *CheckResult {
	desc := occurrenceDesc(name, target, bounds)

	var (
		n   int
		at  []string
		err error
	)
	switch {
	case target.Pattern != "":
		n, at, err = countPattern(got, target.Pattern)
		if err != nil {
			return &CheckResult{Desc: "assert " + name, Hint: fmt.Sprintf("invalid regexp %q: %v", target.Pattern, err)}
		}
	case target.Literal != "":
		n, at = countLiteral(got, target.Literal)
	default:
		return &CheckResult{Desc: "assert " + name, Hint: "a count bound needs a contains or matches matcher to count"}
	}

	if bounds.satisfied(n) {
		return pass(desc)
	}
	return &CheckResult{
		Desc:     desc,
		Expected: fmt.Sprintf("%s to occur %s", name, bounds.phrase()),
		Actual:   fmt.Sprintf("%s in %s", plural.Count(n, "occurrence", "occurrences"), name),
		Hint:     occurrenceHint(name, target, bounds, n, at),
	}
}

// satisfied reports whether an observed count is within the bounds.
func (b countBounds) satisfied(n int) bool {
	if b.Count != nil {
		return n == *b.Count
	}
	if b.Min != nil && n < *b.Min {
		return false
	}
	if b.Max != nil && n > *b.Max {
		return false
	}
	return true
}

// phrase renders the bounds the way the failure message talks about them.
func (b countBounds) phrase() string {
	switch {
	case b.Count != nil:
		return fmt.Sprintf("exactly %s", plural.Count(*b.Count, "time", "times"))
	case b.Min != nil && b.Max != nil:
		return fmt.Sprintf("between %d and %s", *b.Min, plural.Count(*b.Max, "time", "times"))
	case b.Min != nil:
		return fmt.Sprintf("at least %s", plural.Count(*b.Min, "time", "times"))
	case b.Max != nil:
		return fmt.Sprintf("at most %s", plural.Count(*b.Max, "time", "times"))
	default:
		return "any number of times"
	}
}

// label renders the counted thing for a human: a quoted literal or a /regexp/.
func (t occurrenceTarget) label() string {
	if t.Pattern != "" {
		return "/" + t.Pattern + "/"
	}
	return fmt.Sprintf("%q", t.Literal)
}

func occurrenceDesc(name string, target occurrenceTarget, bounds countBounds) string {
	return fmt.Sprintf("assert %s has %s %s occurring %s", name, target.kindWord(), target.label(), bounds.phrase())
}

func (t occurrenceTarget) kindWord() string {
	if t.Pattern != "" {
		return "regexp"
	}
	return "substring"
}

// maxReportedLocations caps how many match positions a failure lists. Four is
// enough to see the shape of a duplicate ("printed twice, at lines 3 and 7")
// without turning a report into a dump when a pattern matched hundreds of times.
const maxReportedLocations = 4

// occurrenceHint says what was counted and, crucially, WHERE — an off-by-one in
// a duplicate-output bug is only diagnosable if the report names the places the
// matcher hit, so the author can see which one is the extra.
func occurrenceHint(name string, target occurrenceTarget, bounds countBounds, n int, at []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s occurs %s in %s, want %s",
		target.label(), plural.Count(n, "time", "times"), name, bounds.phrase())
	if len(at) == 0 {
		return b.String()
	}
	shown := at
	if len(shown) > maxReportedLocations {
		shown = shown[:maxReportedLocations]
	}
	fmt.Fprintf(&b, " (found at %s", strings.Join(shown, ", "))
	if len(at) > len(shown) {
		fmt.Fprintf(&b, ", and %d more", len(at)-len(shown))
	}
	b.WriteString(")")
	return b.String()
}

// countLiteral counts non-overlapping occurrences of sub in got and describes
// where each landed.
func countLiteral(got, sub string) (int, []string) {
	if sub == "" {
		return 0, nil
	}
	var at []string
	n := 0
	for off := 0; ; {
		i := strings.Index(got[off:], sub)
		if i < 0 {
			break
		}
		n++
		at = append(at, location(got, off+i))
		off += i + len(sub)
	}
	return n, at
}

// countPattern counts non-overlapping regexp matches and describes where each
// landed.
func countPattern(got, pattern string) (int, []string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return 0, nil, err
	}
	locs := re.FindAllStringIndex(got, -1)
	at := make([]string, 0, len(locs))
	for _, loc := range locs {
		at = append(at, location(got, loc[0]))
	}
	return len(locs), at, nil
}

// location renders a byte offset as the 1-based line it falls on, which is how
// a reader looks for it in the output the report already printed.
func location(got string, off int) string {
	return fmt.Sprintf("line %d", strings.Count(got[:off], "\n")+1)
}

// checkFileSize evaluates the byte-size bounds on a file assert (#397).
//
// It stats rather than reads: the whole point of a size bound is that the
// content may be huge (or binary, or partially written), and a bound must not
// need the bytes in memory to have an opinion about them. A directory fails
// with the same phrasing exists: uses — a directory has a size, but it is not
// the file the assertion is about.
func checkFileSize(f *spec.FileAssert, path string) *CheckResult {
	desc := fmt.Sprintf("assert file %q size %s", f.Path, sizeBounds(f).phrase())
	// StatNoFollow, not os.Stat: a program under test can plant a symlink at the
	// assertion target, and following it would report the size of a host file
	// outside the workdir — the same disclosure ReadFileNoFollow refuses on the
	// read path (issue #16).
	info, err := security.StatNoFollow(path)
	if err != nil {
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("stat-able file %q", f.Path),
			Actual:   err.Error(),
			Hint:     fmt.Sprintf("could not stat file %q: %v", f.Path, err),
		}
	}
	if info.IsDir() {
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("file %q", f.Path),
			Actual:   fmt.Sprintf("%q is a directory", f.Path),
			Hint:     fmt.Sprintf("%q is a directory, not a file; use a dir: assertion for it", f.Path),
		}
	}
	got := info.Size()
	if sizeBounds(f).satisfied(got) {
		return pass(desc)
	}
	return &CheckResult{
		Desc:     desc,
		Expected: fmt.Sprintf("file %q %s", f.Path, sizeBounds(f).phrase()),
		Actual:   byteCount(got),
		Hint:     sizeHint(f, got),
	}
}

// sizeBounds are the byte bounds an assert asked for, in the same shape the
// occurrence bounds use.
type sizeBoundSet struct {
	Exact *int64
	Min   *int64
	Max   *int64
}

func sizeBounds(f *spec.FileAssert) sizeBoundSet {
	return sizeBoundSet{Exact: f.Size, Min: f.MinSize, Max: f.MaxSize}
}

func (b sizeBoundSet) satisfied(n int64) bool {
	if b.Exact != nil {
		return n == *b.Exact
	}
	if b.Min != nil && n < *b.Min {
		return false
	}
	if b.Max != nil && n > *b.Max {
		return false
	}
	return true
}

func (b sizeBoundSet) phrase() string {
	switch {
	case b.Exact != nil:
		return fmt.Sprintf("is exactly %s", byteCount(*b.Exact))
	case b.Min != nil && b.Max != nil:
		return fmt.Sprintf("is between %s and %s", byteCount(*b.Min), byteCount(*b.Max))
	case b.Min != nil:
		return fmt.Sprintf("is at least %s", byteCount(*b.Min))
	case b.Max != nil:
		return fmt.Sprintf("is at most %s", byteCount(*b.Max))
	default:
		return "is any size"
	}
}

// byteCount renders a length the way the failure text reads it.
func byteCount(n int64) string {
	return plural.Count64(n, "byte", "bytes")
}

// sizeHint states the miss and, for a near miss on an exact bound, names the
// two things that are almost always behind it. A one-byte difference is a
// trailing newline; a difference equal to the line count is CRLF — and the
// author cannot see either in a number alone.
func sizeHint(f *spec.FileAssert, got int64) string {
	base := fmt.Sprintf("file %q %s, want it to %s", f.Path, byteCount(got), sizeBounds(f).phrase())
	if f.Size == nil {
		return base
	}
	switch got - *f.Size {
	case 1:
		return base + " (one byte over: a trailing newline is the usual cause)"
	case -1:
		return base + " (one byte under: a missing trailing newline is the usual cause)"
	}
	return base
}
