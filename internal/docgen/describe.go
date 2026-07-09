package docgen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/nao1215/atago/internal/assertdesc"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/markdown"
)

// codeList renders a contains/not_contains matcher argument. A single element
// is rendered as one inline-code span (byte-identical to the pre-list format); a
// list joins its elements with ", " so the generated doc shows every required
// (or forbidden) substring.
func codeList(subs spec.StringList) string {
	parts := make([]string, len(subs))
	for i, s := range subs {
		parts[i] = markdown.Code(s)
	}
	return strings.Join(parts, ", ")
}

var docgenJSONStyle = assertdesc.JSONStyle{
	Prefix:  func(path string) string { return "at " + markdown.Code(path) },
	Equals:  func(v any) string { return "equals " + markdown.Code(fmt.Sprint(v)) },
	Matches: func(s string) string { return "matches " + markdown.Code("/"+s+"/") },
	Length:  func(n int) string { return fmt.Sprintf("has length %d", n) },
	Compare: func(op string, v any) string { return "is " + markdown.Code(fmt.Sprintf("%s %v", op, v)) },
	Default: "is checked",
}

var docgenYAMLStyle = assertdesc.JSONStyle{
	Prefix:  func(path string) string { return "YAML at " + markdown.Code(path) },
	Equals:  docgenJSONStyle.Equals,
	Matches: docgenJSONStyle.Matches,
	Length:  docgenJSONStyle.Length,
	Compare: docgenJSONStyle.Compare,
	Default: docgenJSONStyle.Default,
}

var docgenStreamStyle = assertdesc.StreamStyle{
	List:      codeList,
	Regex:     func(s string) string { return markdown.Code("/" + s + "/") },
	Equals:    "equals an exact value",
	NotEquals: "does not equal an exact value",
	JSON:      docgenJSONStyle,
	YAML:      docgenYAMLStyle,
	Snapshot:  markdown.Code,
	NoMatcher: "is checked",
}

var docgenFileStyle = assertdesc.FileStyle{
	Path:       markdown.Code,
	List:       codeList,
	JSON:       docgenJSONStyle,
	Snapshot:   markdown.Code,
	Checked:    func(path string) string { return markdown.Code(path) + " is checked" },
	ExactBytes: "equals exact bytes",
}

var docgenHeaderStyle = assertdesc.HeaderStyle{
	Name:  markdown.Code,
	Value: markdown.Code,
	Regex: func(s string) string { return markdown.Code("/" + s + "/") },
	Bare:  func(s string) string { return markdown.Code(s) + " is checked" },
}

var docgenImageStyle = assertdesc.ImageStyle{
	Path:      markdown.Code,
	Format:    markdown.Code,
	SimilarTo: markdown.Code,
	Checked:   func(path string) string { return markdown.Code(path) + " is checked" },
}

var docgenDirStyle = assertdesc.DirStyle{
	Path:    markdown.Code,
	Item:    markdown.Code,
	Token:   markdown.Code,
	Checked: func(path string) string { return markdown.Code(path) + " is checked" },
}

var docgenPDFStyle = assertdesc.PDFStyle{
	Path:    markdown.Code,
	Value:   markdown.Code,
	Stream:  describeStream,
	Checked: func(path string) string { return markdown.Code(path) + " is checked" },
}

var docgenChangesStyle = assertdesc.ChangesStyle{
	Entry: markdown.Code,
	Join:  ", ",
}

var docgenMockStyle = assertdesc.MockStyle{
	Name:  markdown.Code,
	Route: markdown.Code,
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
			return fmt.Sprintf("exit code is not %s", markdown.Code(fmt.Sprint(*a.ExitCode.Not)))
		}
		if len(a.ExitCode.In) > 0 {
			codes := make([]string, len(a.ExitCode.In))
			for i, n := range a.ExitCode.In {
				codes[i] = markdown.Code(fmt.Sprint(n))
			}
			return "exit code is one of " + strings.Join(codes, ", ")
		}
		if a.ExitCode.Equals != nil {
			return fmt.Sprintf("exit code is %s", markdown.Code(fmt.Sprint(*a.ExitCode.Equals)))
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
			return "HTTP status is " + markdown.Code(fmt.Sprint(*a.Status))
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
			return "gRPC status is " + markdown.Code(fmt.Sprint(*a.GRPCStatus))
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
