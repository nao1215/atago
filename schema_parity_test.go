package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/spec"
)

// This file holds the spec-key inventory drift guards. The set of keys a spec
// file may contain is written down FOUR times: the yaml struct tags in
// internal/spec, the per-field validators in internal/loader, the published
// schema/atago.schema.json (with additionalProperties: false), and the website
// Reference "Since" table in website/data/spec_keys.json. Those four have no
// mechanical link, so they rot independently: a struct field the loader accepts
// but the schema omits ships silently, then editors that validate against the
// schema reject a spec the loader would happily run; a schema key with no
// spec_keys.json entry leaves the website Since column silently blank because
// that JSON is regenerated only when someone remembers to run
// website/tools/gen-spec-keys.py by hand. The existing schema_test.go only
// spot-checks a handful of demo specs and hand-written snippets, so none of that
// rot is caught. The two tests below close both gaps by deriving the key
// inventory from each source and comparing the whole set, not samples.

// defBoundaries maps every internal/spec struct type that the JSON Schema models
// as a named `$defs/<name>` definition to that definition's name. A field whose
// (element) type is one of these is a `$ref` in the schema — the schema records
// the field itself and stops, expanding the referenced definition once at the
// top level — so the reflection walk below treats such a field as a leaf and
// walks the type separately with the definition name as its prefix. This mirrors
// exactly what website/tools/gen-spec-keys.py does: top-level `properties` plus
// each `$defs` entry expanded under its own name, never following `$ref`.
func defBoundaries() map[reflect.Type]string {
	return map[reflect.Type]string{
		reflect.TypeOf(spec.Defaults{}):       "defaults",
		reflect.TypeOf(spec.Runner{}):         "runner",
		reflect.TypeOf(spec.Condition{}):      "condition",
		reflect.TypeOf(spec.Scenario{}):       "scenario",
		reflect.TypeOf(spec.Service{}):        "service",
		reflect.TypeOf(spec.Step{}):           "step",
		reflect.TypeOf(spec.CDP{}):            "cdp",
		reflect.TypeOf(spec.Query{}):          "query",
		reflect.TypeOf(spec.GRPC{}):           "grpc",
		reflect.TypeOf(spec.Fixture{}):        "fixture",
		reflect.TypeOf(spec.Run{}):            "run",
		reflect.TypeOf(spec.Retry{}):          "retry",
		reflect.TypeOf(spec.PTY{}):            "pty",
		reflect.TypeOf(spec.MockServer{}):     "mockServer",
		reflect.TypeOf(spec.MockRoute{}):      "mockRoute",
		reflect.TypeOf(spec.MockAssert{}):     "mockAssert",
		reflect.TypeOf(spec.Signal{}):         "signal",
		reflect.TypeOf(spec.HTTP{}):           "http",
		reflect.TypeOf(spec.Assert{}):         "assert",
		reflect.TypeOf(spec.ChangesAssert{}):  "changesAssert",
		reflect.TypeOf(spec.DurationAssert{}): "durationAssert",
		reflect.TypeOf(spec.PDFAssert{}):      "pdf",
		reflect.TypeOf(spec.DirAssert{}):      "dir",
		reflect.TypeOf(spec.ExitCode{}):       "exitCode",
		reflect.TypeOf(spec.StreamAssert{}):   "stream",
		reflect.TypeOf(spec.FileAssert{}):     "file",
		reflect.TypeOf(spec.ImageAssert{}):    "image",
		reflect.TypeOf(spec.JSONAssert{}):     "jsonAssert",
		reflect.TypeOf(spec.HeaderMatch{}):    "headerMatch",
		reflect.TypeOf(spec.Store{}):          "store",
	}
}

// polymorphicKeys lists the sub-keys of the few spec nodes whose YAML shape is
// hand-decoded by a custom UnmarshalYAML rather than expressed with struct tags,
// so plain reflection over their fields sees no yaml names. Each is a scalar-or-
// mapping union; the sub-keys here are the mapping-form keys the schema exposes
// (the scalar form carries no key). Stdin (`run.stdin`, a string / {file} /
// {base64}) and PTYSend (`pty.session[].send`, a string / {key}) are inline
// nodes, so their keys hang off the field path. ExitCode (`exit_code`, an int /
// {not} / {in}) is a `$defs/exitCode` boundary, so its keys hang off the
// definition name; it is applied in collectStructPaths where the definition is
// walked, not here.
func polymorphicKeys() map[reflect.Type][]string {
	return map[reflect.Type][]string{
		reflect.TypeOf(spec.Stdin{}):   {"file", "base64"},
		reflect.TypeOf(spec.PTYSend{}): {"key"},
	}
}

// exitCodeKeys are the mapping-form keys of the polymorphic exit_code node
// (`{not: N}` / `{in: [...]}`); the bare-integer form carries no key. ExitCode
// is a `$defs/exitCode` boundary, so these attach to the definition name.
var exitCodeKeys = []string{"not", "in"}

// collectStructPaths walks the internal/spec struct graph from the root Spec
// type and returns every reachable yaml-tag path in the SAME `$defs`-relative
// shape the schema uses ("run.command", "cdp.actions[].click", "suite.name") —
// the shape website/tools/gen-spec-keys.py emits. It walks the root with an
// empty prefix and each `$defs` boundary type under its definition name,
// stopping at a boundary field (the schema's `$ref`) so a definition is walked
// exactly once. Pointers and slices/maps are dereferenced; `yaml:"-"` fields are
// skipped; the custom-decoded polymorphic nodes contribute their hand-listed
// keys.
func collectStructPaths() map[string]bool {
	boundaries := defBoundaries()
	poly := polymorphicKeys()
	paths := map[string]bool{}

	var walkStruct func(t reflect.Type, prefix string)

	recurse := func(t reflect.Type, path string) {
		for t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		switch t.Kind() {
		case reflect.Struct:
			if _, ok := boundaries[t]; ok {
				return // schema `$ref`: recorded as a leaf, expanded separately
			}
			if subs, ok := poly[t]; ok {
				for _, s := range subs {
					paths[path+"."+s] = true
				}
				return
			}
			walkStruct(t, path)
		case reflect.Slice, reflect.Array:
			el := t.Elem()
			for el.Kind() == reflect.Pointer {
				el = el.Elem()
			}
			if el.Kind() != reflect.Struct {
				return // scalar slice or slice-of-map: a leaf
			}
			if _, ok := boundaries[el]; ok {
				return // e.g. []Step, []Scenario, []JSONAssert: leaf
			}
			if subs, ok := poly[el]; ok {
				for _, s := range subs {
					paths[path+"[]."+s] = true
				}
				return
			}
			walkStruct(el, path+"[]")
		case reflect.Map:
			// Map values are scalar strings or `$ref` structs; the schema records
			// only the field (additionalProperties never adds a path here).
			return
		default:
			// scalar / any: a leaf.
		}
	}

	walkStruct = func(t reflect.Type, prefix string) {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			tag := f.Tag.Get("yaml")
			name, opts, _ := strings.Cut(tag, ",")
			if name == "-" {
				continue
			}
			// Embedded or `,inline` struct: merge its fields under the same prefix.
			if f.Anonymous || (name == "" && strings.Contains(opts, "inline")) {
				et := f.Type
				for et.Kind() == reflect.Pointer {
					et = et.Elem()
				}
				if et.Kind() == reflect.Struct {
					walkStruct(et, prefix)
				}
				continue
			}
			if name == "" {
				continue // untagged non-embedded field (none on the walked types)
			}
			if f.PkgPath != "" {
				continue // unexported
			}
			path := name
			if prefix != "" {
				path = prefix + "." + name
			}
			paths[path] = true
			recurse(f.Type, path)
		}
	}

	walkStruct(reflect.TypeOf(spec.Spec{}), "")
	for typ, prefix := range boundaries {
		if typ == reflect.TypeOf(spec.ExitCode{}) {
			for _, s := range exitCodeKeys {
				paths[prefix+"."+s] = true
			}
			continue
		}
		walkStruct(typ, prefix)
	}
	return paths
}

// collectSchemaPaths ports website/tools/gen-spec-keys.py's extract_paths to Go:
// it collects every property path in schema/atago.schema.json — top-level
// `properties`, then each `$defs` definition expanded under its own name, with
// `[]` for array items and `.*` for an object's additionalProperties, descending
// through oneOf/anyOf but never following `$ref`. Keeping this in lockstep with
// the Python source is what lets the two guards below compare like with like.
func collectSchemaPaths(t *testing.T) map[string]bool {
	t.Helper()
	data, err := os.ReadFile("schema/atago.schema.json")
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("parse schema: %v", err)
	}
	paths := map[string]bool{}
	var walk func(o any, prefix string)
	walk = func(o any, prefix string) {
		m, ok := o.(map[string]any)
		if !ok {
			return
		}
		if props, ok := m["properties"].(map[string]any); ok {
			for k, v := range props {
				p := k
				if prefix != "" {
					p = prefix + "." + k
				}
				paths[p] = true
				walk(v, p)
			}
		}
		if items, ok := m["items"].(map[string]any); ok {
			walk(items, prefix+"[]")
		}
		if ap, ok := m["additionalProperties"].(map[string]any); ok {
			walk(ap, prefix+".*")
		}
		for _, key := range []string{"oneOf", "anyOf"} {
			if arr, ok := m[key].([]any); ok {
				for _, x := range arr {
					walk(x, prefix)
				}
			}
		}
	}
	walk(root, "")
	if defs, ok := root["$defs"].(map[string]any); ok {
		for name, d := range defs {
			walk(d, name)
		}
	}
	return paths
}

// TestSpecSchema_StructParity is the drift guard between the internal/spec
// structs (which, with internal/loader, define exactly what a spec may contain)
// and the published JSON Schema. Both are hand-maintained and have no mechanical
// link, so they rot in opposite, equally bad directions: a struct field the
// loader accepts but the schema forgot means schema-validating editors flag a
// spec the loader runs fine, while a schema key the structs never model means
// the schema promises a field the loader rejects. This walks both inventories in
// the same `$defs`-relative shape and asserts they are equal in both directions,
// so the next field added to a struct (or the schema) fails the build unless its
// counterpart is added too.
//
// The one intentional asymmetry is the `defaults` block: the schema hand-curates
// `defaults.run` / `defaults.service` as a narrower projection of a run / service
// (a default cannot carry a per-step command, stdin, redirect, retry, or a
// service's identifying name/command), while the Go model reuses the full Run /
// Service types. Rather than hand-list those excluded keys, the schema→struct
// direction accepts a `defaults.run.<x>` / `defaults.service.<x>` key exactly
// when `run.<x>` / `service.<x>` is a real modeled path — i.e. the schema's
// defaults really are a projection of the corresponding type.
func TestSpecSchema_StructParity(t *testing.T) {
	structP := collectStructPaths()
	schemaP := collectSchemaPaths(t)

	for p := range structP {
		if !schemaP[p] {
			t.Errorf("struct/loader accepts %q but schema/atago.schema.json has no such property; add it (with a description, like every existing property) or a schema-validating editor will reject specs the loader runs", p)
		}
	}
	for p := range schemaP {
		if structP[p] {
			continue
		}
		// defaults.run.* / defaults.service.* are a curated projection of the run
		// / service types (see the doc comment); accept them when the projected
		// key is a real modeled path.
		if strings.HasPrefix(p, "defaults.run.") || strings.HasPrefix(p, "defaults.service.") {
			if structP[strings.TrimPrefix(p, "defaults.")] {
				continue
			}
		}
		t.Errorf("schema exposes property %q but no internal/spec struct field models it; remove it from the schema or add the field, so the schema never promises a key the loader rejects", p)
	}
}

// TestSpecSchema_SpecKeysComplete guards the website Reference "Since" table
// (website/data/spec_keys.json) against the schema. That JSON is regenerated only
// when someone remembers to run `python3 website/tools/gen-spec-keys.py`, so
// adding a schema key and forgetting the regen leaves the website's Since column
// silently blank for the new key — and a key removed from the schema leaves a
// stale row pointing at a field that no longer exists. This asserts the two key
// sets are equal in both directions, turning that "forgot to regenerate" mistake
// into a build failure with the exact command to fix it.
func TestSpecSchema_SpecKeysComplete(t *testing.T) {
	schemaP := collectSchemaPaths(t)

	data, err := os.ReadFile("website/data/spec_keys.json")
	if err != nil {
		t.Fatalf("read spec_keys.json: %v", err)
	}
	var doc struct {
		Keys map[string]string `json:"keys"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse spec_keys.json: %v", err)
	}

	for p := range schemaP {
		if _, ok := doc.Keys[p]; !ok {
			t.Errorf("schema property %q is missing from website/data/spec_keys.json; regenerate it with `python3 website/tools/gen-spec-keys.py` so the website Since column covers it", p)
		}
	}
	for k := range doc.Keys {
		if !schemaP[k] {
			t.Errorf("website/data/spec_keys.json lists %q, which is no longer a schema property; regenerate it with `python3 website/tools/gen-spec-keys.py`", k)
		}
	}
}
