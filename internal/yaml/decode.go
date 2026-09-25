package yaml

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// Unmarshaler is a type that reads itself from a node, for a value that may be
// written in more than one shape.
type Unmarshaler interface {
	UnmarshalYAML(n *Node) error
}

// maxDecodedNodes bounds how many nodes one decode visits. An alias hands out a
// node that is already in the document, so a few lines of anchors can stand for
// more nodes than any memory holds; the bound turns that into an error.
const maxDecodedNodes = 10_000_000

// ErrTooLarge is the error of a document whose aliases expand it past what one
// decode reads.
var ErrTooLarge = errors.New("the document expands to more values than can be read; check its aliases")

// Unmarshal parses data and decodes its first document into the value v points
// to. A key with no field in a struct is left out.
func Unmarshal(data []byte, v any) error {
	f, err := Parse(data)
	if err != nil {
		return err
	}
	return Decode(f.First(), v, false)
}

// Decode decodes a node into the value v points to. With strict set, a mapping
// key that names no field of a struct is an error. A nil node, which is an
// empty document, leaves v as it is.
func Decode(n *Node, v any, strict bool) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("yaml: cannot decode into %T; a non-nil pointer is needed", v)
	}
	if n == nil {
		return nil
	}
	d := &decoder{strict: strict}
	return d.decode(n, rv.Elem(), "")
}

// Value returns the plain Go value a node holds: nil, bool, int64, uint64,
// float64 or string for a scalar, []any for a sequence and map[string]any for a
// mapping.
func Value(n *Node) (any, error) {
	d := &decoder{}
	return d.generic(n, "")
}

type decoder struct {
	strict bool
	visits int
}

var unmarshalerType = reflect.TypeFor[Unmarshaler]()

func (d *decoder) fail(n *Node, path, format string, args ...any) *Error {
	return &Error{Line: n.Line, Col: n.Col, Path: path, Msg: fmt.Sprintf(format, args...)}
}

// shape reports a node of the wrong kind for the value it is decoded into.
func (d *decoder) shape(n *Node, path, want string) *Error {
	got := describe(n)
	e := d.fail(n, path, "%s: expected %s, not %s", subject(path), want, got)
	e.Want, e.Got = want, n.Kind
	return e
}

// subject names the value an error is about.
func subject(path string) string {
	if path == "" {
		return "the document"
	}
	return path
}

// describe says what a node is in an error: its kind, and for a scalar what
// the plain spelling reads as.
func describe(n *Node) string {
	switch n.Kind {
	case SequenceNode:
		return "a list"
	case MappingNode:
		return "a mapping"
	case ScalarNode:
	}
	if n.Style != PlainStyle || n.Tag == "!!str" {
		return "a string"
	}
	v, _ := plainValue(n.Value)
	switch v.(type) {
	case bool:
		return "a bool"
	case int64, uint64, float64:
		return "a number"
	}
	return "a string"
}

func (d *decoder) visit(n *Node) error {
	d.visits++
	if d.visits > maxDecodedNodes {
		return &Error{Line: n.Line, Col: n.Col, Msg: ErrTooLarge.Error()}
	}
	return nil
}

// decode reads n into rv, which is settable. path names rv in the document,
// for the errors.
func (d *decoder) decode(n *Node, rv reflect.Value, path string) error {
	if n == nil {
		// A key with nothing after it holds null.
		n = nullNode
	}
	if err := d.visit(n); err != nil {
		return err
	}
	if rv.CanAddr() && rv.Addr().Type().Implements(unmarshalerType) {
		if n.IsNull() {
			return nil
		}
		u, _ := rv.Addr().Interface().(Unmarshaler)
		if err := u.UnmarshalYAML(n); err != nil {
			var e *Error
			if errors.As(err, &e) {
				return err
			}
			return &Error{Line: n.Line, Col: n.Col, Path: path, Msg: err.Error()}
		}
		return nil
	}
	switch rv.Kind() {
	case reflect.Pointer:
		if n.IsNull() {
			rv.Set(reflect.Zero(rv.Type()))
			return nil
		}
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return d.decode(n, rv.Elem(), path)
	case reflect.Interface:
		if rv.NumMethod() > 0 {
			return d.fail(n, path, "cannot decode into %s", rv.Type())
		}
		v, err := d.generic(n, path)
		if err != nil {
			return err
		}
		if v == nil {
			rv.Set(reflect.Zero(rv.Type()))
			return nil
		}
		rv.Set(reflect.ValueOf(v))
		return nil
	case reflect.String:
		return d.text(n, rv, path)
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return d.scalar(n, rv, path)
	case reflect.Slice:
		return d.slice(n, rv, path)
	case reflect.Map:
		return d.mapping(n, rv, path)
	case reflect.Struct:
		return d.structure(n, rv, path)
	default:
		return d.fail(n, path, "cannot decode into %s", rv.Type())
	}
}

// text reads a scalar into a string. A plain scalar is the text as written, so
// `1.20` is the four characters and not the number they spell; null is the
// empty string; a !!binary scalar is the bytes its base64 spells.
func (d *decoder) text(n *Node, rv reflect.Value, path string) error {
	if n.Kind != ScalarNode {
		return d.shape(n, path, "a string")
	}
	if n.IsNull() {
		rv.SetString("")
		return nil
	}
	if n.Tag == "!!binary" {
		b, err := decodeBinary(n.Value)
		if err != nil {
			return d.fail(n, path, "the !!binary value is not base64: %v", err)
		}
		rv.SetString(string(b))
		return nil
	}
	rv.SetString(n.Value)
	return nil
}

// decodeBinary reads the base64 of a !!binary scalar, whose text may be broken
// over lines.
func decodeBinary(s string) ([]byte, error) {
	s = strings.Map(func(r rune) rune {
		if r == ' ' || r == '\n' || r == '\t' || r == '\r' {
			return -1
		}
		return r
	}, s)
	return base64.StdEncoding.DecodeString(s)
}

// scalar reads n into a bool or a number. Null is the zero value.
func (d *decoder) scalar(n *Node, rv reflect.Value, path string) error {
	if n.IsNull() {
		rv.Set(reflect.Zero(rv.Type()))
		return nil
	}
	if n.Kind != ScalarNode {
		return d.shape(n, path, scalarName(rv.Kind()))
	}
	switch k := rv.Kind(); {
	case k == reflect.Bool:
		b, ok := parseBool(n.Value)
		if !ok {
			return d.fail(n, path, "%s: expected true or false, not %q", subject(path), n.Value)
		}
		rv.SetBool(b)
	case k == reflect.Float32 || k == reflect.Float64:
		f, ok := parseFloat(n.Value)
		if !ok {
			if i, iok := parseInt(n.Value); iok {
				f, ok = float64(i), true
			}
		}
		if !ok {
			return d.fail(n, path, "%s: expected a number, not %q", subject(path), n.Value)
		}
		rv.SetFloat(f)
	case k >= reflect.Uint && k <= reflect.Uint64:
		i, ok := parseInt(n.Value)
		if !ok || i < 0 {
			return d.fail(n, path, "%s: expected a whole number of 0 or more, not %q", subject(path), n.Value)
		}
		if rv.OverflowUint(uint64(i)) {
			return d.fail(n, path, "%s: %s is too large", subject(path), n.Value)
		}
		rv.SetUint(uint64(i))
	default:
		i, ok := parseInt(n.Value)
		if !ok {
			return d.fail(n, path, "%s: expected a whole number, not %q", subject(path), n.Value)
		}
		if rv.OverflowInt(i) {
			return d.fail(n, path, "%s: %s is too large", subject(path), n.Value)
		}
		rv.SetInt(i)
	}
	return nil
}

// scalarName says what a scalar kind is called in an error.
func scalarName(k reflect.Kind) string {
	switch k {
	case reflect.String:
		return "a string"
	case reflect.Bool:
		return "true or false"
	default:
		return "a number"
	}
}

// slice reads a sequence into a slice.
func (d *decoder) slice(n *Node, rv reflect.Value, path string) error {
	if n.IsNull() {
		rv.Set(reflect.Zero(rv.Type()))
		return nil
	}
	if n.Kind != SequenceNode {
		return d.shape(n, path, "a list")
	}
	out := reflect.MakeSlice(rv.Type(), len(n.Items), len(n.Items))
	for i, item := range n.Items {
		if err := d.decode(item, out.Index(i), indexPath(path, i)); err != nil {
			return err
		}
	}
	rv.Set(out)
	return nil
}

// mapping reads a mapping into a map keyed by strings.
func (d *decoder) mapping(n *Node, rv reflect.Value, path string) error {
	if n.IsNull() {
		rv.Set(reflect.Zero(rv.Type()))
		return nil
	}
	if n.Kind != MappingNode {
		return d.shape(n, path, "a mapping")
	}
	if rv.Type().Key().Kind() != reflect.String {
		return d.fail(n, path, "cannot decode into a map keyed by %s", rv.Type().Key())
	}
	out := reflect.MakeMapWithSize(rv.Type(), len(n.Pairs))
	for _, pair := range n.Pairs {
		if pair.Key.Kind != ScalarNode {
			return d.fail(pair.Key, path, "a mapping key has to be a scalar")
		}
		elem := reflect.New(rv.Type().Elem()).Elem()
		if err := d.decode(pair.Value, elem, joinPath(path, pair.Key.Value)); err != nil {
			return err
		}
		out.SetMapIndex(reflect.ValueOf(pair.Key.Value).Convert(rv.Type().Key()), elem)
	}
	rv.Set(out)
	return nil
}

// structure reads a mapping into a struct by its yaml tags, entry by entry in
// the order written, so the first problem in the document is the one reported.
// A key with no field is an error when strict, and left out otherwise.
func (d *decoder) structure(n *Node, rv reflect.Value, path string) error {
	if n.IsNull() {
		return nil
	}
	if n.Kind != MappingNode {
		return d.shape(n, path, "a mapping")
	}
	fields := structFields(rv.Type())
	for _, pair := range n.Pairs {
		f, ok := fields[pair.Key.Value]
		if !ok {
			if d.strict {
				return &Error{
					Line: pair.Key.Line, Col: pair.Key.Col, Path: path, Key: pair.Key.Value,
					Msg: fmt.Sprintf("unknown field %q", pair.Key.Value),
				}
			}
			continue
		}
		if err := d.decode(pair.Value, fieldByIndex(rv, f), joinPath(path, pair.Key.Value)); err != nil {
			return err
		}
	}
	return nil
}

// fieldByIndex returns the field at index, allocating the pointers to inline
// structs on the way.
func fieldByIndex(rv reflect.Value, index []int) reflect.Value {
	for i, x := range index {
		if i > 0 && rv.Kind() == reflect.Pointer {
			if rv.IsNil() {
				rv.Set(reflect.New(rv.Type().Elem()))
			}
			rv = rv.Elem()
		}
		rv = rv.Field(x)
	}
	return rv
}

// parseBool reads the spellings of a bool in the YAML 1.2 core schema.
func parseBool(s string) (bool, bool) {
	switch s {
	case "true", "True", "TRUE":
		return true, true
	case "false", "False", "FALSE":
		return false, true
	}
	return false, false
}

// parseInt reads an integer: decimal, 0x hexadecimal, 0o or 0b, or a leading
// 0 for octal, with an optional sign and "_" between digits. It reads what
// atago's previous YAML reader did, so a spec or an output that spells
// `mode: 0755` means the number it meant before: 493.
func parseInt(s string) (int64, bool) {
	body, neg := numberBody(s)
	if body == "" {
		return 0, false
	}
	base := 10
	switch {
	case len(body) > 2 && body[0] == '0' && body[1] == 'x':
		base, body = 16, body[2:]
	case len(body) > 2 && body[0] == '0' && body[1] == 'o':
		base, body = 8, body[2:]
	case len(body) > 2 && body[0] == '0' && body[1] == 'b':
		base, body = 2, body[2:]
	case len(body) > 1 && body[0] == '0':
		base, body = 8, body[1:]
	}
	for _, c := range body {
		if !isDigit(c, base) {
			return 0, false
		}
	}
	u, err := strconv.ParseUint(body, base, 64)
	if err != nil {
		return 0, false
	}
	if neg {
		if u > 1<<63 {
			return 0, false
		}
		return -int64(u), true //nolint:gosec // bounded just above
	}
	if u > math.MaxInt64 {
		return 0, false
	}
	return int64(u), true
}

// numberBody strips the sign of a number and the "_" between its digits. A
// "_" before the first digit makes it a word, not a number.
func numberBody(s string) (string, bool) {
	neg := false
	if s != "" && (s[0] == '+' || s[0] == '-') {
		neg = s[0] == '-'
		s = s[1:]
	}
	if s == "" || s[0] == '_' {
		return "", neg
	}
	if strings.IndexByte(s, '_') >= 0 {
		s = strings.ReplaceAll(s, "_", "")
	}
	return s, neg
}

func isDigit(c rune, base int) bool {
	switch base {
	case 16:
		return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
	case 8:
		return c >= '0' && c <= '7'
	case 2:
		return c == '0' || c == '1'
	}
	return c >= '0' && c <= '9'
}

// parseFloat reads a float in the YAML 1.2 core schema: a decimal with a point
// or an exponent, or one of the spellings of infinity and not-a-number.
func parseFloat(s string) (float64, bool) {
	switch s {
	case ".inf", ".Inf", ".INF", "+.inf", "+.Inf", "+.INF":
		return math.Inf(1), true
	case "-.inf", "-.Inf", "-.INF":
		return math.Inf(-1), true
	case ".nan", ".NaN", ".NAN":
		return math.NaN(), true
	}
	if strings.IndexByte(s, '_') >= 0 {
		body, neg := numberBody(s)
		if body == "" {
			return 0, false
		}
		if s = body; neg {
			s = "-" + body
		}
	}
	if !looksFloat(s) {
		return 0, false
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil && !errors.Is(err, strconv.ErrRange) {
		return 0, false
	}
	return f, true
}

// looksFloat reports the core schema's float syntax with a point or an
// exponent, [-+]? ( \.[0-9]+ | [0-9]+ ( \.[0-9]* )? ) ( [eE] [-+]? [0-9]+ )?,
// which strconv alone would widen with spellings YAML reads as strings (Inf,
// 0x1p4).
func looksFloat(s string) bool {
	i := 0
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i++
	}
	digits := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
		digits++
	}
	if i < len(s) && s[i] == '.' {
		i++
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
			digits++
		}
	}
	if digits == 0 {
		return false
	}
	// Digits alone are an integer, and one too large to read stays the word it
	// is rather than turning into a float nobody wrote.
	point := strings.IndexByte(s, '.') >= 0
	if !point && (i == len(s) || s[i] != 'e' && s[i] != 'E') {
		return false
	}
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		i++
		if i < len(s) && (s[i] == '+' || s[i] == '-') {
			i++
		}
		exp := 0
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
			exp++
		}
		if exp == 0 {
			return false
		}
	}
	return i == len(s)
}

// plainValue reads a plain scalar by the core schema: null, a bool, an
// integer, a float, or the string it is. An integer too large for int64 is a
// uint64 when it fits one, and a float otherwise.
func plainValue(s string) (any, bool) {
	if isNullWord(s) {
		return nil, true
	}
	if b, ok := parseBool(s); ok {
		return b, true
	}
	if i, ok := parseInt(s); ok {
		return i, true
	}
	if body, neg := numberBody(s); !neg && body != "" && body[0] != '0' {
		if u, err := strconv.ParseUint(body, 10, 64); err == nil {
			return u, true
		}
	}
	if f, ok := parseFloat(s); ok {
		return f, true
	}
	return s, false
}

// generic turns a node into the plain Go value it holds.
func (d *decoder) generic(n *Node, path string) (any, error) {
	if n == nil {
		return nil, nil //nolint:nilnil // an absent value is null, which is nil
	}
	if err := d.visit(n); err != nil {
		return nil, err
	}
	switch n.Kind {
	case SequenceNode:
		out := make([]any, len(n.Items))
		for i, item := range n.Items {
			v, err := d.generic(item, indexPath(path, i))
			if err != nil {
				return nil, err
			}
			out[i] = v
		}
		return out, nil
	case MappingNode:
		out := make(map[string]any, len(n.Pairs))
		for _, pair := range n.Pairs {
			if pair.Key.Kind != ScalarNode {
				return nil, d.fail(pair.Key, path, "a mapping key has to be a scalar")
			}
			v, err := d.generic(pair.Value, joinPath(path, pair.Key.Value))
			if err != nil {
				return nil, err
			}
			out[pair.Key.Value] = v
		}
		return out, nil
	case ScalarNode:
	}
	return d.tagged(n, path)
}

// tagged reads a scalar by its tag, or by the core schema when it has none or
// a tag this reader does not know.
func (d *decoder) tagged(n *Node, path string) (any, error) {
	switch n.Tag {
	case "!!str", "!":
		// "!" is the non-specific tag, which makes a plain scalar a string.
		return n.Value, nil
	case "!!null":
		return nil, nil //nolint:nilnil // null is nil
	case "!!binary":
		b, err := decodeBinary(n.Value)
		if err != nil {
			return nil, d.fail(n, path, "the !!binary value is not base64: %v", err)
		}
		return string(b), nil
	case "!!bool":
		if b, ok := parseBool(n.Value); ok {
			return b, nil
		}
		return n.Value, nil
	case "!!int":
		if i, ok := parseInt(n.Value); ok {
			return i, nil
		}
		return n.Value, nil
	case "!!float":
		if f, ok := parseFloat(n.Value); ok {
			return f, nil
		}
		if i, ok := parseInt(n.Value); ok {
			return float64(i), nil
		}
		return n.Value, nil
	}
	if n.Style != PlainStyle {
		return n.Value, nil
	}
	v, _ := plainValue(n.Value)
	return v, nil
}

// field is the index path of a struct field, through the inline structs it is
// promoted from.
type field = []int

var fieldCache sync.Map // reflect.Type -> map[string]field

// structFields maps the yaml key of every field of t to its index path. A
// field's key is its `yaml` tag, or its name in lower case; a tag of "-" leaves
// the field out, and ",inline" promotes the fields of a struct field into t.
func structFields(t reflect.Type) map[string]field {
	if m, ok := fieldCache.Load(t); ok {
		fields, _ := m.(map[string]field)
		return fields
	}
	m := map[string]field{}
	collectFields(t, nil, m)
	fieldCache.Store(t, m)
	return m
}

func collectFields(t reflect.Type, prefix []int, m map[string]field) {
	// A struct's own fields come first, so a field written in it wins over one
	// of the same name an inline struct would promote.
	var inline []int
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		name, opts, _ := strings.Cut(f.Tag.Get("yaml"), ",")
		if name == "-" {
			continue
		}
		if hasOption(opts, "inline") {
			inline = append(inline, i)
			continue
		}
		if name == "" {
			name = strings.ToLower(f.Name)
		}
		if _, dup := m[name]; !dup {
			m[name] = append(append([]int(nil), prefix...), i)
		}
	}
	for _, i := range inline {
		ft := t.Field(i).Type
		if ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}
		collectFields(ft, append(append([]int(nil), prefix...), i), m)
	}
}

func hasOption(opts, want string) bool {
	for opts != "" {
		var o string
		o, opts, _ = strings.Cut(opts, ",")
		if o == want {
			return true
		}
	}
	return false
}
