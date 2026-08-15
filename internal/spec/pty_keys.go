package spec

import (
	"strings"
)

// ptyKeySequences maps each named key (#26) to the xterm byte sequence it
// transmits. Documented bytes: enter=\r, tab=\t, esc=\x1b, space=" ",
// backspace=\x7f (DEL, the modern erase), delete=\x1b[3~, insert=\x1b[2~,
// arrows up/down/right/left=\x1b[A/B/C/D, home=\x1b[H, end=\x1b[F,
// pageup=\x1b[5~, pagedown=\x1b[6~, f1-f4=\x1bOP..\x1bOS,
// f5..f12=\x1b[15~,[17~..[21~,[23~,[24~, ctrl-a..ctrl-z=0x01..0x1a,
// plus the punctuation aliases terminals conventionally expose for the
// remaining C0 controls: ctrl-space/ctrl-@=NUL, ctrl-[=ESC, ctrl-\=FS,
// ctrl-]=GS, ctrl-^=RS, ctrl-_=US. For modifier combos whose physical key does
// not have a stable legacy C0 byte (e.g. Ctrl+-), use the terminal's CSI-u key
// event instead so modern TUIs like Yazi see the intended modified key. ctrl-d
// therefore stays the readable alias for the empty-send EOF rule.
//
// Three more families cover the chords real TUIs bind (#376), each the sequence
// an xterm-class terminal sends for that physical key:
//
//   - shift-tab=\x1b[Z (CSI Z, "backtab"), the reverse-focus key every
//     form-style TUI reads. `backtab` is an accepted alias for the same bytes.
//   - alt-<letter>=\x1b<letter>, the ESC prefix a terminal transmits for a Meta
//     chord (readline word operations, helix/emacs-style bindings), plus
//     alt-enter=\x1b\r and alt-backspace=\x1b\x7f. Letters only: a Meta digit
//     has no encoding stable enough to promise.
//   - Modified arrows, in xterm's CSI 1 ; <modifier> <final> form —
//     ctrl-<arrow>=\x1b[1;5A..D (word-wise cursor movement) and
//     shift-<arrow>=\x1b[1;2A..D (selection movement).
var ptyKeySequences = func() map[string]string {
	m := map[string]string{
		"enter":     "\r",
		"tab":       "\t",
		"esc":       "\x1b",
		"space":     " ",
		"backspace": "\x7f",
		"delete":    "\x1b[3~",
		"insert":    "\x1b[2~",
		"shift-tab": "\x1b[Z",
		"backtab":   "\x1b[Z",
		"up":        "\x1b[A",
		"down":      "\x1b[B",
		"right":     "\x1b[C",
		"left":      "\x1b[D",
		"home":      "\x1b[H",
		"end":       "\x1b[F",
		"pageup":    "\x1b[5~",
		"pagedown":  "\x1b[6~",
		"f1":        "\x1bOP",
		"f2":        "\x1bOQ",
		"f3":        "\x1bOR",
		"f4":        "\x1bOS",
		"f5":        "\x1b[15~",
		"f6":        "\x1b[17~",
		"f7":        "\x1b[18~",
		"f8":        "\x1b[19~",
		"f9":        "\x1b[20~",
		"f10":       "\x1b[21~",
		"f11":       "\x1b[23~",
		"f12":       "\x1b[24~",
	}
	for c := byte('a'); c <= 'z'; c++ {
		m["ctrl-"+string(c)] = string([]byte{c - 'a' + 1})
		m["alt-"+string(c)] = "\x1b" + string(c)
	}
	m["alt-enter"] = "\x1b\r"
	m["alt-backspace"] = "\x1b\x7f"
	// Modified arrows share one shape: CSI 1 ; <modifier> <final>, where the
	// modifier is 1+bitmask (shift 1, alt 2, ctrl 4) — so ctrl is 5 and shift 2.
	for i, final := range []string{"A", "B", "C", "D"} {
		name := []string{"up", "down", "right", "left"}[i]
		m["ctrl-"+name] = "\x1b[1;5" + final
		m["shift-"+name] = "\x1b[1;2" + final
	}
	m["ctrl-space"] = "\x00"
	m["ctrl-@"] = "\x00"
	m["ctrl-["] = "\x1b"
	m["ctrl-\\"] = "\x1c"
	m["ctrl-]"] = "\x1d"
	m["ctrl-^"] = "\x1e"
	m["ctrl-_"] = "\x1f"
	// xterm/kitty CSI-u for Ctrl+- (#286): raw 0x1f is Ctrl+_ and does not
	// trigger TUIs that bind the physical hyphen key as a distinct modified key.
	m["ctrl-hyphen"] = "\x1b[45;5u"
	m["ctrl-minus"] = "\x1b[45;5u"
	return m
}()

// ValidPTYKey reports whether name is in the named-key vocabulary (#26).
func ValidPTYKey(name string) bool {
	_, ok := ptyKeySequences[strings.ToLower(strings.TrimSpace(name))]
	return ok
}

// ptyKeyBySequence reverse-maps an xterm byte sequence to its friendly key name
// (#69), preferring the readable name over a ctrl-* alias when a byte is shared
// (e.g. \r is both enter and ctrl-m — enter wins). Built once at init: the
// ctrl-* aliases go in first, then the friendly names overwrite any collision.
var ptyKeyBySequence = func() map[string]string {
	m := make(map[string]string, len(ptyKeySequences))
	m["\x00"] = "ctrl-space"
	for c := byte('a'); c <= 'z'; c++ {
		m[string([]byte{c - 'a' + 1})] = "ctrl-" + string(c)
		// alt-<letter> is two bytes (ESC + the letter), so it never collides
		// with a single-byte ctrl chord or with a CSI sequence. A user who
		// pressed Esc and then typed the letter produces the same bytes, and
		// replaying {key: alt-x} reproduces them exactly either way.
		m["\x1b"+string(c)] = "alt-" + string(c)
	}
	m["\x1c"] = "ctrl-\\"
	m["\x1d"] = "ctrl-]"
	m["\x1e"] = "ctrl-^"
	m["\x1f"] = "ctrl-_"
	m["\x1b[45;5u"] = "ctrl-hyphen"
	// backtab is deliberately absent: it and shift-tab share \x1b[Z, and a
	// recording reads better naming the key a keyboard has.
	for _, name := range []string{
		"enter", "tab", "esc", "space", "backspace", "delete", "insert",
		"shift-tab",
		"up", "down", "right", "left", "home", "end",
		"pageup", "pagedown",
		"alt-enter", "alt-backspace",
		"ctrl-up", "ctrl-down", "ctrl-right", "ctrl-left",
		"shift-up", "shift-down", "shift-right", "shift-left",
		"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11", "f12",
	} {
		m[ptyKeySequences[name]] = name
	}
	return m
}()

// PTYKeyForSequence returns the friendly named key whose xterm sequence exactly
// equals seq (#69), so `atago record --pty` can render a lone control key as
// {key: <name>} instead of an opaque escape. It reports false when no named key
// matches the bytes exactly.
func PTYKeyForSequence(seq string) (string, bool) {
	name, ok := ptyKeyBySequence[seq]
	return name, ok
}

// PTYKeyNames lists the vocabulary for error messages, compactly.
func PTYKeyNames() string {
	return "enter, tab, esc, space, backspace, delete, insert, up, down, left, right, home, end, pageup, pagedown, f1-f12, " +
		"shift-tab (alias backtab), alt-a..alt-z, alt-enter, alt-backspace, " +
		"ctrl-up/ctrl-down/ctrl-left/ctrl-right, shift-up/shift-down/shift-left/shift-right, " +
		"ctrl-a..ctrl-z, ctrl-space/ctrl-@, ctrl-[, ctrl-\\, ctrl-], ctrl-^, ctrl-_, ctrl-hyphen/ctrl-minus"
}
