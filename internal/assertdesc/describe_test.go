package assertdesc

import (
	"fmt"
	"testing"

	"github.com/nao1215/atago/internal/spec"
)

func TestDescribeHeader(t *testing.T) {
	t.Parallel()
	style := HeaderStyle{
		Name:  func(s string) string { return "<" + s + ">" },
		Value: func(s string) string { return fmt.Sprintf("%q", s) },
		Regex: func(s string) string { return "/" + s + "/" },
		Bare:  func(s string) string { return "bare " + s },
	}
	tests := []struct {
		name string
		h    *spec.HeaderMatch
		want string
	}{
		{"contains", &spec.HeaderMatch{Name: "X", Contains: strptr("ok")}, `<X> contains "ok"`},
		{"equals", &spec.HeaderMatch{Name: "X", Equals: strptr("1")}, `<X> equals "1"`},
		{"matches", &spec.HeaderMatch{Name: "Auth", Matches: strptr("^Bearer ")}, `<Auth> matches /^Bearer /`},
		{"bare", &spec.HeaderMatch{Name: "X"}, `bare X`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := DescribeHeader(tt.h, style); got != tt.want {
				t.Fatalf("DescribeHeader() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDescribeStreamAndFile(t *testing.T) {
	t.Parallel()
	jsonStyle := JSONStyle{
		Prefix:  func(path string) string { return "at " + path },
		Equals:  func(v any) string { return fmt.Sprintf("== %v", v) },
		Matches: func(s string) string { return "matches /" + s + "/" },
		Length:  func(n int) string { return fmt.Sprintf("length %d", n) },
		Compare: func(op string, v any) string { return fmt.Sprintf("%s %v", op, v) },
		Default: "checked",
	}
	streamStyle := StreamStyle{
		List:      func(ss spec.StringList) string { return fmt.Sprint([]string(ss)) },
		Regex:     func(s string) string { return "/" + s + "/" },
		Equals:    "equals exact text",
		NotEquals: "does not equal exact text",
		JSON:      jsonStyle,
		YAML: JSONStyle{
			Prefix:  func(path string) string { return "yaml " + path },
			Equals:  jsonStyle.Equals,
			Matches: jsonStyle.Matches,
			Length:  jsonStyle.Length,
			Compare: jsonStyle.Compare,
			Default: jsonStyle.Default,
		},
		Snapshot:  func(s string) string { return s },
		NoMatcher: "(no matcher)",
	}
	fileStyle := FileStyle{
		Path:       func(s string) string { return "[" + s + "]" },
		List:       streamStyle.List,
		JSON:       jsonStyle,
		Snapshot:   func(s string) string { return "<" + s + ">" },
		Checked:    func(s string) string { return "checked " + s },
		ExactBytes: "equals exact bytes",
	}

	if got := DescribeStream(&spec.StreamAssert{JSON: spec.JSONChecks{{Path: "$.n", Gte: f64ptr(3)}}}, streamStyle); got != "at $.n >= 3" {
		t.Fatalf("DescribeStream(JSON) = %q", got)
	}
	if got := DescribeStream(&spec.StreamAssert{Snapshot: "out.snap"}, streamStyle); got != "matches snapshot out.snap" {
		t.Fatalf("DescribeStream(snapshot) = %q", got)
	}
	if got := DescribeFile(&spec.FileAssert{Path: "data.json", JSON: spec.JSONChecks{{Path: "$.ok", Equals: true}}}, fileStyle); got != "[data.json] at $.ok == true" {
		t.Fatalf("DescribeFile(JSON) = %q", got)
	}
	if got := DescribeFile(&spec.FileAssert{Path: "raw.bin"}, fileStyle); got != "checked raw.bin" {
		t.Fatalf("DescribeFile(default) = %q", got)
	}
}

// TestDescribeStreamComposedMatchers pins that every matcher a stream assertion
// sets is described. The text matchers compose at run time (all of them must
// hold), so describing only the first one published a weaker contract than the
// run enforces: `contains` + `not_contains` read as a bare `contains`.
func TestDescribeStreamComposedMatchers(t *testing.T) {
	t.Parallel()
	style := StreamStyle{
		List:      func(ss spec.StringList) string { return fmt.Sprint([]string(ss)) },
		Regex:     func(s string) string { return "/" + s + "/" },
		Equals:    "equals exact text",
		NotEquals: "does not equal exact text",
		Snapshot:  func(s string) string { return s },
		Line:      func(n int) string { return fmt.Sprintf("line %d", n) },
		NoMatcher: "(no matcher)",
	}
	tests := []struct {
		name string
		s    *spec.StreamAssert
		want string
	}{
		{
			name: "contains and not_contains",
			s:    &spec.StreamAssert{Contains: spec.StringList{"ok"}, NotContains: spec.StringList{"panic"}},
			want: "contains [ok], does not contain [panic]",
		},
		{
			name: "contains and matches and not_matches",
			s: &spec.StreamAssert{
				Contains:   spec.StringList{"ok"},
				Matches:    strptr("^ok"),
				NotMatches: strptr("error"),
			},
			want: "contains [ok], matches /^ok/, does not match /error/",
		},
		{
			name: "line selector prefixes the matcher",
			s:    &spec.StreamAssert{Line: intptr(2), Equals: strptr("two")},
			want: "line 2 equals exact text",
		},
		{
			name: "line selector with composed matchers",
			s:    &spec.StreamAssert{Line: intptr(3), Contains: spec.StringList{"a"}, NotContains: spec.StringList{"b"}},
			want: "line 3 contains [a], does not contain [b]",
		},
		{
			name: "no matcher",
			s:    &spec.StreamAssert{},
			want: "(no matcher)",
		},
		{
			name: "line selector survives a matcher-less assertion",
			s:    &spec.StreamAssert{Line: intptr(4)},
			want: "line 4 (no matcher)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := DescribeStream(tt.s, style); got != tt.want {
				t.Fatalf("DescribeStream() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestDescribeFileEveryMatcher pins that each matcher the loader accepts on a
// file assertion has its own phrasing. not_contains and executable used to fall
// through to the Checked fallback, so `atago doc` and `atago explain` published
// "the file is checked" for an assertion that pins content or the exec bit.
func TestDescribeFileEveryMatcher(t *testing.T) {
	t.Parallel()
	style := FileStyle{
		Path:       func(s string) string { return "[" + s + "]" },
		List:       func(ss spec.StringList) string { return fmt.Sprint([]string(ss)) },
		Snapshot:   func(s string) string { return "<" + s + ">" },
		Checked:    func(s string) string { return "checked " + s },
		ExactBytes: "equals exact bytes",
	}
	tests := []struct {
		name string
		f    *spec.FileAssert
		want string
	}{
		{"exists", &spec.FileAssert{Path: "a", Exists: boolptr(true)}, "[a] exists"},
		{"not exists", &spec.FileAssert{Path: "a", Exists: boolptr(false)}, "[a] does not exist"},
		{"contains", &spec.FileAssert{Path: "a", Contains: spec.StringList{"x"}}, "[a] contains [x]"},
		{"not_contains", &spec.FileAssert{Path: "a", NotContains: spec.StringList{"secret"}}, "[a] does not contain [secret]"},
		{"executable", &spec.FileAssert{Path: "a", Executable: boolptr(true)}, "[a] is executable"},
		{"not executable", &spec.FileAssert{Path: "a", Executable: boolptr(false)}, "[a] is not executable"},
		{"equals", &spec.FileAssert{Path: "a", Equals: strptr("x")}, "[a] equals exact bytes"},
		{"equals_file", &spec.FileAssert{Path: "a", EqualsFile: strptr("b")}, "[a] is byte-identical to [b]"},
		{"snapshot", &spec.FileAssert{Path: "a", Snapshot: "s"}, "[a] matches snapshot <s>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := DescribeFile(tt.f, style); got != tt.want {
				t.Fatalf("DescribeFile() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDescribeChangesMockImageDirAndPDF(t *testing.T) {
	t.Parallel()
	if got := DescribeChanges(&spec.ChangesAssert{
		Created:  &spec.StringList{"a.txt"},
		Modified: &spec.StringList{},
	}, ChangesStyle{
		Entry: func(s string) string { return "<" + s + ">" },
		Join:  " | ",
	}); got != "created <a.txt> | modified nothing" {
		t.Fatalf("DescribeChanges() = %q", got)
	}

	if got := DescribeMock(&spec.MockAssert{Name: "api", Method: "get", Path: "/v1", Count: intptr(2)}, MockStyle{
		Name:  func(s string) string { return "<" + s + ">" },
		Route: func(s string) string { return "[" + s + "]" },
		Count: func(n int) string { return fmt.Sprintf(" x%d", n) },
	}); got != "mock <api> received [GET /v1] x2" {
		t.Fatalf("DescribeMock() = %q", got)
	}

	streamStyle := StreamStyle{
		List:      func(ss spec.StringList) string { return fmt.Sprint([]string(ss)) },
		Regex:     func(s string) string { return "/" + s + "/" },
		Equals:    "equals exact text",
		NotEquals: "does not equal exact text",
		JSON:      JSONStyle{Prefix: func(path string) string { return path }, Equals: func(v any) string { return fmt.Sprint(v) }, Matches: func(s string) string { return s }, Length: func(n int) string { return fmt.Sprint(n) }, Compare: func(op string, v any) string { return fmt.Sprintf("%s %v", op, v) }, Default: "checked"},
		YAML:      JSONStyle{Prefix: func(path string) string { return path }, Equals: func(v any) string { return fmt.Sprint(v) }, Matches: func(s string) string { return s }, Length: func(n int) string { return fmt.Sprint(n) }, Compare: func(op string, v any) string { return fmt.Sprintf("%s %v", op, v) }, Default: "checked"},
		Snapshot:  func(s string) string { return s },
		NoMatcher: "(no matcher)",
	}

	if got := DescribeImage(&spec.ImageAssert{Path: "out.png", Format: "png", Alpha: boolptr(true), SimilarTo: "base.png"}, ImageStyle{
		Path:      func(s string) string { return "[" + s + "]" },
		Format:    func(s string) string { return "<" + s + ">" },
		SimilarTo: func(s string) string { return "(" + s + ")" },
		Checked:   func(s string) string { return "checked " + s },
	}); got != "[out.png] is <png>, has alpha, similar to (base.png)" {
		t.Fatalf("DescribeImage() = %q", got)
	}

	if got := DescribeDir(&spec.DirAssert{Path: "site", Contains: []string{"index.html"}, Snapshot: "tree.snap", Recursive: true}, DirStyle{
		Path:    func(s string) string { return "[" + s + "]" },
		Item:    func(s string) string { return "<" + s + ">" },
		Token:   func(s string) string { return "(" + s + ")" },
		Checked: func(s string) string { return "checked " + s },
	}); got != "[site] contains <index.html>, tree matches snapshot (tree.snap), (recursive)" {
		t.Fatalf("DescribeDir() = %q", got)
	}

	if got := DescribePDF(&spec.PDFAssert{
		Path:     "r.pdf",
		Pages:    intptr(3),
		Metadata: map[string]string{"title": "Q1"},
		Text:     &spec.StreamAssert{Contains: spec.StringList{"total"}},
	}, PDFStyle{
		Path:    func(s string) string { return "[" + s + "]" },
		Value:   func(s string) string { return "<" + s + ">" },
		Stream:  func(s *spec.StreamAssert) string { return DescribeStream(s, streamStyle) },
		Checked: func(s string) string { return "checked " + s },
	}); got != "[r.pdf] 3 pages, title contains <Q1>, text contains [total]" {
		t.Fatalf("DescribePDF() = %q", got)
	}
}

func strptr(s string) *string   { return &s }
func intptr(n int) *int         { return &n }
func boolptr(b bool) *bool      { return &b }
func f64ptr(f float64) *float64 { return &f }
