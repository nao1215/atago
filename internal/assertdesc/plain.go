package assertdesc

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nao1215/atago/internal/spec"
)

// Describe renders one plain-prose phrase per target an assert sets — "exit code
// is 0", `stdout contains "ready"`, "dir \"site\" has 3 entries" — in the order
// SetTargets reports them.
//
// It is the describer `atago explain` prints and the one `atago manifest`
// records, so the two cannot answer differently about the same assertion. The
// manifest used to reduce every assert to its target name, which made two
// different assertions on one target indistinguishable — and a suite whose
// assertions were gutted produced a byte-identical manifest. `atago doc` renders
// the same facts through the Markdown styles in docgen.
func Describe(a *spec.Assert) []string {
	if a == nil {
		return nil
	}
	targets := a.SetTargets()
	if len(targets) == 0 {
		return []string{"(invalid assertion)"}
	}
	out := make([]string, 0, len(targets))
	for _, t := range targets {
		out = append(out, DescribeTarget(a, t))
	}
	return out
}

// DescribeTarget renders the phrase for one target of an assert.
func DescribeTarget(a *spec.Assert, target spec.AssertTarget) string {
	switch target {
	case spec.AssertExitCode:
		if a.ExitCode.Not != nil {
			return fmt.Sprintf("exit code is not %d", *a.ExitCode.Not)
		}
		if len(a.ExitCode.In) > 0 {
			return "exit code in " + intList(a.ExitCode.In)
		}
		if a.ExitCode.Equals != nil {
			return fmt.Sprintf("exit code is %d", *a.ExitCode.Equals)
		}
		return "exit code"
	case spec.AssertMock:
		return DescribeMock(a.Mock, plainMockStyle)
	case spec.AssertScreen:
		return "screen " + DescribeScreen(a.Screen)
	case spec.AssertDuration:
		return "completes " + a.Duration.DescribeDuration()
	case spec.AssertChanges:
		return "changed exactly " + DescribeChanges(a.Changes, plainChangesStyle)
	case spec.AssertStdout:
		return "stdout " + DescribeStreamPlain(a.Stdout)
	case spec.AssertStderr:
		return "stderr " + DescribeStreamPlain(a.Stderr)
	case spec.AssertFile:
		return "file " + DescribeFile(a.File, plainFileStyle)
	case spec.AssertImage:
		return "image " + DescribeImage(a.Image, plainImageStyle)
	case spec.AssertDir:
		return "dir " + DescribeDir(a.Dir, plainDirStyle)
	case spec.AssertPDF:
		return "pdf " + DescribePDF(a.PDF, plainPDFStyle)
	case spec.AssertStatus:
		if a.Status != nil {
			return fmt.Sprintf("HTTP status is %d", *a.Status)
		}
		return "HTTP status"
	case spec.AssertHeader:
		if a.Header != nil {
			return "header " + DescribeHeader(a.Header, plainHeaderStyle)
		}
		return "header"
	case spec.AssertBody:
		return "body " + DescribeStreamPlain(a.Body)
	case spec.AssertRows:
		return "rows " + DescribeStreamPlain(a.Rows)
	case spec.AssertGRPCStatus:
		if a.GRPCStatus != nil {
			return fmt.Sprintf("gRPC status is %d", *a.GRPCStatus)
		}
		return "gRPC status"
	case spec.AssertMessage:
		return "message " + DescribeStreamPlain(a.Message)
	case spec.AssertValue:
		return "value " + DescribeStreamPlain(a.Value)
	default:
		return string(target)
	}
}

// DescribeStreamPlain renders a stream matcher in the plain style.
func DescribeStreamPlain(s *spec.StreamAssert) string {
	return DescribeStream(s, plainStreamStyle)
}

// DescribeScreen renders a rendered-screen assertion: its stream matcher, plus
// the attribute entries (#382) that say how the text is drawn.
func DescribeScreen(s *spec.ScreenAssert) string {
	if s == nil {
		return ""
	}
	parts := []string{}
	if desc := DescribeStreamPlain(&s.StreamAssert); desc != "" {
		parts = append(parts, desc)
	}
	for i := range s.Attrs {
		parts = append(parts, "shows "+s.Attrs[i].Describe())
	}
	return strings.Join(parts, " and ")
}

// intList renders an accepted exit-code set as "[0, 2]" (#19).
func intList(ns []int) string {
	parts := make([]string, len(ns))
	for i, n := range ns {
		parts[i] = strconv.Itoa(n)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

var plainJSONStyle = JSONStyle{
	Prefix:  func(path string) string { return "JSON " + path },
	Equals:  func(v any) string { return "== " + JSONValueText(v) },
	Matches: func(s string) string { return fmt.Sprintf("matches /%s/", s) },
	Length:  func(n int) string { return fmt.Sprintf("length %d", n) },
	Compare: func(op string, v any) string { return fmt.Sprintf("%s %v", op, v) },
	Default: "",
}

var plainYAMLStyle = plainJSONStyle.WithPrefix(func(path string) string {
	return "YAML " + path
})

var plainStreamStyle = StreamStyle{
	List:      spec.StringList.Quoted,
	Regex:     func(s string) string { return fmt.Sprintf("/%s/", s) },
	Equals:    "equals exact text",
	NotEquals: "does not equal exact text",
	JSON:      plainJSONStyle,
	YAML:      plainYAMLStyle,
	Snapshot:  func(s string) string { return s },
	Line:      func(n int) string { return fmt.Sprintf("line %d", n) },
	NoMatcher: "(no matcher)",
}

var plainFileStyle = FileStyle{
	Path:       func(s string) string { return fmt.Sprintf("%q", s) },
	List:       spec.StringList.Quoted,
	JSON:       plainJSONStyle,
	Snapshot:   func(s string) string { return s },
	Checked:    func(path string) string { return path },
	ExactBytes: "equals exact bytes",
}

var plainHeaderStyle = HeaderStyle{
	Name:  func(s string) string { return fmt.Sprintf("%q", s) },
	Value: func(s string) string { return fmt.Sprintf("%q", s) },
	Regex: func(s string) string { return fmt.Sprintf("/%s/", s) },
	Bare:  func(s string) string { return fmt.Sprintf("%q", s) },
}

var plainImageStyle = ImageStyle{
	Path:      func(s string) string { return fmt.Sprintf("%q", s) },
	Format:    func(s string) string { return s },
	SimilarTo: func(s string) string { return s },
	Checked:   func(path string) string { return fmt.Sprintf("%q is checked", path) },
}

var plainDirStyle = DirStyle{
	Path:    func(s string) string { return fmt.Sprintf("%q", s) },
	Item:    func(s string) string { return s },
	Token:   func(s string) string { return s },
	Checked: func(path string) string { return fmt.Sprintf("%q is checked", path) },
}

var plainPDFStyle = PDFStyle{
	Path:    func(s string) string { return fmt.Sprintf("%q", s) },
	Value:   func(s string) string { return fmt.Sprintf("%q", s) },
	Stream:  DescribeStreamPlain,
	Checked: func(path string) string { return fmt.Sprintf("%q is checked", path) },
}

var plainChangesStyle = ChangesStyle{
	Entry: func(s string) string { return s },
	Join:  "; ",
}

var plainMockStyle = MockStyle{
	Name:  func(s string) string { return s },
	Route: func(s string) string { return s },
	Count: func(n int) string { return fmt.Sprintf(" x%d", n) },
}
