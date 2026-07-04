package cli

import (
	"reflect"
	"testing"

	shellwords "github.com/mattn/go-shellwords"
)

// TestShellJoin_RoundTrip proves shellJoin's output re-tokenizes to the same
// argv through the POSIX tokenizer the cmd runner uses (#30 follow-up): argv
// boundaries must survive spaces and quotes. A no-whitespace token — Windows
// paths included — passes through verbatim (asserted separately).
func TestShellJoin_RoundTrip(t *testing.T) {
	t.Parallel()
	cases := [][]string{
		{"mytool", "convert", "input.txt"},
		{"mytool", "--msg", "hello world"},
		{"tool", "it's"},
		{"tool", `she said "hi"`},
		{"tool", "spaced arg", "--flag=with space"},
		{"tool", ""}, // an empty argument must survive as a distinct entry
	}
	for _, argv := range cases {
		joined := shellJoin(argv)
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

// TestShellJoin_PassesThroughPlainTokens proves whitespace-free tokens are
// left verbatim so a Windows path never gets quotes the native tokenizer
// would treat as part of the filename (the e2e-windows regression this fixes).
func TestShellJoin_PassesThroughPlainTokens(t *testing.T) {
	t.Parallel()
	got := shellJoin([]string{`C:\Users\runner\atago.exe`, "version"})
	want := `C:\Users\runner\atago.exe version`
	if got != want {
		t.Errorf("shellJoin = %q, want %q (plain tokens must not be quoted)", got, want)
	}
}
