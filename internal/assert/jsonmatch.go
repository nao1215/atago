package assert

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/goccy/go-yaml"
	"github.com/nao1215/atago/internal/spec"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
)

// checkJSON parses data as JSON, selects nodes with a JSONPath, and applies the
// configured matcher.
func checkJSON(desc, name string, data []byte, j *spec.JSONAssert) *CheckResult {
	// oj.Parse treats empty/whitespace input as a valid (nil) document, which
	// would otherwise yield a misleading "selected no value" instead of telling
	// the user the stream was empty. Surfaced by dogfooding `gup list --json`,
	// which prints nothing when no tools are installed.
	if strings.TrimSpace(string(data)) == "" {
		return &CheckResult{
			Desc:     desc,
			Expected: "valid JSON",
			Actual:   "(empty)",
			Hint:     fmt.Sprintf("%s was empty, so it is not valid JSON", name),
		}
	}
	parsed, err := oj.Parse(data)
	if err != nil {
		return &CheckResult{
			Desc:     desc,
			Expected: "valid JSON",
			Actual:   excerpt(string(data)),
			Hint:     fmt.Sprintf("%s was not valid JSON: %v", name, err),
		}
	}
	return applyJSONMatch(desc, parsed, j)
}

// checkYAML parses data as YAML and applies the same JSONPath matcher logic as
// checkJSON, since a YAML document decodes to the same generic value model
// (maps/slices/scalars) that the JSONPath engine walks (issue #9).
func checkYAML(desc, name string, data []byte, j *spec.JSONAssert) *CheckResult {
	if strings.TrimSpace(string(data)) == "" {
		return &CheckResult{
			Desc:     desc,
			Expected: "valid YAML",
			Actual:   "(empty)",
			Hint:     fmt.Sprintf("%s was empty, so it is not valid YAML", name),
		}
	}
	var parsed any
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return &CheckResult{
			Desc:     desc,
			Expected: "valid YAML",
			Actual:   excerpt(string(data)),
			Hint:     fmt.Sprintf("%s was not valid YAML: %v", name, err),
		}
	}
	return applyJSONMatch(desc, parsed, j)
}

// applyJSONMatch selects nodes from an already-decoded generic value with a
// JSONPath and applies the configured matcher. Shared by checkJSON/checkYAML.
func applyJSONMatch(desc string, parsed any, j *spec.JSONAssert) *CheckResult {
	expr, err := jp.ParseString(j.Path)
	if err != nil {
		return &CheckResult{Desc: desc, Hint: fmt.Sprintf("invalid JSON path %q: %v", j.Path, err)}
	}
	nodes := expr.Get(parsed)

	switch {
	case j.Length != nil:
		return jsonLength(desc, j.Path, nodes, *j.Length)
	case j.Matches != nil:
		return jsonMatches(desc, j.Path, nodes, *j.Matches)
	case j.Gt != nil:
		return jsonCompare(desc, j.Path, nodes, "gt", *j.Gt)
	case j.Gte != nil:
		return jsonCompare(desc, j.Path, nodes, "gte", *j.Gte)
	case j.Lt != nil:
		return jsonCompare(desc, j.Path, nodes, "lt", *j.Lt)
	case j.Lte != nil:
		return jsonCompare(desc, j.Path, nodes, "lte", *j.Lte)
	case j.Equals != nil:
		return jsonEquals(desc, j.Path, nodes, j.Equals)
	default:
		return &CheckResult{Desc: desc, Hint: "json matcher must set equals/matches/length/gt/gte/lt/lte"}
	}
}

// jsonCompare asserts a numeric ordering (gt/gte/lt/lte) on the single selected
// node. The node must be a number or a numeric string; anything else is a clear
// failure rather than a silent pass.
func jsonCompare(desc, path string, nodes []any, op string, want float64) *CheckResult {
	node, cr := single(desc, path, nodes)
	if cr != nil {
		return cr
	}
	got, ok := toFloat(node)
	if !ok {
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("%s is a number %s %v", path, opSymbol(op), want),
			Actual:   fmt.Sprintf("%s = %v", path, node),
			Hint:     fmt.Sprintf("value at %s is not numeric, so it cannot be compared with %s", path, op),
		}
	}
	var okCmp bool
	switch op {
	case "gt":
		okCmp = got > want
	case "gte":
		okCmp = got >= want
	case "lt":
		okCmp = got < want
	case "lte":
		okCmp = got <= want
	}
	d := fmt.Sprintf("%s %s %v", desc, opSymbol(op), want)
	if okCmp {
		return pass(d)
	}
	return &CheckResult{
		Desc:     d,
		Expected: fmt.Sprintf("%s %s %v", path, opSymbol(op), want),
		Actual:   fmt.Sprintf("%s = %v", path, formatNum(got)),
		Hint:     fmt.Sprintf("value at %s (%v) is not %s %v", path, formatNum(got), op, want),
	}
}

func opSymbol(op string) string {
	switch op {
	case "gt":
		return ">"
	case "gte":
		return ">="
	case "lt":
		return "<"
	case "lte":
		return "<="
	}
	return op
}

// formatNum prints an integral float without a trailing ".0" so failure messages
// read naturally (e.g. "3" not "3.000000").
func formatNum(f float64) string {
	if f == float64(int64(f)) {
		return fmt.Sprintf("%d", int64(f))
	}
	return fmt.Sprintf("%g", f)
}

func jsonEquals(desc, path string, nodes []any, want any) *CheckResult {
	node, cr := single(desc, path, nodes)
	if cr != nil {
		return cr
	}
	d := fmt.Sprintf("%s == %v", desc, want)
	if valuesEqual(node, want) {
		return pass(d)
	}
	return &CheckResult{
		Desc:     d,
		Expected: fmt.Sprintf("%s == %v", path, want),
		Actual:   fmt.Sprintf("%s = %v", path, node),
		Hint:     fmt.Sprintf("value at %s did not equal %v", path, want),
	}
}

func jsonMatches(desc, path string, nodes []any, pattern string) *CheckResult {
	node, cr := single(desc, path, nodes)
	if cr != nil {
		return cr
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return &CheckResult{Desc: desc, Hint: fmt.Sprintf("invalid regexp %q: %v", pattern, err)}
	}
	got := fmt.Sprintf("%v", node)
	if re.MatchString(got) {
		return pass(desc)
	}
	return &CheckResult{
		Desc:     desc,
		Expected: fmt.Sprintf("%s matches /%s/", path, pattern),
		Actual:   fmt.Sprintf("%s = %q", path, got),
		Hint:     fmt.Sprintf("value at %s did not match /%s/", path, pattern),
	}
}

func jsonLength(desc, path string, nodes []any, want int) *CheckResult {
	node, cr := single(desc, path, nodes)
	if cr != nil {
		return cr
	}
	got, ok := lengthOf(node)
	if !ok {
		return &CheckResult{
			Desc: desc,
			Hint: fmt.Sprintf("value at %s is not an array, object, or string and has no length", path),
		}
	}
	if got == want {
		return pass(fmt.Sprintf("%s length == %d", desc, want))
	}
	return &CheckResult{
		Desc:     desc,
		Expected: fmt.Sprintf("%s length == %d", path, want),
		Actual:   fmt.Sprintf("%s length == %d", path, got),
		Hint:     fmt.Sprintf("length at %s was %d, expected %d", path, got, want),
	}
}

// single requires exactly one matched node and returns a failure CheckResult
// otherwise.
func single(desc, path string, nodes []any) (any, *CheckResult) {
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("a value at %s", path),
			Actual:   "no match",
			Hint:     fmt.Sprintf("JSON path %s selected no value", path),
		}
	default:
		return nil, &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("a single value at %s", path),
			Actual:   fmt.Sprintf("%d matches", len(nodes)),
			Hint:     fmt.Sprintf("JSON path %s selected %d values; use a more specific path", path, len(nodes)),
		}
	}
}

func lengthOf(v any) (int, bool) {
	switch t := v.(type) {
	case []any:
		return len(t), true
	case map[string]any:
		return len(t), true
	case string:
		// Count characters, not bytes: a spec author asking for the length of a
		// string means "how many characters", so a multi-byte value like "café"
		// is length 4, not 5. Array/object length is element count either way.
		return utf8.RuneCountInString(t), true
	default:
		return 0, false
	}
}

// valuesEqual compares a selected JSON node with a YAML-decoded expected value.
// It normalizes numeric types (int/float and numeric/string-number) and, for
// objects and arrays, recurses so nested structures compare by value rather
// than by fmt.Sprintf output. Map comparison is key-order independent (#40).
func valuesEqual(node, want any) bool {
	// Numeric normalization (int 2 == 2.0, and a numeric string vs a number) is
	// intended, but only when at least ONE side is a genuine number: two DIFFERENT
	// strings that merely parse to the same float ("2" vs "2.0", "1e3" vs "1000")
	// must not be reported equal by an exact `equals`. Gating on a real numeric
	// operand keeps number/numeric-string equality while making string-vs-string
	// byte-exact.
	if isNumericKind(node) || isNumericKind(want) {
		if nf, ok := toFloat(node); ok {
			if wf, ok := toFloat(want); ok {
				return nf == wf
			}
		}
	}
	switch n := node.(type) {
	case map[string]any:
		w, ok := want.(map[string]any)
		if !ok || len(n) != len(w) {
			return false
		}
		for k, nv := range n {
			wv, present := w[k]
			if !present || !valuesEqual(nv, wv) {
				return false
			}
		}
		return true
	case []any:
		w, ok := want.([]any)
		if !ok || len(n) != len(w) {
			return false
		}
		for i := range n {
			if !valuesEqual(n[i], w[i]) {
				return false
			}
		}
		return true
	}
	return fmt.Sprintf("%v", node) == fmt.Sprintf("%v", want)
}

// isNumericKind reports whether v is a genuine numeric type (not a numeric
// string). It gates valuesEqual's numeric coercion so string-vs-string equality
// stays byte-exact.
func isNumericKind(v any) bool {
	switch v.(type) {
	case int, int64, float64, float32:
		return true
	default:
		return false
	}
}

func toFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case string:
		// strconv.ParseFloat requires the WHOLE (trimmed) string to be a valid
		// float. fmt.Sscanf("%g") used to be used here, but it stops at the first
		// non-numeric byte and reports success on the prefix, so "1.2.3" parsed as
		// 1.2 and "3abc" as 3 — making version strings compare equal and string
		// fields silently pass numeric matchers. Requiring full consumption fixes
		// both.
		if f, err := strconv.ParseFloat(strings.TrimSpace(t), 64); err == nil {
			return f, true
		}
	}
	return 0, false
}
