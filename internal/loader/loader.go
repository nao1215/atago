// Package loader reads a atago YAML file and turns it into a validated
// *spec.Spec. Validation happens in layers: YAML parse, then
// schema/semantic checks. Errors carry the file path and, for parse failures,
// the line/column the YAML reader reports.
package loader

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nao1215/atago/internal/diag"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/yaml"
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
	s, _, err := loadBytesWithProject(path, data, proj)
	return s, err
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
func explicitTagError(doc *yaml.Node) string {
	found := firstTag(doc, map[*yaml.Node]bool{})
	if found == nil {
		return ""
	}
	where := fmt.Sprintf("[%d:%d] ", found.Line, found.Col)
	if found.Tag == "!" {
		return where + "explicit YAML tag is not supported in a spec: remove the tag"
	}
	return fmt.Sprintf("%sexplicit YAML tag %q is not supported in a spec: remove the tag", where, found.Tag)
}

// firstTag returns the first node in document order that carries a tag other
// than !!binary. seen keeps an aliased node from being walked twice.
func firstTag(n *yaml.Node, seen map[*yaml.Node]bool) *yaml.Node {
	if n == nil || seen[n] {
		return nil
	}
	seen[n] = true
	if n.Tag != "" && n.Tag != binaryTag {
		return n
	}
	for _, item := range n.Items {
		if t := firstTag(item, seen); t != nil {
			return t
		}
	}
	for _, pair := range n.Pairs {
		if t := firstTag(pair.Key, seen); t != nil {
			return t
		}
		if t := firstTag(pair.Value, seen); t != nil {
			return t
		}
	}
	return nil
}

// LoadBytes parses and validates spec bytes, labeling errors with path.
func LoadBytes(path string, data []byte) (*spec.Spec, error) {
	s, _, err := loadBytesWithProject(path, data, nil)
	return s, err
}

// loadBytesWithProject is LoadBytes with a directory manifest layered beneath
// the spec's own values (#392). It also returns the document the spec was
// decoded from, which the source locator answers from.
func loadBytesWithProject(path string, data []byte, proj *Project) (*spec.Spec, *yaml.Node, error) {
	f, err := yaml.Parse(data)
	if err != nil {
		return nil, nil, &Error{Path: path, Kind: KindParse, Code: diag.YAMLSyntax, Msg: formatYAMLError(err, sourceOf(data))}
	}
	doc := firstDocument(f)
	if doc == nil {
		return nil, nil, &Error{Path: path, Kind: KindParse, Code: diag.SpecEmpty, Msg: "spec is empty: expected a YAML document with version, suite, and scenarios"}
	}
	if msg := explicitTagError(doc); msg != "" {
		return nil, nil, &Error{Path: path, Kind: KindParse, Code: diag.YAMLTag, Msg: msg}
	}
	var s spec.Spec
	// A plain scalar decodes into a text field as the text written, so
	// `contains: 1.20` asserts the four characters 1.20 and `env: {V: 007}`
	// exports 007; only a field typed as a number or `any` reads what the
	// scalar spells.
	if err := yaml.Decode(doc, &s, true); err != nil {
		return nil, nil, &Error{Path: path, Kind: KindParse, Code: classifyYAMLError(err), Msg: formatYAMLError(err, f.Source())}
	}
	// Record each scenario's authored index before matrix expansion, so every
	// expanded instance can be traced back to its authored source location (#80).
	for i := range s.Scenarios {
		s.Scenarios[i].SourceIndex = i
	}
	// A forall block is a matrix whose rows are generated (#656): validate the
	// generators while they are still there to name, then draw the rows, so
	// everything below this point — matrix validation included — works on the
	// concrete inputs the scenario will actually run.
	if errs := validateForall(&s); len(errs) > 0 {
		return nil, nil, &Error{Path: path, Kind: KindValidation, Msg: joinErrors(errs)}
	}
	expandForall(&s)
	// Validate matrix shape on the raw spec, then expand each matrix scenario into
	// concrete instances so the remaining validation and the engine only ever see
	// plain scenarios.
	if errs := validateMatrix(&s); len(errs) > 0 {
		return nil, nil, &Error{Path: path, Kind: KindValidation, Msg: joinErrors(errs)}
	}
	expandMatrix(&s)
	// A directory manifest is the weakest layer, applied before the file's own
	// defaults are expanded so both go through one merge path (#392).
	applyProject(&s, proj)
	// Expand the top-level defaults into the concrete scenario/step/service model
	// so validation and the engine only ever see fully-resolved scenarios.
	applyDefaults(&s)
	if errs := validate(&s); len(errs) > 0 {
		return nil, nil, &Error{Path: path, Kind: KindValidation, Msg: joinErrors(errs)}
	}
	return &s, doc, nil
}

// firstDocument returns the root of the file's first document. A spec or a
// manifest is that document: an empty one (comments or a bare `---`) is empty,
// even if a later document has content.
func firstDocument(f *yaml.File) *yaml.Node {
	if len(f.Docs) == 0 {
		return nil
	}
	return f.Docs[0].Root
}

// sourceOf returns data as the reader sees it, which is what an error's line
// and column count in: without a byte-order mark, with CRLF read as LF.
func sourceOf(data []byte) []byte {
	data = bytes.TrimPrefix(data, []byte("\xEF\xBB\xBF"))
	if bytes.IndexByte(data, '\r') >= 0 {
		data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
		data = bytes.ReplaceAll(data, []byte("\r"), []byte("\n"))
	}
	return data
}

// formatYAMLError renders a reader error with the excerpt of src it points at,
// and appends a hint for the mistakes that have a known fix.
func formatYAMLError(err error, src []byte) string {
	msg := yaml.FormatError(err, src)
	var ye *yaml.Error
	if !errors.As(err, &ye) {
		return msg
	}
	if ye.Key != "" {
		return suggestUnknownField(msg, ye.Key)
	}
	if ye.Want != "" {
		return suggestShape(msg, ye)
	}
	return msg
}

// classifyYAMLError picks the diagnostic for a decode failure of a document
// that parsed: a key the schema does not define, or a value written in a shape
// its key cannot take. A document that is not YAML at all never gets here; the
// loader reports it as a syntax error when the reader cannot parse it.
func classifyYAMLError(err error) diag.Code {
	var ye *yaml.Error
	if errors.As(err, &ye) && ye.Key != "" {
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
