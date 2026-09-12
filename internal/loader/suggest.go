package loader

import (
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/nao1215/atago/internal/spec"
)

// unknownFieldRe matches the strict-decode error goccy/go-yaml emits for a
// field the spec model does not define.
var unknownFieldRe = regexp.MustCompile(`unknown field "([^"]+)"`)

// suggestUnknownField appends a "did you mean" hint to an unknown-field parse
// error when the misspelled name is close to a real spec field. A typo like
// `asserts:` or `stdut:` is the most common first-five-minutes mistake, and
// the raw error alone makes the user diff the docs by eye.
func suggestUnknownField(msg string) string {
	m := unknownFieldRe.FindStringSubmatch(msg)
	if m == nil {
		return msg
	}
	// The name is a real spec field, just not valid here: almost always an
	// indentation slip (e.g. `command:` dedented out of its `run:` block).
	if isKnownField(m[1]) {
		return fmt.Sprintf("%s\nhint: %q is a spec field, but not at this position — check the indentation and nesting", msg, m[1])
	}
	best, ok := closestField(m[1])
	if !ok {
		return msg
	}
	return fmt.Sprintf("%s\nhint: did you mean %q?", msg, best)
}

// scalarWhereMappingRe matches the strict-decode error goccy emits when a
// scalar (or sequence) sits where the model wants a mapping.
var scalarWhereMappingRe = regexp.MustCompile(`(?:string|int|uint|float|bool|sequence) was used where mapping is expected`)

// streamTargetRe finds a stream-assertion target on the snippet line the error
// marker (`> N | ...`) points at, so surrounding context lines cannot trigger a
// misleading hint.
var streamTargetRe = regexp.MustCompile(`(?m)^>\s*\d+ \|.*?\b(stdout|stderr|body|rows|message|value)\s*:`)

// suggestScalarMatcher appends a hint when a stream assertion was written as a
// bare scalar — `stdout: hello` instead of `stdout: {contains: hello}` — which
// is the single most common first-spec mistake. The generic decoder message
// ("string was used where mapping is expected") is positionally accurate but
// gives no clue what shape is wanted.
func suggestScalarMatcher(msg string) string {
	if !scalarWhereMappingRe.MatchString(msg) {
		return msg
	}
	m := streamTargetRe.FindStringSubmatch(msg)
	if m == nil {
		return msg
	}
	t := m[1]
	return fmt.Sprintf("%s\nhint: %s must set a matcher mapping, e.g. %s: {contains: \"...\"} or %s: {equals: \"...\"}", msg, t, t, t)
}

// collectionIntoScalarRe matches goccy's decode error for a list or a mapping
// written where the model holds a single value. The message names the Go type
// and the OUTERMOST struct field ("Spec.Scenarios"), never the key the author
// wrote, so on its own it points at the wrong place.
var collectionIntoScalarRe = regexp.MustCompile(`cannot unmarshal (\[\]interface \{\}|map\[string\]interface \{\}) into Go struct field \S+ of type (\S+)`)

// gutterRe strips the excerpt's line gutter ("  10 | " / "> 11 | ") so the YAML
// on that line can be read.
var gutterRe = regexp.MustCompile(`^(>?\s*\d+ \| )`)

// blockKeyRe matches a mapping key that opens a block (no value on its line),
// which is the key a following sequence or mapping belongs to.
var blockKeyRe = regexp.MustCompile(`^(\s*)([A-Za-z_][A-Za-z0-9_-]*):\s*$`)

// flowKeyRe matches a mapping key whose value opens a flow collection on the
// same line (`matches: ["a", "b"]`).
var flowKeyRe = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_-]*):\s*[\[{]`)

// suggestCollectionShape appends a hint when a list or a mapping was written
// where one value belongs — `matches: [a, b]` for a key that takes one pattern,
// or a nested mapping under a key that takes a string. The decoder's own message
// names a Go type and the outermost struct field ("Spec.Scenarios of type
// string"), so it tells the author neither which key is wrong nor what shape it
// wants; the excerpt already points at the right line, and this reads the key
// off it.
func suggestCollectionShape(msg string) string {
	m := collectionIntoScalarRe.FindStringSubmatch(msg)
	if m == nil {
		return msg
	}
	key, ok := keyAtMarker(msg)
	if !ok {
		return msg
	}
	written := "a list"
	if strings.HasPrefix(m[1], "map[") {
		written = "a mapping"
	}
	hint := fmt.Sprintf("%s\nhint: %q takes %s, not %s", msg, key, singleValuePhrase(m[2]), written)
	// A list of patterns is the common shape of this mistake, and the substring
	// matchers next to it are the ones that do take a list — so say which is
	// which rather than leaving the author to guess that one key differs.
	if written == "a list" && (key == "matches" || key == "not_matches") {
		hint += "; write one assert per pattern (the substring matchers \"contains\" and \"not_contains\" are the ones that take a list)"
	}
	return hint
}

// singleValuePhrase renders the Go type the decoder wanted as something a spec
// author recognizes.
func singleValuePhrase(goType string) string {
	switch goType {
	case "string":
		return "a single value"
	case "int", "int64", "uint", "uint64", "float64":
		return "a single number"
	case "bool":
		return "true or false"
	default:
		return "a single value"
	}
}

// keyAtMarker reads the spec key the error's excerpt points at: the key on the
// marked line when the collection was written in flow style, and otherwise the
// nearest block key above it that opens the collection.
func keyAtMarker(msg string) (string, bool) {
	lines := strings.Split(msg, "\n")
	marked := -1
	for i, l := range lines {
		if strings.HasPrefix(l, ">") && gutterRe.MatchString(l) {
			marked = i
			break
		}
	}
	if marked < 0 {
		return "", false
	}
	content := gutterRe.ReplaceAllString(lines[marked], "")
	if m := flowKeyRe.FindAllStringSubmatch(content, -1); m != nil {
		// The innermost flow key on the line is the one whose value the caret
		// sits in: `- assert: {stdout: {matches: [...]}}`.
		return m[len(m)-1][1], true
	}
	markedIndent := len(content) - len(strings.TrimLeft(content, " "))
	for i := marked - 1; i >= 0; i-- {
		if !gutterRe.MatchString(lines[i]) {
			continue
		}
		above := gutterRe.ReplaceAllString(lines[i], "")
		km := blockKeyRe.FindStringSubmatch(above)
		if km == nil {
			continue
		}
		if len(km[1]) < markedIndent {
			return km[2], true
		}
	}
	return "", false
}

// isKnownField reports whether name is a field somewhere in the spec model.
func isKnownField(name string) bool {
	return slices.Contains(fieldVocabulary(), strings.ToLower(name))
}

// closestField returns the spec field name nearest to the typo, if any is
// within an edit distance small enough to be a plausible slip.
func closestField(typo string) (string, bool) {
	lower := strings.ToLower(typo)
	best, bestDist := "", 3 // allow up to 2 edits; anything further is guessing
	for _, name := range fieldVocabulary() {
		if name == lower {
			continue // same name: the field exists elsewhere; a distance-0 hint would mislead
		}
		if d := editDistance(lower, name); d < bestDist {
			best, bestDist = name, d
		}
	}
	return best, best != ""
}

var (
	vocabOnce sync.Once
	vocab     []string
)

// fieldVocabulary collects every yaml field name reachable from spec.Spec by
// walking the struct tags reflectively, so the hint list can never drift from
// the model.
func fieldVocabulary() []string {
	vocabOnce.Do(func() {
		seen := map[string]bool{}
		visited := map[reflect.Type]bool{}
		var walk func(t reflect.Type)
		walk = func(t reflect.Type) {
			for t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice || t.Kind() == reflect.Array || t.Kind() == reflect.Map {
				t = t.Elem()
			}
			if t.Kind() != reflect.Struct || visited[t] {
				return
			}
			visited[t] = true
			for f := range t.Fields() {
				tag, _, _ := strings.Cut(f.Tag.Get("yaml"), ",")
				if tag != "" && tag != "-" {
					seen[tag] = true
				}
				walk(f.Type)
			}
		}
		walk(reflect.TypeFor[spec.Spec]())
		for name := range seen {
			vocab = append(vocab, name)
		}
	})
	return vocab
}

// editDistance is the Levenshtein distance between two short field names.
func editDistance(a, b string) int {
	if a == b {
		return 0
	}
	la, lb := len(a), len(b)
	prev := make([]int, lb+1)
	cur := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		cur[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[lb]
}
