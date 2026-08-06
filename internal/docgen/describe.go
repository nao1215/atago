package docgen

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/nao1215/atago/internal/assertdesc"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/markdown"
)

// code renders a spec-supplied string as an inline-code span, writing a value
// the reader cannot see as a Go-quoted literal instead. A matcher may
// legitimately assert on a control character — `not_contains: "\r"` is how a
// spec guards a golden against a stray carriage return — and rendering that raw
// produced a span that read as empty, and put a bare CR into the committed
// Markdown. The console failure block already quotes such a value (`does not
// contain "\r"`); the generated docs now say the same thing. Printable text,
// including non-ASCII, is rendered exactly as authored.
func code(s string) string {
	if strings.ContainsFunc(s, unicode.IsControl) {
		return markdown.Code(strconv.Quote(s))
	}
	return markdown.Code(s)
}

// codeRegex renders a regex matcher's pattern as an inline-code span. Control
// characters are escaped so the span stays on one line — a pattern containing a
// literal newline used to break the span across two, which is not inline code at
// all — but the rest of the pattern is left byte-for-byte as authored.
// strconv.Quote would double every backslash (`\.` into `\\.`) and stop the
// published pattern from reading as the regex the spec wrote.
func codeRegex(pattern string) string {
	return markdown.Code("/" + escapeInvisible(pattern) + "/")
}

// escapeInvisible rewrites the runes a reader cannot see — CR, LF, tab, ESC, and
// the other control characters — as their Go escapes, leaving every other byte
// exactly as authored.
func escapeInvisible(s string) string {
	if !strings.ContainsFunc(s, unicode.IsControl) {
		return s
	}
	var b strings.Builder
	for _, r := range s {
		if unicode.IsControl(r) {
			// QuoteRune wraps its result in single quotes ('\r'); a control rune is
			// never a quote itself, so trimming them yields the bare escape.
			b.WriteString(strings.Trim(strconv.QuoteRune(r), "'"))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// codeList renders a contains/not_contains matcher argument. A single element
// is rendered as one inline-code span (byte-identical to the pre-list format); a
// list joins its elements with ", " so the generated doc shows every required
// (or forbidden) substring.
func codeList(subs spec.StringList) string {
	parts := make([]string, len(subs))
	for i, s := range subs {
		parts[i] = code(s)
	}
	return strings.Join(parts, ", ")
}

var docgenJSONStyle = assertdesc.JSONStyle{
	Prefix:  func(path string) string { return "at " + code(path) },
	Equals:  func(v any) string { return "equals " + code(assertdesc.JSONValueText(v)) },
	Matches: func(s string) string { return "matches " + codeRegex(s) },
	Length:  func(n int) string { return fmt.Sprintf("has length %d", n) },
	Compare: func(op string, v any) string { return "is " + code(fmt.Sprintf("%s %v", op, v)) },
	Default: "is checked",
}

var docgenYAMLStyle = docgenJSONStyle.WithPrefix(func(path string) string {
	return "YAML at " + code(path)
})

var docgenStreamStyle = assertdesc.StreamStyle{
	List:      codeList,
	Regex:     codeRegex,
	Equals:    "equals an exact value",
	NotEquals: "does not equal an exact value",
	JSON:      docgenJSONStyle,
	YAML:      docgenYAMLStyle,
	Snapshot:  code,
	Line:      func(n int) string { return "line " + code(fmt.Sprint(n)) },
	NoMatcher: "is checked",
}

var docgenFileStyle = assertdesc.FileStyle{
	Path:       code,
	List:       codeList,
	JSON:       docgenJSONStyle,
	Snapshot:   code,
	Checked:    func(path string) string { return code(path) + " is checked" },
	ExactBytes: "equals exact bytes",
}

var docgenHeaderStyle = assertdesc.HeaderStyle{
	Name:  code,
	Value: code,
	Regex: codeRegex,
	Bare:  func(s string) string { return code(s) + " is checked" },
}

var docgenImageStyle = assertdesc.ImageStyle{
	Path:      code,
	Format:    code,
	SimilarTo: code,
	Checked:   func(path string) string { return code(path) + " is checked" },
}

var docgenDirStyle = assertdesc.DirStyle{
	Path:    code,
	Item:    code,
	Token:   code,
	Checked: func(path string) string { return code(path) + " is checked" },
}

var docgenPDFStyle = assertdesc.PDFStyle{
	Path:    code,
	Value:   code,
	Stream:  describeStream,
	Checked: func(path string) string { return code(path) + " is checked" },
}

var docgenChangesStyle = assertdesc.ChangesStyle{
	Entry: code,
	Join:  ", ",
}

var docgenMockStyle = assertdesc.MockStyle{
	Name:  code,
	Route: code,
	Count: func(n int) string { return fmt.Sprintf(" exactly %d time(s)", n) },
}

// describeAsserts renders an assertion as one Markdown "Then" bullet per target.
// An assert may set several targets (exit_code + stdout + …); each is its own
// independent check, so each gets its own bullet.
func describeAsserts(a *spec.Assert) []string {
	targets := a.SetTargets()
	if len(targets) == 0 {
		return []string{"_(invalid assertion)_"}
	}
	bullets := make([]string, 0, len(targets))
	for _, t := range targets {
		bullets = append(bullets, describeTarget(a, t))
	}
	return bullets
}

// describeTarget renders a single assertion target as a Markdown "Then" bullet.
func describeTarget(a *spec.Assert, target spec.AssertTarget) string {
	switch target {
	case spec.AssertExitCode:
		if a.ExitCode.Not != nil {
			return fmt.Sprintf("exit code is not %s", code(fmt.Sprint(*a.ExitCode.Not)))
		}
		if len(a.ExitCode.In) > 0 {
			codes := make([]string, len(a.ExitCode.In))
			for i, n := range a.ExitCode.In {
				codes[i] = code(fmt.Sprint(n))
			}
			return "exit code is one of " + strings.Join(codes, ", ")
		}
		if a.ExitCode.Equals != nil {
			return fmt.Sprintf("exit code is %s", code(fmt.Sprint(*a.ExitCode.Equals)))
		}
		return "exit code is checked"
	case spec.AssertMock:
		return describeMockAssert(a.Mock)
	case spec.AssertScreen:
		return "rendered screen " + describeStream(a.Screen)
	case spec.AssertDuration:
		return "completes " + a.Duration.DescribeDuration()
	case spec.AssertChanges:
		return "the step changed exactly " + describeChanges(a.Changes)
	case spec.AssertStdout:
		return "stdout " + describeStream(a.Stdout)
	case spec.AssertStderr:
		return "stderr " + describeStream(a.Stderr)
	case spec.AssertFile:
		return "file " + describeFile(a.File)
	case spec.AssertImage:
		return "image " + describeImage(a.Image)
	case spec.AssertDir:
		return "dir " + describeDir(a.Dir)
	case spec.AssertPDF:
		return "pdf " + describePDF(a.PDF)
	case spec.AssertStatus:
		if a.Status != nil {
			return "HTTP status is " + code(fmt.Sprint(*a.Status))
		}
		return "HTTP status is checked"
	case spec.AssertHeader:
		if a.Header != nil {
			return "header " + describeHeader(a.Header)
		}
		return "header is checked"
	case spec.AssertBody:
		return "body " + describeStream(a.Body)
	case spec.AssertRows:
		return "rows " + describeStream(a.Rows)
	case spec.AssertGRPCStatus:
		if a.GRPCStatus != nil {
			return "gRPC status is " + code(fmt.Sprint(*a.GRPCStatus))
		}
		return "gRPC status is checked"
	case spec.AssertMessage:
		return "message " + describeStream(a.Message)
	case spec.AssertValue:
		return "value " + describeStream(a.Value)
	default:
		return string(target)
	}
}

// describeChanges renders a workdir-delta assertion (#70) as a compact phrase
// listing each set category. `modified: []` renders as "modified nothing".
func describeChanges(c *spec.ChangesAssert) string {
	return assertdesc.DescribeChanges(c, docgenChangesStyle)
}

func describeHeader(h *spec.HeaderMatch) string {
	return assertdesc.DescribeHeader(h, docgenHeaderStyle)
}

func describeImage(im *spec.ImageAssert) string {
	return assertdesc.DescribeImage(im, docgenImageStyle)
}

// describeDir renders a directory/tree assertion (#74) as a compact phrase
// listing each set constraint.
func describeDir(d *spec.DirAssert) string {
	return assertdesc.DescribeDir(d, docgenDirStyle)
}

// describePDF renders a PDF assertion (#73) as a compact phrase.
func describePDF(p *spec.PDFAssert) string {
	return assertdesc.DescribePDF(p, docgenPDFStyle)
}

func describeStream(s *spec.StreamAssert) string {
	return assertdesc.DescribeStream(s, docgenStreamStyle)
}

func describeFile(f *spec.FileAssert) string {
	return assertdesc.DescribeFile(f, docgenFileStyle)
}

func describeMockAssert(m *spec.MockAssert) string {
	return assertdesc.DescribeMock(m, docgenMockStyle)
}

func sortedEnvKeys(env map[string]string) []string {
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
