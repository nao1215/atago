package ptyrun

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/nao1215/atago/internal/spec"
)

// FuzzRenderScreen attacks the vt100 screen-rendering path (#27). RenderScreen
// feeds the ENTIRE raw pty transcript — arbitrary bytes chosen by the program
// under test — straight into the third-party vt10x escape parser whenever a
// spec asserts `screen:`. A panic in that parser would crash atago mid-suite,
// so the fuzzer explores pathological ANSI (huge CSI params, truncated escapes,
// unterminated OSC, alt-screen toggles, SGR spam, invalid UTF-8, NUL bytes,
// very long lines, CR/LF mixes) across a range of terminal geometries.
//
// Invariants asserted (not just "no panic"), derived from what the code
// actually guarantees:
//   - line budget: vt10x.State.String emits exactly `rows` rows, and
//     RenderScreen only strips trailing blank rows, so the rendered screen
//     never has MORE than `rows` lines;
//   - column budget: String emits exactly `cols` runes per row and RenderScreen
//     only TrimRights each line, so every rendered line is at most `cols` runes
//     wide (matching TestRenderScreen_TruncatesAtCols);
//   - valid UTF-8: the emulator drops invalid UTF-8 input bytes before they
//     reach a cell and back-fills untouched cells with a space, so the returned
//     screen is always valid UTF-8 even though the transcript need not be.
func FuzzRenderScreen(f *testing.F) {
	// Well-formed baselines (mirror screen_test.go).
	f.Add([]byte("loading...\rdone.     \r\nnext\r\n"), 24, 80)
	f.Add([]byte("garbage\r\n\x1b[2J\x1b[HMain Menu\r\n> Settings\r\n"), 10, 40)

	// Huge / negative / empty CSI parameters.
	f.Add([]byte("\x1b[99999999;99999999H"), 24, 80)
	f.Add([]byte("\x1b[-1;-1Hx"), 24, 80)
	f.Add([]byte("\x1b[;;;;;;Hx\x1b[999999999999999999999999999999m"), 5, 5)
	// Regression: a negative DCH/ICH parameter drove vt10x's deleteChars /
	// insertBlanks into slice-bounds-out-of-range panics (found by this fuzzer);
	// sanitizeTranscript now discards these malformed sequences.
	f.Add([]byte("\x1b[-10P"), 76, 28)
	f.Add([]byte("\x1b[-5@"), 24, 80)
	f.Add([]byte("abc\x1b[-99Pdef"), 8, 12)

	// Truncated / bare escape sequences at end of stream.
	f.Add([]byte("text\x1b["), 24, 80)
	f.Add([]byte("text\x1b]"), 24, 80)
	f.Add([]byte("text\x1b"), 24, 80)
	f.Add([]byte("\x1b[38;2;"), 24, 80)

	// OSC without terminator, then more input.
	f.Add([]byte("\x1b]0;title with no ST or BEL and lots of text after it"), 24, 80)
	f.Add([]byte("\x1b]0;title\x07visible"), 24, 80)

	// Alternate-screen enter/exit.
	f.Add([]byte("main\x1b[?1049hALT SCREEN\x1b[?1049lback"), 8, 20)
	f.Add([]byte("\x1b[?47h\x1b[2Jalt\x1b[?47l"), 8, 20)

	// SGR spam.
	f.Add([]byte(strings.Repeat("\x1b[1;31;42;4;7m", 500)+"x"), 4, 10)

	// Invalid UTF-8 and lone continuation bytes.
	f.Add([]byte{0xff, 0xfe, '\n', 0x80, 0x80, 'a'}, 24, 80)
	f.Add([]byte{0xc3, 0x28, 0xe2, 0x82, '\n'}, 24, 80) // invalid + truncated multibyte

	// NUL bytes.
	f.Add([]byte("a\x00b\x00c\r\n\x00\x00"), 24, 80)

	// Very long line (forces wrapping) and a tall geometry.
	f.Add([]byte(strings.Repeat("W", 4096)), 200, 300)

	// CR / LF mixes.
	f.Add([]byte("a\rb\nc\r\nd\n\re"), 6, 6)

	// Realistic htop-style frame with a menu bar, meters, and a process row.
	f.Add([]byte("\x1b[2J\x1b[H\x1b[46;30m 1 \x1b[0m[\x1b[42m||||    \x1b[0m 55%]\r\n"+
		"\x1b[46;30m 2 \x1b[0m[\x1b[41m||||||| \x1b[0m 88%]\r\n"+
		"  PID USER      CPU% MEM%  Command\r\n"+
		"\x1b[7m 1234 nao       12.0  3.4  atago run suite.atago.yaml\x1b[0m\r\n"), 24, 80)

	// Degenerate geometries the harness must clamp.
	f.Add([]byte("clamp me"), 0, 0)
	f.Add([]byte("clamp me"), -100, -100)

	f.Fuzz(func(t *testing.T, transcript []byte, rows, cols int) {
		// Explore geometry too, but keep it terminal-shaped: vt10x requires at
		// least 1x1, and a huge screen just wastes fuzz time, so clamp to a sane
		// band rather than rejecting the input.
		rows = clampGeom(rows)
		cols = clampGeom(cols)

		got := RenderScreen(transcript, &spec.PTY{Rows: rows, Cols: cols})

		if !utf8.ValidString(got) {
			t.Fatalf("RenderScreen returned invalid UTF-8: %q\ntranscript=%q rows=%d cols=%d", got, transcript, rows, cols)
		}

		if got == "" {
			return // an empty screen has zero lines; both budgets hold trivially.
		}
		lines := strings.Split(got, "\n")
		if len(lines) > rows {
			t.Fatalf("screen has %d lines, exceeds rows=%d\ntranscript=%q cols=%d\nscreen=%q", len(lines), rows, transcript, cols, got)
		}
		for i, l := range lines {
			if n := utf8.RuneCountInString(l); n > cols {
				t.Fatalf("line %d is %d runes wide, exceeds cols=%d\ntranscript=%q rows=%d\nline=%q", i, n, cols, transcript, rows, l)
			}
		}
	})
}

// clampGeom keeps a fuzzed terminal dimension in the inclusive band 1..500.
func clampGeom(v int) int {
	if v < 1 {
		return 1
	}
	if v > 500 {
		return 500
	}
	return v
}
