package docgen

import (
	"fmt"
	"io"
	"strings"
)

// mdDoc builds a Markdown document out of blocks: headings, paragraphs, list
// items and fenced code. The generator needs no more than that, and a writer
// of its own keeps a table renderer, and the dependencies it brings, out of a
// binary that starts on every test run.
type mdDoc struct {
	blocks []string
}

func (m *mdDoc) H1(text string) { m.blocks = append(m.blocks, "# "+text) }
func (m *mdDoc) H2(text string) { m.blocks = append(m.blocks, "## "+text) }
func (m *mdDoc) H3(text string) { m.blocks = append(m.blocks, "### "+text) }
func (m *mdDoc) H4(text string) { m.blocks = append(m.blocks, "#### "+text) }

func (m *mdDoc) H2f(format string, args ...any) { m.H2(fmt.Sprintf(format, args...)) }
func (m *mdDoc) H3f(format string, args ...any) { m.H3(fmt.Sprintf(format, args...)) }

// PlainText adds a paragraph, or any text written as it is.
func (m *mdDoc) PlainText(text string) { m.blocks = append(m.blocks, text) }

func (m *mdDoc) PlainTextf(format string, args ...any) { m.PlainText(fmt.Sprintf(format, args...)) }

// BulletList adds one "- " item per text.
func (m *mdDoc) BulletList(texts ...string) {
	for _, t := range texts {
		m.blocks = append(m.blocks, "- "+t)
	}
}

// CodeBlocks adds a fenced code block in lang.
func (m *mdDoc) CodeBlocks(lang, text string) {
	m.blocks = append(m.blocks, "```"+lang+"\n"+text+"\n```")
}

// String joins the blocks one per line, with a blank line only where Markdown
// needs one to end a block: after a list that the next block is not an item
// of, and after a quote, which would otherwise swallow the next line.
func (m *mdDoc) String() string {
	var b strings.Builder
	for i, block := range m.blocks {
		if i > 0 {
			b.WriteByte('\n')
			if needsBlankLine(m.blocks[i-1], block) {
				b.WriteByte('\n')
			}
		}
		b.WriteString(block)
	}
	return b.String()
}

// Build writes the document to w, ending with a line break.
func (m *mdDoc) Build(w io.Writer) error {
	out := m.String()
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	_, err := io.WriteString(w, out)
	return err
}

// needsBlankLine reports whether a blank line has to separate two blocks.
func needsBlankLine(prev, next string) bool {
	if strings.TrimSpace(prev) == "" || strings.TrimSpace(next) == "" || strings.HasSuffix(prev, "\n") {
		return false
	}
	if strings.HasPrefix(next, "<!--") || strings.HasPrefix(prev, "<!--") {
		return false
	}
	if strings.HasPrefix(prev, ">") {
		return true
	}
	prevList := listKind(prev)
	return prevList != "" && prevList != listKind(next)
}

// listKind names the kind of list item a block is, or "" for none.
func listKind(block string) string {
	trimmed := strings.TrimLeft(block, " ")
	switch {
	case strings.HasPrefix(trimmed, "- [ ] "), strings.HasPrefix(trimmed, "- [x] "):
		return "checkbox"
	case strings.HasPrefix(trimmed, "- "), strings.HasPrefix(trimmed, "* "), strings.HasPrefix(trimmed, "+ "):
		return "bullet"
	}
	digits := 0
	for digits < len(trimmed) && trimmed[digits] >= '0' && trimmed[digits] <= '9' {
		digits++
	}
	if digits > 0 && strings.HasPrefix(trimmed[digits:], ". ") {
		return "ordered"
	}
	return ""
}
