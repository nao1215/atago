package diag

import (
	"strings"
	"testing"
)

// FuzzParse checks the code parser against arbitrary text. It is reachable from
// the command line (`atago explain <arg>`), so it takes whatever a user types:
// the guarantee is that it never panics and never claims an unregistered code.
func FuzzParse(f *testing.F) {
	for _, seed := range []string{
		"", "A", "AT", "ATG", "ATG2", "ATG2201", "atg2201", "  ATG2201  ",
		"ATG22015", "ATG-201", "ATG 2201", "ATGxxxx", "ATG9999", "ATG0000",
		"ATG+201", "ATG 2201\n", "ÀTG2201", "ATG²²⁰¹", "ATG٢٢٠١",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		c, ok := Parse(s)
		if !ok {
			return
		}
		if _, found := Lookup(c); !found {
			t.Fatalf("Parse(%q) returned %v, which is not registered", s, c)
		}
		// A parsed code must round-trip: whatever spelling was accepted, the
		// canonical form parses back to the same code.
		again, ok2 := Parse(c.String())
		if !ok2 || again != c {
			t.Fatalf("Parse(%q) -> %v, but %q does not parse back to it", s, c, c.String())
		}
	})
}

// FuzzCodes checks the scanner that finds codes inside a message. It runs over
// error text that embeds a program's own output, so the input is untrusted:
// it must never panic, never report an unregistered code, and never report the
// same code twice.
func FuzzCodes(f *testing.F) {
	for _, seed := range []string{
		"", "ATG2201", "spec.yaml: ATG2201: suite.name is required",
		"ATG2201 ATG2201", "ATG22015", "xATG2201x", "ATG2201ATG2502",
		"ATG9999 ATG2201", "\x00ATG2201\x00", strings.Repeat("ATG2201 ", 50),
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		got := Codes(s)
		seen := map[Code]bool{}
		for _, c := range got {
			if _, found := Lookup(c); !found {
				t.Fatalf("Codes(%q) reported %v, which is not registered", s, c)
			}
			if seen[c] {
				t.Fatalf("Codes(%q) reported %v twice", s, c)
			}
			seen[c] = true
		}
	})
}
