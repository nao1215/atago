package spec

import (
	"testing"
)

// TestPTYKeySequences_GoldenTable pins the exact bytes each named key
// transmits (#26) — the documented contract TUI specs rely on.
func TestPTYKeySequences_GoldenTable(t *testing.T) {
	t.Parallel()
	want := map[string]string{
		"enter":     "\r",
		"tab":       "\t",
		"esc":       "\x1b",
		"space":     " ",
		"backspace": "\x7f",
		"delete":    "\x1b[3~",
		"up":        "\x1b[A",
		"down":      "\x1b[B",
		"right":     "\x1b[C",
		"left":      "\x1b[D",
		"home":      "\x1b[H",
		"end":       "\x1b[F",
		"pageup":    "\x1b[5~",
		"pagedown":  "\x1b[6~",
		"f1":        "\x1bOP",
		"f4":        "\x1bOS",
		"f5":        "\x1b[15~",
		"f12":       "\x1b[24~",
		"ctrl-a":    "\x01",
		"ctrl-c":    "\x03",
		"ctrl-d":    "\x04",
		"ctrl-z":    "\x1a",
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
