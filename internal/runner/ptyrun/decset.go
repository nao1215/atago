package ptyrun

import (
	"strconv"

	"github.com/nao1215/atago/internal/diag"
)

// DEC private modes atago tracks from the program's own output. A terminal
// feature like bracketed paste exists only once the application asks for it, so
// the request in the output stream is the only honest signal that sending the
// corresponding input is meaningful (#378).
const (
	// decsetBracketedPaste wraps pasted input in the PasteStart/PasteEnd
	// markers. A program that never requests it reads those markers as
	// ordinary characters.
	decsetBracketedPaste = 2004
	// Mouse tracking modes (#381). A program requests one of these to be told
	// about clicks at all: 1000 reports button press/release, 1002 adds motion
	// while a button is down, and 1003 reports every motion.
	decsetMouseX11    = 1000
	decsetMouseButton = 1002
	decsetMouseAny    = 1003
	// decsetMouseSGR is the ENCODING the reports arrive in. Without it a
	// terminal falls back to the legacy X10 form, whose coordinates are single
	// bytes and cannot address a screen wider than 223 columns; atago only
	// sends SGR, so a program tracking without it is one atago cannot drive.
	decsetMouseSGR = 1006
)

// checkMouseMode reports why a mouse event cannot be delivered, or nil when it
// can (#381). The two answers are different mistakes and deserve different
// messages: a program that never asked about the mouse wants a keyboard-driven
// spec, while one tracking in the legacy encoding is a program atago cannot
// drive at all — worth knowing about rather than silently mis-sending.
func checkMouseMode(term *transcriptDrain, command string, idx int) error {
	tracking := term.modeEnabled(decsetMouseX11) ||
		term.modeEnabled(decsetMouseButton) ||
		term.modeEnabled(decsetMouseAny)
	switch {
	case !tracking:
		return diag.TerminalModeMismatch.Errorf("pty %q: session[%d] sends a mouse event, but the program has not enabled mouse reporting "+
			"(it never wrote ESC [?1000h, ESC [?1002h, or ESC [?1003h, or turned tracking back off). "+
			"Programs that only read the keyboard want a key send instead; "+
			"if this one does enable tracking, wait for it with an expect or expect_screen before clicking",
			command, idx)
	case !term.modeEnabled(decsetMouseSGR):
		return diag.TerminalModeMismatch.Errorf("pty %q: session[%d] sends a mouse event, but the program tracks the mouse in the legacy X10 "+
			"encoding (it enabled tracking without ESC [?1006h). atago only sends SGR reports, whose "+
			"coordinates are decimal and can address the whole screen; the X10 form packs them into "+
			"single bytes and cannot. Modern TUI toolkits request SGR — please open an issue if you "+
			"hit a real program that does not",
			command, idx)
	}
	return nil
}

// decsetMode is one observed DEC private mode transition.
type decsetMode struct {
	Param   int
	Enabled bool
}

// decsetScanner incrementally recognizes DEC private mode set/reset requests —
// `CSI ? Pm h` and `CSI ? Pm l` — in the program's output. It keeps state
// across calls because a pty read can split an escape sequence at any byte, and
// it handles the multi-parameter form (`CSI ? 1049 ; 2004 h`, which is how most
// TUI toolkits turn several modes on at once) by reporting one transition per
// parameter.
//
// The recognition is deliberately narrow: anything that is not a well-formed
// private-mode sequence resets the scanner rather than being guessed at. A
// missed request costs an actionable error message; a wrongly inferred one
// would let atago send input the program cannot interpret.
type decsetScanner struct {
	state decsetState
	buf   []byte
}

type decsetState uint8

const (
	decsetNormal decsetState = iota
	decsetESC
	decsetCSI
	decsetPrivate
)

// maxDECSETParamBytes bounds the parameter buffer. Real requests are a handful
// of bytes; a longer run is not a private-mode sequence atago should be
// accumulating.
const maxDECSETParamBytes = 64

func (s *decsetScanner) consume(chunk []byte) []decsetMode {
	var out []decsetMode
	for _, b := range chunk {
		switch s.state {
		case decsetNormal:
			if b == 0x1b {
				s.state = decsetESC
			}
		case decsetESC:
			switch b {
			case '[':
				s.state = decsetCSI
				s.buf = s.buf[:0]
			case 0x1b:
				// A new ESC restarts escape parsing.
			default:
				s.state = decsetNormal
			}
		case decsetCSI:
			switch b {
			case '?':
				s.state = decsetPrivate
			case 0x1b:
				s.state = decsetESC
			default:
				// Not a private-mode sequence; nothing here to track.
				s.state = decsetNormal
			}
		case decsetPrivate:
			switch {
			case b == 0x1b:
				s.state = decsetESC
				s.buf = s.buf[:0]
			case (b >= '0' && b <= '9') || b == ';':
				if len(s.buf) >= maxDECSETParamBytes {
					s.state = decsetNormal
					s.buf = s.buf[:0]
					continue
				}
				s.buf = append(s.buf, b)
			case b >= 0x40 && b <= 0x7e:
				if b == 'h' || b == 'l' {
					out = append(out, parseDECSETParams(s.buf, b == 'h')...)
				}
				s.state = decsetNormal
				s.buf = s.buf[:0]
			default:
				s.state = decsetNormal
				s.buf = s.buf[:0]
			}
		}
	}
	return out
}

// parseDECSETParams splits a `;`-separated parameter list into one transition
// per parameter. An empty or unparsable element is skipped rather than guessed
// at: a terminal treats a missing parameter as a default, and there is no
// default private mode worth inventing.
func parseDECSETParams(body []byte, enabled bool) []decsetMode {
	var out []decsetMode
	start := 0
	for i := 0; i <= len(body); i++ {
		if i < len(body) && body[i] != ';' {
			continue
		}
		if field := body[start:i]; len(field) > 0 {
			if n, err := strconv.Atoi(string(field)); err == nil {
				out = append(out, decsetMode{Param: n, Enabled: enabled})
			}
		}
		start = i + 1
	}
	return out
}
