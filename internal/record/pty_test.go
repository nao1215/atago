package record

import (
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/loader"
	"github.com/nao1215/atago/internal/spec"
)

func outSeg(s string) PTYSegment { return PTYSegment{Output: []byte(s)} }
func inSeg(s string) PTYSegment  { return PTYSegment{Input: []byte(s)} }

// loadGenerated asserts the generated spec both validates (GeneratePTY does this
// internally) and re-loads to a single pty step, returning that step.
func loadGenerated(t *testing.T, data []byte) *spec.PTY {
	t.Helper()
	s, err := loader.LoadBytes("gen.atago.yaml", data)
	if err != nil {
		t.Fatalf("generated spec does not load: %v\n%s", err, data)
	}
	if len(s.Scenarios) != 1 || len(s.Scenarios[0].Steps) == 0 || s.Scenarios[0].Steps[0].PTY == nil {
		t.Fatalf("generated spec has no leading pty step:\n%s", data)
	}
	return s.Scenarios[0].Steps[0].PTY
}

func TestGeneratePTY(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		rec         PTYRecording
		wantContain []string
		wantAbsent  []string
	}{
		{
			name: "prompt at line end becomes expect, input becomes literal send",
			rec: PTYRecording{
				Command:  "sh -c 'printf name:; read n; echo hi $n'",
				ExitCode: 0,
				Segments: []PTYSegment{
					outSeg("name: "),
					inSeg("world\r"),
					outSeg("world\r\nhi world\r\n"),
				},
			},
			wantContain: []string{"- expect:", "name:", "- send:", "world", "exit_code: 0", "hi world"},
		},
		{
			name: "a lone control key maps to a named key",
			rec: PTYRecording{
				Command:  "cat",
				ExitCode: 0,
				Segments: []PTYSegment{
					outSeg("> "),
					inSeg("\r"),
					outSeg("\r\n"),
				},
			},
			wantContain: []string{"send: {key: enter}"},
		},
		{
			name: "ctrl-c maps to its named key",
			rec: PTYRecording{
				Command:  "top",
				ExitCode: 130,
				Segments: []PTYSegment{
					outSeg("load: 0.1\n"),
					inSeg("\x03"),
				},
			},
			wantContain: []string{"send: {key: ctrl-c}", "exit_code: 130"},
		},
		{
			name: "multi-line output between inputs anchors on the last stable line",
			rec: PTYRecording{
				Command:  "wizard",
				ExitCode: 0,
				Segments: []PTYSegment{
					outSeg("Welcome\nStep 1 of 2\nProject name: "),
					inSeg("demo\r"),
					outSeg("created demo/\n"),
				},
			},
			// The expect anchors on "Project name:", not the earlier lines.
			wantContain: []string{"Project name", "created demo"},
			wantAbsent:  []string{"expect: \"Welcome"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := GeneratePTY(tt.rec, Options{SuiteName: "gen"})
			if err != nil {
				t.Fatalf("GeneratePTY: %v", err)
			}
			s := string(got)
			for _, want := range tt.wantContain {
				if !strings.Contains(s, want) {
					t.Errorf("generated spec missing %q:\n%s", want, s)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(s, absent) {
					t.Errorf("generated spec unexpectedly contains %q:\n%s", absent, s)
				}
			}
			loadGenerated(t, got) // proves the round-trip guarantee
		})
	}
}

// TestGeneratePTY_EscapesVariableReferences proves the pty round-trip law for
// text containing ${...}: the engine expands ${name} in the pty command and in
// literal send text at replay, so both must carry the $${...} literal escape.
// Expect anchors must NOT be $$-escaped: regexp.QuoteMeta already renders the
// reference as \$\{...\} which the expander ignores, and double-escaping would
// make the compiled pattern miss the transcript.
func TestGeneratePTY_EscapesVariableReferences(t *testing.T) {
	t.Parallel()
	rec := PTYRecording{
		Command:  "sh -c 'read x; echo got'",
		ExitCode: 0,
		Segments: []PTYSegment{
			outSeg("enter ${VAR} value: "),
			inSeg("literal ${HOME}\r"),
			outSeg("got\r\n"),
		},
	}
	got, err := GeneratePTY(rec, Options{SuiteName: "esc"})
	if err != nil {
		t.Fatalf("GeneratePTY: %v", err)
	}
	// Assert on the DECODED session (yamlScalar doubles backslashes on the
	// wire, so string-matching the file text would test the quoting, not the
	// semantics).
	pty := loadGenerated(t, got)
	var expect, send string
	for _, a := range pty.Session {
		if a.Expect != "" {
			expect = a.Expect
		}
		if a.Send != nil && a.Send.Text != nil {
			send = *a.Send.Text
		}
	}
	if send != "literal $${HOME}\n" {
		t.Errorf("literal send = %q, want the $${...} escape carried through", send)
	}
	// The expect stays QuoteMeta-only: \$\{VAR\} matches the literal prompt
	// and is already inert to the expander; $$-escaping it too would make the
	// compiled pattern miss the transcript.
	if expect != `enter \$\{VAR\} value:` {
		t.Errorf("expect = %q, want QuoteMeta-escaped exactly once", expect)
	}
}

// TestGeneratePTY_EchoOffNeverLeaksSecret proves an echo-off (password) input
// burst produces an ${env:...} placeholder and that the literal secret never
// reaches the generated YAML (#69).
func TestGeneratePTY_EchoOffNeverLeaksSecret(t *testing.T) {
	t.Parallel()
	const secret = "hunter2SuperSecret"
	rec := PTYRecording{
		Command:  "login",
		ExitCode: 0,
		Segments: []PTYSegment{
			outSeg("Username: "),
			inSeg("alice\r"),
			outSeg("alice\r\nPassword: "),
			{Input: []byte(secret + "\r"), EchoOff: true},
			outSeg("\r\nwelcome alice\r\n"),
		},
	}
	got, err := GeneratePTY(rec, Options{SuiteName: "login"})
	if err != nil {
		t.Fatalf("GeneratePTY: %v", err)
	}
	s := string(got)
	if strings.Contains(s, secret) {
		t.Fatalf("SECRET LEAKED into generated spec:\n%s", s)
	}
	if !strings.Contains(s, "${env:ATAGO_SECRET_1}") {
		t.Errorf("expected an ${env:...} placeholder for the secret input:\n%s", s)
	}
	if !strings.Contains(s, "echo was off") {
		t.Errorf("expected a comment explaining the redacted secret:\n%s", s)
	}
	// The non-secret username still round-trips as a normal send.
	if !strings.Contains(s, "alice") {
		t.Errorf("non-secret input should still be recorded:\n%s", s)
	}
	loadGenerated(t, got)
}

// TestStableLine strips ANSI control sequences and returns the last visible line.
func TestStableLine(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want string
	}{
		{"name: ", "name:"},
		{"\x1b[2J\x1b[H> ", ">"},
		{"line one\r\nline two\r\n", "line two"},
		{"\x1b[32mgreen prompt\x1b[0m ", "green prompt"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := stableLine([]byte(tt.in)); got != tt.want {
			t.Errorf("stableLine(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
