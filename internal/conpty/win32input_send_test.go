package conpty

import "testing"

func TestWin32InputSend(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"a lone ESC is one Escape key press", "\x1b", "\x1b[27;0;27;1;0;1_\x1b[27;0;27;0;0;1_"},
		{"a letter carries its virtual-key code", "q", "\x1b[81;0;113;1;0;1_\x1b[81;0;113;0;0;1_"},
		{"an upper-case letter is typed with Shift", "Q", "\x1b[81;0;81;1;16;1_\x1b[81;0;81;0;16;1_"},
		{"a digit and a space", "1 ", "\x1b[49;0;49;1;0;1_\x1b[49;0;49;0;0;1_\x1b[32;0;32;1;0;1_\x1b[32;0;32;0;0;1_"},
		{"CR is Enter", "\r", "\x1b[13;0;13;1;0;1_\x1b[13;0;13;0;0;1_"},
		{"TAB is Tab", "\t", "\x1b[9;0;9;1;0;1_\x1b[9;0;9;0;0;1_"},
		{"DEL is Backspace", "\x7f", "\x1b[8;0;8;1;0;1_\x1b[8;0;8;0;0;1_"},
		{"a C0 control is its Ctrl chord", "\x03", "\x1b[67;0;3;1;8;1_\x1b[67;0;3;0;8;1_"},
		{"NUL is Ctrl+Space", "\x00", "\x1b[32;0;0;1;8;1_\x1b[32;0;0;0;8;1_"},
		{"a character with no key of its own", "é", "\x1b[0;0;233;1;0;1_\x1b[0;0;233;0;0;1_"},
		{
			"a character outside the BMP is its two surrogates",
			"\U0001F44D",
			"\x1b[0;0;55357;1;0;1_\x1b[0;0;55357;0;0;1_\x1b[0;0;56397;1;0;1_\x1b[0;0;56397;0;0;1_",
		},
		{"ESC and a letter is Alt and that letter", "\x1bq", "\x1b[81;0;113;1;2;1_\x1b[81;0;113;0;2;1_"},
		{"ESC and CR is Alt+Enter", "\x1b\r", "\x1b[13;0;13;1;2;1_\x1b[13;0;13;0;2;1_"},
		{"ESC twice is two Escape presses", "\x1b\x1b", "\x1b[27;0;27;1;0;1_\x1b[27;0;27;0;0;1_\x1b[27;0;27;1;0;1_\x1b[27;0;27;0;0;1_"},
		{"ESC and a digit is Escape, then the digit", "\x1b1", "\x1b[27;0;27;1;0;1_\x1b[27;0;27;0;0;1_\x1b[49;0;49;1;0;1_\x1b[49;0;49;0;0;1_"},
		{"an arrow", "\x1b[A", "\x1b[38;0;0;1;256;1_\x1b[38;0;0;0;256;1_"},
		{"Ctrl and an arrow", "\x1b[1;5D", "\x1b[37;0;0;1;264;1_\x1b[37;0;0;0;264;1_"},
		{"Shift and an arrow", "\x1b[1;2C", "\x1b[39;0;0;1;272;1_\x1b[39;0;0;0;272;1_"},
		{"Home and End", "\x1b[H\x1b[F", "\x1b[36;0;0;1;256;1_\x1b[36;0;0;0;256;1_\x1b[35;0;0;1;256;1_\x1b[35;0;0;0;256;1_"},
		{"Delete", "\x1b[3~", "\x1b[46;0;0;1;256;1_\x1b[46;0;0;0;256;1_"},
		{"Page Down", "\x1b[6~", "\x1b[34;0;0;1;256;1_\x1b[34;0;0;0;256;1_"},
		{"F1 in its SS3 form", "\x1bOP", "\x1b[112;0;0;1;0;1_\x1b[112;0;0;0;0;1_"},
		{"F12", "\x1b[24~", "\x1b[123;0;0;1;0;1_\x1b[123;0;0;0;0;1_"},
		{"Shift+Tab", "\x1b[Z", "\x1b[9;0;9;1;16;1_\x1b[9;0;9;0;16;1_"},
		{"a CSI-u key", "\x1b[45;5u", "\x1b[189;0;45;1;8;1_\x1b[189;0;45;0;8;1_"},
		{"a bracketed-paste marker passes through", "\x1b[200~a\x1b[201~", "\x1b[200~\x1b[65;0;97;1;0;1_\x1b[65;0;97;0;0;1_\x1b[201~"},
		{"a mouse report passes through", "\x1b[<0;5;12M", "\x1b[<0;5;12M"},
		{"a CSI cut short is Escape, then its characters", "\x1b[", "\x1b[27;0;27;1;0;1_\x1b[27;0;27;0;0;1_\x1b[0;0;91;1;0;1_\x1b[0;0;91;0;0;1_"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := string(Win32InputSend([]byte(c.in))); got != c.want {
				t.Errorf("Win32InputSend(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
