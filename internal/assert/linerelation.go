package assert

import (
	"fmt"
	"sort"
	"strings"

	"github.com/nao1215/atago/internal/plural"
	"github.com/nao1215/atago/internal/spec"
)

// checkLineRelation evaluates the two line relations (#658) — `permutation_of`
// (the same lines in any order) and `subset_of` (every line was in the
// reference) — against a stream.
//
// Both compare LINES as multisets, which is the relation a CLI's output
// actually satisfies when its order is not part of the contract. Counting
// matters: output that printed a line twice where the reference has it once is
// neither a permutation nor a subset, because printing something twice is the
// bug a set comparison would hide.
//
// The comparison runs on CRLF-folded text, like every other stream text
// matcher, so one spec means the same thing under cmd.exe and under sh. The
// lines themselves are otherwise untouched — a blank line inside the output is
// a line, and leading spaces belong to the line they are on — so the matcher
// never quietly normalizes what it is comparing.
func checkLineRelation(name, got, want string, exact bool) *CheckResult {
	relation := "subset_of"
	claim := "every line is in the reference"
	if exact {
		relation = "permutation_of"
		claim = "the same lines in any order"
	}
	desc := fmt.Sprintf("assert %s %s", name, relation)

	actual := splitLines(foldCRLF(got))
	reference := splitLines(foldCRLF(want))
	extra, missing := lineDifference(actual, reference)
	if !exact {
		// A subset may leave reference lines out; only lines the output added
		// break the relation.
		missing = nil
	}
	if len(extra) == 0 && len(missing) == 0 {
		return pass(desc)
	}
	return &CheckResult{
		Desc:             desc,
		Expected:         excerpt(want),
		Actual:           excerpt(got),
		Hint:             relationHint(name, claim, extra, missing),
		ArtifactExpected: []byte(want),
	}
}

// lineDifference reports the multiset difference both ways: the lines the
// output has that the reference does not account for, and the reference lines
// the output never produced. Each is sorted and reported with its multiplicity,
// because "twice" is part of what went wrong.
func lineDifference(actual, reference []string) (extra, missing []string) {
	counts := make(map[string]int, len(reference))
	for _, line := range reference {
		counts[line]++
	}
	for _, line := range actual {
		if counts[line] > 0 {
			counts[line]--
			continue
		}
		extra = append(extra, line)
	}
	for line, n := range counts {
		for range n {
			missing = append(missing, line)
		}
	}
	sort.Strings(extra)
	sort.Strings(missing)
	return extra, missing
}

// relationHint names the lines that broke the relation. A relation that only
// answered "no" would leave the reader comparing two blocks of text by eye —
// which is exactly the work the order-insensitive comparison was reached for.
func relationHint(name, claim string, extra, missing []string) string {
	var parts []string
	if len(missing) > 0 {
		parts = append(parts, fmt.Sprintf("%s missing from %s: %s",
			plural.Count(len(missing), "line", "lines"), name, quoteLines(missing)))
	}
	if len(extra) > 0 {
		parts = append(parts, fmt.Sprintf("%s in %s that the reference does not have: %s",
			plural.Count(len(extra), "line", "lines"), name, quoteLines(extra)))
	}
	return fmt.Sprintf("expected %s to have %s; %s", name, claim, strings.Join(parts, "; "))
}

// hintLineLimit is how many differing lines a hint names before it stops. A
// failure that prints a thousand lines is a wall the reader scrolls past; the
// full streams are in the Expected/Actual blocks and the artifacts directory.
const hintLineLimit = 10

// quoteLines renders differing lines for a hint, quoted so trailing spaces and
// an empty line are visible, and capped so one failure cannot flood the report.
func quoteLines(lines []string) string {
	shown := lines
	suffix := ""
	if len(shown) > hintLineLimit {
		suffix = fmt.Sprintf(" (and %d more)", len(shown)-hintLineLimit)
		shown = shown[:hintLineLimit]
	}
	return spec.StringList(shown).Quoted() + suffix
}
