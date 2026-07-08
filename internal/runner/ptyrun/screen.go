package ptyrun

import (
	"bytes"
	"strings"
	"unicode/utf8"

	"github.com/hinshun/vt10x"
	"github.com/nao1215/atago/internal/spec"
)

// RenderScreen replays a pty transcript through a vt10x terminal emulator and
// returns the final rendered screen as plain text (#27): what the user
// actually SEES after every cursor move, overwrite, and erase — the signal a
// raw transcript scatters across redraws. Trailing whitespace is stripped per
// line and trailing blank lines are dropped; colors and attributes are out of
// scope in v1.
func RenderScreen(transcript []byte, p *spec.PTY) string {
	rows, cols := defaultRows, defaultCols
	if p.Rows > 0 {
		rows = p.Rows
	}
	if p.Cols > 0 {
		cols = p.Cols
	}
	term := vt10x.New(vt10x.WithSize(cols, rows))
	writeTranscript(term, sanitizeTranscript(transcript))

	lines := strings.Split(term.String(), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	// Drop trailing blank rows: a 24-row screen showing two lines snapshots
	// as two lines, not twenty-four.
	end := len(lines)
	for end > 0 && lines[end-1] == "" {
		end--
	}
	return strings.Join(lines[:end], "\n")
}

// writeTranscript feeds the transcript to the emulator, containing panics
// from vt10x's escape parser. The transcript is arbitrary bytes chosen by the
// program under test, and unmaintained vt10x runs strconv.Atoi over CSI
// parameters and feeds the result straight into slice arithmetic — a crash
// there must not take down the whole atago process mid-suite. The known-bad
// shapes are defused up front by sanitizeTranscript, which preserves the rest
// of the frame; this recover is the backstop for whatever shape the fuzzer has
// not met yet. On panic the screen state built so far still renders — vt10x
// mutates cells as it parses and releases its lock via defer during unwind —
// so the assertion compares against everything drawn before the malformed
// sequence.
func writeTranscript(term vt10x.Terminal, transcript []byte) {
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

// sanitizeTranscript defuses the transcript shapes that crash or hang vt10x's
// parser, mirroring exactly what its Write loop and state machine will see:
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
func sanitizeTranscript(b []byte) []byte {
	if bytes.IndexByte(b, 0x1b) < 0 {
		return b // no ESC: nothing can start a CSI sequence.
	}
	out := make([]byte, 0, len(b))
	i := 0
scan:
	for i < len(b) {
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
		hasMinus, hasWideRune := false, false
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
				hasMinus, hasWideRune = false, false
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
			body = append(body, c)
			k += sz
		}
		switch {
		case finalByte == 0:
			// Truncated trailing CSI: it can never dispatch, copy verbatim.
			out = append(out, b[i:k]...)
		case hasMinus || hasWideRune:
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
	return out
}

// clampDigitRuns rewrites every digit run longer than maxCSIParamDigits to
// all-nines of that width, bounding the work a loop-per-count CSI handler can
// be asked to do while leaving every legitimate parameter untouched.
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
			out = append(out, bytes.Repeat([]byte{'9'}, maxCSIParamDigits)...)
		} else {
			out = append(out, body[i:j]...)
		}
		i = j
	}
	return out
}
