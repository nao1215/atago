package assert

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
)

// checkPDF evaluates a PDF assertion (#73). It is deliberately black-box and
// content-oriented: it inspects the page count, the Info-dictionary metadata,
// and the text extracted from content streams — without a layout engine. Every
// set field is an independent constraint; the first failing one is reported. The
// PDF path is confined to the scenario workdir.
func checkPDF(p *spec.PDFAssert, env Env) *CheckResult {
	path, err := security.ResolveWorkdirPath("assert.pdf.path", env.Workdir, p.Path)
	if err != nil {
		return &CheckResult{Desc: fmt.Sprintf("assert pdf %q", p.Path), Hint: err.Error()}
	}
	data, rerr := os.ReadFile(path) //nolint:gosec // path is confined to the workdir above
	if rerr != nil {
		return &CheckResult{
			Desc:     fmt.Sprintf("assert pdf %q", p.Path),
			Expected: fmt.Sprintf("readable PDF %q", p.Path),
			Actual:   rerr.Error(),
			Hint:     fmt.Sprintf("could not read %q: %v", p.Path, rerr),
		}
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		return &CheckResult{
			Desc:     fmt.Sprintf("assert pdf %q", p.Path),
			Expected: "a PDF file (%PDF- header)",
			Actual:   "missing %PDF- header",
			Hint:     fmt.Sprintf("%q does not look like a PDF", p.Path),
		}
	}

	doc := parsePDF(data)

	if cr := checkPDFPages(p, doc); cr != nil {
		return cr
	}
	if cr := checkPDFMetadata(p, doc); cr != nil {
		return cr
	}
	if p.Text != nil {
		// Reuse the standard stream matchers against the extracted text so pdf text
		// checks share contains/matches/equals/snapshot semantics with every other
		// stream target.
		return checkStream("pdf text", p.Text, []byte(doc.text), true, env)
	}
	return pass(fmt.Sprintf("assert pdf %q", p.Path))
}

func checkPDFPages(p *spec.PDFAssert, doc pdfDoc) *CheckResult {
	if p.Pages == nil && p.MinPages == nil && p.MaxPages == nil {
		return nil
	}
	n := doc.pages
	fail := func(want string) *CheckResult {
		return &CheckResult{
			Desc:     fmt.Sprintf("assert pdf %q page count", p.Path),
			Expected: want,
			Actual:   fmt.Sprintf("%d pages", n),
			Hint:     fmt.Sprintf("PDF %q has %d pages, expected %s", p.Path, n, want),
		}
	}
	if p.Pages != nil && n != *p.Pages {
		return fail(fmt.Sprintf("exactly %d pages", *p.Pages))
	}
	if p.MinPages != nil && n < *p.MinPages {
		return fail(fmt.Sprintf("at least %d pages", *p.MinPages))
	}
	if p.MaxPages != nil && n > *p.MaxPages {
		return fail(fmt.Sprintf("at most %d pages", *p.MaxPages))
	}
	return nil
}

func checkPDFMetadata(p *spec.PDFAssert, doc pdfDoc) *CheckResult {
	for _, field := range sortedMetadataKeys(p.Metadata) {
		want := p.Metadata[field]
		got, ok := doc.metadata[strings.ToLower(field)]
		if !ok || !strings.Contains(got, want) {
			return &CheckResult{
				Desc:     fmt.Sprintf("assert pdf %q metadata %s", p.Path, field),
				Expected: fmt.Sprintf("%s contains %q", field, want),
				Actual:   metadataActual(got, ok),
				Hint:     fmt.Sprintf("PDF %q metadata %s does not contain %q", p.Path, field, want),
			}
		}
	}
	return nil
}

func metadataActual(got string, ok bool) string {
	if !ok {
		return "field not present"
	}
	return excerpt(got)
}

func sortedMetadataKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// Deterministic order so a failing metadata check is stable.
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}

// pdfDoc is the extracted, black-box view of a PDF: its page count, Info
// metadata (lower-cased keys), and concatenated content-stream text.
type pdfDoc struct {
	pages    int
	metadata map[string]string
	text     string
}

var (
	// A page object is `/Type /Page` not followed by another letter (so it does
	// not match `/Type /Pages`). Whitespace between token parts is flexible.
	rePageObj  = regexp.MustCompile(`/Type\s*/Page(?:[^s]|$)`)
	reCount    = regexp.MustCompile(`/Count\s+(\d+)`)
	reStream   = regexp.MustCompile(`(?s)stream\r?\n(.*?)\r?\nendstream`)
	reMetaItem = regexp.MustCompile(`/(Title|Author|Subject|Keywords|Creator|Producer)\s*\(`)
	reTextOp   = regexp.MustCompile(`\((?:[^()\\]|\\.)*\)`)
)

// parsePDF extracts a black-box view of a PDF. It is lenient by design: it reads
// classic (non-object-stream) PDFs and decompresses FlateDecode content streams
// with the standard library, which covers the common generated-PDF case without a
// third-party dependency.
func parsePDF(data []byte) pdfDoc {
	doc := pdfDoc{metadata: map[string]string{}}

	// Page count: prefer counting page objects; fall back to the Pages /Count.
	if m := rePageObj.FindAll(data, -1); len(m) > 0 {
		doc.pages = len(m)
	} else if c := reCount.FindSubmatch(data); c != nil {
		if n, err := strconv.Atoi(string(c[1])); err == nil {
			doc.pages = n
		}
	}

	// Content-stream text: decode every stream (raw + zlib/Flate) and pull the
	// parenthesized string literals that feed the text-showing operators.
	var text strings.Builder
	for _, m := range reStream.FindAllSubmatch(data, -1) {
		raw := m[1]
		decoded := raw
		if inflated, err := inflate(raw); err == nil {
			decoded = inflated
		}
		for _, lit := range reTextOp.FindAll(decoded, -1) {
			text.WriteString(decodePDFString(lit))
		}
		text.WriteByte(' ')
	}
	doc.text = strings.TrimSpace(text.String())

	// Info metadata: read the parenthesized value after each known field name.
	for _, loc := range reMetaItem.FindAllSubmatchIndex(data, -1) {
		field := strings.ToLower(string(data[loc[2]:loc[3]]))
		// loc[1] is just past the opening "(" of the value.
		if val, ok := readPDFString(data, loc[1]-1); ok {
			doc.metadata[field] = val
		}
	}
	return doc
}

// inflate decompresses a zlib/FlateDecode stream. A non-Flate (raw) stream
// returns an error so the caller keeps the raw bytes.
func inflate(b []byte) ([]byte, error) {
	zr, err := zlib.NewReader(bytes.NewReader(bytes.TrimLeft(b, "\r\n")))
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	out, err := io.ReadAll(zr)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// readPDFString reads a PDF literal string starting at the opening parenthesis at
// index open, honoring balanced parentheses and backslash escapes.
func readPDFString(data []byte, open int) (string, bool) {
	if open < 0 || open >= len(data) || data[open] != '(' {
		return "", false
	}
	depth := 0
	var b strings.Builder
	for i := open; i < len(data); i++ {
		c := data[i]
		switch c {
		case '\\':
			if i+1 < len(data) {
				b.WriteByte(unescapePDFByte(data[i+1]))
				i++
			}
		case '(':
			if depth > 0 {
				b.WriteByte(c)
			}
			depth++
		case ')':
			depth--
			if depth == 0 {
				return b.String(), true
			}
			b.WriteByte(c)
		default:
			b.WriteByte(c)
		}
	}
	return b.String(), true
}

// decodePDFString decodes a full "(…)" literal (including the delimiters).
func decodePDFString(lit []byte) string {
	s, _ := readPDFString(lit, 0)
	return s
}

func unescapePDFByte(c byte) byte {
	switch c {
	case 'n':
		return '\n'
	case 'r':
		return '\r'
	case 't':
		return '\t'
	default:
		return c
	}
}
