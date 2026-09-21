package spec

import (
	"reflect"
	"strings"
	"testing"
)

// literalAssertFields records the assertion fields WalkAssertStrings
// deliberately does NOT expand, and why. Everything else that holds
// author-written text must be visited: the walker is the single field list the
// engine expands with and that explain/manifest collect variables from, so a
// field it skips is a field where `${name}` reaches the comparison as the
// literal six characters — and where the variable never shows up in the
// summaries either.
//
// Paths are yaml keys from the assert root, with `[]` for a list element and
// `.*` for a map value.
var literalAssertFields = map[string]string{
	"header.name":        "a protocol header name, not user text",
	"mock.name":          "references a mock server declared in the spec, checked against the declared set at load time",
	"file.snapshot":      "names a committed golden by a stable name; the file it points at is the value",
	"stdout.snapshot":    "same as file.snapshot",
	"stderr.snapshot":    "same as file.snapshot",
	"body.snapshot":      "same as file.snapshot",
	"rows.snapshot":      "same as file.snapshot",
	"message.snapshot":   "same as file.snapshot",
	"value.snapshot":     "same as file.snapshot",
	"screen.snapshot":    "same as file.snapshot",
	"pdf.text.snapshot":  "same as file.snapshot",
	"mock.body.snapshot": "same as file.snapshot",
	"image.format":       "an image format name (png/jpeg/…), matched against a fixed set",
	"dir.snapshot":       "same as file.snapshot",
	"mock.method":        "an HTTP method, matched case-insensitively against what the client sent",
	"mock.header.name":   "a protocol header name, as in header.name",
	"duration.lt":        "a Go duration the loader parses at load time, before any variable exists",
	"duration.lte":       "a Go duration the loader parses at load time, before any variable exists",
	"duration.gt":        "a Go duration the loader parses at load time, before any variable exists",
	"duration.gte":       "a Go duration the loader parses at load time, before any variable exists",
	"screen.attrs[].fg":  "a color name from a fixed set",
	"screen.attrs[].bg":  "a color name from a fixed set",
}

// TestWalkAssertStrings_VisitsEveryTextField fills every text field an assert
// can carry, walks it, and requires each one to have been visited or recorded
// above with a reason. It is the structural guard for the walk-forgetting bug
// family: `file.equals_file: "${workdir}/want.txt"` was compared against a path
// containing those literal characters for as long as the matcher existed,
// because the File branch of the walker listed four of its six text fields.
func TestWalkAssertStrings_VisitsEveryTextField(t *testing.T) {
	t.Parallel()
	const sentinel = "P"
	filled := &Assert{}
	fillText(reflect.ValueOf(filled).Elem(), sentinel, 0)
	walked := WalkAssertStrings(filled, func(s string) string { return s + "!" })

	var missing []string
	seen := map[string]bool{}
	compareText(reflect.ValueOf(filled).Elem(), reflect.ValueOf(walked).Elem(), "", func(path string, visited bool) {
		seen[path] = true
		reason, recorded := literalAssertFields[path]
		switch {
		case visited && recorded:
			t.Errorf("%s is visited by WalkAssertStrings but recorded as literal (%q); drop the entry", path, reason)
		case !visited && !recorded:
			missing = append(missing, path)
		}
	})
	for path := range literalAssertFields {
		if !seen[path] {
			t.Errorf("literalAssertFields records %q, which is no longer a text field of an assert; remove the entry", path)
		}
	}
	for _, path := range missing {
		t.Errorf("assert field %q holds author-written text that WalkAssertStrings never visits, so a ${name} there is compared literally and never reported by explain/manifest; visit it, or record it in literalAssertFields with the reason it is deliberately literal", path)
	}
}

// fillText sets every string-shaped leaf under v to sentinel, allocating the
// pointers, slices and maps it has to pass through. depth stops the recursion
// on a self-referential type; the assert model has none, and the bound is
// cheap insurance.
func fillText(v reflect.Value, sentinel string, depth int) {
	if depth > 6 || !v.CanSet() {
		return
	}
	switch v.Kind() {
	case reflect.String:
		v.SetString(sentinel)
	case reflect.Pointer:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		fillText(v.Elem(), sentinel, depth+1)
	case reflect.Struct:
		for i := range v.NumField() {
			if !v.Type().Field(i).IsExported() {
				continue
			}
			if tag, _, _ := strings.Cut(v.Type().Field(i).Tag.Get("yaml"), ","); tag == "-" {
				continue
			}
			fillText(v.Field(i), sentinel, depth+1)
		}
	case reflect.Slice:
		one := reflect.MakeSlice(v.Type(), 1, 1)
		fillText(one.Index(0), sentinel, depth+1)
		v.Set(one)
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return
		}
		m := reflect.MakeMap(v.Type())
		val := reflect.New(v.Type().Elem()).Elem()
		fillText(val, sentinel, depth+1)
		m.SetMapIndex(reflect.ValueOf(sentinel).Convert(v.Type().Key()), val)
		v.Set(m)
	default:
		// Numbers, booleans and `any` payloads carry no interpolatable text.
	}
}

// compareText walks the filled value and the walked copy together and reports
// each string leaf as visited or not, by yaml path.
func compareText(filled, walked reflect.Value, path string, report func(string, bool)) {
	if !filled.IsValid() || !walked.IsValid() {
		return
	}
	switch filled.Kind() {
	case reflect.String:
		report(path, walked.Kind() == reflect.String && walked.String() != filled.String())
	case reflect.Pointer:
		if filled.IsNil() || walked.IsNil() {
			return
		}
		compareText(filled.Elem(), walked.Elem(), path, report)
	case reflect.Struct:
		for i := range filled.NumField() {
			f := filled.Type().Field(i)
			if !f.IsExported() {
				continue
			}
			tag, _, _ := strings.Cut(f.Tag.Get("yaml"), ",")
			if tag == "-" {
				continue
			}
			sub := path
			// An embedded or inline struct (ScreenAssert's StreamAssert) shares
			// its parent's path, exactly as its keys share the parent's mapping.
			if tag != "" && !f.Anonymous {
				sub = joinPath(path, tag)
			}
			compareText(filled.Field(i), walked.Field(i), sub, report)
		}
	case reflect.Slice:
		if filled.Len() == 0 || walked.Len() != filled.Len() {
			return
		}
		suffix := "[]"
		if filled.Type().Elem().Kind() == reflect.String {
			suffix = ""
		}
		compareText(filled.Index(0), walked.Index(0), path+suffix, report)
	case reflect.Map:
		for _, k := range filled.MapKeys() {
			compareText(filled.MapIndex(k), walked.MapIndex(k), path+".*", report)
		}
	default:
	}
}

// joinPath appends a yaml key to a dotted path.
func joinPath(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}
