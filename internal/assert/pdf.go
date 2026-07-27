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
	rePageObj = regexp.MustCompile(`/Type\s*/Page(?:[^s]|$)`)
	reCount   = regexp.MustCompile(`/Count\s+(\d+)`)
	// A stream's payload starts just after the `stream` keyword and its EOL.
	reStreamStart = regexp.MustCompile(`stream\r?\n`)
	// The EOL before `endstream` is recommended by ISO 32000 but not required,
	// and real producers (Ghostscript) omit it. Requiring it made the scan run
	// past the true end of a stream and swallow the following objects, so a
	// document's later pages contributed no text at all.
	reEndStream = regexp.MustCompile(`\r?\n?endstream`)
	// A direct (non-indirect) stream length: authoritative when present.
	reDirectLength = regexp.MustCompile(`/Length\s+(\d+)\s*(?:/|>>)`)
	// /Type /ObjStm marks a stream that holds document objects (PDF 1.5+), which
	// is where a modern writer puts the Info dictionary.
	reObjStm = regexp.MustCompile(`/Type\s*/ObjStm`)
	// A metadata value is either a "(…)" literal string or a "<…>" hex string.
	reMetaItem = regexp.MustCompile(`/(Title|Author|Subject|Keywords|Creator|Producer)\s*[(<]`)
	// A text operand is a "(…)" literal string or a "<…>" hex string (ISO 32000
	// §7.3.4.3). The hex alternative excludes a leading "<<" dictionary opener.
	reTextOp = regexp.MustCompile(`\((?:[^()\\]|\\.)*\)|<[0-9A-Fa-f\s]*>`)
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
	// parenthesized string literals that feed the text-showing operators. The
	// decoded streams are kept: since PDF 1.5 a writer may pack the document's
	// objects — including the Info dictionary — into a compressed object stream,
	// so metadata has to be looked for inside them too.
	var text strings.Builder
	var objectStreams [][]byte
	for _, st := range pdfStreams(data) {
		decoded := st.payload
		if inflated, err := inflate(st.payload); err == nil {
			decoded = inflated
			// Only a stream that declares /Type /ObjStm holds document objects.
			// Any other decompressed stream is page content, where a literal
			// "/Title(...)" drawn on the page would otherwise be mistaken for
			// the document's metadata.
			if reObjStm.Match(st.dict) {
				objectStreams = append(objectStreams, decoded)
			}
		}
		for _, lit := range reTextOp.FindAll(decoded, -1) {
			text.WriteString(decodePDFString(lit))
		}
		text.WriteByte(' ')
	}
	doc.text = strings.TrimSpace(text.String())

	// Info metadata: read the value after each known field name. The raw bytes
	// come first (a PDF 1.4-style writer leaves the Info dictionary in the
	// clear); anything still missing is looked for in the decompressed object
	// streams, which is where Ghostscript 10, LaTeX, and Word put it.
	readMetadata(doc.metadata, data, false)
	for _, decoded := range objectStreams {
		readMetadata(doc.metadata, decoded, true)
	}
	return doc
}

// dictWindow is how far back from a `stream` keyword the object's dictionary is
// looked for. A dictionary is a short header; a window keeps the scan cheap and
// cannot wander into the previous object's payload for any realistic PDF.
const dictWindow = 512

// pdfStream is one stream object: its payload and the dictionary that declared
// it (used to tell an object stream from page content).
type pdfStream struct {
	dict    []byte
	payload []byte
}

// pdfStreams walks the file and returns every stream object. The payload's end
// comes from the dictionary's /Length when that is a direct integer, which is
// authoritative and survives a payload that happens to contain the bytes
// "endstream"; otherwise — a producer that writes an indirect length, as
// Ghostscript does for page content — it falls back to the next `endstream`
// keyword. Scanning resumes past each stream, so the keyword search can never
// start inside a payload it has already consumed.
func pdfStreams(data []byte) []pdfStream {
	var out []pdfStream
	for pos := 0; pos < len(data); {
		loc := reStreamStart.FindIndex(data[pos:])
		if loc == nil {
			break
		}
		keywordAt := pos + loc[0]
		payloadAt := pos + loc[1]
		dictFrom := keywordAt - dictWindow
		if dictFrom < 0 {
			dictFrom = 0
		}
		dict := data[dictFrom:keywordAt]

		end, resume := streamEnd(data, dict, payloadAt)
		if end < 0 {
			break
		}
		out = append(out, pdfStream{dict: dict, payload: data[payloadAt:end]})
		pos = resume
	}
	return out
}

// streamEnd returns where the payload starting at payloadAt ends and where the
// scan should resume. It prefers the declared /Length, verifying that the
// `endstream` keyword really follows, and falls back to the keyword search.
// (-1, 0) means no terminator was found at all.
func streamEnd(data, dict []byte, payloadAt int) (end, resume int) {
	if m := reDirectLength.FindSubmatch(dict); m != nil {
		if n, err := strconv.Atoi(string(m[1])); err == nil && n >= 0 && payloadAt+n <= len(data) {
			if tail := reEndStream.FindIndex(data[payloadAt+n:]); tail != nil && tail[0] == 0 {
				return payloadAt + n, payloadAt + n + tail[1]
			}
		}
	}
	tail := reEndStream.FindIndex(data[payloadAt:])
	if tail == nil {
		return -1, 0
	}
	return payloadAt + tail[0], payloadAt + tail[1]
}

// readMetadata pulls the known Info fields out of buf into meta. When keepFirst
// is set, a field already found is not overwritten, so the uncompressed document
// trailer keeps precedence over whatever a stream repeats.
func readMetadata(meta map[string]string, buf []byte, keepFirst bool) {
	for _, loc := range reMetaItem.FindAllSubmatchIndex(buf, -1) {
		field := strings.ToLower(string(buf[loc[2]:loc[3]]))
		if keepFirst {
			if _, ok := meta[field]; ok {
				continue
			}
		}
		// loc[1] is just past the opening delimiter ("(" or "<") of the value.
		if val, ok := readPDFStringOrHex(buf, loc[1]-1); ok {
			meta[field] = val
		}
	}
}

// maxPDFStreamBytes caps how many bytes a single FlateDecode stream may inflate
// to. atago runs the (untrusted) output of the CLI under test through pdf
// assertions, and a FlateDecode "zip bomb" — a few hundred KB of zeros — inflates
// to hundreds of megabytes, which would OOM-kill atago. The cap keeps a
// malicious or degenerate stream from exhausting memory; genuine generated PDFs
// have content streams far below this ceiling.
const maxPDFStreamBytes = 64 << 20 // 64 MiB

// inflate decompresses a zlib/FlateDecode stream. A non-Flate (raw) stream
// returns an error so the caller keeps the raw bytes. Decompression is bounded
// by maxPDFStreamBytes: a stream that would inflate past the cap returns an
// error rather than allocating unbounded memory.
func inflate(b []byte) ([]byte, error) {
	zr, err := zlib.NewReader(bytes.NewReader(bytes.TrimLeft(b, "\r\n")))
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	// Read one byte past the cap so we can tell "exactly at the cap" from
	// "over the cap".
	out, err := io.ReadAll(io.LimitReader(zr, maxPDFStreamBytes+1))
	if err != nil {
		return nil, err
	}
	if len(out) > maxPDFStreamBytes {
		return nil, fmt.Errorf("decompressed PDF stream exceeds the %d-byte cap", maxPDFStreamBytes)
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
			if i+1 >= len(data) {
				break // trailing backslash: nothing to escape
			}
			nc := data[i+1]
			switch {
			case nc >= '0' && nc <= '7':
				// Octal escape \ddd (1–3 digits): the canonical way PDFs encode a
				// non-ASCII byte in a literal string. Consume up to three octal
				// digits and emit the resulting byte.
				val := 0
				n := 0
				for n < 3 && i+1 < len(data) && data[i+1] >= '0' && data[i+1] <= '7' {
					val = val*8 + int(data[i+1]-'0')
					i++
					n++
				}
				b.WriteByte(byte(val))
			case nc == '\n':
				// Backslash-newline is a line continuation: both bytes are dropped.
				i++
			case nc == '\r':
				// Same, tolerating a CRLF line ending after the backslash.
				i++
				if i+1 < len(data) && data[i+1] == '\n' {
					i++
				}
			default:
				b.WriteByte(unescapePDFByte(nc))
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

// decodePDFString decodes a full string operand (including the delimiters):
// either a "(…)" literal or a "<…>" hex string.
func decodePDFString(lit []byte) string {
	if len(lit) > 0 && lit[0] == '<' {
		return decodePDFHexString(lit)
	}
	s, _ := readPDFString(lit, 0)
	return s
}

// readPDFStringOrHex reads a PDF string value starting at the opening delimiter
// at index open: a "(…)" literal or a "<…>" hex string.
func readPDFStringOrHex(data []byte, open int) (string, bool) {
	if open < 0 || open >= len(data) {
		return "", false
	}
	if data[open] == '<' {
		return readPDFHexString(data, open)
	}
	return readPDFString(data, open)
}

// readPDFHexString reads a PDF hex string (ISO 32000 §7.3.4.3) that starts at
// the "<" at index open and ends at the matching ">". A leading "<<" is a
// dictionary opener, not a hex string.
func readPDFHexString(data []byte, open int) (string, bool) {
	if open < 0 || open >= len(data) || data[open] != '<' {
		return "", false
	}
	if open+1 < len(data) && data[open+1] == '<' {
		return "", false
	}
	for i := open + 1; i < len(data); i++ {
		c := data[i]
		if c == '>' {
			return decodePDFHexString(data[open : i+1]), true
		}
		if !isHexDigit(c) && !isPDFSpace(c) {
			return "", false
		}
	}
	return "", false
}

// decodePDFHexString decodes a "<…>" hex string. Per spec, whitespace inside is
// ignored and an odd number of hex digits is padded with a trailing 0.
func decodePDFHexString(lit []byte) string {
	digits := make([]byte, 0, len(lit))
	for _, c := range lit {
		if isHexDigit(c) {
			digits = append(digits, c)
		}
	}
	if len(digits)%2 == 1 {
		digits = append(digits, '0')
	}
	var b strings.Builder
	for i := 0; i+1 < len(digits); i += 2 {
		b.WriteByte(hexNibble(digits[i])<<4 | hexNibble(digits[i+1]))
	}
	return b.String()
}

func isHexDigit(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func isPDFSpace(c byte) bool {
	switch c {
	case ' ', '\t', '\r', '\n', '\f', 0:
		return true
	default:
		return false
	}
}

func hexNibble(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	default:
		return c - 'A' + 10
	}
}

func unescapePDFByte(c byte) byte {
	switch c {
	case 'n':
		return '\n'
	case 'r':
		return '\r'
	case 't':
		return '\t'
	case 'b':
		return '\b'
	case 'f':
		return '\f'
	default:
		// \( \) \\ and any other escaped byte pass through literally.
		return c
	}
}
