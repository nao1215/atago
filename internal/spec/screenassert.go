package spec

import "strings"

// ScreenAssert is the rendered-terminal assertion target (#27, #382): the
// stream matchers over the screen's text, plus attribute checks over the colors
// and styling the emulator tracked for it.
//
// The text matchers and `attrs` compose — both must hold — because they answer
// different questions about the same frame: what it says, and how it looks.
type ScreenAssert struct {
	StreamAssert `yaml:",inline"`

	// Attrs checks how text is drawn (#382): the error line is red, the selected
	// row is reverse-video, `--no-color` really did leave the frame uncolored.
	// Every entry must hold.
	Attrs []ScreenAttr `yaml:"attrs,omitempty"`
}

// ScreenAttr is one "this text is drawn like this" claim (#382).
//
// It is position-free by default: the entry passes when at least ONE occurrence
// of Text on the screen has every one of its cells carrying the demanded
// attributes. That keeps an assertion about styling from breaking every time the
// layout shifts, which is the failure mode that would make the feature more
// trouble than it is worth. Pin a row with Row when the position is the point.
type ScreenAttr struct {
	// Text is the literal substring whose cells are checked. Required.
	Text string `yaml:"text"`
	// Row restricts the search to one 1-based screen row, addressed the same way
	// `line:` addresses the text matchers.
	Row int `yaml:"row,omitempty"`
	// FG and BG name a color: an ANSI name (`red`, `bright-red`), a 256-palette
	// index (`203`), or `default` — the terminal's own color, which is how a
	// `--no-color` contract becomes assertable.
	FG string `yaml:"fg,omitempty"`
	BG string `yaml:"bg,omitempty"`
	// The attribute bits vt10x tracks. Each is a *bool so that `bold: false` is a
	// real claim ("this text must NOT be bold") rather than indistinguishable
	// from not asking. Dim and strikethrough are absent because the emulator does
	// not track them, and reporting them as false would be a lie.
	Bold      *bool `yaml:"bold,omitempty"`
	Italic    *bool `yaml:"italic,omitempty"`
	Underline *bool `yaml:"underline,omitempty"`
	Reverse   *bool `yaml:"reverse,omitempty"`
	Blink     *bool `yaml:"blink,omitempty"`
}

// HasAttribute reports whether the entry demands anything at all beyond naming
// text, which the loader requires.
func (a *ScreenAttr) HasAttribute() bool {
	return a.FG != "" || a.BG != "" ||
		a.Bold != nil || a.Italic != nil || a.Underline != nil ||
		a.Reverse != nil || a.Blink != nil
}

// Describe renders the entry as a human phrase, shared by explain and doc so
// the two never drift: `"ERROR" in bold red`, `"README.md" reverse on row 4`.
func (a *ScreenAttr) Describe() string {
	var parts []string
	for _, s := range []struct {
		want *bool
		name string
	}{
		{a.Bold, "bold"},
		{a.Italic, "italic"},
		{a.Underline, "underlined"},
		{a.Reverse, "reverse"},
		{a.Blink, "blinking"},
	} {
		if s.want == nil {
			continue
		}
		if *s.want {
			parts = append(parts, s.name)
		} else {
			parts = append(parts, "not "+s.name)
		}
	}
	if a.FG != "" {
		parts = append(parts, a.FG)
	}
	if a.BG != "" {
		parts = append(parts, "on "+a.BG)
	}
	phrase := `"` + a.Text + `"`
	if len(parts) > 0 {
		phrase += " in " + strings.Join(parts, " ")
	}
	if a.Row > 0 {
		phrase += " on row " + itoa(a.Row)
	}
	return phrase
}

// itoa avoids pulling strconv in for one call in a description helper.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
