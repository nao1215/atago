package ptyrun

import (
	"reflect"
	"testing"
)

// TestDECSETScanner covers the shapes a real program's output carries: the mode
// on and off, several modes requested at once, and a request split across pty
// reads — plus the near misses that must NOT be read as a request (#378).
func TestDECSETScanner(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		chunks []string
		want   []decsetMode
	}{
		{
			name:   "enable",
			chunks: []string{"\x1b[?2004h"},
			want:   []decsetMode{{Param: 2004, Enabled: true}},
		},
		{
			name:   "disable",
			chunks: []string{"\x1b[?2004l"},
			want:   []decsetMode{{Param: 2004, Enabled: false}},
		},
		{
			// How a TUI toolkit usually turns its modes on: one request, several
			// parameters. Reading only the first would miss bracketed paste.
			name:   "several modes in one request",
			chunks: []string{"\x1b[?1049;1006;2004h"},
			want: []decsetMode{
				{Param: 1049, Enabled: true},
				{Param: 1006, Enabled: true},
				{Param: 2004, Enabled: true},
			},
		},
		{
			// A pty read can end anywhere, including mid-sequence.
			name:   "split across reads",
			chunks: []string{"out\x1b[?20", "0", "4h more"},
			want:   []decsetMode{{Param: 2004, Enabled: true}},
		},
		{
			name:   "surrounded by ordinary output",
			chunks: []string{"before\x1b[?2004hafter\r\n"},
			want:   []decsetMode{{Param: 2004, Enabled: true}},
		},
		{
			// Not a private-mode request: no `?`, so this is an ordinary CSI.
			name:   "public mode is not tracked",
			chunks: []string{"\x1b[4h"},
			want:   nil,
		},
		{
			// A private-mode sequence with some other final byte is a different
			// request (a query, a report) and sets nothing.
			name:   "other final byte sets nothing",
			chunks: []string{"\x1b[?2004$p"},
			want:   nil,
		},
		{
			// An ESC restarts escape parsing, so the abandoned parameters must
			// not survive into whatever comes next.
			name:   "esc abandons the pending request",
			chunks: []string{"\x1b[?2004\x1b[?1000h"},
			want:   []decsetMode{{Param: 1000, Enabled: true}},
		},
		{
			name:   "empty parameter is skipped",
			chunks: []string{"\x1b[?;2004h"},
			want:   []decsetMode{{Param: 2004, Enabled: true}},
		},
		{
			name:   "an absurd parameter run is dropped rather than accumulated",
			chunks: []string{"\x1b[?" + string(make([]byte, 0)) + "11111111111111111111111111111111111111111111111111111111111111111111h"},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var s decsetScanner
			var got []decsetMode
			for _, c := range tt.chunks {
				got = append(got, s.consume([]byte(c))...)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("modes = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestDECSETScanner_LastWins proves the tracked state follows the program: a
// mode turned on and then off is off, which is what makes "the program stopped
// accepting pastes" a reportable condition rather than a silent one.
func TestDECSETScanner_LastWins(t *testing.T) {
	t.Parallel()
	var s decsetScanner
	got := s.consume([]byte("\x1b[?2004h...\x1b[?2004l"))
	want := []decsetMode{
		{Param: 2004, Enabled: true},
		{Param: 2004, Enabled: false},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("modes = %+v, want %+v", got, want)
	}
}
