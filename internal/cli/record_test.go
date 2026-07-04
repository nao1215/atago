package cli

import (
	"reflect"
	"testing"

	shellwords "github.com/mattn/go-shellwords"
)

// TestShellJoin_RoundTrip proves shellJoin's output re-tokenizes to the same
// argv (#30 follow-up): argv boundaries must survive spaces, quotes, and
// shell metacharacters.
func TestShellJoin_RoundTrip(t *testing.T) {
	t.Parallel()
	cases := [][]string{
		{"mytool", "convert", "input.txt"},
		{"mytool", "--msg", "hello world"},
		{"tool", "it's"},
		{"tool", `she said "hi"`},
		{"tool", "a$b", "c;d", "e|f", ""},
		{"tool", "spaced arg", "--flag=with space"},
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
