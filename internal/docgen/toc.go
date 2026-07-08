package docgen

import (
	"fmt"
	"sort"
	"strings"
)

// summary aggregates document-wide counts for the top-of-file summary block
// (#66): how many suites and scenarios the document covers and how often each
// tag appears. It stays deterministic — tags are emitted in sorted order — so the
// committed docs are byte-stable.
type summary struct {
	suites    int
	scenarios int
	tagCounts map[string]int
}

func computeSummary(sources []Source) summary {
	s := summary{tagCounts: map[string]int{}}
	for _, src := range sources {
		s.suites++
		for i := range src.Spec.Scenarios {
			s.scenarios++
			for _, tag := range src.Spec.Scenarios[i].Tags {
				s.tagCounts[tag]++
			}
		}
	}
	return s
}

// tagLine renders the tag summary as a deterministic, sorted list like
// "`smoke` (3), `network` (2)". Empty when the document has no tags.
func (s summary) tagLine() string {
	if len(s.tagCounts) == 0 {
		return ""
	}
	tags := make([]string, 0, len(s.tagCounts))
	for t := range s.tagCounts {
		tags = append(tags, t)
	}
	sort.Strings(tags)
	parts := make([]string, 0, len(tags))
	for _, t := range tags {
		parts = append(parts, fmt.Sprintf("`%s` (%d)", t, s.tagCounts[t]))
	}
	return strings.Join(parts, ", ")
}

// pluralize returns "N thing" / "N things".
func pluralize(n int, thing string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, thing)
	}
	return fmt.Sprintf("%d %ss", n, thing)
}

// anchorer assigns GitHub-compatible heading anchors, deduplicating repeats in
// document order the way GitHub does (a repeated slug gets a "-1", "-2", …
// suffix). It is used to build a table of contents whose links resolve to the
// generated suite/scenario headings.
type anchorer struct {
	used map[string]int
}

func newAnchorer() *anchorer { return &anchorer{used: map[string]int{}} }

// anchor slugifies text the way GitHub does for heading anchors: lowercase, drop
// characters that are not letters/digits/space/hyphen, then convert spaces to
// hyphens. Repeated slugs get an incrementing numeric suffix.
func (a *anchorer) anchor(text string) string {
	slug := slugify(text)
	n := a.used[slug]
	a.used[slug]++
	if n == 0 {
		return slug
	}
	return fmt.Sprintf("%s-%d", slug, n)
}

// Anchors assigns GitHub-compatible heading anchors to headings in document
// order, applying the same slugging and duplicate-suffix rules the generated
// docs use. Exported so drift guards over hand-written docs (doc/cookbook.md)
// resolve anchors exactly the way this package does, instead of keeping a
// second slugger that could drift.
func Anchors(headings []string) []string {
	a := newAnchorer()
	out := make([]string, len(headings))
	for i, h := range headings {
		out[i] = a.anchor(h)
	}
	return out
}

func slugify(text string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(text) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-':
			b.WriteRune('-')
		case r == '_':
			b.WriteRune('_')
			// GitHub keeps underscores; other punctuation is dropped.
		}
	}
	return b.String()
}

// tableOfContents renders a nested TOC linking each suite and its scenarios. The
// anchors are computed in the same order the headings are later emitted, so the
// links resolve. The returned anchorer is reused so the emitted headings and the
// TOC agree on suffixing.
func tableOfContents(sources []Source) (string, *anchorer) {
	// First pass: reserve the anchors this document's headings will consume, in
	// emission order, so TOC links match. The H1 title and the "Contents" heading
	// come first, exactly as writeHeader emits them.
	a := newAnchorer()
	a.anchor("atago Behavior Specs")
	a.anchor("Summary")
	a.anchor("Contents")

	var b strings.Builder
	for _, src := range sources {
		suiteAnchor := a.anchor(src.Spec.Suite.Name)
		fmt.Fprintf(&b, "- [%s](#%s) — %s\n",
			mdEscape(src.Spec.Suite.Name), suiteAnchor, pluralize(len(src.Spec.Scenarios), "scenario"))
		for i := range src.Spec.Scenarios {
			scName := src.Spec.Scenarios[i].Name
			scAnchor := a.anchor("Scenario: " + scName)
			fmt.Fprintf(&b, "  - [%s](#%s)\n", mdEscape(scName), scAnchor)
		}
	}
	return b.String(), a
}

// mdEscape escapes the few Markdown link-text characters that would otherwise
// break a list link. Brackets are the practical concern for scenario names.
func mdEscape(s string) string {
	r := strings.NewReplacer("[", "\\[", "]", "\\]")
	return r.Replace(s)
}
