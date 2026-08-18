// Package loader reads a atago YAML file and turns it into a validated
// *spec.Spec. Validation happens in layers: YAML parse, then
// schema/semantic checks. Errors carry the file path and, for parse failures,
// the line/column reported by goccy/go-yaml.
package loader

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/nao1215/atago/internal/diag"
	"github.com/nao1215/atago/internal/spec"
)

// Kind classifies why loading failed, so callers can map it to an exit code
// (exit 2 = spec parse error).
type Kind int

const (
	// KindParse is a YAML syntax or decode error.
	KindParse Kind = iota
	// KindValidation is a schema or semantic error in an otherwise-parseable file.
	KindValidation
)

// Error is a loader failure annotated with the source path and kind.
//
// Code names the diagnostic. A validation failure reports several problems at
// once, each carrying its own code inside Msg, so Code is set only when the
// whole failure is one diagnostic — which is every parse error, since parsing
// stops at the first one.
type Error struct {
	Path string
	Kind Kind
	Code diag.Code
	Msg  string
}

func (e *Error) Error() string {
	msg := e.Msg
	if e.Code != 0 {
		msg = e.Code.Annotate(msg)
	}
	if e.Path == "" {
		return msg
	}
	return fmt.Sprintf("%s: %s", e.Path, msg)
}

// Load reads and validates the spec file at path.
func Load(path string) (*spec.Spec, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path comes from user-specified spec args
	if err != nil {
		return nil, &Error{Path: path, Kind: KindValidation, Code: diag.SpecUnreadable, Msg: err.Error()}
	}
	// A directory-level manifest (#392) is configuration for a TREE of specs, so
	// it is discovered from the spec's own location rather than passed in: every
	// command that reads a spec — run, explain, doc, list, manifest — has to
	// resolve the same configuration, or the file means different things
	// depending on which subcommand opened it.
	proj, perr := FindProject(filepath.Dir(path))
	if perr != nil {
		return nil, perr
	}
	return loadBytesWithProject(path, data, proj)
}

// binaryTag is the one explicit YAML tag a spec may carry. `atago record`
// writes it when a captured stream is not valid UTF-8, so recorded specs must
// keep loading; every other tag is an authoring mistake.
const binaryTag = "!!binary"

// explicitTagError describes the first unsupported explicit YAML tag in a spec
// document, or returns "" when the document has none. atago's schema is closed
// and fully typed — every field's Go type already fixes how its value is read —
// so a tag can only ever restate or contradict the schema, and no spec,
// example, or doc in this repository authors one.
//
// Turning tags away up front is also what keeps the decoder's nil-panic path
// out of everyday use. When a tagged node lands on a list field,
// goccy/go-yaml v1.19.2 panics: ast.TagNode.ArrayRange returns a nil
// *ArrayNodeIter whenever the tag's value is not a sequence, and decodeSlice
// dereferences it without a nil check (decode.go:1593 -> ast.go:1543).
// decodeSpec recovers from that, but internal/loader is then the only package
// here that raises a real runtime nil dereference on every test run, and it is
// also the only package that has crashed Windows CI with Go GC heap-corruption
// fatal errors ("found pointer to free object"). Not provoking the panic costs
// nothing and removes that correlation.
//
// A `!!binary` over a list field still reaches the panic, because record's
// output has to stay loadable; decodeSpec's recover remains the backstop for
// that residue.
//
// A parse failure yields no usable AST; that case reports "" so the decoder can
// produce its own, better-positioned syntax error.
func explicitTagError(data []byte) string {
	f, err := parser.ParseBytes(data, 0)
	if err != nil || f == nil || len(f.Docs) == 0 {
		return ""
	}
	// Decode reads a single document, so only that document can reach the panic.
	v := &tagFinder{}
	ast.Walk(v, f.Docs[0])
	if v.found == nil || v.found.Start == nil {
		return ""
	}
	pos := v.found.Start.Position
	name := strings.TrimSpace(v.found.Start.Value)
	where := ""
	if pos != nil {
		where = fmt.Sprintf("[%d:%d] ", pos.Line, pos.Column)
	}
	if name == "" || name == "!" {
		return where + "explicit YAML tag is not supported in a spec: remove the tag"
	}
	return fmt.Sprintf("%sexplicit YAML tag %q is not supported in a spec: remove the tag", where, name)
}

// tagFinder is an ast.Visitor that stops at the first unsupported explicit tag
// it meets.
type tagFinder struct {
	found *ast.TagNode
}

// Visit records the first unsupported *ast.TagNode and then stops descending.
// A `!!binary` scalar is skipped over rather than accepted wholesale: the walk
// continues into its children so a tag nested under it is still caught. Walk
// also calls Visit(nil) after a node's children, which the type assertion
// ignores.
func (v *tagFinder) Visit(n ast.Node) ast.Visitor {
	if v.found != nil {
		return nil
	}
	tag, ok := n.(*ast.TagNode)
	if !ok {
		return v
	}
	if tag.Start != nil && strings.TrimSpace(tag.Start.Value) == binaryTag {
		return v
	}
	v.found = tag
	return nil
}

// decodeSpec decodes one YAML document into s, converting a panic from the
// third-party decoder into an ordinary error. goccy/go-yaml can nil-panic on
// some malformed input (a bare `!` tag over a broken mapping, found by
// FuzzLoadBytes); atago's contract is that loading untrusted spec bytes never
// crashes the process, so recover here and let the caller report a parse error.
// explicitTagError already turns away the inputs known to reach that panic, so
// this stays as a backstop for any shape not yet found.
func decodeSpec(dec *yaml.Decoder, s *spec.Spec) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// Report a spec problem, not the raw runtime panic: "nil pointer
			// dereference" reads like an atago crash, but the cause is malformed
			// input the third-party decoder mishandled.
			err = fmt.Errorf("malformed YAML: the document could not be parsed")
		}
	}()
	return dec.Decode(s)
}

// LoadBytes parses and validates spec bytes, labeling errors with path.
func LoadBytes(path string, data []byte) (*spec.Spec, error) {
	return loadBytesWithProject(path, data, nil)
}

// loadBytesWithProject is LoadBytes with a directory manifest layered beneath
// the spec's own values (#392).
func loadBytesWithProject(path string, data []byte, proj *Project) (*spec.Spec, error) {
	// Strip a leading UTF-8 byte-order mark. Windows/Notepad-family editors emit
	// one routinely, and goccy/go-yaml does not skip it: it glues the BOM onto the
	// first key, so a correctly-authored spec failed with a confusing "unknown
	// field" error naming a field the author wrote right. Most YAML tooling strips
	// a single leading BOM transparently.
	data = bytes.TrimPrefix(data, []byte("\ufeff"))
	if msg := explicitTagError(data); msg != "" {
		return nil, &Error{Path: path, Kind: KindParse, Code: diag.YAMLTag, Msg: msg}
	}
	var s spec.Spec
	dec := yaml.NewDecoder(bytes.NewReader(data), yaml.Strict())
	if err := decodeSpec(dec, &s); err != nil {
		// An empty document (empty file, whitespace, or comments only) decodes to
		// io.EOF, whose bare "EOF" tells the user nothing. Name the problem and
		// what a spec needs instead.
		if errors.Is(err, io.EOF) {
			return nil, &Error{Path: path, Kind: KindParse, Code: diag.SpecEmpty, Msg: "spec is empty: expected a YAML document with version, suite, and scenarios"}
		}
		msg := formatYAMLError(err)
		return nil, &Error{Path: path, Kind: KindParse, Code: classifyYAMLError(data, msg), Msg: msg}
	}
	// Restore the source text of any plain scalar that landed in a text field
	// before anything reads those values: YAML types an unquoted 1.20 as a float,
	// and a spec that compares literals must mean the characters the author
	// wrote. It runs while the model still mirrors the document one-to-one,
	// which is what lets the two be walked together.
	applyVerbatimScalars(&s, data)
	// Record each scenario's authored index before matrix expansion, so every
	// expanded instance can be traced back to its authored source location (#80).
	for i := range s.Scenarios {
		s.Scenarios[i].SourceIndex = i
	}
	// Validate matrix shape on the raw spec, then expand each matrix scenario into
	// concrete instances so the remaining validation and the engine only ever see
	// plain scenarios.
	if errs := validateMatrix(&s); len(errs) > 0 {
		return nil, &Error{Path: path, Kind: KindValidation, Msg: joinErrors(errs)}
	}
	expandMatrix(&s)
	// A directory manifest is the weakest layer, applied before the file's own
	// defaults are expanded so both go through one merge path (#392).
	applyProject(&s, proj)
	// Expand the top-level defaults into the concrete scenario/step/service model
	// so validation and the engine only ever see fully-resolved scenarios.
	applyDefaults(&s)
	if errs := validate(&s); len(errs) > 0 {
		return nil, &Error{Path: path, Kind: KindValidation, Msg: joinErrors(errs)}
	}
	return &s, nil
}

// formatYAMLError renders goccy errors with position context when available,
// and appends a did-you-mean hint for misspelled field names.
func formatYAMLError(err error) string {
	var yerr yaml.Error
	if errors.As(err, &yerr) {
		return suggestCollectionShape(suggestScalarMatcher(suggestUnknownField(yaml.FormatError(err, false, true))))
	}
	return suggestCollectionShape(suggestScalarMatcher(suggestUnknownField(err.Error())))
}

// classifyYAMLError picks the diagnostic for a decode failure. goccy reports
// all of them through one error type, but they are three different mistakes
// with three different fixes: a document that is not YAML at all, a key the
// schema does not define, and a value written in a shape its key cannot take.
//
// The first is separated structurally rather than by matching the message: if
// the parser cannot build a document from the bytes, the problem is syntax; if
// it can, the YAML was fine and the spec model is what rejected it. That keeps
// the classification correct for the schema errors raised by the spec package's
// own unmarshalers, whose wording this file has no business knowing.
func classifyYAMLError(data []byte, msg string) diag.Code {
	if _, err := parser.ParseBytes(data, 0); err != nil {
		return diag.YAMLSyntax
	}
	if unknownFieldRe.MatchString(msg) {
		return diag.UnknownKey
	}
	return diag.WrongValueShape
}

func joinErrors(errs []string) string {
	if len(errs) == 1 {
		return errs[0]
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d validation errors:", len(errs))
	for _, e := range errs {
		b.WriteString("\n  - ")
		b.WriteString(e)
	}
	return b.String()
}
