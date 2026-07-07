package assert

import (
	"bytes"
	"compress/zlib"
	"os"
	"path/filepath"
	"testing"

	"github.com/nao1215/atago/internal/spec"
)

// minimalPDF is a hand-written, uncompressed 1-page PDF with an Info dictionary
// and a text-showing content stream. The atago PDF inspector is lenient (it does
// not require a valid xref table), so this is enough to exercise page count,
// metadata, and text extraction.
const minimalPDF = `%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>
endobj
4 0 obj
<< /Length 46 >>
stream
BT /F1 24 Tf 72 700 Td (Hello atago report) Tj ET
endstream
endobj
6 0 obj
<< /Title (Resume of Ada) /Author (Ada Lovelace) >>
endobj
trailer
<< /Root 1 0 R /Info 6 0 R >>
%%EOF
`

// twoPagePDF has two page objects.
const twoPagePDF = `%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R 4 0 R] /Count 2 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R >>
endobj
4 0 obj
<< /Type /Page /Parent 2 0 R >>
endobj
%%EOF
`

// hexStringPDF is a 1-page PDF whose /Title and content-stream text-showing
// operator use PDF hex-string literals (<...>) instead of (...): a form written
// by LibreOffice/Word/wkhtmltopdf for Unicode. /Title <5265706F7274> is
// "Report" and <48656C6C6F> Tj is "Hello".
const hexStringPDF = `%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>
endobj
4 0 obj
<< /Length 40 >>
stream
BT /F1 24 Tf 72 700 Td <48656C6C6F> Tj ET
endstream
endobj
6 0 obj
<< /Title <5265706F7274> /Author (Ada Lovelace) >>
endobj
trailer
<< /Root 1 0 R /Info 6 0 R >>
%%EOF
`

func writePDF(t *testing.T, body string) string {
	t.Helper()
	wd := t.TempDir()
	if err := os.WriteFile(filepath.Join(wd, "doc.pdf"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return wd
}

func checkPDFOK(t *testing.T, wd string, p *spec.PDFAssert) *CheckResult {
	t.Helper()
	return Check(&spec.Assert{PDF: p}, nil, Env{Workdir: wd})
}

func TestParsePDF_Basics(t *testing.T) {
	doc := parsePDF([]byte(minimalPDF))
	if doc.pages != 1 {
		t.Errorf("pages = %d, want 1", doc.pages)
	}
	if doc.metadata["title"] != "Resume of Ada" {
		t.Errorf("title = %q, want %q", doc.metadata["title"], "Resume of Ada")
	}
	if doc.metadata["author"] != "Ada Lovelace" {
		t.Errorf("author = %q", doc.metadata["author"])
	}
	if got := doc.text; got == "" || !contains(got, "Hello atago report") {
		t.Errorf("text = %q, want it to contain the shown string", got)
	}
}

func TestCheckPDF_Pages(t *testing.T) {
	wd1 := writePDF(t, minimalPDF)
	if cr := checkPDFOK(t, wd1, &spec.PDFAssert{Path: "doc.pdf", Pages: ptrInt(1)}); !cr.OK {
		t.Errorf("pages 1 should pass: %+v", cr)
	}
	if cr := checkPDFOK(t, wd1, &spec.PDFAssert{Path: "doc.pdf", Pages: ptrInt(2)}); cr.OK {
		t.Error("pages 2 should fail on a 1-page PDF")
	}
	wd2 := writePDF(t, twoPagePDF)
	if cr := checkPDFOK(t, wd2, &spec.PDFAssert{Path: "doc.pdf", Pages: ptrInt(2)}); !cr.OK {
		t.Errorf("pages 2 should pass on 2-page PDF: %+v", cr)
	}
	if cr := checkPDFOK(t, wd2, &spec.PDFAssert{Path: "doc.pdf", MinPages: ptrInt(1), MaxPages: ptrInt(3)}); !cr.OK {
		t.Errorf("page bounds should pass: %+v", cr)
	}
}

func TestCheckPDF_MetadataAndText(t *testing.T) {
	wd := writePDF(t, minimalPDF)
	if cr := checkPDFOK(t, wd, &spec.PDFAssert{Path: "doc.pdf", Metadata: map[string]string{"title": "Resume"}}); !cr.OK {
		t.Errorf("metadata title contains should pass: %+v", cr)
	}
	if cr := checkPDFOK(t, wd, &spec.PDFAssert{Path: "doc.pdf", Metadata: map[string]string{"author": "Nobody"}}); cr.OK {
		t.Error("wrong author should fail")
	}
	if cr := checkPDFOK(t, wd, &spec.PDFAssert{Path: "doc.pdf", Text: &spec.StreamAssert{Contains: spec.StringList{"Hello atago"}}}); !cr.OK {
		t.Errorf("text contains should pass: %+v", cr)
	}
	if cr := checkPDFOK(t, wd, &spec.PDFAssert{Path: "doc.pdf", Text: &spec.StreamAssert{Contains: spec.StringList{"Goodbye"}}}); cr.OK {
		t.Error("absent text should fail")
	}
}

// TestParsePDF_HexStrings pins decoding of PDF hex-string literals (ISO 32000
// §7.3.4.3) in both the metadata and the text-operator path. A PDF that encodes
// its /Title and shown text as <...> hex must still satisfy metadata and text
// assertions; otherwise a valid PDF reports a misleading "field not present".
func TestParsePDF_HexStrings(t *testing.T) {
	doc := parsePDF([]byte(hexStringPDF))
	if doc.metadata["title"] != "Report" {
		t.Errorf("title = %q, want %q", doc.metadata["title"], "Report")
	}
	if !contains(doc.text, "Hello") {
		t.Errorf("text = %q, want it to contain %q", doc.text, "Hello")
	}
}

func TestCheckPDF_HexStrings(t *testing.T) {
	wd := writePDF(t, hexStringPDF)
	if cr := checkPDFOK(t, wd, &spec.PDFAssert{Path: "doc.pdf", Metadata: map[string]string{"title": "Report"}}); !cr.OK {
		t.Errorf("hex-string metadata title should match: %+v", cr)
	}
	if cr := checkPDFOK(t, wd, &spec.PDFAssert{Path: "doc.pdf", Text: &spec.StreamAssert{Contains: spec.StringList{"Hello"}}}); !cr.OK {
		t.Errorf("hex-string text should contain Hello: %+v", cr)
	}
}

func TestCheckPDF_NotAPDF(t *testing.T) {
	wd := t.TempDir()
	if err := os.WriteFile(filepath.Join(wd, "doc.pdf"), []byte("not a pdf"), 0o600); err != nil {
		t.Fatal(err)
	}
	if cr := checkPDFOK(t, wd, &spec.PDFAssert{Path: "doc.pdf", Pages: ptrInt(1)}); cr.OK {
		t.Error("a non-PDF file should fail")
	}
}

func TestCheckPDF_PathConfinement(t *testing.T) {
	wd := writePDF(t, minimalPDF)
	if cr := checkPDFOK(t, wd, &spec.PDFAssert{Path: "../escape.pdf", Pages: ptrInt(1)}); cr.OK {
		t.Error("a path escaping the workdir must be rejected")
	}
}

// TestInflate_DecompressionBombCapped proves the FlateDecode inflate path
// refuses a decompression bomb: a tiny zlib stream that would inflate past the
// cap returns an error instead of allocating unbounded memory (atago runs
// untrusted CLI output through pdf assertions). A normal small stream still
// inflates correctly.
func TestInflate_DecompressionBombCapped(t *testing.T) {
	t.Parallel()

	// A highly compressible zero buffer that inflates well beyond the cap.
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	zeros := make([]byte, maxPDFStreamBytes+(1<<20)) // cap + 1 MiB
	if _, err := zw.Write(zeros); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := inflate(buf.Bytes()); err == nil {
		t.Errorf("inflate should reject a stream that inflates past the %d-byte cap", maxPDFStreamBytes)
	}

	// A normal small stream still round-trips.
	var small bytes.Buffer
	sw := zlib.NewWriter(&small)
	if _, err := sw.Write([]byte("hello pdf stream")); err != nil {
		t.Fatal(err)
	}
	if err := sw.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := inflate(small.Bytes())
	if err != nil {
		t.Fatalf("inflate of a small stream should succeed: %v", err)
	}
	if string(got) != "hello pdf stream" {
		t.Errorf("inflate = %q, want %q", got, "hello pdf stream")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// TestDecodePDFString_Escapes pins PDF literal-string decoding (ISO 32000
// §7.3.4.2) as used by the `pdf: {text: {...}}` assertion. Beyond the \n \r \t
// already handled, PDFs written by common generators (pandoc, LaTeX, wkhtmltopdf)
// escape non-ASCII bytes as octal \ddd and use \b \f and backslash-newline line
// continuations. Decoding those wrong makes a valid pdf text assertion silently
// fail to match, so each case here is a correctness guard, not cosmetics.
func TestDecodePDFString_Escapes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		lit  string
		want string
	}{
		{"plain", "(hello)", "hello"},
		{"newline-tab", "(a\\nb\\tc)", "a\nb\tc"},
		// \351 is octal for 0xE9 — 'é' in Latin-1/PDFDocEncoding. It must decode to
		// that single byte, not the literal text "351".
		{"octal-latin1", "(caf\\351)", "caf\xe9"},
		{"octal-short", "(\\7a)", "\x07a"},
		{"octal-three", "(\\101\\102\\103)", "ABC"},
		// \b and \f are defined escapes.
		{"backspace-formfeed", "(x\\by\\fz)", "x\by\fz"},
		// Escaped parentheses and backslash pass through literally.
		{"escaped-parens", "(a\\(b\\)c\\\\d)", "a(b)c\\d"},
		// A backslash immediately before a newline is a line continuation: it and
		// the newline vanish, joining the two source lines.
		{"line-continuation", "(line1\\\nline2)", "line1line2"},
	}
	for _, c := range cases {
		if got := decodePDFString([]byte(c.lit)); got != c.want {
			t.Errorf("%s: decodePDFString(%q) = %q, want %q", c.name, c.lit, got, c.want)
		}
	}
}
