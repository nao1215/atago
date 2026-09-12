package loader

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/diag"
	"github.com/nao1215/atago/internal/spec"
)

func validateAssert(add addFunc, where string, a *spec.Assert, mockNames map[string]bool) {
	targets := a.SetTargets()
	if len(targets) == 0 {
		add(diag.ChooseAtLeastOne, "%s.assert: must set at least one assertion target (got none)", where)
		return
	}
	// Each set target is an independent check and all must hold, so validate every
	// one of them (an assert may combine, e.g., exit_code + stdout + file).
	for _, t := range targets {
		validateAssertTarget(add, where, a, t, mockNames)
	}
}

// validateAssertTarget checks the shape of a single assertion target family.
func validateAssertTarget(add addFunc, where string, a *spec.Assert, target spec.AssertTarget, mockNames map[string]bool) {
	switch target {
	case spec.AssertExitCode:
		// Scalar/mapping shape enforced by ExitCode.UnmarshalYAML; the one-of
		// rule and the `in` set contents are semantic checks (#19).
		validateExitCode(add, where+".assert.exit_code", a.ExitCode)
	case spec.AssertStdout:
		validateStream(add, where+".assert.stdout", a.Stdout)
	case spec.AssertStderr:
		validateStream(add, where+".assert.stderr", a.Stderr)
	case spec.AssertFile:
		validateFile(add, where+".assert.file", a.File)
	case spec.AssertBody:
		validateStream(add, where+".assert.body", a.Body)
	case spec.AssertHeader:
		validateHeaderMatch(add, where+".assert.header", a.Header)
	case spec.AssertRows:
		validateStream(add, where+".assert.rows", a.Rows)
	case spec.AssertMessage:
		validateStream(add, where+".assert.message", a.Message)
	case spec.AssertValue:
		validateStream(add, where+".assert.value", a.Value)
	case spec.AssertImage:
		validateImage(add, where+".assert.image", a.Image)
	case spec.AssertDir:
		validateDir(add, where+".assert.dir", a.Dir)
	case spec.AssertMock:
		validateMockAssert(add, where+".assert.mock", a.Mock, mockNames)
	case spec.AssertScreen:
		validateScreen(add, where+".assert.screen", a.Screen)
	case spec.AssertDuration:
		validateDuration(add, where+".assert.duration", a.Duration)
	case spec.AssertChanges:
		validateChanges(add, where+".assert.changes", a.Changes)
	case spec.AssertPDF:
		validatePDF(add, where+".assert.pdf", a.PDF)
	case spec.AssertStatus, spec.AssertGRPCStatus:
		// Both are a bare *int; there is no shape to validate beyond the decode.
	}
}

// validateExitCode checks the exit_code assertion (#19): exactly one of the
// bare-int / {not} / {in} forms, and an `in` set that is non-empty with unique
// values — a duplicated code is authoring confusion, not a wider contract.
func validateExitCode(add addFunc, where string, e *spec.ExitCode) {
	set := 0
	if e.Equals != nil {
		set++
	}
	if e.Not != nil {
		set++
	}
	if e.In != nil {
		set++
	}
	if set == 0 {
		add(diag.ChooseExactlyOne, "%s must be an int, {not: int}, or {in: [int, ...]}", where)
		return
	}
	if set > 1 {
		add(diag.ChooseExactlyOne, "%s: set exactly one of a bare int, not, or in", where)
		return
	}
	if e.In != nil {
		if len(e.In) == 0 {
			add(diag.EmptyList, "%s.in must list at least one accepted exit code", where)
		}
		seen := make(map[int]bool, len(e.In))
		for _, n := range e.In {
			if seen[n] {
				add(diag.DuplicateEntry, "%s.in lists %d more than once", where, n)
			}
			seen[n] = true
		}
	}
}

// streamExclusiveMatchers pin the whole stream, so each is used on its own. The
// text matchers (contains/not_contains/matches/not_matches) may be combined —
// they all have to hold — which is the common "output has X but not Y" shape.
var streamExclusiveMatchers = map[string]bool{
	"empty": true, "equals": true, "not_equals": true,
	"json": true, "yaml": true, "snapshot": true,
}

func validateStream(add addFunc, where string, s *spec.StreamAssert) {
	// trim is a store-only selector (#158), not an assertion matcher.
	if s.Trim != nil {
		add(diag.KeyNotHere, "%s: trim is only valid in a store source, not an assertion", where)
	}
	matchers := s.SetMatchers()
	if len(matchers) == 0 {
		add(diag.ChooseAtLeastOne, "%s: must set at least one matcher (empty/contains/not_contains/matches/not_matches/equals/not_equals/json/yaml/snapshot)", where)
	}
	// A whole-stream matcher cannot be combined with anything else; only the
	// text matchers compose.
	var exclusive []string
	for _, m := range matchers {
		if streamExclusiveMatchers[m] {
			exclusive = append(exclusive, m)
		}
	}
	if len(exclusive) > 0 && len(matchers) > 1 {
		add(diag.ExclusiveKeys, "%s: %v cannot be combined with another matcher (only contains/not_contains/matches/not_matches may be combined)", where, exclusive)
	}
	validateJSONChecks(add, where+".json", s.JSON)
	validateJSONChecks(add, where+".yaml", s.YAML)
	if s.Matches != nil {
		validateRegexp(add, where, "matches", *s.Matches)
	}
	if s.NotMatches != nil {
		validateRegexp(add, where, "not_matches", *s.NotMatches)
	}
	validateStringList(add, where, "contains", s.Contains)
	validateStringList(add, where, "not_contains", s.NotContains)
	validateContainsOverlap(add, where, s.Contains, s.NotContains)
	if s.Matches != nil && s.NotMatches != nil && *s.Matches == *s.NotMatches {
		add(diag.ExclusiveKeys, "%s: matches and not_matches are the same pattern %q, so no output can satisfy both", where, *s.Matches)
	}

	if s.HasCount() {
		validateStreamCount(add, where, s, matchers)
	}

	if s.Line != nil {
		if *s.Line < 1 {
			add(diag.OutOfRange, "%s.line must be >= 1 (got %d)", where, *s.Line)
		}
		// line narrows the stream to one line, so it only composes with the
		// text matchers — json/snapshot operate on the whole document.
		if len(s.JSON) > 0 || len(s.YAML) > 0 || s.Snapshot != "" {
			add(diag.ExclusiveKeys, "%s.line cannot be combined with json/yaml/snapshot (use contains/matches/equals/empty)", where)
		}
	}
}

// validateStreamCount checks a stream assert's occurrence bounds (#396). A
// count bound needs exactly one countable matcher to count; contains and
// matches both compose freely without it, so the ambiguity only appears once a
// bound is written, and it has to be refused there rather than resolved by a
// precedence rule nobody would remember.
func validateStreamCount(add addFunc, where string, s *spec.StreamAssert, matchers []string) {
	validateCountBounds(add, where, s.Count, s.MinCount, s.MaxCount)
	countable := 0
	if s.Contains != nil {
		countable++
		if len(s.Contains) > 1 {
			add(diag.ExclusiveKeys, "%s: count applies to one substring; contains lists %d (write one assert per substring)", where, len(s.Contains))
		}
	}
	if s.Matches != nil {
		countable++
		// Zero-width matches are counted, so an empty-matching pattern under a
		// bound reports len(input)+1 occurrences of nothing: `max_count: 0` — the
		// documented way to say "never" — can then never pass, and the reported
		// count is about the length of the stream rather than about the pattern.
		// The pattern is legitimate without a bound, which is why the refusal
		// lives here rather than in validateRegexp.
		if matchesEmpty(*s.Matches) {
			add(diag.VacuousMatcher, "%s.matches %q matches the empty string, so a count bound counts one zero-width match per position rather than real occurrences; require at least one character (e.g. \"q+\")", where, *s.Matches)
		}
	}
	switch {
	case countable == 0:
		add(diag.KeyNeedsAnother, "%s: count/min_count/max_count need a contains or matches matcher to count", where)
	case countable > 1:
		add(diag.ChooseExactlyOne, "%s: count/min_count/max_count need exactly one countable matcher, but both contains and matches are set", where)
	}
	// The other text matchers stay meaningful next to a count, but the
	// whole-stream ones do not: `equals` already pins every byte.
	for _, m := range matchers {
		if streamExclusiveMatchers[m] {
			add(diag.ExclusiveKeys, "%s: count/min_count/max_count cannot be combined with %s", where, m)
		}
	}
}

// validateCountBounds checks the shape shared by every occurrence bound (#396):
// counts are non-negative, the exact form excludes the range form, and a range
// has to be satisfiable. An unsatisfiable range (min above max) is an authoring
// mistake that would otherwise fail at runtime with a message about the output
// rather than about the spec.
func validateCountBounds(add addFunc, where string, exact, minCount, maxCount *int) {
	for name, v := range map[string]*int{"count": exact, "min_count": minCount, "max_count": maxCount} {
		if v != nil && *v < 0 {
			add(diag.NegativeValue, "%s.%s must be >= 0 (got %d)", where, name, *v)
		}
	}
	if exact != nil && (minCount != nil || maxCount != nil) {
		add(diag.ExclusiveKeys, "%s: count is the exact form; use min_count/max_count for a range, not both", where)
	}
	if minCount != nil && maxCount != nil && *minCount > *maxCount {
		add(diag.EmptyInterval, "%s: min_count %d is greater than max_count %d, so nothing can satisfy it", where, *minCount, *maxCount)
	}
}

// validateSizeBounds is validateCountBounds for the byte-size family (#397).
func validateSizeBounds(add addFunc, where string, exact, minSize, maxSize *int64) {
	for name, v := range map[string]*int64{"size": exact, "min_size": minSize, "max_size": maxSize} {
		if v != nil && *v < 0 {
			add(diag.NegativeValue, "%s.%s must be >= 0 (got %d)", where, name, *v)
		}
	}
	if exact != nil && (minSize != nil || maxSize != nil) {
		add(diag.ExclusiveKeys, "%s: size is the exact form; use min_size/max_size for a range, not both", where)
	}
	if minSize != nil && maxSize != nil && *minSize > *maxSize {
		add(diag.EmptyInterval, "%s: min_size %d is greater than max_size %d, so nothing can satisfy it", where, *minSize, *maxSize)
	}
}

// requireExactlyOne reports a matcher block that set none or several of its
// alternatives. The list of alternatives is written once here per call site:
// spelling it twice — for "none" and for "several" — is how the two messages
// drift apart when a matcher is added.
func requireExactlyOne(add addFunc, where, alternatives string, n int) {
	requireOneOf(add, where, alternatives, alternatives, n)
}

// requireOneOf is requireExactlyOne for a block whose two rules cover different
// key sets: satisfied names every key that satisfies the at-least-one rule, and
// exclusive names only the keys that may not appear together. They differ where
// a family composes instead of competing — a file assert is satisfied by a size
// bound alone, which no content matcher excludes — and naming the exclusive set
// in the "none" message sends the author looking for a matcher they may not
// need.
func requireOneOf(add addFunc, where, satisfied, exclusive string, n int) {
	switch {
	case n == 0:
		add(diag.ChooseExactlyOne, "%s: must set one of %s", where, satisfied)
	case n > 1:
		add(diag.ChooseExactlyOne, "%s: must set exactly one of %s", where, exclusive)
	}
}

// fileContentMatchers are the file assert's mutually exclusive content matchers:
// each makes its own claim about the bytes, so two of them on one assert is an
// authoring mistake.
const fileContentMatchers = "exists/contains/not_contains/executable/equals/equals_file/json/snapshot"

// fileSizeMatchers are the file assert's size bounds. They compose with a
// content matcher and stand alone (#397), so they satisfy the at-least-one rule
// without joining the exclusive family.
const fileSizeMatchers = "size/min_size/max_size"

func validateFile(add addFunc, where string, f *spec.FileAssert) {
	if f.Path == "" {
		add(diag.RequiredKey, "%s.path is required", where)
	}
	// text is a store-only selector (#158), not an assertion matcher.
	if f.Text != nil {
		add(diag.KeyNotHere, "%s: text is only valid in a store source, not an assertion", where)
	}
	n := 0
	if f.Exists != nil {
		n++
	}
	if f.Contains != nil {
		n++
		validateStringList(add, where, "contains", f.Contains)
	}
	if f.NotContains != nil {
		n++
		validateStringList(add, where, "not_contains", f.NotContains)
	}
	if f.Executable != nil {
		n++
	}
	if f.Equals != nil {
		n++
	}
	if f.EqualsFile != nil {
		n++
		if *f.EqualsFile == "" {
			add(diag.EmptyValue, "%s.equals_file must not be empty", where)
		}
	}
	if len(f.JSON) > 0 {
		n++
		validateJSONChecks(add, where+".json", f.JSON)
	}
	if f.Snapshot != "" {
		n++
	}

	// The size family composes with the content matchers and can also stand
	// alone (#397): "the failed run left a zero-byte file" is a complete claim.
	// So it satisfies the at-least-one rule without joining the one-of family.
	if f.HasSize() {
		validateSizeBounds(add, where, f.Size, f.MinSize, f.MaxSize)
		// A size bound has to stat a regular file; `exists: false` demands there
		// be nothing to stat. Nothing can satisfy both, so it is an authoring
		// mistake rather than a contract.
		if f.Exists != nil && !*f.Exists {
			add(diag.ExclusiveKeys, "%s: size/min_size/max_size cannot be combined with exists: false — an absent file has no size", where)
		}
		if f.Snapshot != "" {
			add(diag.ExclusiveKeys, "%s: size/min_size/max_size cannot be combined with snapshot — a snapshot already pins every byte", where)
		}
		if n > 1 {
			add(diag.ChooseExactlyOne, "%s: must set exactly one of %s", where, fileContentMatchers)
		}
	} else {
		requireOneOf(add, where, fileContentMatchers+"/"+fileSizeMatchers, fileContentMatchers, n)
	}

	// A count bound counts the single `contains` substring (#396). A file assert
	// has no `matches`, so contains is the only countable matcher here.
	if f.HasCount() {
		validateCountBounds(add, where, f.Count, f.MinCount, f.MaxCount)
		switch {
		case f.Contains == nil:
			add(diag.KeyNeedsAnother, "%s: count/min_count/max_count need a contains matcher to count", where)
		case len(f.Contains) > 1:
			add(diag.ExclusiveKeys, "%s: count applies to one substring; contains lists %d (write one assert per substring)", where, len(f.Contains))
		}
	}
}

func validateHeaderMatch(add addFunc, where string, h *spec.HeaderMatch) {
	if h.Name == "" {
		add(diag.RequiredKey, "%s.name is required", where)
	}
	n := 0
	if h.Contains != nil {
		n++
	}
	if h.Equals != nil {
		n++
	}
	if h.Matches != nil {
		n++
		validateRegexp(add, where, "matches", *h.Matches)
	}
	requireExactlyOne(add, where, "contains/equals/matches", n)
}

// validateJSONChecks validates each check in a json/yaml matcher list (#156).
// The empty-list case is already rejected by JSONChecks.UnmarshalYAML; a nil
// list (the matcher is absent) is a no-op. Each entry names its index in the
// error location so a spec author can find the failing check.
func validateJSONChecks(add addFunc, where string, list spec.JSONChecks) {
	if len(list) == 1 {
		validateJSON(add, where, &list[0])
		return
	}
	for i := range list {
		validateJSON(add, fmt.Sprintf("%s[%d]", where, i), &list[i])
	}
}

func validateJSON(add addFunc, where string, j *spec.JSONAssert) {
	if j.Path == "" {
		add(diag.RequiredKey, "%s.path is required", where)
	}
	n := 0
	if j.HasEquals() {
		n++
	}
	if j.Matches != nil {
		n++
		validateRegexp(add, where, "matches", *j.Matches)
	}
	if j.Length != nil {
		n++
		if *j.Length < 0 {
			add(diag.NegativeValue, "%s.length must be >= 0 (got %d); no array, object, or string has a negative length", where, *j.Length)
		}
	}
	if j.Gt != nil {
		n++
	}
	if j.Gte != nil {
		n++
	}
	if j.Lt != nil {
		n++
	}
	if j.Lte != nil {
		n++
	}
	requireExactlyOne(add, where, "equals/matches/length/gt/gte/lt/lte", n)
}

// validateStringList rejects an explicitly-empty contains/not_contains list
// (`contains: []`), which would otherwise decode to a present-but-empty matcher
// that trivially passes, and rejects any empty-string element: `contains: ""`
// is an always-true no-op and `not_contains: ""` can never pass (every string
// contains the empty substring), so either is an authoring mistake — caught at
// load time like the empty-list case and like validateChanges' empty entries.
func validateStringList(add addFunc, where, key string, l spec.StringList) {
	if l != nil && len(l) == 0 {
		add(diag.EmptyValue, "%s.%s must not be empty", where, key)
	}
	for i, s := range l {
		if s == "" {
			add(diag.VacuousMatcher, "%s.%s[%d] is an empty string, which matches everything (contains) or nothing (not_contains); remove it or give a real substring", where, key, i)
		}
	}
}

// validateContainsOverlap refuses a contains/not_contains pair of one assert
// that shares an entry: no output can both hold and not hold the same
// substring, so the assertion runs and then fails with a message about the
// program under test ("the substring %q was unexpectedly present"), which sends
// a reviewer to the CLI instead of to the spec. Which half reports depends on
// the observed output, so the same broken spec accuses different things on
// different days.
//
// The comparison is literal apart from CRLF folding, which is what the engine
// does to a needle before matching: comparing raw here would give the loader
// and the engine different answers for a spec authored in a CRLF editor.
// `contains: [abc]` against `not_contains: [abcd]` is satisfiable and keeps
// loading.
func validateContainsOverlap(add addFunc, where string, contains, notContains []string) {
	if len(contains) == 0 || len(notContains) == 0 {
		return
	}
	forbidden := make(map[string]bool, len(notContains))
	for _, s := range notContains {
		forbidden[spec.FoldCRLF(s)] = true
	}
	reported := map[string]bool{}
	for _, s := range contains {
		folded := spec.FoldCRLF(s)
		if forbidden[folded] && !reported[folded] {
			reported[folded] = true
			add(diag.ExclusiveKeys, "%s: contains and not_contains both list %q, so nothing can satisfy the assertion", where, s)
		}
	}
}

// validateRegexp rejects an empty pattern and an uncompilable one for a
// matches/not_matches matcher. An empty regexp matches everything, so `matches:
// ""` is an always-true no-op and `not_matches: ""` can never pass — either is
// an authoring mistake, caught at load time like the empty-string contains/
// not_contains case in validateStringList.
func validateRegexp(add addFunc, where, key, pattern string) {
	if pattern == "" {
		add(diag.VacuousMatcher, "%s.%s must not be an empty regexp, which matches everything (matches) or nothing (not_matches); remove it or give a real pattern", where, key)
		return
	}
	if _, err := regexp.Compile(pattern); err != nil {
		add(diag.BadRegexp, "%s.%s %q is not a valid regexp: %v", where, key, pattern, err)
		return
	}
	// A not_matches pattern that matches the empty string (e.g. "z*", "a?",
	// "(foo)?", "x{0,3}") matches at position 0 of every input, so not_matches
	// can never pass — the same trap as the empty pattern above. Caught at load
	// time rather than failing at runtime with a confusing "unexpectedly matched"
	// hint. An empty-matching pattern under a bare matches is legitimate (it
	// matches everything, which the author may intend); the places where it is
	// not are refused by their own callers through matchesEmpty.
	if key == "not_matches" && matchesEmpty(pattern) {
		add(diag.VacuousMatcher, "%s.%s %q matches the empty string, so not_matches can never pass; anchor it or require at least one character (e.g. \"z+\")", where, key, pattern)
	}
}

// matchesEmpty reports whether a pattern matches the empty string — the
// `*`-where-`+`-was-meant shape ("q*", "[0-9]*", "(foo)?", "x{0,3}"). It is the
// one predicate every caller asks, so the rules cannot drift: a pattern that
// matches nothing at all still matches at position 0 of every input, which
// makes an occurrence count meaningless (one zero-width match per byte plus
// one), shreds a scrub placeholder between every byte, and captures the empty
// string from a stream that plainly contains the value.
//
// A pattern that does not compile answers false; its own diagnostic says so.
func matchesEmpty(pattern string) bool {
	re, err := regexp.Compile(pattern)
	return err == nil && re.MatchString("")
}

// validateDuration checks a duration assert (#31): at least one bound, lt/lte
// and gt/gte mutually exclusive, every bound a valid Go duration, and any
// interval non-empty (lower < upper).
func validateDuration(add addFunc, where string, d *spec.DurationAssert) {
	parse := func(field, val string) (time.Duration, bool) {
		if val == "" {
			return 0, false
		}
		dur, err := time.ParseDuration(val)
		if err != nil {
			add(diag.BadDuration, "%s.%s %q is not a valid duration (e.g. \"2s\", \"100ms\")", where, field, val)
			return 0, false
		}
		// A measured wall-clock duration is never negative, so a negative bound
		// is always an authoring mistake: gt/gte trivially hold and lt/lte can
		// never pass. Caught at load time like the empty-interval case below.
		if dur < 0 {
			add(diag.NegativeValue, "%s.%s must not be negative (got %q); a measured duration is never below zero", where, field, val)
			return 0, false
		}
		return dur, true
	}
	lt, ltOK := parse("lt", d.LT)
	lte, lteOK := parse("lte", d.LTE)
	gt, gtOK := parse("gt", d.GT)
	gte, gteOK := parse("gte", d.GTE)

	if d.LT == "" && d.LTE == "" && d.GT == "" && d.GTE == "" {
		add(diag.ChooseAtLeastOne, "%s: set at least one bound (lt/lte/gt/gte)", where)
		return
	}
	if d.LT != "" && d.LTE != "" {
		add(diag.ExclusiveKeys, "%s: set only one upper bound (lt or lte, not both)", where)
	}
	if d.GT != "" && d.GTE != "" {
		add(diag.ExclusiveKeys, "%s: set only one lower bound (gt or gte, not both)", where)
	}

	// The interval, when both ends are present, must be non-empty. lte/gte
	// endpoints may touch (lte == gte is the single-point interval); strict
	// bounds must leave room.
	upper, upOK := lt, ltOK
	upStrict := true
	if lteOK {
		upper, upOK, upStrict = lte, true, false
	}
	lower, lowOK := gt, gtOK
	lowStrict := true
	if gteOK {
		lower, lowOK, lowStrict = gte, true, false
	}
	if upOK && lowOK {
		if (upStrict || lowStrict) && lower >= upper {
			add(diag.EmptyInterval, "%s: the bounds form an empty interval (lower %s is not below upper %s)", where, lower, upper)
		} else if !upStrict && !lowStrict && lower > upper {
			add(diag.EmptyInterval, "%s: the bounds form an empty interval (lower %s exceeds upper %s)", where, lower, upper)
		}
	}
}

// validateMockAssert checks a `mock:` assertion (#24): a declared server
// name (listed on a miss, mirroring the unknown-runner message), a sane
// count, and well-formed header/body matchers. A nil mockNames (retry.until)
// skips the declared-name check.
func validateMockAssert(add addFunc, where string, m *spec.MockAssert, mockNames map[string]bool) {
	switch {
	case m.Name == "":
		add(diag.RequiredKey, "%s.name is required (the mock server whose recorded requests to check)", where)
	case mockNames != nil && !mockNames[m.Name]:
		declared := "none"
		if len(mockNames) > 0 {
			declared = strings.Join(slices.Sorted(maps.Keys(mockNames)), ", ")
		}
		add(diag.NameNotDeclared, "%s.name %q is not a declared mock server (declared: %s)", where, m.Name, declared)
	}
	if m.Count != nil {
		if *m.Count < 0 {
			add(diag.NegativeValue, "%s.count must be >= 0 (got %d)", where, *m.Count)
		}
		if *m.Count == 0 && (m.Header != nil || m.Body != nil) {
			add(diag.ExclusiveKeys, "%s: count: 0 cannot be combined with header/body matchers (there is no request to match)", where)
		}
	}
	if m.Header != nil {
		validateHeaderMatch(add, where+".header", m.Header)
	}
	if m.Body != nil {
		validateStream(add, where+".body", m.Body)
	}
}
