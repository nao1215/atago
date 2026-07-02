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
