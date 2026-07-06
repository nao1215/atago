package loader

import (
	"regexp"

	"github.com/nao1215/atago/internal/spec"
	"github.com/ohler55/ojg/jp"
)

// reservedVarNames are the built-in variables the engine seeds into every
// scenario store (${atago} binary path, ${workdir}, ${suitedir}). A user store
// or matrix variable that reuses one silently shadows it — pointing ${workdir}
// outside the isolated temp dir while cleanup still targets the real one — so
// the collision is rejected at load time.
var reservedVarNames = map[string]bool{"atago": true, "workdir": true, "suitedir": true}

func validateStore(add func(string, ...any), where string, s *spec.Store) {
	if s.Name == "" {
		add("%s.store.name is required", where)
	}
	if reservedVarNames[s.Name] {
		add("%s.store.name %q shadows a built-in variable (atago/workdir/suitedir); choose another name", where, s.Name)
	}
	if s.From == nil {
		add("%s.store.from is required", where)
		return
	}
	n := 0
	if s.From.Stdout != nil {
		n++
	}
	if s.From.Body != nil {
		n++
	}
	if s.From.File != nil {
		n++
	}
	if s.From.Header != "" {
		n++
	}
	if s.From.Rows != nil {
		n++
	}
	if s.From.Message != nil {
		n++
	}
	if s.From.Value != nil {
		n++
	}
	switch n {
	case 0:
		add("%s.store.from must set one of stdout/body/file/header/rows/message/value", where)
	case 1:
	default:
		add("%s.store.from must set exactly one source", where)
	}

	// A store selector extracts a value via a json path or a matches regexp
	// (unlike a full assert). Validate the regexp/path at load time so a typo
	// fails with a positioned message instead of aborting mid-run, matching how
	// assert streams validate their regexp/path.
	for _, sel := range []struct {
		name string
		s    *spec.StreamAssert
	}{
		{"stdout", s.From.Stdout},
		{"body", s.From.Body},
		{"rows", s.From.Rows},
		{"message", s.From.Message},
		{"value", s.From.Value},
	} {
		if sel.s != nil {
			validateStoreSelector(add, where+".store.from."+sel.name, sel.s)
		}
	}
	if s.From.File != nil {
		validateStoreFileSelector(add, where+".store.from.file", s.From.File)
	}
}

// validateStoreFileSelector checks a store.from.file selector: exactly one of a
// json path (extract a value) or text: true (capture the whole file verbatim,
// #158).
func validateStoreFileSelector(add func(string, ...any), where string, f *spec.FileAssert) {
	n := 0
	if len(f.JSON) > 0 {
		n++
		validateStoreJSONSelector(add, where+".json", f.JSON)
	}
	if f.Text != nil {
		n++
	}
	switch n {
	case 0:
		add("%s must set a json path or text: true to capture a value", where)
	case 1:
	default:
		add("%s must set exactly one selector (json or text)", where)
	}
}

// validateStoreJSONSelector checks the json selector of a store source: it
// captures exactly one value, so a list of checks (#156) is rejected — a store
// needs a single JSONPath, not several assertions.
func validateStoreJSONSelector(add func(string, ...any), where string, list spec.JSONChecks) {
	if len(list) > 1 {
		add("%s must select a single value with one json path, not a list of checks", where)
		return
	}
	validateStoreJSONPath(add, where, list[0].Path)
}

// validateStoreSelector checks a store.from stream selector: it must carry
// exactly one of a json path, a matches regexp, or trim (capture the whole
// stream, #158) — the selectors the extractor understands — and whichever is
// present must be well-formed.
func validateStoreSelector(add func(string, ...any), where string, s *spec.StreamAssert) {
	n := 0
	if len(s.JSON) > 0 {
		n++
		validateStoreJSONSelector(add, where+".json", s.JSON)
	}
	if s.Matches != nil {
		n++
		if _, err := regexp.Compile(*s.Matches); err != nil {
			add("%s.matches %q is not a valid regexp: %v", where, *s.Matches, err)
		}
	}
	if s.Trim != nil {
		n++
	}
	switch n {
	case 0:
		add("%s must set a json path, a matches regexp, or trim to extract a value", where)
	case 1:
	default:
		add("%s must set exactly one selector (json, matches, or trim)", where)
	}
}

// validateStoreJSONPath compile-checks a store selector's JSON path.
func validateStoreJSONPath(add func(string, ...any), where, path string) {
	if path == "" {
		add("%s.path is required", where)
		return
	}
	if _, err := jp.ParseString(path); err != nil {
		add("%s.path %q is not a valid JSON path: %v", where, path, err)
	}
}
