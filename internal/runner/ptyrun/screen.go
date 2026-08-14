package ptyrun

import (
	"bytes"
	"image/color"
	"io"
	"strings"
	"unicode/utf8"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/vt"

	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// screenResize is one mid-session terminal size change, anchored at the byte
// offset in the RAW transcript where it took effect (#379).
type screenResize struct {
	offset     int
	rows, cols int
}

// RenderScreen replays a pty transcript through a terminal emulator and returns
// the final rendered screen as plain text (#27): what the user actually SEES
// after every cursor move, overwrite, and erase — the signal a raw transcript
// scatters across redraws. Trailing whitespace is stripped per line and trailing
// blank lines are dropped.
//
// A wide character (CJK, emoji) occupies two terminal columns, so cursor
// addressing after it depends on the emulator modeling that width. The x/vt
// emulator does; a program that positions a label just past a Japanese string
// with `\x1b[row;colH` lands it where the terminal put it, not two columns early,
// and overwriting one half of a wide cell blanks it the way a terminal does
// (#432). One edge remains upstream: a wide character that must AUTOWRAP at the
// right margin (no explicit newline, the char straddling the last column) is
// dropped rather than carried to the next line. TUIs position with cursor
// addressing and explicit newlines, which render correctly.
func RenderScreen(transcript []byte, p *spec.PTY) string {
	return renderScreenResized(transcript, p, nil)
}

// renderScreenResized is RenderScreen for a session that changed size while it
// ran (#379): the transcript is replayed in pieces, resizing the emulator at
// each recorded boundary, so every part of the frame is drawn under the size it
// was actually produced under. With no resizes it is one piece and behaves
// exactly as before.
//
// The whole transcript is sanitized ONCE and the pieces are cut from the result.
// Sanitizing each raw piece separately would be a different thing entirely: a
// cut inside an escape sequence hides it from the scanner, and vt10x's stateful
// parser reassembles it on the far side — which is how a clamped repeat count
// gets back its quadrillion steps. sanitizeTranscriptMarks translates the
// offsets instead, landing every cut on a boundary between whole units.
func renderScreenResized(transcript []byte, p *spec.PTY, resizes []screenResize) string {
	text, _ := renderScreenCells(transcript, p, resizes)
	return text
}

// renderScreenCells is the one renderer: it replays the transcript and returns
// BOTH the plain-text screen and the same screen with its colors and attributes
// (#382). The text is derived from the cells rather than from the emulator's own
// String(), so `attrs:` and the text matchers can never disagree about what is
// on row 3 — a mismatch there would make an attribute failure point at the wrong
// line, which is worse than not having the feature.
func renderScreenCells(transcript []byte, p *spec.PTY, resizes []screenResize) (string, [][]runner.ScreenCell) {
	rows, cols := defaultRows, defaultCols
	if p.Rows > 0 {
		rows = p.Rows
	}
	if p.Cols > 0 {
		cols = p.Cols
	}
	term := vt.NewEmulator(cols, rows)

	// The emulator answers a device query in the transcript (a program's `\x1b[6n`
	// cursor-position report, DA1) and even a resize by writing the reply to an
	// internal pipe meant for the child's stdin. This render has no child, so
	// nothing would drain that pipe and the very next Write or Resize would block
	// forever. A goroutine drains and discards the replies for the emulator's
	// lifetime.
	//
	// It is closed by shutting the pipe's WRITE end directly (InputPipe returns it)
	// rather than by Emulator.Close: Close also flips an unsynchronized `closed`
	// flag that Read reads, which the race detector flags. Closing the pipe writer
	// EOFs the blocked Read through io.Pipe's own synchronization and leaves that
	// flag untouched.
	drainDone := make(chan struct{})
	go func() {
		defer close(drainDone)
		buf := make([]byte, 4096)
		for {
			if _, err := term.Read(buf); err != nil {
				return
			}
		}
	}()
	defer func() {
		if pw, ok := term.InputPipe().(*io.PipeWriter); ok {
			_ = pw.Close()
		}
		<-drainDone
	}()

	// Offsets must be ascending and inside the transcript for the translation
	// to mean anything; a recorded offset can be neither when the screen is
	// rendered from a snapshot shorter than the one the resize was recorded
	// against.
	marks := make([]int, len(resizes))
	prev := 0
	for i, r := range resizes {
		prev = min(max(r.offset, prev), len(transcript))
		marks[i] = prev
	}
	sanitized, cuts := sanitizeTranscriptMarks(transcript, marks)

	at := 0
	for i, r := range resizes {
		cut := min(max(cuts[i], at), len(sanitized))
		writeTranscript(term, sanitized[at:cut])
		// Resize takes width (cols) first; getting that backwards silently
		// transposes every frame after a resize.
		term.Resize(r.cols, r.rows)
		at = cut
	}
	writeTranscript(term, sanitized[at:])

	// Read the grid out of the emulator once and build both views from it. The
	// emulator's own String() is not consulted: two independent reads of the same
	// state could disagree about a cell, and the whole value of the attribute
	// matchers is that a row number means the same thing in both.
	//
	// A wide character occupies two columns: the first holds the grapheme, the
	// second is a continuation cell (Width 0). The continuation is dropped so one
	// logical cell holds one grapheme — the text row reads "日本語", not "日 本 語",
	// and the attrs matcher keeps matching one cell per rune of its query text.
	curCols, curRows := term.Width(), term.Height()
	grid := make([][]runner.ScreenCell, 0, curRows)
	for y := 0; y < curRows; y++ {
		row := make([]runner.ScreenCell, 0, curCols)
		for x := 0; x < curCols; x++ {
			c := term.CellAt(x, y)
			if c == nil || c.Width == 0 {
				continue
			}
			row = append(row, glyphCell(c))
		}
		grid = append(grid, row)
	}

	lines := make([]string, len(grid))
	for i, row := range grid {
		var b strings.Builder
		for _, c := range row {
			b.WriteRune(c.Rune)
		}
		lines[i] = strings.TrimRight(b.String(), " \t")
	}
	// Drop trailing blank rows: a 24-row screen showing two lines snapshots
	// as two lines, not twenty-four. The cell grid is trimmed the same way, so
	// row N of the text is row N of the cells.
	end := len(lines)
	for end > 0 && lines[end-1] == "" {
		end--
	}
	return strings.Join(lines[:end], "\n"), grid[:end]
}

// glyphCell converts one emulator cell into the cell the assertion layer reads.
// A cell the program never wrote carries empty content, which renders as a space
// — matching what the screen shows — and is not a character anyone asserts on.
// Only the first rune of the grapheme cluster is kept in Rune, so the attrs
// matcher matches one cell per rune of its query text; the plain-text screen is
// built from the same runes.
func glyphCell(c *uv.Cell) runner.ScreenCell {
	r := ' '
	if c.Content != "" {
		r, _ = utf8.DecodeRuneInString(c.Content)
	}
	attrs := c.Style.Attrs
	return runner.ScreenCell{
		Rune:      r,
		FG:        colorToIndex(c.Style.Fg),
		BG:        colorToIndex(c.Style.Bg),
		Bold:      attrs&uv.AttrBold != 0,
		Italic:    attrs&uv.AttrItalic != 0,
		Underline: c.Style.Underline != uv.UnderlineNone,
		Reverse:   attrs&uv.AttrReverse != 0,
		Blink:     attrs&(uv.AttrBlink|uv.AttrRapidBlink) != 0,
	}
}

// colorToIndex maps an emulator color to atago's palette-index model (#382): a
// cell with no color set reads as DefaultColor, an ANSI (0..15) or xterm-256
// (0..255) color keeps its index, and a 24-bit color is quantized to the nearest
// xterm-256 index — atago's screen model has no truecolor slot, and quantizing
// keeps a truecolor cell assertable rather than dropping its color entirely.
func colorToIndex(c color.Color) uint32 {
	switch v := c.(type) {
	case nil:
		return runner.DefaultColor
	case ansi.BasicColor:
		return uint32(v)
	case ansi.IndexedColor:
		return uint32(v)
	default:
		return uint32(ansi.Convert256(c))
	}
}

// writeTranscript feeds the transcript to the emulator, containing any panic
// from its escape parser. The transcript is arbitrary bytes chosen by the
// program under test, and a crash there must not take down the whole atago
// process mid-suite. The shapes that make an emulator loop for minutes on an
// enormous CSI count are defused up front by sanitizeTranscript, which preserves
// the rest of the frame; this recover is the backstop for whatever shape the
// fuzzer has not met yet. On panic the screen state built so far still renders,
// so the assertion compares against everything drawn before the malformed
// sequence.
func writeTranscript(term *vt.Emulator, transcript []byte) {
	defer func() { _ = recover() }()
	_, _ = term.Write(transcript)
}

// maxCSIParamDigits bounds a CSI numeric parameter before the transcript
// reaches vt10x: any digit run longer than this is clamped to all-nines.
// Legitimate parameters never get near it — rows/cols top out at 4 digits and
// SGR components at 3 — but several vt10x handlers loop PARAM times (CBT/CHT
// tab stepping, scroll counts), so an adversarial "CSI 80111111110 Z" would
// otherwise spin the emulator for minutes. 9999 iterations are instant.
const maxCSIParamDigits = 4

// sanitizeTranscriptMarks defuses the transcript shapes that crash or hang
// vt10x's parser, mirroring exactly what its Write loop and state machine will
// see:
//
//   - Write silently DROPS lone invalid-UTF-8 bytes without touching parser
//     state, and handleControlCodes makes NUL/ENQ/XON/XOFF/DEL (and friends)
//     transparent to an escape in progress — so "ESC \x00 [" still opens a
//     CSI sequence and the scanner must look through those bytes too.
//   - csiEscape.put runs strconv.Atoi over each ';'-separated parameter, so a
//     '-' (not a valid ECMA-48 parameter byte; a conformant terminal ignores
//     the sequence) yields a NEGATIVE count that panics the slice arithmetic
//     in deleteChars (CSI P / DCH) and insertBlanks (CSI @ / ICH). Sequences
//     carrying one are dropped wholesale, as are sequences whose parameters
//     contain non-ASCII runes (vt10x would truncate those to their low byte).
//   - Loop-per-count handlers (CBT, CHT, scrolls) execute an absurd repeat
//     count one step at a time, hanging the run for minutes; digit runs longer
//     than maxCSIParamDigits are clamped to all-nines.
//   - ESC inside a CSI restarts escape parsing and CAN/SUB reset the parameter
//     buffer without leaving CSI state; the scan follows both so its notion of
//     "the parameters vt10x will Atoi" never drifts from the real parser.
//
// Clean sequences — including OSC runs and truncated trailing escapes — pass
// through byte-for-byte.
//
// It also translates offsets: given raw transcript offsets in ascending order,
// it returns where each one lands in the sanitized output (#379). Mid-session
// resizes need that. The screen is rendered by replaying the transcript in
// pieces and resizing the emulator between them, and the pieces must be cut
// from the SANITIZED bytes, not the raw ones. Sanitizing each raw piece
// separately looks equivalent and is not: a cut inside an escape sequence hides
// that sequence from the scan above, so `\x1b[80111111110Z` split down the
// middle passes through as a truncated CSI plus ordinary text, reassembles
// inside vt10x's stateful parser, and hangs the render on a quadrillion-step
// tab. FuzzRenderScreen finds that in seconds.
//
// Each mark is translated to the output length at the moment the scan reaches
// (or passes) it, which is always a boundary between whole units — so the
// pieces the caller cuts never split a sequence either. Pass a nil marks slice
// to sanitize alone.
func sanitizeTranscriptMarks(b []byte, marks []int) ([]byte, []int) {
	if bytes.IndexByte(b, 0x1b) < 0 {
		// No ESC: nothing can start a CSI sequence, so the bytes pass through
		// unchanged and every offset means the same thing in both buffers.
		translated := make([]int, len(marks))
		for k, m := range marks {
			translated[k] = min(max(m, 0), len(b))
		}
		return b, translated
	}
	out := make([]byte, 0, len(b))
	translated := make([]int, 0, len(marks))
	next := 0
	// flushMarks records every mark the scan has now reached. It runs at the top
	// of each iteration, where `i` sits on the first byte of the next unit and
	// `out` holds only whole units — which is what makes a cut here safe.
	flushMarks := func(i int) {
		for next < len(marks) && marks[next] <= i {
			translated = append(translated, len(out))
			next++
		}
	}
	i := 0
scan:
	for i < len(b) {
		flushMarks(i)
		if b[i] != 0x1b {
			out = append(out, b[i])
			i++
			continue
		}
		// ESC: find the rune that decides the escape kind, looking through the
		// bytes vt10x's Write/handleControlCodes make transparent.
		j := i + 1
		for j < len(b) {
			r, sz := utf8.DecodeRune(b[j:])
			switch {
			case r == utf8.RuneError && sz == 1:
				j++ // invalid byte: dropped by Write before the parser sees it.
				continue
			case r == 0x1b:
				// A second ESC restarts escape parsing: the first is inert.
				out = append(out, b[i:j]...)
				i = j
				continue scan
			case r < 0x20 || r == 0x7f:
				j += sz // control code: handled out-of-band, escape state kept.
				continue
			}
			break
		}
		if j >= len(b) || b[j] != '[' {
			// Not a CSI (or a truncated trailing ESC): copy the ESC and rescan
			// from the next byte, so non-CSI escapes pass through untouched.
			out = append(out, b[i])
			i++
			continue
		}
		// CSI body: scan effective runes until the final byte (0x40..0x7E),
		// tracking exactly the parameter bytes vt10x will accumulate.
		body := make([]byte, 0, 16)
		var controls []byte // side-effect control codes seen inside the sequence
		hasMinus, hasWideRune, hasMidMarker := false, false, false
		finalByte := byte(0)
		k := j + 1
		for k < len(b) {
			r, sz := utf8.DecodeRune(b[k:])
			if r == utf8.RuneError && sz == 1 {
				k++ // transparent to vt10x's parser, transparent to the scan.
				continue
			}
			if r == 0x1b {
				// ESC restarts escape parsing: the sequence so far can never
				// dispatch, so drop its parameter bytes — copying them verbatim
				// would leave an OPEN CSI in the output if the aborting escape
				// later gets dropped as malformed, and then ordinary text (say a
				// final-range 'Z') would finalize the stale parameters into a
				// quadrillion-step CBT (found by FuzzRenderScreen). Only the
				// embedded control codes vt10x already executed are kept.
				out = append(out, controls...)
				i = k
				continue scan
			}
			if r == 0x18 || r == 0x1a {
				// CAN/SUB reset the parameter buffer but STAY in CSI state.
				body = body[:0]
				hasMinus, hasWideRune, hasMidMarker = false, false, false
				k += sz
				continue
			}
			if r < 0x20 || r == 0x7f {
				// Other control codes are transparent to the CSI state but DO
				// execute (tab, CR, LF move the cursor mid-sequence); remember
				// them so a dropped sequence still replays its side effects.
				controls = append(controls, byte(r&0x7f)) // r < 0x20 or == 0x7f here
				k += sz
				continue
			}
			if r > 0x7e {
				// vt10x truncates the rune to its low byte — nonsense that can
				// even finalize the sequence. Mirror the boundary, drop later.
				hasWideRune = true
				if low := byte(r & 0xff); low >= 0x40 && low <= 0x7e {
					finalByte = low
					k += sz
					break
				}
				k += sz
				continue
			}
			c := byte(r)
			if c >= 0x40 && c <= 0x7e {
				finalByte = c
				k += sz
				break
			}
			if c == '-' {
				hasMinus = true
			}
			// A private-marker byte (< = > ?) is well-formed only as the FIRST
			// parameter byte (CSI ? 25 h, CSI < 0 ; 0 M). One that appears after a
			// parameter has begun makes the sequence malformed — a conformant
			// terminal ignores it — and x/vt's parser handles the malformed shape
			// (CSI 999 ? 999 ? 999 X) in time quadratic in the parameters, seconds
			// for a few thousand. Drop it like a negative parameter.
			if c >= '<' && c <= '?' && len(body) > 0 {
				hasMidMarker = true
			}
			body = append(body, c)
			k += sz
		}
		switch {
		case finalByte == 0:
			// Truncated trailing CSI: it can never dispatch, copy verbatim.
			out = append(out, b[i:k]...)
		case hasMinus || hasWideRune || hasMidMarker:
			// Malformed parameters: a conformant terminal ignores the whole
			// sequence, so drop it — replaying only the control codes it
			// carried — and keep the surrounding frame intact.
			out = append(out, controls...)
		case len(clampDigitRuns(body)) != len(body):
			// An oversized repeat count: replay embedded control codes, then
			// re-emit the sequence with the digit runs clamped.
			out = append(out, controls...)
			out = append(out, 0x1b, '[')
			out = append(out, clampDigitRuns(body)...)
			out = append(out, finalByte)
		default:
			// Clean sequence: byte-for-byte, side-effect bytes included.
			out = append(out, b[i:k]...)
		}
		i = k
	}
	// Anything still pending sat at or past the end of the transcript.
	for next < len(marks) {
		translated = append(translated, len(out))
		next++
	}
	return out, translated
}

// clampDigitRuns truncates every digit run longer than maxCSIParamDigits to its
// first maxCSIParamDigits digits, bounding the value a loop-per-count CSI handler
// can be asked to iterate while leaving every legitimate parameter untouched.
//
// Truncating rather than replacing with all-nines matters: a zero-padded run
// ("00000") is the harmless value 0, and rewriting it to 9999 would MANUFACTURE
// an expensive count out of nothing — which is exactly how a fuzzer turned
// "\x1b[00000?...X" into a multi-second render.
func clampDigitRuns(body []byte) []byte {
	out := make([]byte, 0, len(body))
	for i := 0; i < len(body); {
		c := body[i]
		if c < '0' || c > '9' {
			out = append(out, c)
			i++
			continue
		}
		j := i
		for j < len(body) && body[j] >= '0' && body[j] <= '9' {
			j++
		}
		if j-i > maxCSIParamDigits {
			out = append(out, body[i:i+maxCSIParamDigits]...)
		} else {
			out = append(out, body[i:j]...)
		}
		i = j
	}
	return out
}
