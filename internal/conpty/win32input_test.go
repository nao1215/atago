package conpty

import "testing"

func TestWin32InputKeys(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"one ASCII character", "G", "\x1b[0;0;71;1;0;1_"},
		{
			"an ESC is a key press, not the start of a sequence",
			"\x1b_",
			"\x1b[0;0;27;1;0;1_\x1b[0;0;95;1;0;1_",
		},
		{
			"a character outside the BMP is two UTF-16 code units",
			"\U0001F600",
			"\x1b[0;0;55357;1;0;1_\x1b[0;0;56832;1;0;1_",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := string(Win32InputKeys([]byte(c.in))); got != c.want {
				t.Errorf("Win32InputKeys(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
