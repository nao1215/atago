// Package schematest compares a JSON writer's Go structs against the JSON
// Schema published for that writer's output.
//
// atago ships schema/report.schema.json and schema/manifest.schema.json for
// editors and CI consumers to validate against, and both set
// "additionalProperties": false at every level. Nothing linked them to the
// structs that produce the documents, so every field added to a writer after
// the schema was written made the schema reject real output — and the only
// guard, a conformance check over a committed example, stayed green because the
// example never exercised the new field (#496).
//
// CheckParity closes that gap by deriving the key inventory from both sides and
// comparing the whole set in both directions, so the next added field fails at
// the source rather than when an example happens to include it.
package schematest

import (
	"encoding/json"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// CheckParity asserts that the JSON Schema in schemaJSON describes exactly the
// fields document's type emits: every `json:` tag has a schema property, and
// every schema property has a field. document is a zero value of the writer's
// top-level document type, and schemaPath only names the schema in failure
// messages (the caller reads it, so the path stays where the test can see it).
//
// It walks the two trees together, following $ref into $defs and items into
// array element types, and reports each mismatch with the JSON path it found.
// A type is walked once: the schema models a Go type as one definition, so a
// second visit would only repeat the same comparison.
func CheckParity(t *testing.T, schemaPath string, schemaJSON []byte, document any) {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal(schemaJSON, &root); err != nil {
		t.Fatalf("parse %s: %v", schemaPath, err)
	}
	w := &walker{t: t, schemaPath: schemaPath, root: root, seen: map[reflect.Type]bool{}}
	w.walk(root, reflect.TypeOf(document), "")
}

type walker struct {
	t          *testing.T
	schemaPath string
	root       map[string]any
	seen       map[reflect.Type]bool
}

// walk compares one schema node against one Go type, recursing into struct
// fields and array elements.
func (w *walker) walk(node map[string]any, typ reflect.Type, path string) {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	node = w.resolve(node)
	if node == nil {
		return
	}
	switch typ.Kind() {
	case reflect.Slice, reflect.Array:
		items, _ := node["items"].(map[string]any)
		if items == nil {
			w.t.Errorf("%s: %s is an array in Go but the schema declares no items", w.schemaPath, path)
			return
		}
		w.walk(items, typ.Elem(), path+"[]")
	case reflect.Struct:
		w.walkStruct(node, typ, path)
	default:
		// A scalar or a map: the schema constrains it with type /
		// additionalProperties, which carries no key inventory to compare.
	}
}

func (w *walker) walkStruct(node map[string]any, typ reflect.Type, path string) {
	if w.seen[typ] {
		return
	}
	w.seen[typ] = true
	props, _ := node["properties"].(map[string]any)
	fields := jsonFields(typ)
	for _, name := range slices.Sorted(maps.Keys(fields)) {
		if _, ok := props[name]; !ok {
			w.t.Errorf("%s: %s emits %q, which the schema does not declare — a real document carrying it fails validation (additionalProperties: false)",
				w.schemaPath, joinPath(path, name), name)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(props)) {
		if _, ok := fields[name]; !ok {
			w.t.Errorf("%s: the schema declares %s, which %s never emits — drop it or the writer field that went with it",
				w.schemaPath, joinPath(path, name), typ)
		}
	}
	for name, field := range fields {
		prop, _ := props[name].(map[string]any)
		if prop == nil {
			continue
		}
		w.walk(prop, field.Type, joinPath(path, name))
	}
}

// resolve follows a $ref to the definition it names, so a walk never has to
// know whether a node is written inline or shared.
func (w *walker) resolve(node map[string]any) map[string]any {
	ref, ok := node["$ref"].(string)
	if !ok {
		return node
	}
	name, found := strings.CutPrefix(ref, "#/$defs/")
	if !found {
		w.t.Errorf("%s: unsupported $ref %q (only #/$defs/<name> is walked)", w.schemaPath, ref)
		return nil
	}
	defs, _ := w.root["$defs"].(map[string]any)
	def, _ := defs[name].(map[string]any)
	if def == nil {
		w.t.Errorf("%s: $ref %q names no definition", w.schemaPath, ref)
		return nil
	}
	return def
}

// jsonFields returns the struct's serialized fields by their JSON name.
func jsonFields(typ reflect.Type) map[string]reflect.StructField {
	out := make(map[string]reflect.StructField, typ.NumField())
	for i := range typ.NumField() {
		f := typ.Field(i)
		if !f.IsExported() {
			continue
		}
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			continue
		}
		out[name] = f
	}
	return out
}

func joinPath(path, name string) string {
	if path == "" {
		return name
	}
	return path + "." + name
}
