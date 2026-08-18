package loader

import (
	"reflect"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/nao1215/atago/internal/spec"
)

// applyVerbatimScalars restores the source text of every plain scalar that
// landed in a text field. YAML types an unquoted `1.20` as a float and an
// unquoted `TRUE` as a boolean, and decoding one of those into a Go string
// keeps the parsed VALUE and drops the text: `contains: 1.20` became the
// substring "1.2" (which matches output that never contained 1.20), `env: {V:
// 007}` exported "7" to the program under test, and a fixture written as
// `content: 0x10` wrote "16". atago compares literals, so in a text field the
// characters in the file are the contract, and quoting a value must not change
// what it means.
//
// It runs on the decoded spec BEFORE matrix expansion and defaults merging,
// while the Go value graph still mirrors the document one-to-one, and walks the
// two together: a repair is made only where the document has a non-string
// scalar and the model has a text field at the same place. Fields whose Go type
// is `any` are left alone — `json: {equals: true}` compares against the JSON
// boolean on purpose — and so are aliases, explicit tags (`!!binary`, which
// `atago record` writes), and YAML null, whose text is not what "unset" means.
//
// A document that no longer parses, or that never matched the model, simply
// leaves the decoded values as they are: this pass improves fidelity and must
// never be able to fail a load that would otherwise succeed.
func applyVerbatimScalars(s *spec.Spec, data []byte) {
	f, err := parser.ParseBytes(data, 0)
	if err != nil || f == nil || len(f.Docs) == 0 || f.Docs[0] == nil {
		return
	}
	repairValue(reflect.ValueOf(s), f.Docs[0].Body)
}

// verbatimScalar returns the source text of a scalar node whose YAML type is
// not string, which is exactly the case where decoding into a text field loses
// the text. A string scalar (plain, quoted, or block) already decoded verbatim,
// and a null carries no text a field could take.
func verbatimScalar(n ast.Node) (string, bool) {
	switch n.(type) {
	case *ast.BoolNode, *ast.IntegerNode, *ast.FloatNode, *ast.InfinityNode, *ast.NanNode:
		tk := n.GetToken()
		if tk == nil {
			return "", false
		}
		return tk.Value, true
	default:
		return "", false
	}
}

// repairValue walks one model value against the document node it was decoded
// from.
func repairValue(v reflect.Value, n ast.Node) {
	if n == nil || !v.IsValid() {
		return
	}
	switch node := n.(type) {
	case *ast.DocumentNode:
		repairValue(v, node.Body)
		return
	case *ast.AnchorNode:
		repairValue(v, node.Value)
		return
	case *ast.AliasNode, *ast.TagNode:
		// An alias resolves to a node elsewhere in the document (repaired
		// there), and a tag states the type explicitly, which is the one case
		// where the author asked for something other than the text.
		return
	}
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return
		}
		repairValue(v.Elem(), n)
	case reflect.Interface:
		// An `any` field keeps YAML's typing: it is what lets a json matcher
		// tell the string "true" from the boolean true.
	case reflect.String:
		if raw, ok := verbatimScalar(n); ok && v.CanSet() {
			v.SetString(raw)
		}
	case reflect.Struct:
		repairStruct(v, n)
	case reflect.Slice, reflect.Array:
		repairSlice(v, n)
	case reflect.Map:
		repairMap(v, n)
	default:
		// A number, a bool, or a duration-typed field decoded as itself.
	}
}

// repairSlice walks a sequence node element-wise. A node that is not a sequence
// is the scalar-or-single form of a polymorphic field — `contains: warn` for a
// StringList, one mapping for JSONChecks — which decodes to exactly one
// element, so the whole node belongs to that element.
func repairSlice(v reflect.Value, n ast.Node) {
	if seq, ok := n.(*ast.SequenceNode); ok {
		for i, item := range seq.Values {
			if i >= v.Len() {
				return
			}
			repairValue(v.Index(i), item)
		}
		return
	}
	if v.Len() == 1 {
		repairValue(v.Index(0), n)
	}
}

// repairMap walks a mapping node against a map value. Both halves of an entry
// can lose their text, so a key written `TRUE` is moved back to the name the
// author typed, and a map value is rewritten through SetMapIndex because a map
// element is not addressable.
func repairMap(v reflect.Value, n ast.Node) {
	if v.IsNil() {
		return
	}
	if v.Type().Key().Kind() != reflect.String {
		return
	}
	for _, entry := range mappingEntries(n) {
		key, ok := decodedKey(entry.Key)
		if !ok {
			continue
		}
		cur := v.MapIndex(reflect.ValueOf(key).Convert(v.Type().Key()))
		if !cur.IsValid() {
			continue
		}
		// Repair the value into an addressable copy, then store it back.
		box := reflect.New(v.Type().Elem())
		box.Elem().Set(cur)
		repairValue(box.Elem(), entry.Value)
		v.SetMapIndex(reflect.ValueOf(key).Convert(v.Type().Key()), box.Elem())

		if raw, isScalar := verbatimScalar(entry.Key); isScalar && raw != key {
			zero := reflect.Value{}
			v.SetMapIndex(reflect.ValueOf(key).Convert(v.Type().Key()), zero)
			v.SetMapIndex(reflect.ValueOf(raw).Convert(v.Type().Key()), box.Elem())
		}
	}
}

// decodedKey returns the string a mapping key decoded to, which is what the map
// is keyed by — `TRUE` decodes to "true" for the same reason a value does.
func decodedKey(n ast.Node) (string, bool) {
	if n == nil {
		return "", false
	}
	var k string
	if err := yaml.NodeToValue(n, &k); err != nil {
		return "", false
	}
	return k, true
}

// repairStruct walks a mapping node against a struct. Most spec types describe
// their mapping with yaml tags; the few that decode themselves are handled from
// customDecodes, which also names the field a scalar form lands in.
func repairStruct(v reflect.Value, n ast.Node) {
	custom, hasCustom := customDecodes[v.Type()]
	if raw, ok := verbatimScalar(n); ok {
		// A scalar where a mapping was expected is the scalar form of a
		// polymorphic node; only a type that declares one has a field for it.
		if hasCustom {
			repairScalarForm(v, custom, raw)
		}
		return
	}
	entries := mappingEntries(n)
	if len(entries) == 0 {
		return
	}
	fields := yamlFields(v.Type())
	for _, entry := range entries {
		key, ok := decodedKey(entry.Key)
		if !ok {
			continue
		}
		if hasCustom && len(custom.keys) > 0 {
			name, known := custom.keys[key]
			if !known {
				continue
			}
			if f := v.FieldByName(name); f.IsValid() {
				repairValue(f, entry.Value)
			}
			continue
		}
		index, known := fields[key]
		if !known {
			continue
		}
		repairValue(v.FieldByIndex(index), entry.Value)
	}
}

// repairScalarForm assigns the raw text to the field a polymorphic type's
// scalar form decodes into. A type that declares no such field wrote its scalar
// somewhere this walk cannot see, so nothing is repaired.
func repairScalarForm(v reflect.Value, custom customDecode, raw string) {
	if custom.scalarField == "" {
		return
	}
	if f := v.FieldByName(custom.scalarField); f.IsValid() && f.CanSet() {
		setScalarField(f, raw)
	}
}

// setScalarField assigns raw text to the field a polymorphic node's scalar form
// decodes into, which is a string on some types and a *string on others (the
// pointer is how they tell "written empty" from "not written").
func setScalarField(f reflect.Value, raw string) {
	switch f.Kind() {
	case reflect.String:
		f.SetString(raw)
	case reflect.Pointer:
		if f.Type().Elem().Kind() != reflect.String {
			return
		}
		if f.IsNil() {
			f.Set(reflect.New(f.Type().Elem()))
		}
		f.Elem().SetString(raw)
	default:
	}
}

// mappingEntries flattens the node shapes goccy uses for a mapping: a block or
// flow mapping, and the single-entry node a one-key mapping can decode to.
func mappingEntries(n ast.Node) []*ast.MappingValueNode {
	switch node := n.(type) {
	case *ast.MappingNode:
		return node.Values
	case *ast.MappingValueNode:
		return []*ast.MappingValueNode{node}
	default:
		return nil
	}
}

// yamlFields maps a struct's yaml keys to field index paths, descending into
// embedded and `,inline` structs the way the decoder does.
func yamlFields(t reflect.Type) map[string][]int {
	out := map[string][]int{}
	var walk func(t reflect.Type, prefix []int)
	walk = func(t reflect.Type, prefix []int) {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" && !f.Anonymous {
				continue // unexported
			}
			name, opts, _ := strings.Cut(f.Tag.Get("yaml"), ",")
			index := append(append([]int(nil), prefix...), i)
			if name == "-" {
				continue
			}
			if f.Anonymous || (name == "" && strings.Contains(opts, "inline")) {
				et := f.Type
				for et.Kind() == reflect.Pointer {
					et = et.Elem()
				}
				if et.Kind() == reflect.Struct {
					walk(et, index)
					continue
				}
			}
			if name == "" {
				continue
			}
			out[name] = index
		}
	}
	walk(t, nil)
	return out
}

// customDecode describes how a spec type that decodes itself maps a document
// node onto its fields. The decoder calls the type's own UnmarshalYAML, so the
// yaml tags on such a type say nothing about the shapes it accepts, and this
// walk would otherwise skip past the text it holds.
type customDecode struct {
	// scalarField is the field the scalar form decodes into.
	scalarField string
	// keys maps the mapping form's keys to fields, for the types whose mapping
	// is hand-decoded and whose fields therefore carry no yaml tags. When empty,
	// the struct's yaml tags describe the mapping form.
	keys map[string]string
	// why states how this type's text is reached, so "nobody looked at this
	// one" cannot pass as a decision. Required on every entry.
	why string
}

// customDecodes covers every spec type that decodes itself. A type missing from
// here silently keeps a retyped scalar, so TestVerbatim_EveryCustomDecodeIsCovered
// walks the spec model and fails when one is added without an entry.
var customDecodes = map[reflect.Type]customDecode{
	// `stdin: 007` is text fed to the child, and the mapping form names a path
	// or a base64 payload.
	reflect.TypeFor[spec.Stdin](): {
		scalarField: "Inline",
		keys:        map[string]string{"file": "File", "base64": "Base64"},
		why:         "a scalar is inline stdin; the mapping form is hand-decoded and carries no yaml tags",
	},
	// `send: 007` types those three characters into the terminal; `paste:` is
	// text too, while a key name and a mouse report are never numbers.
	reflect.TypeFor[spec.PTYSend](): {
		scalarField: "Text",
		keys:        map[string]string{"paste": "Paste", "mouse": "Mouse"},
		why:         "a scalar is typed input and paste is text; a key name and a mouse report are hand-decoded",
	},
	// `exec:` is a command line in either form, and its mapping carries yaml
	// tags the generic walk can follow.
	reflect.TypeFor[spec.PTYExec](): {
		scalarField: "Command",
		why:         "a scalar is the command line; the mapping form carries yaml tags",
	},
	// A json check's mapping carries yaml tags; `equals` is `any` and keeps
	// YAML's typing, which the generic walk already respects.
	reflect.TypeFor[spec.JSONAssert](): {why: "the mapping form carries yaml tags, and equals is `any`, which keeps YAML's typing"},
	reflect.TypeFor[spec.ExitCode]():   {why: "a number, {not: N}, or {in: [...]}: every field is an int, so there is no text to lose"},
	// A scalar or a sequence of them; repairSlice takes the scalar form as the
	// single element it decoded to.
	reflect.TypeFor[spec.StringList](): {why: "a slice: the sequence walk repairs the scalar and list forms alike"},
	reflect.TypeFor[spec.JSONChecks](): {why: "a slice: one mapping or a sequence of them, both reached by the sequence walk"},
}
