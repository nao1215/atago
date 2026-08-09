package ptyrun

import "strconv"

// DEC private modes atago tracks from the program's own output. A terminal
// feature like bracketed paste exists only once the application asks for it, so
// the request in the output stream is the only honest signal that sending the
// corresponding input is meaningful (#378).
const (
	// decsetBracketedPaste wraps pasted input in the PasteStart/PasteEnd
	// markers. A program that never requests it reads those markers as
	// ordinary characters.
	decsetBracketedPaste = 2004
)

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
