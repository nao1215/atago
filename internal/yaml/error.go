package yaml

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

// Error is text that is not YAML, or a document that does not fit the value it
// is decoded into. Line and Col are where the problem is, both counted from 1.
type Error struct {
	Line, Col int
	Msg       string
	// Path names the value in the document the problem is at,
	// "scenarios[2].steps[0].run", when the error comes from decoding.
	Path string
	// Key is the mapping key an unknown-field error names.
	Key string
	// Want and Got describe a value of the wrong shape: Want is what the field
	// takes ("a mapping", "a string"), Got the node the document has there.
	Want string
	Got  Kind
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%d:%d] %s", e.Line, e.Col, e.Msg)
}

// contextLines is how many lines an excerpt shows before and after the line an
// error is on.
const contextLines = 3

// Format renders the error with an excerpt of src, the text it was found in:
// the line it is on, marked, a caret under the column, and the lines around it.
func (e *Error) Format(src []byte) string {
	if src == nil || e.Line < 1 {
		return e.Error()
	}
	lines := bytes.Split(src, []byte("\n"))
	if e.Line > len(lines) {
		return e.Error()
	}
	var b strings.Builder
	b.WriteString(e.Error())
	first := max(1, e.Line-contextLines)
	last := min(len(lines), e.Line+contextLines)
	// A text that ends with a line break has an empty last "line" after it,
	// which is not a line of the text.
	if last == len(lines) && len(lines[last-1]) == 0 && last > e.Line {
		last--
	}
	for n := first; n <= last; n++ {
		marker := " "
		if n == e.Line {
			marker = ">"
		}
		fmt.Fprintf(&b, "\n%s%3d | %s", marker, n, lines[n-1])
		if n == e.Line {
			b.WriteString("\n")
			b.WriteString(strings.Repeat(" ", len(">  1 | ")+max(e.Col-1, 0)))
			b.WriteString("^")
		}
	}
	return b.String()
}

// FormatError renders err with an excerpt of src when it is, or wraps, an
// *Error, and as err.Error() otherwise.
func FormatError(err error, src []byte) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Format(src)
	}
	return err.Error()
}
