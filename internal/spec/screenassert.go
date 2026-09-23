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

	// Images checks the images on screen now. They are recorded only by a pty
	// step with `graphics: kitty`, which is also what lets a program that
	// requires image support start at all. An image counts from when it is
	// transmitted until a delete command takes it off the screen by id,
	// number, id range, or all; a delete by screen position, z-index, or
	// animation frame is not followed.
	Images *ScreenImages `yaml:"images,omitempty"`
}

// ScreenImages checks the images on a pty step's screen, in the order they
// were sent. Every set field must hold.
type ScreenImages struct {
	// Count is the exact number of images on screen.
	Count *int `yaml:"count,omitempty"`
	// MinCount is the minimum number of images on screen.
	MinCount *int `yaml:"min_count,omitempty"`
	// Contains lists images that must be among those on screen: an entry holds
	// when at least one of them meets every constraint the entry sets. Matching
	// is by content rather than position because a program that loads images
	// concurrently draws them in no fixed order.
	Contains []ScreenImage `yaml:"contains,omitempty"`
}

// ScreenImage is one `screen.images.contains` entry: the pixel constraints of
// the `image:` assertion, applied to a drawn image instead of a file.
type ScreenImage struct {
	Width     *int     `yaml:"width,omitempty"`
	Height    *int     `yaml:"height,omitempty"`
	MinWidth  *int     `yaml:"min_width,omitempty"`
	MaxWidth  *int     `yaml:"max_width,omitempty"`
	MinHeight *int     `yaml:"min_height,omitempty"`
	MaxHeight *int     `yaml:"max_height,omitempty"`
	Alpha     *bool    `yaml:"alpha,omitempty"`
	SimilarTo string   `yaml:"similar_to,omitempty"`
	MaxDiff   *float64 `yaml:"max_diff,omitempty"`
}

// ImageAssert expresses the entry as the `image:` assertion it borrows its
// constraints from, naming the checked image `label`.
func (si *ScreenImage) ImageAssert(label string) *ImageAssert {
	return &ImageAssert{
		Path:      label,
		Width:     si.Width,
		Height:    si.Height,
		MinWidth:  si.MinWidth,
		MaxWidth:  si.MaxWidth,
		MinHeight: si.MinHeight,
		MaxHeight: si.MaxHeight,
		Alpha:     si.Alpha,
		SimilarTo: si.SimilarTo,
		MaxDiff:   si.MaxDiff,
	}
}

// Describe renders the images check as a phrase for explain and doc.
func (si *ScreenImages) Describe() string {
	var parts []string
	if si.Count != nil {
		parts = append(parts, "exactly "+itoa(*si.Count)+" image(s)")
	}
	if si.MinCount != nil {
		parts = append(parts, "at least "+itoa(*si.MinCount)+" image(s)")
	}
	for i := range si.Contains {
		parts = append(parts, "an image "+si.Contains[i].Describe())
	}
	if len(parts) == 0 {
		return "images"
	}
	return "shows " + strings.Join(parts, " and ")
}

// Describe renders the entry's constraints as a phrase.
func (si *ScreenImage) Describe() string {
	var parts []string
	dim := func(name string, v *int) {
		if v != nil {
			parts = append(parts, name+" "+itoa(*v)+"px")
		}
	}
	dim("width", si.Width)
	dim("height", si.Height)
	dim("width >=", si.MinWidth)
	dim("width <=", si.MaxWidth)
	dim("height >=", si.MinHeight)
	dim("height <=", si.MaxHeight)
	if si.Alpha != nil {
		if *si.Alpha {
			parts = append(parts, "with transparency")
		} else {
			parts = append(parts, "without transparency")
		}
	}
	if si.SimilarTo != "" {
		parts = append(parts, "like "+si.SimilarTo)
	}
	return strings.Join(parts, ", ")
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
