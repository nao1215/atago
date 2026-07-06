package assert

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/goccy/go-yaml"
	"github.com/nao1215/atago/internal/spec"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
)

// checkJSONChecks applies every JSONPath check in the list to data (all must
// hold, #156), parsing it as JSON — or as YAML when isYAML is set. The payload
// is parsed ONCE and every check runs against the same decoded value. It
// returns the first failing check, or the last passing result when they all
// pass. A single-element list produces byte-identical output to the pre-list
// scalar form so the common case is unchanged.
func checkJSONChecks(desc, name string, data []byte, checks spec.JSONChecks, isYAML bool) *CheckResult {
	parsed, cr := parseDoc(desc, name, data, isYAML)
	if cr != nil {
		return cr
	}
	var last *CheckResult
	for i := range checks {
		d := desc
		if len(checks) > 1 {
			d = fmt.Sprintf("%s[%d]", desc, i)
		}
		r := applyJSONMatch(d, parsed, &checks[i])
		if !r.OK {
			return r
		}
		last = r
	}
	return last
}

// parseDoc decodes data as JSON (or YAML when isYAML is set) into the generic
// value model the JSONPath engine walks, returning a failure CheckResult for
// empty or malformed input. It is the shared parse step behind checkJSON/
// checkYAML and the list form, so the payload is decoded once per assertion.
func parseDoc(desc, name string, data []byte, isYAML bool) (any, *CheckResult) {
	kind := "JSON"
	if isYAML {
		kind = "YAML"
	}
	// oj.Parse treats empty/whitespace input as a valid (nil) document, which
	// would otherwise yield a misleading "selected no value" instead of telling
	// the user the stream was empty. Surfaced by dogfooding `gup list --json`,
	// which prints nothing when no tools are installed.
	if strings.TrimSpace(string(data)) == "" {
		return nil, &CheckResult{
			Desc:     desc,
			Expected: "valid " + kind,
			Actual:   "(empty)",
			Hint:     fmt.Sprintf("%s was empty, so it is not valid %s", name, kind),
		}
	}
	var parsed any
	var err error
	if isYAML {
		err = yaml.Unmarshal(data, &parsed)
	} else {
		parsed, err = oj.Parse(data)
	}
	if err != nil {
		return nil, &CheckResult{
			Desc:     desc,
			Expected: "valid " + kind,
			Actual:   excerpt(string(data)),
			Hint:     fmt.Sprintf("%s was not valid %s: %v", name, kind, err),
		}
	}
	return parsed, nil
}

// checkJSON parses data as JSON, selects nodes with a JSONPath, and applies the
// configured matcher.
func checkJSON(desc, name string, data []byte, j *spec.JSONAssert) *CheckResult {
	parsed, cr := parseDoc(desc, name, data, false)
	if cr != nil {
		return cr
	}
	return applyJSONMatch(desc, parsed, j)
}

// checkYAML parses data as YAML and applies the same JSONPath matcher logic as
// checkJSON, since a YAML document decodes to the same generic value model
// (maps/slices/scalars) that the JSONPath engine walks (issue #9).
func checkYAML(desc, name string, data []byte, j *spec.JSONAssert) *CheckResult {
	parsed, cr := parseDoc(desc, name, data, true)
	if cr != nil {
		return cr
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
	cmp, ok := numericCmp(node, want)
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
		okCmp = cmp > 0
	case "gte":
		okCmp = cmp >= 0
	case "lt":
		okCmp = cmp < 0
	case "lte":
		okCmp = cmp <= 0
	}
	d := fmt.Sprintf("%s %s %v", desc, opSymbol(op), want)
	if okCmp {
		return pass(d)
	}
	return &CheckResult{
		Desc:     d,
		Expected: fmt.Sprintf("%s %s %v", path, opSymbol(op), want),
		Actual:   fmt.Sprintf("%s = %s", path, renderNode(node)),
		Hint:     fmt.Sprintf("value at %s (%s) is not %s %v", path, renderNode(node), op, want),
	}
}

// numericCmp returns the sign of node−want (−1, 0, +1) for a numeric ordering.
// It prefers exact big.Int arithmetic so a JSON integer beyond 2^53 compares
// exactly — the same precision equals uses — instead of collapsing distinct
// values to the same float64; it falls back to float64 for fractional operands.
// ok is false when node is not numeric (nor a numeric string).
func numericCmp(node any, want float64) (int, bool) {
	if ni, ok := toBigInt(node); ok {
		if wf := big.NewFloat(want); wf.IsInt() {
			wi, _ := wf.Int(nil)
			return ni.Cmp(wi), true
		}
	}
	nf, ok := toFloat(node)
	if !ok {
		return 0, false
	}
	switch {
	case nf < want:
		return -1, true
	case nf > want:
		return 1, true
	default:
		return 0, true
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
		Actual:   fmt.Sprintf("%s = %s", path, renderNode(node)),
		Hint:     fmt.Sprintf("value at %s did not equal %v", path, want),
	}
}

// renderNode renders a selected JSON value for a matcher subject or a failure
// message. Floating-point numbers print without scientific notation — a whole
// number like 1000000.0 reads as "1000000", not "1e+06" — so `matches` sees the
// digits a spec author wrote and a failure message stays readable. Everything
// else falls back to the default Go rendering (json.Number already carries its
// exact digits as a string).
func renderNode(v any) string {
	switch t := v.(type) {
	case float64:
		return formatNum(t)
	case float32:
		return formatNum(float64(t))
	default:
		return fmt.Sprintf("%v", v)
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
	got := renderNode(node)
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
		// Prefer exact integer comparison. Two distinct integers beyond 2^53 (a
		// JSON id, or a value beyond int64 that decodes as json.Number) round to
		// the same float64, so a float compare would report them equal. big.Int
		// keeps every digit.
		if ni, ok := toBigInt(node); ok {
			if wi, ok := toBigInt(want); ok {
				return ni.Cmp(wi) == 0
			}
		}
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
	case string:
		// A non-numeric node vs a string is a string match (the string-vs-number
		// case was already handled by the numeric branch above). Comparing by
		// fmt.Sprintf would let a JSON string equal a like-printed non-string.
		w, ok := want.(string)
		return ok && n == w
	case bool:
		// A JSON boolean equals only a boolean: `true` must not equal the string
		// "true", which the old fmt.Sprintf fallback reported as a false pass.
		w, ok := want.(bool)
		return ok && n == w
	case nil:
		return want == nil
	}
	return fmt.Sprintf("%v", node) == fmt.Sprintf("%v", want)
}

// isNumericKind reports whether v is a genuine number (not an arbitrary numeric
// string). It gates valuesEqual's numeric coercion so string-vs-string equality
// stays byte-exact. Unsigned kinds are included because goccy/go-yaml decodes a
// large integer that overflows int64 as uint64, and json.Number because oj
// decodes an integer beyond int64/uint64 as one — both are real numbers from a
// parsed document, so an `equals` against a numeric spec value must compare them
// numerically (via toBigInt/toFloat), not lexically.
func isNumericKind(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, json.Number:
		return true
	default:
		return false
	}
}

func toFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case int:
		return float64(t), true
	case int8:
		return float64(t), true
	case int16:
		return float64(t), true
	case int32:
		return float64(t), true
	case int64:
		return float64(t), true
	case uint:
		return float64(t), true
	case uint8:
		return float64(t), true
	case uint16:
		return float64(t), true
	case uint32:
		return float64(t), true
	case uint64:
		return float64(t), true
	case float32:
		return float64(t), true
	case float64:
		return t, true
	case json.Number:
		// oj decodes a JSON integer beyond int64/uint64 range as json.Number
		// (a string carrying the exact digits). Without this case the numeric
		// matchers (gt/gte/lt/lte) reject a perfectly valid large number as
		// "not numeric".
		if f, err := strconv.ParseFloat(strings.TrimSpace(string(t)), 64); err == nil {
			return f, true
		}
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

// toBigInt returns v as an exact arbitrary-precision integer when it is an
// integer value: a signed/unsigned integer kind, an integer-valued float, a
// json.Number, or a numeric string with no fractional part. It is the exact
// half of the numeric comparison — non-integers (2.5) return false so the
// caller falls back to float comparison. Using big.Int keeps large integers
// (JSON ids beyond 2^53, values beyond uint64) byte-exact where float64 would
// silently collapse distinct values.
func toBigInt(v any) (*big.Int, bool) {
	switch t := v.(type) {
	case int:
		return big.NewInt(int64(t)), true
	case int8:
		return big.NewInt(int64(t)), true
	case int16:
		return big.NewInt(int64(t)), true
	case int32:
		return big.NewInt(int64(t)), true
	case int64:
		return big.NewInt(t), true
	case uint:
		return new(big.Int).SetUint64(uint64(t)), true
	case uint8:
		return new(big.Int).SetUint64(uint64(t)), true
	case uint16:
		return new(big.Int).SetUint64(uint64(t)), true
	case uint32:
		return new(big.Int).SetUint64(uint64(t)), true
	case uint64:
		return new(big.Int).SetUint64(t), true
	case float32:
		return floatToBigInt(float64(t))
	case float64:
		return floatToBigInt(t)
	case json.Number:
		if i, ok := new(big.Int).SetString(strings.TrimSpace(string(t)), 10); ok {
			return i, true
		}
	case string:
		if i, ok := new(big.Int).SetString(strings.TrimSpace(t), 10); ok {
			return i, true
		}
	}
	return nil, false
}

// floatToBigInt returns a finite, integer-valued float as an exact integer.
// A fractional or non-finite value returns false so numeric equality falls back
// to float comparison.
func floatToBigInt(f float64) (*big.Int, bool) {
	if math.IsInf(f, 0) || math.IsNaN(f) || f != math.Trunc(f) {
		return nil, false
	}
	bi, _ := big.NewFloat(f).Int(nil)
	return bi, true
}
