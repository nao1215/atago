package loader

import (
	"maps"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/spec"
)

func validateAssert(add func(string, ...any), where string, a *spec.Assert, mockNames map[string]bool) {
	targets := a.SetTargets()
	if len(targets) == 0 {
		add("%s.assert: must set at least one assertion target (got none)", where)
		return
	}
	// Each set target is an independent check and all must hold, so validate every
	// one of them (an assert may combine, e.g., exit_code + stdout + file).
	for _, t := range targets {
		validateAssertTarget(add, where, a, t, mockNames)
	}
}

// validateAssertTarget checks the shape of a single assertion target family.
func validateAssertTarget(add func(string, ...any), where string, a *spec.Assert, target spec.AssertTarget, mockNames map[string]bool) {
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
		validateStream(add, where+".assert.screen", a.Screen)
	case spec.AssertDuration:
		validateDuration(add, where+".assert.duration", a.Duration)
	case spec.AssertChanges:
		validateChanges(add, where+".assert.changes", a.Changes)
	case spec.AssertPDF:
		validatePDF(add, where+".assert.pdf", a.PDF)
	case spec.AssertGRPCStatus:
		// grpc_status is a bare int; no further shape to validate.
	}
}

// validateExitCode checks the exit_code assertion (#19): exactly one of the
// bare-int / {not} / {in} forms, and an `in` set that is non-empty with unique
// values — a duplicated code is authoring confusion, not a wider contract.
func validateExitCode(add func(string, ...any), where string, e *spec.ExitCode) {
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
		add("%s must be an int, {not: int}, or {in: [int, ...]}", where)
		return
	}
	if set > 1 {
		add("%s: set exactly one of a bare int, not, or in", where)
		return
	}
	if e.In != nil {
		if len(e.In) == 0 {
			add("%s.in must list at least one accepted exit code", where)
		}
		seen := make(map[int]bool, len(e.In))
		for _, n := range e.In {
			if seen[n] {
				add("%s.in lists %d more than once", where, n)
			}
			seen[n] = true
		}
	}
}

func validateStream(add func(string, ...any), where string, s *spec.StreamAssert) {
	matchers := s.SetMatchers()
	switch len(matchers) {
	case 0:
		add("%s: must set exactly one matcher (empty/contains/not_contains/matches/not_matches/equals/not_equals/json/yaml/snapshot)", where)
	case 1:
		if s.JSON != nil {
			validateJSON(add, where+".json", s.JSON)
		}
		if s.YAML != nil {
			validateJSON(add, where+".yaml", s.YAML)
		}
		if s.Matches != nil {
			validateRegexp(add, where, "matches", *s.Matches)
		}
		if s.NotMatches != nil {
			validateRegexp(add, where, "not_matches", *s.NotMatches)
		}
	default:
		add("%s: must set exactly one matcher, but set %v", where, matchers)
	}
	validateStringList(add, where, "contains", s.Contains)
	validateStringList(add, where, "not_contains", s.NotContains)

	if s.Line != nil {
		if *s.Line < 1 {
			add("%s.line must be >= 1 (got %d)", where, *s.Line)
		}
		// line narrows the stream to one line, so it only composes with the
		// text matchers — json/snapshot operate on the whole document.
		if s.JSON != nil || s.YAML != nil || s.Snapshot != "" {
			add("%s.line cannot be combined with json/yaml/snapshot (use contains/matches/equals/empty)", where)
		}
	}
}

func validateFile(add func(string, ...any), where string, f *spec.FileAssert) {
	if f.Path == "" {
		add("%s.path is required", where)
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
	if f.JSON != nil {
		n++
		validateJSON(add, where+".json", f.JSON)
	}
	if f.Snapshot != "" {
		n++
	}
	if n == 0 {
		add("%s: must set one of exists/contains/not_contains/executable/json/snapshot", where)
	} else if n > 1 {
		add("%s: must set exactly one of exists/contains/not_contains/executable/json/snapshot", where)
	}
}

func validateHeaderMatch(add func(string, ...any), where string, h *spec.HeaderMatch) {
	if h.Name == "" {
		add("%s.name is required", where)
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
	switch n {
	case 0:
		add("%s: must set one of contains/equals/matches", where)
	case 1:
	default:
		add("%s: must set exactly one of contains/equals/matches", where)
	}
}

func validateJSON(add func(string, ...any), where string, j *spec.JSONAssert) {
	if j.Path == "" {
		add("%s.path is required", where)
	}
	n := 0
	if j.Equals != nil {
		n++
	}
	if j.Matches != nil {
		n++
		validateRegexp(add, where, "matches", *j.Matches)
	}
	if j.Length != nil {
		n++
		if *j.Length < 0 {
			add("%s.length must be >= 0 (got %d); no array, object, or string has a negative length", where, *j.Length)
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
	if n == 0 {
		add("%s: must set one of equals/matches/length/gt/gte/lt/lte", where)
	} else if n > 1 {
		add("%s: must set exactly one of equals/matches/length/gt/gte/lt/lte", where)
	}
}

// validateStringList rejects an explicitly-empty contains/not_contains list
// (`contains: []`), which would otherwise decode to a present-but-empty matcher
// that trivially passes, and rejects any empty-string element: `contains: ""`
// is an always-true no-op and `not_contains: ""` can never pass (every string
// contains the empty substring), so either is an authoring mistake — caught at
// load time like the empty-list case and like validateChanges' empty entries.
func validateStringList(add func(string, ...any), where, key string, l spec.StringList) {
	if l != nil && len(l) == 0 {
		add("%s.%s must not be empty", where, key)
	}
	for i, s := range l {
		if s == "" {
			add("%s.%s[%d] is an empty string, which matches everything (contains) or nothing (not_contains); remove it or give a real substring", where, key, i)
		}
	}
}

// validateRegexp rejects an empty pattern and an uncompilable one for a
// matches/not_matches matcher. An empty regexp matches everything, so `matches:
// ""` is an always-true no-op and `not_matches: ""` can never pass — either is
// an authoring mistake, caught at load time like the empty-string contains/
// not_contains case in validateStringList.
func validateRegexp(add func(string, ...any), where, key, pattern string) {
	if pattern == "" {
		add("%s.%s must not be an empty regexp, which matches everything (matches) or nothing (not_matches); remove it or give a real pattern", where, key)
		return
	}
	if _, err := regexp.Compile(pattern); err != nil {
		add("%s.%s %q is not a valid regexp: %v", where, key, pattern, err)
	}
}

// validateDuration checks a duration assert (#31): at least one bound, lt/lte
// and gt/gte mutually exclusive, every bound a valid Go duration, and any
// interval non-empty (lower < upper).
func validateDuration(add func(string, ...any), where string, d *spec.DurationAssert) {
	parse := func(field, val string) (time.Duration, bool) {
		if val == "" {
			return 0, false
		}
		dur, err := time.ParseDuration(val)
		if err != nil {
			add("%s.%s %q is not a valid duration (e.g. \"2s\", \"100ms\")", where, field, val)
			return 0, false
		}
		// A measured wall-clock duration is never negative, so a negative bound
		// is always an authoring mistake: gt/gte trivially hold and lt/lte can
		// never pass. Caught at load time like the empty-interval case below.
		if dur < 0 {
			add("%s.%s must not be negative (got %q); a measured duration is never below zero", where, field, val)
			return 0, false
		}
		return dur, true
	}
	lt, ltOK := parse("lt", d.LT)
	lte, lteOK := parse("lte", d.LTE)
	gt, gtOK := parse("gt", d.GT)
	gte, gteOK := parse("gte", d.GTE)

	if d.LT == "" && d.LTE == "" && d.GT == "" && d.GTE == "" {
		add("%s: set at least one bound (lt/lte/gt/gte)", where)
		return
	}
	if d.LT != "" && d.LTE != "" {
		add("%s: set only one upper bound (lt or lte, not both)", where)
	}
	if d.GT != "" && d.GTE != "" {
		add("%s: set only one lower bound (gt or gte, not both)", where)
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
			add("%s: the bounds form an empty interval (lower %s is not below upper %s)", where, lower, upper)
		} else if !upStrict && !lowStrict && lower > upper {
			add("%s: the bounds form an empty interval (lower %s exceeds upper %s)", where, lower, upper)
		}
	}
}

// validateMockAssert checks a `mock:` assertion (#24): a declared server
// name (listed on a miss, mirroring the unknown-runner message), a sane
// count, and well-formed header/body matchers. A nil mockNames (retry.until)
// skips the declared-name check.
func validateMockAssert(add func(string, ...any), where string, m *spec.MockAssert, mockNames map[string]bool) {
	switch {
	case m.Name == "":
		add("%s.name is required (the mock server whose recorded requests to check)", where)
	case mockNames != nil && !mockNames[m.Name]:
		declared := "none"
		if len(mockNames) > 0 {
			declared = strings.Join(slices.Sorted(maps.Keys(mockNames)), ", ")
		}
		add("%s.name %q is not a declared mock server (declared: %s)", where, m.Name, declared)
	}
	if m.Count != nil {
		if *m.Count < 0 {
			add("%s.count must be >= 0 (got %d)", where, *m.Count)
		}
		if *m.Count == 0 && (m.Header != nil || m.Body != nil) {
			add("%s: count: 0 cannot be combined with header/body matchers (there is no request to match)", where)
		}
	}
	if m.Header != nil {
		validateHeaderMatch(add, where+".header", m.Header)
	}
	if m.Body != nil {
		validateStream(add, where+".body", m.Body)
	}
}
