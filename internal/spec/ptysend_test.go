package spec

import (
	"strings"
	"testing"
)

// TestPTYKeySequences_GoldenTable pins the exact bytes each named key
// transmits (#26) — the documented contract TUI specs rely on.
func TestPTYKeySequences_GoldenTable(t *testing.T) {
	t.Parallel()
	want := map[string]string{
		"enter":       "\r",
		"tab":         "\t",
		"esc":         "\x1b",
		"space":       " ",
		"backspace":   "\x7f",
		"delete":      "\x1b[3~",
		"up":          "\x1b[A",
		"down":        "\x1b[B",
		"right":       "\x1b[C",
		"left":        "\x1b[D",
		"home":        "\x1b[H",
		"end":         "\x1b[F",
		"pageup":      "\x1b[5~",
		"pagedown":    "\x1b[6~",
		"f1":          "\x1bOP",
		"f4":          "\x1bOS",
		"f5":          "\x1b[15~",
		"f12":         "\x1b[24~",
		"ctrl-space":  "\x00",
		"ctrl-@":      "\x00",
		"ctrl-a":      "\x01",
		"ctrl-c":      "\x03",
		"ctrl-d":      "\x04",
		"ctrl-[":      "\x1b",
		"ctrl-\\":     "\x1c",
		"ctrl-]":      "\x1d",
		"ctrl-^":      "\x1e",
		"ctrl-_":      "\x1f",
		"ctrl-hyphen": "\x1b[45;5u",
		"ctrl-minus":  "\x1b[45;5u",
		"ctrl-z":      "\x1a",

		// The editing/navigation keys real TUIs bind (#376).
		"shift-tab":     "\x1b[Z",
		"backtab":       "\x1b[Z",
		"insert":        "\x1b[2~",
		"alt-a":         "\x1ba",
		"alt-b":         "\x1bb",
		"alt-z":         "\x1bz",
		"alt-enter":     "\x1b\r",
		"alt-backspace": "\x1b\x7f",
		"ctrl-up":       "\x1b[1;5A",
		"ctrl-down":     "\x1b[1;5B",
		"ctrl-right":    "\x1b[1;5C",
		"ctrl-left":     "\x1b[1;5D",
		"shift-up":      "\x1b[1;2A",
		"shift-down":    "\x1b[1;2B",
		"shift-right":   "\x1b[1;2C",
		"shift-left":    "\x1b[1;2D",
	}
	for name, bytes := range want {
		got := (&PTYSend{Key: name}).Bytes()
		if string(got) != bytes {
			t.Errorf("key %s = %q, want %q", name, got, bytes)
		}
	}
	// The whole vocabulary is valid; a typo is not.
	for name := range want {
		if !ValidPTYKey(name) {
			t.Errorf("ValidPTYKey(%s) = false", name)
		}
	}
	if ValidPTYKey("entr") {
		t.Error("ValidPTYKey(entr) = true, want false")
	}
}

// TestPTYKeySequences_AltCoversEveryLetter pins the whole Meta family (#376):
// alt-<letter> is the ESC prefix a terminal transmits for a Meta chord, which
// is what readline word operations and helix/emacs-style bindings read.
func TestPTYKeySequences_AltCoversEveryLetter(t *testing.T) {
	t.Parallel()
	for c := byte('a'); c <= 'z'; c++ {
		name := "alt-" + string(c)
		want := "\x1b" + string(c)
		if got := (&PTYSend{Key: name}).Bytes(); string(got) != want {
			t.Errorf("key %s = %q, want %q", name, got, want)
		}
		if !ValidPTYKey(name) {
			t.Errorf("ValidPTYKey(%s) = false", name)
		}
	}
	// Meta is letters only: a digit chord has no stable legacy encoding, so it
	// must stay a load error rather than silently sending nothing.
	if ValidPTYKey("alt-1") {
		t.Error("ValidPTYKey(alt-1) = true, want false")
	}
}

// TestPTYKeyBySequence_CoversEveryKey is the drift guard between the two
// directions of the vocabulary: `atago record --pty` renders a captured control
// sequence through PTYKeyForSequence (#69), so a key added to the forward table
// and forgotten in the reverse one would record as an opaque escape instead of
// the readable {key: <name>} the author could have written. Every name must
// reverse-map to a name that transmits the SAME bytes — which lets documented
// aliases (ctrl-minus/ctrl-hyphen, backtab/shift-tab) resolve to whichever
// spelling the reverse table prefers without weakening the guard.
func TestPTYKeyBySequence_CoversEveryKey(t *testing.T) {
	t.Parallel()
	for name, seq := range ptyKeySequences {
		got, ok := PTYKeyForSequence(seq)
		if !ok {
			t.Errorf("key %s transmits %q, which PTYKeyForSequence does not name; add it to ptyKeyBySequence", name, seq)
			continue
		}
		if back := ptyKeySequences[got]; back != seq {
			t.Errorf("key %s transmits %q but reverse-maps to %s, which transmits %q", name, seq, got, back)
		}
	}
}

// TestPTYKeyForSequence_PrefersFriendlyName pins which spelling wins when two
// names share a byte sequence: the readable one, so a recording says
// {key: shift-tab} rather than {key: backtab} and {key: enter} rather than
// {key: ctrl-m}.
func TestPTYKeyForSequence_PrefersFriendlyName(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		seq  string
		want string
	}{
		{"\r", "enter"},
		{"\x1b", "esc"},
		{"\x1b[Z", "shift-tab"},
		{"\x1b[45;5u", "ctrl-hyphen"},
		{"\x1bb", "alt-b"},
		{"\x1b[1;5D", "ctrl-left"},
		{"\x1b[2~", "insert"},
	} {
		got, ok := PTYKeyForSequence(tc.seq)
		if !ok || got != tc.want {
			t.Errorf("PTYKeyForSequence(%q) = %q, %v; want %q, true", tc.seq, got, ok, tc.want)
		}
	}
}

// TestPTYKeyNames_ListsEveryFamily keeps the loader's "not a supported key"
// message a complete answer: a spec author reading it must find every family
// they could have meant. The e2e suite pins its "enter, tab" prefix, so the
// added families go at the end.
func TestPTYKeyNames_ListsEveryFamily(t *testing.T) {
	t.Parallel()
	names := PTYKeyNames()
	if !strings.HasPrefix(names, "enter, tab") {
		t.Errorf("PTYKeyNames should still start with %q: %q", "enter, tab", names)
	}
	for _, want := range []string{
		"insert", "shift-tab", "backtab", "alt-a..alt-z",
		"alt-enter", "alt-backspace", "ctrl-left", "shift-left",
	} {
		if !strings.Contains(names, want) {
			t.Errorf("PTYKeyNames does not mention %q: %q", want, names)
		}
	}
}

// TestPTYSend_Times pins the repeat form (#377): a navigation key pressed N
// times is one action, not N copies of the same line.
func TestPTYSend_Times(t *testing.T) {
	t.Parallel()
	if got := (&PTYSend{Key: "left", Times: 3}).Bytes(); string(got) != "\x1b[D\x1b[D\x1b[D" {
		t.Errorf("times 3 = %q", got)
	}
	// An unset or explicit 1 is exactly one press, so adding the field cannot
	// change what an existing spec sends.
	for _, n := range []int{0, 1} {
		if got := (&PTYSend{Key: "left", Times: n}).Bytes(); string(got) != "\x1b[D" {
			t.Errorf("times %d = %q, want one press", n, got)
		}
	}
	if got := (&PTYSend{Key: "left", Times: 16}).Label(); got != "press left x16" {
		t.Errorf("label = %q", got)
	}
	if got := (&PTYSend{Key: "left", Times: 1}).Label(); got != "press left" {
		t.Errorf("label = %q", got)
	}
}

// TestPTYSend_Paste pins the bracketed-paste form (#378): the text goes out
// wrapped in the markers a terminal puts around a paste, so a program that
// distinguishes pasted input from typed input sees a paste.
func TestPTYSend_Paste(t *testing.T) {
	t.Parallel()
	p := &PTYSend{Paste: strPtr("SELECT 1;\nSELECT 2;\n")}
	want := "\x1b[200~SELECT 1;\nSELECT 2;\n\x1b[201~"
	if got := p.Bytes(); string(got) != want {
		t.Errorf("paste bytes = %q, want %q", got, want)
	}
	// An empty paste is an empty paste, NOT the empty-send EOF rule: that rule
	// belongs to the scalar form alone.
	empty := &PTYSend{Paste: strPtr("")}
	if got := empty.Bytes(); string(got) != "\x1b[200~\x1b[201~" {
		t.Errorf("empty paste = %q, want just the markers", got)
	}
	if got := p.Label(); got != `paste "SELECT 1;\nSELECT 2;\n"` {
		t.Errorf("label = %q", got)
	}
}

func strPtr(s string) *string { return &s }

// TestPTYSend_TextAndEOF proves the scalar form and the historical
// empty-string EOF rule survive the polymorphic type.
func TestPTYSend_TextAndEOF(t *testing.T) {
	t.Parallel()
	if got := SendText("hello\n").Bytes(); string(got) != "hello\n" {
		t.Errorf("text bytes = %q", got)
	}
	if got := SendText("").Bytes(); string(got) != "\x04" {
		t.Errorf("empty send = %q, want ^D", got)
	}
	if got := SendText("").Label(); got != "send EOF (^D)" {
		t.Errorf("label = %q", got)
	}
	if got := (&PTYSend{Key: "enter"}).Label(); got != "press enter" {
		t.Errorf("label = %q", got)
	}
}
