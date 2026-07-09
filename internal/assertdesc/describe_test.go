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
		tt := tt
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
