package conpty

import (
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// Virtual-key codes and control-key-state flags of the Windows console input
// model, as a key press carries them in a KEY_EVENT_RECORD.
const (
	vkBack     = 0x08
	vkTab      = 0x09
	vkReturn   = 0x0D
	vkEscape   = 0x1B
	vkSpace    = 0x20
	vkPrior    = 0x21
	vkNext     = 0x22
	vkEnd      = 0x23
	vkHome     = 0x24
	vkLeft     = 0x25
	vkUp       = 0x26
	vkRight    = 0x27
	vkDown     = 0x28
	vkInsert   = 0x2D
	vkDelete   = 0x2E
	vkF1       = 0x70
	vkOEMMinus = 0xBD
	vkOEM5     = 0xDC // the \ | key on a US layout
	vkOEM6     = 0xDD // the ] } key on a US layout

	stateLeftAlt  = 0x0002
	stateLeftCtrl = 0x0008
	stateShift    = 0x0010
	stateEnhanced = 0x0100
)

// win32Key is one key press in win32-input-mode terms.
type win32Key struct {
	vk    uint16
	char  uint16
	state uint32
}

// press appends the key going down and coming back up, as a keyboard does. The
// release is not decoration: the host folds identical key-down records that
// follow each other into one, so three Esc presses without releases between
// them reached the program as two.
func (k win32Key) press(out []byte) []byte {
	for _, down := range []byte{'1', '0'} {
		out = append(out, "\x1b["...)
		out = strconv.AppendUint(out, uint64(k.vk), 10)
		out = append(out, ";0;"...)
		out = strconv.AppendUint(out, uint64(k.char), 10)
		out = append(out, ';', down, ';')
		out = strconv.AppendUint(out, uint64(k.state), 10)
		out = append(out, ";1_"...)
	}
	return out
}

// Win32InputSend encodes what a terminal user types, given as the bytes an
// xterm-class terminal would send, as the key presses Windows Terminal would
// hand the bundled console host for the same keys.
//
// That host parses raw input as a terminal's and gets two things wrong for a
// program reading console input records: a lone ESC is held as the start of an
// escape sequence (so an Esc press, and the key after it, never arrive), and a
// character outside the BMP is dropped. A key press in win32-input-mode is not
// parsed, it becomes one KEY_EVENT record, so every key goes out that way:
//
//   - text as one key press per UTF-16 code unit, a character outside the BMP
//     as its two surrogates, each with the virtual-key code of its key where a
//     US layout has one;
//   - a lone ESC as VK_ESCAPE, CR as VK_RETURN, TAB as VK_TAB, DEL as VK_BACK,
//     and the other C0 controls as their Ctrl chords;
//   - ESC followed by a letter, CR or DEL as that key with Alt;
//   - the CSI and SS3 sequences of the cursor, editing and function keys, with
//     xterm's modifier parameter, as those keys with their modifiers.
//
// Any other escape sequence (a bracketed-paste marker, a mouse report, a focus
// event) goes out unchanged, for the host to parse as it would from a terminal.
// Each key goes down and comes back up, as it does from Windows Terminal.
func Win32InputSend(p []byte) []byte {
	out := make([]byte, 0, len(p)*20)
	s := string(p)
	for i := 0; i < len(s); {
		if s[i] != 0x1b {
			r, size := utf8.DecodeRuneInString(s[i:])
			for _, u := range utf16.Encode([]rune{r}) {
				out = keyForUnit(u).press(out)
			}
			i += size
			continue
		}
		n, keys, raw := escapeAt(s[i:])
		for _, k := range keys {
			out = k.press(out)
		}
		out = append(out, raw...)
		i += n
	}
	return out
}

// escapeAt reads the escape at the start of s (s[0] is ESC) and reports how
// many bytes it spans and either the keys it stands for or, for a sequence that
// is not a key, the bytes to pass through.
func escapeAt(s string) (int, []win32Key, string) {
	esc := win32Key{vk: vkEscape, char: 0x1b}
	if len(s) == 1 {
		return 1, []win32Key{esc}, ""
	}
	switch c := s[1]; {
	case c == '[':
		if n, ok := csiLen(s); ok {
			if k, ok := csiKey(s[2:n]); ok {
				return n, []win32Key{k}, ""
			}
			return n, nil, s[:n]
		}
	case c == 'O' && len(s) >= 3 && s[2] >= 'P' && s[2] <= 'S':
		return 3, []win32Key{{vk: vkF1 + uint16(s[2]-'P')}}, ""
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c == '\r', c == 0x7f:
		k := keyForUnit(uint16(c))
		k.state |= stateLeftAlt
		return 2, []win32Key{k}, ""
	}
	return 1, []win32Key{esc}, ""
}

// csiLen reports the length of the CSI sequence at the start of s, or false
// when s ends before its final byte.
func csiLen(s string) (int, bool) {
	for i := 2; i < len(s); i++ {
		if s[i] >= 0x40 && s[i] <= 0x7e {
			return i + 1, true
		}
		if s[i] < 0x20 || s[i] > 0x3f {
			return 0, false
		}
	}
	return 0, false
}

// csiKey maps the body of a CSI sequence (after `ESC [`, final byte included)
// to the key an xterm-class terminal sends it for.
func csiKey(body string) (win32Key, bool) {
	final := body[len(body)-1]
	params := strings.Split(body[:len(body)-1], ";")
	num := func(i int) (int, bool) {
		if i >= len(params) {
			return 1, true
		}
		if params[i] == "" {
			return 1, true
		}
		n, err := strconv.Atoi(params[i])
		return n, err == nil
	}
	mods, ok := num(1)
	if !ok || mods < 1 || mods > 16 {
		return win32Key{}, false
	}
	state := modifierState(mods)
	var k win32Key
	switch final {
	case 'A', 'B', 'C', 'D', 'H', 'F':
		k = win32Key{vk: map[byte]uint16{'A': vkUp, 'B': vkDown, 'C': vkRight, 'D': vkLeft, 'H': vkHome, 'F': vkEnd}[final], state: stateEnhanced}
	case 'P', 'Q', 'R', 'S':
		k = win32Key{vk: vkF1 + uint16(final-'P')}
	case 'Z':
		k = win32Key{vk: vkTab, char: '\t', state: stateShift}
	case '~':
		code, ok := num(0)
		if !ok {
			return win32Key{}, false
		}
		vk, ok := tildeKeys[code]
		if !ok {
			return win32Key{}, false
		}
		k = win32Key{vk: vk}
		if vk < vkF1 {
			k.state = stateEnhanced
		}
	case 'u':
		cp, ok := num(0)
		if !ok || cp < 0x20 || cp > 0x7e {
			return win32Key{}, false
		}
		k = keyForUnit(uint16(cp))
	default:
		return win32Key{}, false
	}
	k.state |= state
	return k, true
}

// tildeKeys maps the first parameter of a `CSI <n> ~` key to its key.
var tildeKeys = map[int]uint16{
	2: vkInsert, 3: vkDelete, 5: vkPrior, 6: vkNext,
	15: vkF1 + 4, 17: vkF1 + 5, 18: vkF1 + 6, 19: vkF1 + 7, 20: vkF1 + 8, 21: vkF1 + 9,
	23: vkF1 + 10, 24: vkF1 + 11,
}

// modifierState decodes xterm's modifier parameter, 1 plus a bitmask of
// Shift (1), Alt (2) and Ctrl (4).
func modifierState(m int) uint32 {
	var s uint32
	bits := m - 1
	if bits&1 != 0 {
		s |= stateShift
	}
	if bits&2 != 0 {
		s |= stateLeftAlt
	}
	if bits&4 != 0 {
		s |= stateLeftCtrl
	}
	return s
}

// keyForUnit is the key press that types one UTF-16 code unit.
func keyForUnit(u uint16) win32Key {
	switch {
	case u == '\r':
		return win32Key{vk: vkReturn, char: u}
	case u == '\t':
		return win32Key{vk: vkTab, char: u}
	case u == 0x7f:
		return win32Key{vk: vkBack, char: vkBack}
	case u == 0:
		return win32Key{vk: vkSpace, state: stateLeftCtrl}
	case u >= 1 && u <= 26:
		return win32Key{vk: 'A' + u - 1, char: u, state: stateLeftCtrl}
	case u == 0x1b:
		return win32Key{vk: vkEscape, char: u}
	case u == 0x1c:
		return win32Key{vk: vkOEM5, char: u, state: stateLeftCtrl}
	case u == 0x1d:
		return win32Key{vk: vkOEM6, char: u, state: stateLeftCtrl}
	case u == 0x1e:
		return win32Key{vk: '6', char: u, state: stateLeftCtrl | stateShift}
	case u == 0x1f:
		return win32Key{vk: vkOEMMinus, char: u, state: stateLeftCtrl | stateShift}
	case u == ' ':
		return win32Key{vk: vkSpace, char: u}
	case u == '-':
		return win32Key{vk: vkOEMMinus, char: u}
	case u >= 'a' && u <= 'z':
		return win32Key{vk: u - 'a' + 'A', char: u}
	case u >= 'A' && u <= 'Z':
		return win32Key{vk: u, char: u, state: stateShift}
	case u >= '0' && u <= '9':
		return win32Key{vk: u, char: u}
	}
	return win32Key{char: u}
}
