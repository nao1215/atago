package spec

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
)

// Assert checks externally observable behavior. Exactly one target family is set.
type Assert struct {
	ExitCode *ExitCode     `yaml:"exit_code,omitempty"`
	Stdout   *StreamAssert `yaml:"stdout,omitempty"`
	Stderr   *StreamAssert `yaml:"stderr,omitempty"`
	File     *FileAssert   `yaml:"file,omitempty"`

	// HTTP assertion targets.
	Status *int          `yaml:"status,omitempty"`
	Header *HeaderMatch  `yaml:"header,omitempty"`
	Body   *StreamAssert `yaml:"body,omitempty"`

	// Rows is the db assertion target: the query result rows as a JSON array,
	// matched with the stream matchers (json path/length, contains, …).
	Rows *StreamAssert `yaml:"rows,omitempty"`

	// gRPC assertion targets: GRPCStatus checks the numeric status
	// code; Message matches the response message (as JSON) with the stream matchers.
	GRPCStatus *int          `yaml:"grpc_status,omitempty"`
	Message    *StreamAssert `yaml:"message,omitempty"`

	// Value is the browser assertion target: the value captured by the
	// last text/eval action, matched with the stream matchers.
	Value *StreamAssert `yaml:"value,omitempty"`

	// Image is the image assertion target: it inspects a generated
	// image file's decoded properties (format, dimensions, alpha) and can compare
	// its pixels against a baseline image.
	Image *ImageAssert `yaml:"image,omitempty"`

	// Dir is the directory/tree assertion target (#74): black-box checks over a
	// generated directory — existence, expected/forbidden children, entry counts,
	// and glob coverage — for multi-file generators (static sites, scaffolds,
	// extracted archives). It is deliberately declarative and non-recursive.
	Dir *DirAssert `yaml:"dir,omitempty"`

	// PDF is the PDF assertion target (#73): a small, black-box, content-oriented
	// surface for generated PDFs — page count, Info metadata fields, and extracted
	// text — without depending on a specific layout engine.
	PDF *PDFAssert `yaml:"pdf,omitempty"`

	// Mock is the mock-server assertion target (#24): what the CLI under test
	// actually sent to a declared mock server — request count, and header/body
	// matchers applied to the last matching recorded request.
	Mock *MockAssert `yaml:"mock,omitempty"`

	// Screen is the rendered-terminal assertion target (#27), valid after a
	// pty step: the transcript replayed through a vt10x emulator sized by the
	// step's rows/cols, asserted as plain text with the stream matchers
	// (line.n addresses screen rows 1-based). The raw transcript stays on
	// stdout.
	Screen *StreamAssert `yaml:"screen,omitempty"`

	// Duration is the wall-clock assertion target (#31), valid after a
	// measurable step (run/http/query/grpc/pty): it bounds how long that step
	// took with lt/lte/gt/gte Go-duration bounds.
	Duration *DurationAssert `yaml:"duration,omitempty"`

	// Changes is the workdir-delta assertion target (#70), valid after an
	// immediately preceding run/pty step: it pins exactly which files that step
	// created, modified, and deleted in the scenario workdir.
	Changes *ChangesAssert `yaml:"changes,omitempty"`
}

// ChangesAssert pins the exact delta the immediately preceding run/pty step
// made to the scenario workdir (#70). The delta is content-based (hash, not
// mtime). Each set field is EXHAUSTIVE in both directions: every observed path
// must be matched by an entry (an exact workdir-relative path or a /-glob) and
// every entry must match at least one observed path — so `modified: []` asserts
// "modified nothing". An omitted (nil) field is unconstrained.
//
// Entries are doublestar globs, always /-separated (#76): a single `*` matches
// within one path segment, while `**` matches across `/` at any depth
// (`site/**` covers the whole tree under site/, `dist/**/*.css` composes with a
// suffix). A backslash escapes a literal metacharacter — `a\[1\].txt` matches
// the file `a[1].txt`; because entries are always /-separated the escape is
// portable across operating systems.
type ChangesAssert struct {
	Created  *StringList `yaml:"created,omitempty"`
	Modified *StringList `yaml:"modified,omitempty"`
	Deleted  *StringList `yaml:"deleted,omitempty"`
}

// DurationAssert bounds a step's measured wall-clock time (#31). At least one
// bound must be set; lt/lte are mutually exclusive, as are gt/gte, and any
// pair must form a non-empty interval (validated at load time). Values are Go
// duration strings ("2s", "100ms").
type DurationAssert struct {
	// LT / LTE are the upper bound (exclusive / inclusive).
	LT  string `yaml:"lt,omitempty"`
	LTE string `yaml:"lte,omitempty"`
	// GT / GTE are the lower bound (exclusive / inclusive).
	GT  string `yaml:"gt,omitempty"`
	GTE string `yaml:"gte,omitempty"`
}

// DescribeDuration renders a duration assert's bounds as a human phrase (#31),
// shared by explain and doc so the two never drift.
func (d *DurationAssert) DescribeDuration() string {
	var parts []string
	if d.LT != "" {
		parts = append(parts, "in under "+d.LT)
	}
	if d.LTE != "" {
		parts = append(parts, "in at most "+d.LTE)
	}
	if d.GT != "" {
		parts = append(parts, "in over "+d.GT)
	}
	if d.GTE != "" {
		parts = append(parts, "in at least "+d.GTE)
	}
	return strings.Join(parts, " and ")
}

// PDFAssert checks a generated PDF file (#73). Like ImageAssert/DirAssert, every
// set field is an independent constraint and all must hold; at least one (besides
// Path) must be set. The surface is intentionally small and content-oriented:
// page count, Info dictionary metadata, and extracted text — not layout.
type PDFAssert struct {
	// Path is the PDF under test, resolved against the scenario workdir when
	// relative (like FileAssert.Path).
	Path string `yaml:"path"`
	// Pages asserts the exact page count; MinPages/MaxPages assert bounds.
	Pages    *int `yaml:"pages,omitempty"`
	MinPages *int `yaml:"min_pages,omitempty"`
	MaxPages *int `yaml:"max_pages,omitempty"`
	// Metadata maps an Info-dictionary field (title, author, subject, keywords,
	// creator, producer) to a substring the field's value must contain. Keys are
	// matched case-insensitively.
	Metadata map[string]string `yaml:"metadata,omitempty"`
	// Text applies the standard stream matchers (contains/matches/equals/…) to the
	// text extracted from the PDF's content streams (raw and FlateDecode-decoded).
	Text *StreamAssert `yaml:"text,omitempty"`
}

// DirAssert checks a generated directory tree (#74). Like ImageAssert, every
// field that is set is a separate constraint and all of them must hold, because
// a directory has several independent observable properties. At least one
// constraint (besides Path) must be set. Child paths are confined to the
// directory and may not escape it.
type DirAssert struct {
	// Path is the directory under test, resolved against the scenario workdir when
	// relative (like FileAssert.Path).
	Path string `yaml:"path"`
	// Exists asserts the path exists and is a directory (exists:false asserts it
	// is absent).
	Exists *bool `yaml:"exists,omitempty"`
	// Contains lists child paths (relative to Path) that must exist. A child may
	// name a nested path (e.g. "assets/app.css"); it must stay within Path.
	Contains []string `yaml:"contains,omitempty"`
	// NotContains lists child paths (relative to Path) that must NOT exist.
	NotContains []string `yaml:"not_contains,omitempty"`
	// Count asserts the exact number of direct entries in the directory.
	Count *int `yaml:"count,omitempty"`
	// MinCount / MaxCount assert bounds on the number of direct entries.
	MinCount *int `yaml:"min_count,omitempty"`
	MaxCount *int `yaml:"max_count,omitempty"`
	// Glob asserts that at least one direct entry matches this shell glob pattern
	// (filepath.Match semantics, e.g. "*.html").
	Glob string `yaml:"glob,omitempty"`
	// Recursive makes Contains/NotContains accept slash-separated relative
	// paths anywhere in the tree, and Count/MinCount/MaxCount/Glob apply to
	// the whole walk (counts see FILES only; Glob matches each entry's
	// relative path, or its basename for patterns without "/") (#25).
	Recursive bool `yaml:"recursive,omitempty"`
	// Snapshot compares the whole tree against a golden manifest (#25):
	// sorted /-separated relative paths, one line per entry — `dir <path>`,
	// `file <path> sha256:<hash>` (hashed byte-exact: CRLF is NOT normalized
	// inside file content), or `link <path> -> <target>` (not traversed).
	// No mode/mtime (not portable). Refresh with --update-snapshots.
	Snapshot string `yaml:"snapshot,omitempty"`
	// Ignore lists glob patterns excluded from the recursive walk and the
	// snapshot manifest ("*.log", ".git/**"). A pattern without "/" also
	// matches basenames at any depth; a "<dir>/**" pattern prunes the whole
	// subtree.
	Ignore []string `yaml:"ignore,omitempty"`
}

// ImageAssert checks a generated image file. Unlike the one-of
// stream/file targets, every field that is set is a separate constraint and all
// of them must hold, because an image has several independent observable
// attributes. At least one constraint must be set.
type ImageAssert struct {
	// Path is the image file under test, resolved against the scenario workdir
	// when relative (like FileAssert.Path).
	Path string `yaml:"path"`
	// Format asserts the encoded image format, detected from the file's content:
	// png, jpeg, gif, webp, bmp, tiff, avif, or svg.
	Format string `yaml:"format,omitempty"`
	// Width / Height assert the exact pixel dimensions.
	Width  *int `yaml:"width,omitempty"`
	Height *int `yaml:"height,omitempty"`
	// MinWidth / MaxWidth / MinHeight / MaxHeight assert dimension bounds.
	MinWidth  *int `yaml:"min_width,omitempty"`
	MaxWidth  *int `yaml:"max_width,omitempty"`
	MinHeight *int `yaml:"min_height,omitempty"`
	MaxHeight *int `yaml:"max_height,omitempty"`
	// Alpha asserts whether the image actually carries transparency (any
	// non-opaque pixel). It scans decoded pixels rather than the in-memory color
	// model, so an opaque truecolor PNG/BMP correctly reports alpha=false.
	Alpha *bool `yaml:"alpha,omitempty"`
	// SimilarTo compares the decoded pixels against a baseline image. A relative
	// path resolves against the spec file's directory (like a committed
	// snapshot); use an absolute or ${workdir}-prefixed path to compare against
	// another generated file. Both images must share dimensions.
	SimilarTo string `yaml:"similar_to,omitempty"`
	// MaxDiff is the maximum allowed normalized mean per-pixel difference (0..1)
	// for SimilarTo. It defaults to 0 (an exact pixel match); lossy formats need a
	// small tolerance such as 0.02.
	MaxDiff *float64 `yaml:"max_diff,omitempty"`
}

// ExitCode accepts a bare integer, {not: <int>}, or {in: [<int>, ...]} (#19).
// The `in` set is the contract shape real CLIs document (grep's 0/1,
// terraform plan -detailed-exitcode's 0/2): membership in an accepted set.
type ExitCode struct {
	Equals *int
	Not    *int
	In     []int
}

// UnmarshalYAML decodes exit_code as a scalar int, a {not: int} map, or an
// {in: [int, ...]} map. Anything else gets a purpose-built error: the generic
// decoder message ("string was used where mapping is expected", positioned at
// the sub-node) reads like an internal failure, and a spec author needs to
// know the accepted shapes.
func (e *ExitCode) UnmarshalYAML(b []byte) error {
	if n, err := strconv.Atoi(trimYAMLScalar(string(b))); err == nil {
		e.Equals = &n
		return nil
	}
	// A YAML-quoted integer (exit_code: "0" / '2') is still an integer to the
	// author, but the raw-bytes Atoi above sees the surrounding quotes and fails.
	// A plain int decode unquotes it, so a quoted scalar is accepted instead of
	// falling through to the misleading "must be an integer … got \"0\"" error.
	// Mapping/sequence forms ({not:…}, {in:[…]}) fail this decode and fall
	// through to the shape-specific handling below.
	var n int
	if err := yaml.Unmarshal(b, &n); err == nil {
		e.Equals = &n
		return nil
	}
	var m struct {
		Not *int  `yaml:"not"`
		In  []int `yaml:"in"`
	}
	if err := yaml.Unmarshal(b, &m); err != nil {
		return fmt.Errorf("exit_code must be an integer (exit_code: 0), a negation (exit_code: {not: 0}), or a set (exit_code: {in: [0, 2]}), got %q", strings.TrimSpace(string(b)))
	}
	e.Not = m.Not
	e.In = m.In
	return nil
}

// MarshalYAML emits the same shape UnmarshalYAML accepts — a bare integer, a
// {not: <int>} map, or an {in: [<int>, ...]} map — so a loaded exit_code
// round-trips. The default struct marshal writes all three fields at once
// (equals/not/in), and the unmarshaler's inner mapping decode then silently
// drops `equals` and rejects the empty `in`, losing the assertion.
func (e ExitCode) MarshalYAML() (any, error) {
	switch {
	case e.Equals != nil:
		return *e.Equals, nil
	case e.Not != nil:
		return map[string]int{"not": *e.Not}, nil
	case len(e.In) > 0:
		return map[string][]int{"in": e.In}, nil
	default:
		return nil, nil
	}
}

// StreamAssert matches a captured text stream (stdout/stderr/body). One matcher.
//
// Line is an optional 1-based selector: when set, the matcher is
// applied to that single line of the stream instead of the whole stream. It is
// not itself a matcher, so exactly one of empty/contains/matches/equals must
// still be set. Line does not compose with json/snapshot (those operate on the
// whole document).
type StreamAssert struct {
	Line        *int        `yaml:"line,omitempty"`
	Empty       *bool       `yaml:"empty,omitempty"`
	Contains    StringList  `yaml:"contains,omitempty"`
	NotContains StringList  `yaml:"not_contains,omitempty"`
	Matches     *string     `yaml:"matches,omitempty"`
	NotMatches  *string     `yaml:"not_matches,omitempty"`
	Equals      *string     `yaml:"equals,omitempty"`
	NotEquals   *string     `yaml:"not_equals,omitempty"`
	JSON        *JSONAssert `yaml:"json,omitempty"`
	YAML        *JSONAssert `yaml:"yaml,omitempty"`
	Snapshot    string      `yaml:"snapshot,omitempty"`
}

// StringList is a matcher argument that accepts either a single YAML scalar
// string or a sequence of strings. It backs the `contains` / `not_contains`
// matchers on stream and file assertions so one matcher can require (or forbid)
// several substrings without repeating the assert block. A scalar decodes to a
// one-element list and keeps byte-identical behavior with the pre-list format;
// a sequence decodes to its elements. `contains` requires every element to be
// present, `not_contains` requires every element to be absent, and either way
// the whole list counts as a single matcher (the one-of matcher rule is
// unchanged).
type StringList []string

// UnmarshalYAML accepts a scalar string or a sequence of strings. It uses the
// interface-based decoder (not the raw-bytes form) so escapes like "\x1b" are
// resolved by goccy's parser once, rather than re-tokenized from node bytes.
func (l *StringList) UnmarshalYAML(unmarshal func(any) error) error {
	var one string
	if err := unmarshal(&one); err == nil {
		*l = StringList{one}
		return nil
	}
	var many []string
	if err := unmarshal(&many); err != nil {
		return err
	}
	*l = StringList(many)
	return nil
}

// FileAssert checks a generated file.
type FileAssert struct {
	Path        string      `yaml:"path"`
	Exists      *bool       `yaml:"exists,omitempty"`
	Contains    StringList  `yaml:"contains,omitempty"`
	NotContains StringList  `yaml:"not_contains,omitempty"`
	Executable  *bool       `yaml:"executable,omitempty"`
	JSON        *JSONAssert `yaml:"json,omitempty"`
	Snapshot    string      `yaml:"snapshot,omitempty"`
}

// JSONAssert matches a value selected by a JSONPath. One matcher.
//
// Gt/Gte/Lt/Lte assert a numeric bound on the selected value (which must be a
// number, or a numeric string). They exist because tools routinely emit
// non-deterministic-but-bounded metrics — a processed-record count, a coverage
// figure, a duration — where an exact `equals` is impossible but "at least N"
// or "below N" is exactly the contract worth pinning (surfaced dogfooding runn's
// coverage/loadt metrics).
type JSONAssert struct {
	Path    string   `yaml:"path"`
	Equals  any      `yaml:"equals,omitempty"`
	Matches *string  `yaml:"matches,omitempty"`
	Length  *int     `yaml:"length,omitempty"`
	Gt      *float64 `yaml:"gt,omitempty"`
	Gte     *float64 `yaml:"gte,omitempty"`
	Lt      *float64 `yaml:"lt,omitempty"`
	Lte     *float64 `yaml:"lte,omitempty"`
}

// HeaderMatch checks an HTTP header (response headers on the `header` target,
// recorded request headers on the `mock` target). Exactly one matcher.
type HeaderMatch struct {
	Name     string  `yaml:"name"`
	Contains *string `yaml:"contains,omitempty"`
	Equals   *string `yaml:"equals,omitempty"`
	// Matches applies a regexp — the natural shape for auth headers
	// ("^Bearer ") (#24).
	Matches *string `yaml:"matches,omitempty"`
}

// Store captures a value into the variable store (post-MVP).
type Store struct {
	Name string     `yaml:"name"`
	From *StoreFrom `yaml:"from"`
}

// StoreFrom selects where a stored value comes from. Exactly one source is set.
// Stdout/Body extract via a json/regex selector; File reads a generated file via
// a json selector; Header captures an HTTP response header value by name.
type StoreFrom struct {
	Stdout  *StreamAssert `yaml:"stdout,omitempty"`
	Body    *StreamAssert `yaml:"body,omitempty"`
	File    *FileAssert   `yaml:"file,omitempty"`
	Header  string        `yaml:"header,omitempty"`
	Rows    *StreamAssert `yaml:"rows,omitempty"`
	Message *StreamAssert `yaml:"message,omitempty"`
	Value   *StreamAssert `yaml:"value,omitempty"`
}

// trimYAMLScalar strips surrounding whitespace/newlines from a raw scalar node.
func trimYAMLScalar(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\n' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\n' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
