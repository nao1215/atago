package cli

import (
	"reflect"
	"testing"

	shellwords "github.com/mattn/go-shellwords"
)

// TestPOSIXJoin_RoundTrip proves posixJoin's output re-tokenizes to the same
// argv through go-shellwords — the tokenizer the cmd runner uses on POSIX
// (#30 follow-up). Boundaries must survive spaces, quotes, the metacharacters
// go-shellwords treats as separators, and backslashes.
func TestPOSIXJoin_RoundTrip(t *testing.T) {
	t.Parallel()
	cases := [][]string{
		{"mytool", "convert", "input.txt"},
		{"mytool", "--msg", "hello world"},
		{"tool", "it's"},
		{"tool", `she said "hi"`},
		{"tool", "spaced arg", "--flag=with space"},
		{"grep", "foo|bar"},
		{"tool", "a>b", "c<d", "e;f", "g&h"},
		{"tool", `C:\tmp\file`},
		{"tool", "$HOME", "*.go"},
		{"tool", ""}, // an empty argument must survive as a distinct entry
	}
	for _, argv := range cases {
		joined := posixJoin(argv)
		got, err := shellwords.Parse(joined)
		if err != nil {
			t.Errorf("%v -> %q does not tokenize: %v", argv, joined, err)
			continue
		}
		if !reflect.DeepEqual(got, argv) {
			t.Errorf("round-trip changed argv: %v -> %q -> %v", argv, joined, got)
		}
	}
}

// TestWindowsJoin_Boundaries proves windowsJoin leaves a whitespace-free path
// verbatim (the e2e-windows regression this fixes: single quotes had been
// treated as part of the filename) while quoting whitespace-bearing tokens.
func TestWindowsJoin_Boundaries(t *testing.T) {
	t.Parallel()
	if got := windowsJoin([]string{`C:\Users\runner\atago.exe`, "version"}); got != `C:\Users\runner\atago.exe version` {
		t.Errorf("plain Windows path must pass through: %q", got)
	}
	if got := windowsJoin([]string{`C:\Program Files\tool.exe`, "run"}); got != `"C:\Program Files\tool.exe" run` {
		t.Errorf("spaced path must be double-quoted verbatim (backslashes literal): %q", got)
	}
}
