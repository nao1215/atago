package yaml

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Marshaler is a type that writes itself as the value MarshalYAML returns, for
// a value that has more than one shape.
type Marshaler interface {
	MarshalYAML() (any, error)
}

// MapItem is one entry of a MapSlice.
type MapItem struct {
	Key   string
	Value any
}

// MapSlice is a mapping whose entries are written in the order given.
type MapSlice []MapItem

// Quote renders s as one YAML scalar that fits on the line after a key or a
// dash and reads back as exactly s: bare when it can be, double-quoted with
// escapes otherwise. Text that is not valid UTF-8 cannot be a YAML string at
// all; its invalid bytes are written as \x escapes, which read back as the
// code points U+0080 to U+00FF rather than the bytes.
func Quote(s string) string {
	if plainSafe(s) {
		return s
	}
	return doubleQuote(s)
}

// plainSafe reports whether s can be written bare: it would read back as the
// same string and not as null, a bool, a number, a comment, a key or anything
// else YAML gives a meaning to.
func plainSafe(s string) bool {
	if s == "" || !utf8.ValidString(s) {
		return false
	}
	if s[0] == ' ' || s[len(s)-1] == ' ' || strings.HasPrefix(s, "...") || strings.HasPrefix(s, "---") {
		return false
	}
	switch s[0] {
	case '-', '?', ':':
		// "-x" is a plain scalar, but "- x" and a lone "-" are not; quoting
		// all three keeps the rule simple.
		return false
	case ',', '[', ']', '{', '}', '#', '&', '*', '!', '|', '>', '\'', '"', '%', '@', '`', '\t':
		return false
	}
	if strings.Contains(s, ": ") || strings.Contains(s, " #") || strings.HasSuffix(s, ":") {
		return false
	}
	for _, r := range s {
		if r < 0x20 || r == 0x7f || r == 0x85 || r == 0xfeff || r == 0x2028 || r == 0x2029 {
			return false
		}
	}
	// A bare word that reads as null, a bool or a number is written quoted, so
	// it stays the string it is.
	if _, typed := plainValue(s); typed {
		return false
	}
	if _, ok := parseFloat(s); ok {
		return false
	}
	return true
}

// doubleQuote writes s as a double-quoted scalar on one line.
func doubleQuote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			fmt.Fprintf(&b, `\x%02x`, s[i])
			i++
			continue
		}
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		case '\r':
			b.WriteString(`\r`)
		case 0x85:
			b.WriteString(`\N`)
		case 0x2028:
			b.WriteString(`\L`)
		case 0x2029:
			b.WriteString(`\P`)
		case 0xfeff:
			b.WriteString(`\ufeff`)
		default:
			if r < 0x20 || r == 0x7f {
				fmt.Fprintf(&b, `\x%02x`, r)
			} else {
				b.WriteRune(r)
			}
		}
		i += size
	}
	b.WriteByte('"')
	return b.String()
}

// Marshal writes v as a YAML document in block style: a struct as a mapping of
// its yaml-tagged fields in the order declared, a map with its keys sorted, a
// slice as a sequence, and a Marshaler as the value it returns.
func Marshal(v any) ([]byte, error) {
	e := &encoder{}
	if err := e.value(reflect.ValueOf(v), 0, atLine); err != nil {
		return nil, err
	}
	out := e.buf.Bytes()
	if len(out) == 0 || out[len(out)-1] != '\n' {
		out = append(out, '\n')
	}
	return out, nil
}

type encoder struct {
	buf   bytes.Buffer
	depth int
}

var marshalerType = reflect.TypeFor[Marshaler]()

// entry is one key and value of a mapping being written.
type entry struct {
	key string
	val reflect.Value
}

// place is where the value being written starts.
type place int

const (
	// atLine is the start of a line: a collection's entries are written at
	// the indentation given.
	atLine place = iota
	// afterKey follows "key:": a scalar continues the line, and a collection
	// starts on the next.
	afterKey
	// afterDash follows "-": a scalar continues the line, and so does the
	// first entry of a collection, the rest lining up under it.
	afterDash
)

// value writes v, whose collection entries stand at column indent.
func (e *encoder) value(v reflect.Value, indent int, at place) error {
	e.depth++
	defer func() { e.depth-- }()
	if e.depth > MaxDepth {
		return fmt.Errorf("yaml: the value nests deeper than %d levels", MaxDepth)
	}
	v, err := resolve(v)
	if err != nil {
		return err
	}
	if !v.IsValid() {
		e.scalar("null", at)
		return nil
	}
	if ms, ok := v.Interface().(MapSlice); ok {
		entries := make([]entry, len(ms))
		for i, it := range ms {
			entries[i] = entry{it.Key, reflect.ValueOf(it.Value)}
		}
		return e.mapping(entries, indent, at)
	}
	switch v.Kind() {
	case reflect.String:
		e.scalar(Quote(v.String()), at)
	case reflect.Bool:
		e.scalar(strconv.FormatBool(v.Bool()), at)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		e.scalar(strconv.FormatInt(v.Int(), 10), at)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		e.scalar(strconv.FormatUint(v.Uint(), 10), at)
	case reflect.Float32, reflect.Float64:
		e.scalar(formatFloat(v.Float()), at)
	case reflect.Slice, reflect.Array:
		items := make([]reflect.Value, v.Len())
		for i := range items {
			items[i] = v.Index(i)
		}
		return e.sequence(items, indent, at)
	case reflect.Map:
		keys := v.MapKeys()
		slices.SortFunc(keys, func(a, b reflect.Value) int { return strings.Compare(fmt.Sprint(a), fmt.Sprint(b)) })
		entries := make([]entry, len(keys))
		for i, k := range keys {
			entries[i] = entry{fmt.Sprint(k.Interface()), v.MapIndex(k)}
		}
		return e.mapping(entries, indent, at)
	case reflect.Struct:
		return e.mapping(structEntries(v), indent, at)
	default:
		return fmt.Errorf("yaml: cannot marshal %s", v.Type())
	}
	return nil
}

// resolve follows pointers and interfaces and calls a Marshaler, until v is a
// plain value or invalid (null).
func resolve(v reflect.Value) (reflect.Value, error) {
	for range MaxDepth {
		if !v.IsValid() {
			return v, nil
		}
		if v.Type().Implements(marshalerType) && (v.Kind() != reflect.Pointer || !v.IsNil()) {
			m, _ := v.Interface().(Marshaler)
			out, err := m.MarshalYAML()
			if err != nil {
				return v, err
			}
			v = reflect.ValueOf(out)
			continue
		}
		switch v.Kind() {
		case reflect.Pointer, reflect.Interface:
			if v.IsNil() {
				return reflect.Value{}, nil
			}
			v = v.Elem()
		default:
			return v, nil
		}
	}
	return v, fmt.Errorf("yaml: the value refers to itself")
}

func formatFloat(f float64) string {
	switch {
	case math.IsInf(f, 1):
		return ".inf"
	case math.IsInf(f, -1):
		return "-.inf"
	case math.IsNaN(f):
		return ".nan"
	}
	s := strconv.FormatFloat(f, 'g', -1, 64)
	if !strings.ContainsAny(s, ".eE") {
		s += ".0"
	}
	return s
}

func (e *encoder) scalar(text string, at place) {
	if at != atLine {
		e.buf.WriteByte(' ')
	}
	e.buf.WriteString(text)
	e.buf.WriteByte('\n')
}

// lead writes what comes before the i-th entry of a collection at indent.
func (e *encoder) lead(i, indent int, at place) {
	switch {
	case i == 0 && at == afterDash:
		e.buf.WriteByte(' ')
		return
	case i == 0 && at == afterKey:
		e.buf.WriteByte('\n')
	}
	for range indent {
		e.buf.WriteByte(' ')
	}
}

// sequence writes items one per "- " line at indent.
func (e *encoder) sequence(items []reflect.Value, indent int, at place) error {
	if len(items) == 0 {
		e.scalar("[]", at)
		return nil
	}
	for i, item := range items {
		e.lead(i, indent, at)
		e.buf.WriteByte('-')
		if err := e.value(item, indent+2, afterDash); err != nil {
			return err
		}
	}
	return nil
}

// mapping writes entries one per "key:" line at indent.
func (e *encoder) mapping(entries []entry, indent int, at place) error {
	if len(entries) == 0 {
		e.scalar("{}", at)
		return nil
	}
	for i, en := range entries {
		e.lead(i, indent, at)
		e.buf.WriteString(Quote(en.key))
		e.buf.WriteByte(':')
		if err := e.value(en.val, indent+2, afterKey); err != nil {
			return err
		}
	}
	return nil
}

// isZeroer is a value that says itself whether omitempty leaves it out.
type isZeroer interface{ IsZero() bool }

var isZeroerType = reflect.TypeFor[isZeroer]()

// structEntries lists the fields of a struct to write: by yaml tag, in the
// order declared, with ",inline" structs flattened into it and ",omitempty"
// fields left out when they are empty.
func structEntries(v reflect.Value) []entry {
	var out []entry
	t := v.Type()
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		name, opts, _ := strings.Cut(f.Tag.Get("yaml"), ",")
		if name == "-" {
			continue
		}
		fv := v.Field(i)
		if hasOption(opts, "inline") {
			if fv.Kind() == reflect.Pointer {
				if fv.IsNil() {
					continue
				}
				fv = fv.Elem()
			}
			out = append(out, structEntries(fv)...)
			continue
		}
		if hasOption(opts, "omitempty") && isEmpty(fv) {
			continue
		}
		if name == "" {
			name = strings.ToLower(f.Name)
		}
		out = append(out, entry{name, fv})
	}
	return out
}

// isEmpty reports a value omitempty leaves out: one that says it is zero, a nil
// pointer, an empty collection or string, or the zero value.
func isEmpty(v reflect.Value) bool {
	if v.Type().Implements(isZeroerType) && (v.Kind() != reflect.Pointer || !v.IsNil()) {
		z, _ := v.Interface().(isZeroer)
		return z.IsZero()
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Map, reflect.String, reflect.Array:
		return v.Len() == 0
	case reflect.Pointer, reflect.Interface:
		return v.IsNil()
	}
	return v.IsZero()
}
