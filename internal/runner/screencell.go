package runner

// ScreenCell is one cell of a rendered terminal screen, with the colors and
// attributes the emulator tracked for it (#382). It exists so a `screen:`
// assertion can check the visual grammar a TUI relies on — the error line is
// red, the selected row is reverse-video, `--no-color` really did turn colors
// off — none of which survives the plain-text render.
//
// It lives here rather than in the pty runner because runner.Result carries it
// and the pty runner imports this package, not the other way round.
type ScreenCell struct {
	// Content is the grapheme cluster drawn in this cell — usually a single rune,
	// but a base plus combining marks or a ZWJ emoji sequence is one cluster in
	// one cell, kept whole so a `screen:` text match or an `attrs:` query naming it
	// lines up cell-for-cluster (#437).
	Content string
	// FG and BG are the emulator's colors: 0..15 for the ANSI palette, 16..255
	// for the xterm palette, and DefaultColor for "the terminal's own", which is
	// what makes an uncolored screen assertable.
	FG, BG uint32
	// The attribute bits the emulator tracks. Anything it does not track (dim,
	// strikethrough) is deliberately absent rather than reported as false.
	Bold      bool
	Italic    bool
	Underline bool
	Reverse   bool
	Blink     bool
}

// DefaultColor marks a cell drawn in the terminal's own foreground or
// background — no color was set. It matches vt10x's DefaultFG, and DefaultFG+1
// is its DefaultBG; both map here so a spec can say `fg: default` without
// knowing which one a given cell carries.
const DefaultColor uint32 = 1 << 24
