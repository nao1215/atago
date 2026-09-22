package conpty

import (
	"strconv"
	"unicode/utf16"
)

// Win32InputKeys encodes p as win32-input-mode key presses, one
// `CSI 0 ; 0 ; <char> ; 1 ; 0 ; 1 _` per UTF-16 code unit.
//
// It is how a reply reaches a program through the bundled console host
// verbatim. That host parses its input as a terminal's, and drops what it does
// not know: the kitty graphics answer (an APC string, `ESC _ G ... ESC \`) never
// arrives. A key press whose character is ESC, `_`, or `G` is not parsed, it is
// handed to the program as that character, so the program reads the same bytes
// a terminal would have sent. The host accepts this encoding without being
// asked, because it is the one Windows Terminal speaks to it.
func Win32InputKeys(p []byte) []byte {
	out := make([]byte, 0, len(p)*16)
	for _, r := range string(p) {
		for _, u := range utf16.Encode([]rune{r}) {
			out = append(out, "\x1b[0;0;"...)
			out = strconv.AppendUint(out, uint64(u), 10)
			out = append(out, ";1;0;1_"...)
		}
	}
	return out
}
