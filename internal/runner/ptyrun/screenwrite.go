package ptyrun

import (
	"unicode/utf8"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/vt"
)

// crlf is the wrap atago injects on the emulator's behalf: CR returns to column
// 0 and LF indexes down (scrolling at the bottom margin, honoring the scroll
// region), which is exactly what a terminal's autowrap does. Going through the
// emulator's own control codes rather than absolute cursor addressing keeps
// scroll regions and origin mode behaving as the program set them, and both
// codes clear the emulator's pending-wrap state, so the wrap can never fire
// twice.
var crlf = []byte("\r\n")

// screenWriter feeds a sanitized transcript to the terminal emulator, repairing
// the two places where the emulator's grapheme handling differs from a real
// terminal. Both are upstream (github.com/charmbracelet/x/vt), still present in
// its latest revision, and both are corrected here rather than lived with,
// because each one silently changes what a `screen:` assertion compares against:
//
//   - Pending wrap is armed only from the LAST column (#503). A wide character
//     reaches the right margin from the second-to-last one, so after `日本` fills
//     a four-column row the emulator leaves wrap disarmed: the next character is
//     written INSIDE the row, overwriting the second half of `本` and blanking
//     its first half, and a wide character that no longer fits is dropped
//     entirely. This writer arms the wrap itself for a wide character that
//     reaches the margin, and wraps one that does not fit.
//
//   - An ASCII character is committed to the screen the moment it arrives
//     (#505), so a combining mark that follows can never join it and is dropped:
//     a program that writes a decomposed `é` renders `e`. This writer keeps the
//     whole grapheme cluster and puts it back into the cell the base landed in.
//     A non-ASCII base (`か` + dakuten) and a ZWJ emoji sequence go through the
//     emulator's grapheme buffer and were never affected.
//
// The transcript is decoded with the same library the emulator parses it with
// (x/ansi), so the units this writer sees are the units the emulator will see.
// Everything that needs no correction is batched and handed over untouched, so
// ordinary output reaches the emulator exactly as before.
type screenWriter struct {
	term   *vt.Emulator
	parser *ansi.Parser
	// state is the decoder's escape-sequence state, carried across writes
	// because a resize splits one transcript into several (#379).
	state byte
	// autowrap tracks DECAWM. With autowrap off a terminal keeps overwriting
	// the last column instead of wrapping, so every correction below is off too.
	autowrap bool
	// pending is the wrap the emulator failed to arm after a wide character
	// filled the row, and pendingAt where the cursor stood when it was armed:
	// anything that moves the cursor before the next character cancels the
	// wrap, exactly as a terminal cancels its own.
	pending   bool
	pendingAt uv.Position
	batch     []byte
}

func newScreenWriter(term *vt.Emulator) *screenWriter {
	return &screenWriter{term: term, parser: ansi.NewParser(), autowrap: true}
}

// write feeds one piece of the transcript to the emulator, containing any panic
// from its escape parser. The transcript is arbitrary bytes chosen by the
// program under test, and a crash there must not take down the whole atago
// process mid-suite. The shapes that make an emulator loop for minutes on an
// enormous CSI count are defused up front by sanitizeTranscript, which preserves
// the rest of the frame; this recover is the backstop for whatever shape the
// fuzzer has not met yet. On panic the screen state built so far still renders,
// so the assertion compares against everything drawn before the malformed
// sequence.
func (w *screenWriter) write(transcript []byte) {
	defer func() { _ = recover() }()
	defer w.flush()
	for len(transcript) > 0 {
		seq, width, n, next := ansi.DecodeSequence(transcript, w.state, w.parser)
		if n <= 0 {
			// The decoder consumed nothing (an input shape it cannot advance
			// over). Hand the byte to the emulator and move on: a render that
			// stops early is a wrong screen, and a render that does not move is
			// a hung suite.
			w.batch = append(w.batch, transcript[0])
			transcript = transcript[1:]
			continue
		}
		w.state = next
		if width <= 0 {
			w.control(seq)
			transcript = transcript[n:]
			continue
		}
		// The decoder returns an ASCII character on its own, splitting the very
		// cluster whose tail the emulator then drops. Re-read it as a cluster so
		// a combining mark stays with its base.
		//
		// Only a cluster that is valid UTF-8 is taken: the grapheme segmenter
		// will happily carry a stray continuation byte along with the character
		// before it, and the emulator DROPS invalid UTF-8 rather than printing
		// it — so keeping such a "cluster" would put bytes on the screen that
		// the terminal never showed (found by FuzzRenderScreen).
		cluster, clusterWidth := seq, width
		if len(seq) == 1 && isASCIIPrintable(seq[0]) {
			if c, cw := ansi.FirstGraphemeCluster(transcript, ansi.GraphemeWidth); len(c) > len(seq) && utf8.Valid(c) {
				cluster, clusterWidth, n = c, cw, len(c)
			}
		}
		w.printable(cluster, clusterWidth)
		transcript = transcript[n:]
	}
}

// control passes an escape sequence or control byte through untouched, noting
// only the ones that change whether the corrections apply at all.
func (w *screenWriter) control(seq []byte) {
	w.trackAutoWrap(seq)
	w.batch = append(w.batch, seq...)
}

// trackAutoWrap follows DECAWM (CSI ? 7 h / l) and the full reset that restores
// it. The emulator keeps this mode privately, so the one place that can know it
// is the stream both of them read.
func (w *screenWriter) trackAutoWrap(seq []byte) {
	if len(seq) == 2 && seq[0] == ansi.ESC && seq[1] == 'c' {
		w.autowrap = true // RIS restores every mode to its default
		return
	}
	cmd := ansi.Cmd(w.parser.Command())
	final := cmd.Final()
	if cmd.Prefix() != '?' || (final != 'h' && final != 'l') {
		return
	}
	for _, p := range w.parser.Params() {
		if p.Param(0) == 7 {
			w.autowrap = final == 'h'
		}
	}
}

// printable writes one grapheme cluster, correcting the right margin around it.
func (w *screenWriter) printable(cluster []byte, width int) {
	repair := needsMarkRepair(cluster)
	if repair {
		// An ASCII base is one column, and the base is all that reaches the
		// emulator below — so the margin arithmetic follows the base, whatever
		// width the segmenter reports for the cluster as a whole.
		width = 1
	}
	wide := w.autowrap && width > 1
	if !repair && !w.pending && !wide {
		// Nothing to correct around this character: batch it and let the
		// emulator lay it out exactly as it does today.
		w.batch = append(w.batch, cluster...)
		return
	}
	// Everything below reads the emulator's cursor, so the batch has to be in
	// it first.
	w.flush()
	if w.pending {
		// A sequence that moved the cursor since the wrap was armed cancels it,
		// the way a terminal cancels its own pending wrap.
		if w.term.CursorPosition() == w.pendingAt {
			w.writeNow(crlf)
		}
		w.pending = false
	}
	pos := w.term.CursorPosition()
	cols := w.term.Width()
	if wide && pos.X+width > cols {
		// A wide character that no longer fits: a terminal starts it on the
		// next row, where the emulator drops it.
		w.writeNow(crlf)
		pos = w.term.CursorPosition()
	}
	if repair {
		// Only the base reaches the emulator. The marks it would drop anyway
		// are not inert on the way out: a mark arriving while pending wrap is
		// armed makes the emulator index — moving the cursor to the next row,
		// and scrolling the screen when the base sat on the last one — for a
		// grapheme that a terminal does not move the cursor for at all.
		w.writeNow(cluster[:1])
		w.repairCell(cluster, width)
	} else {
		w.writeNow(cluster)
	}
	if wide && pos.X+width >= cols {
		w.pending = true
		w.pendingAt = w.term.CursorPosition()
	}
}

// repairCell puts the whole cluster into the cell the base just landed in.
//
// The cell is found from the cursor AFTER the write rather than from where the
// write started, so a base that wrapped to the next row (scrolling the screen
// with it) is still found: the cursor sits just past the base, or still on it
// when the base filled the last column. The base character has to actually be
// there — a cluster the emulator placed some other way is left alone rather
// than overwritten with content that would not belong to it.
func (w *screenWriter) repairCell(cluster []byte, width int) {
	baseChar := string(cluster[:1]) // an ASCII base is one byte; see needsMarkRepair
	at := w.term.CursorPosition()
	for _, x := range []int{at.X - width, at.X} {
		if x < 0 {
			continue
		}
		c := w.term.CellAt(x, at.Y)
		if c == nil || c.Content != baseChar {
			continue
		}
		fixed := *c
		fixed.Content = string(cluster)
		w.term.SetCell(x, at.Y, &fixed)
		return
	}
}

// needsMarkRepair reports whether the emulator will drop part of this cluster:
// it commits an ASCII base to the screen before the rest of the cluster has
// arrived, so any cluster longer than its ASCII base loses the remainder.
func needsMarkRepair(cluster []byte) bool {
	return len(cluster) > 1 && isASCIIPrintable(cluster[0])
}

// isASCIIPrintable reports whether b takes the emulator's (and the decoder's)
// single-character fast path.
func isASCIIPrintable(b byte) bool {
	return b >= 0x20 && b < 0x7f
}

// flush hands the batched units to the emulator. Units are only ever appended
// whole, so a flush never splits an escape sequence or a grapheme cluster —
// which matters, because the emulator resolves its grapheme buffer at the end
// of every write.
func (w *screenWriter) flush() {
	if len(w.batch) == 0 {
		return
	}
	w.writeNow(w.batch)
	w.batch = w.batch[:0]
}

func (w *screenWriter) writeNow(b []byte) {
	_, _ = w.term.Write(b)
}
