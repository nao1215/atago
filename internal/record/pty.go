package record

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nao1215/atago/internal/buildinfo"
	"github.com/nao1215/atago/internal/loader"
	"github.com/nao1215/atago/internal/spec"
)

// PTYSegment is one chronological chunk of a recorded interactive session:
// either program output (program → terminal) or a burst of user input
// (keystrokes → program). Exactly one of Output/Input is set (#69).
type PTYSegment struct {
	// Output is a run of bytes the program wrote to the terminal.
	Output []byte
	// Input is one burst of bytes the user typed (nil for an output segment).
	Input []byte
	// EchoOff marks an input burst typed while terminal echo was disabled — a
	// password prompt. Its literal bytes must never reach the generated spec.
	EchoOff bool
}

// PTYRecording is a captured interactive `atago record --pty` session: the
// command that ran, the recording terminal's geometry, the observed exit code,
// and the ordered output/input segments the session generator turns into a
// declarative pty step (#69).
type PTYRecording struct {
	Command  string
	Shell    bool
	Rows     int
	Cols     int
	ExitCode int
	Segments []PTYSegment
}

// AppendOutput records a run of program output, coalescing it with a trailing
// output segment so consecutive writes form one chunk.
func (r *PTYRecording) AppendOutput(b []byte) {
	if n := len(r.Segments); n > 0 && r.Segments[n-1].Input == nil {
		r.Segments[n-1].Output = append(r.Segments[n-1].Output, b...)
		return
	}
	r.Segments = append(r.Segments, PTYSegment{Output: append([]byte(nil), b...)})
}

// AppendInput records one burst of user input, tagged with whether terminal
// echo was off (a secret prompt) at the time it was typed. Consecutive input
// bursts with the same echo state are coalesced (mirroring AppendOutput): raw
// mode delivers one keystroke per read, so a typed line arrives as many one-byte
// bursts, and without coalescing an N-character password would render as N
// separate ${env:ATAGO_SECRET_n} placeholders instead of one.
func (r *PTYRecording) AppendInput(b []byte, echoOff bool) {
	if n := len(r.Segments); n > 0 && r.Segments[n-1].Input != nil && r.Segments[n-1].EchoOff == echoOff {
		r.Segments[n-1].Input = append(r.Segments[n-1].Input, b...)
		return
	}
	r.Segments = append(r.Segments, PTYSegment{Input: append([]byte(nil), b...), EchoOff: echoOff})
}

// ansiPattern matches the terminal control sequences that carry no visible
// text: CSI (ESC [ ... final), OSC (ESC ] ... BEL/ST), and two-byte ESC forms.
// Stripping them yields the plain prompt text an expect should anchor on.
var ansiPattern = regexp.MustCompile(`\x1b\[[0-9:;?]*[ -/]*[@-~]|\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)|\x1b[@-Z\\-_]`)

// GeneratePTY renders a spec skeleton whose single pty: step replays the
// recorded session as expect/send pairs, and proves it loads cleanly before
// returning it (the same round-trip guarantee plain record gives) (#69).
func GeneratePTY(rec PTYRecording, opts Options) ([]byte, error) {
	var b strings.Builder
	b.WriteString(buildinfo.SchemaHeader())
	b.WriteString("version: \"1\"\n\n")
	b.WriteString("# Recorded by `atago record --pty` — a starting point, not a verdict:\n")
	b.WriteString("# each send replays a burst you typed, and each expect anchors on the\n")
	b.WriteString("# prompt that preceded it. Tighten the matchers to pin what you care about.\n")
	fmt.Fprintf(&b, "suite:\n  name: %s\n\n", yamlScalar(opts.SuiteName))
	b.WriteString("scenarios:\n")
	fmt.Fprintf(&b, "  - name: %s # TODO: describe the behavior\n", yamlScalar(rec.Command))
	b.WriteString("    steps:\n")
	b.WriteString("      - pty:\n")
	if rec.Shell {
		b.WriteString("          shell: true\n")
	}
	fmt.Fprintf(&b, "          command: %s\n", yamlScalar(escapeVarRefs(rec.Command)))
	if rec.Rows > 0 {
		fmt.Fprintf(&b, "          rows: %d\n", rec.Rows)
	}
	if rec.Cols > 0 {
		fmt.Fprintf(&b, "          cols: %d\n", rec.Cols)
	}

	session, lastOutput := renderSession(&rec)
	if len(session) > 0 {
		b.WriteString("          session:\n")
		for _, line := range session {
			b.WriteString(line)
		}
	}

	b.WriteString("      - assert:\n")
	fmt.Fprintf(&b, "          exit_code: %d\n", rec.ExitCode)
	if anchor := stableLine(lastOutput); anchor != "" {
		b.WriteString("      - assert:\n")
		b.WriteString("          stdout:\n")
		fmt.Fprintf(&b, "            contains: %s # last stable line of the transcript\n", yamlScalar(escapeVarRefs(anchor)))
	}

	out := []byte(b.String())
	if _, err := loader.LoadBytes("recorded.atago.yaml", out); err != nil {
		return nil, fmt.Errorf("generated spec does not validate (this is an atago bug, please report it): %w", err)
	}
	return out, nil
}

// renderSession walks the recorded segments and emits the YAML session lines:
// each input burst becomes a send, preceded by an expect derived from the last
// stable line of the output before it. It returns the session lines and the
// trailing output (after the final input) for the closing assertion (#69).
func renderSession(rec *PTYRecording) (lines []string, trailingOutput []byte) {
	var pending []byte
	secretN := 0
	for _, seg := range rec.Segments {
		if seg.Input == nil {
			pending = append(pending, seg.Output...)
			continue
		}
		if anchor := stableLine(pending); anchor != "" {
			lines = append(lines, fmt.Sprintf("            - expect: %s\n", yamlScalar(regexp.QuoteMeta(anchor))))
		}
		lines = append(lines, renderSend(seg, &secretN)...)
		pending = nil
	}
	return lines, pending
}

// renderSend renders one input burst as a send entry: an ${env:...} placeholder
// for echo-off (secret) input, a named key for a lone control key, or literal
// text otherwise. The literal secret is never emitted (#69).
func renderSend(seg PTYSegment, secretN *int) []string {
	if seg.EchoOff {
		*secretN++
		name := fmt.Sprintf("ATAGO_SECRET_%d", *secretN)
		suffix := ""
		if endsWithNewline(seg.Input) {
			suffix = "\n"
		}
		return []string{
			"            # secret input (terminal echo was off): the literal value is NOT recorded.\n",
			fmt.Sprintf("            # set %s in the environment and add its value to `secrets:` to mask it.\n", name),
			fmt.Sprintf("            - send: %s\n", yamlDoubleQuoted("${env:"+name+"}"+suffix)),
		}
	}
	if key, ok := spec.PTYKeyForSequence(string(seg.Input)); ok {
		return []string{fmt.Sprintf("            - send: {key: %s}\n", key)}
	}
	// Typed text is raw: escape ${...} so the replay engine types the literal
	// bytes the user typed instead of expanding them (the secret placeholder
	// above is the one send that MUST stay a live reference).
	return []string{fmt.Sprintf("            - send: %s\n", yamlDoubleQuoted(escapeVarRefs(literalSend(seg.Input))))}
}

// yamlDoubleQuoted renders s as a YAML double-quoted flow scalar, escaping
// control characters (notably \n and \r) so a multi-line send stays on one line
// instead of becoming a block scalar that would break the session list (#69).
func yamlDoubleQuoted(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				fmt.Fprintf(&b, `\x%02x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}

// literalSend renders a printable input burst for a send scalar: carriage
// returns become "\n" (the readable Enter the pty examples use) and other
// control bytes are left for yamlScalar to escape.
func literalSend(b []byte) string {
	s := string(b)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

// endsWithNewline reports whether an input burst ended with Enter (CR or LF),
// so a secret placeholder can reproduce the line submission without the value.
func endsWithNewline(b []byte) bool {
	return len(b) > 0 && (b[len(b)-1] == '\r' || b[len(b)-1] == '\n')
}

// stableLine returns the conservative literal an expect/contains anchors on: the
// longest run of plain text on the last visible line of the transcript that
// carries no ANSI sequence or control byte (#69). The returned run is a VERBATIM
// substring of the raw transcript — ANSI sequences are turned into a delimiter,
// not stripped and concatenated — so an anchor built from it actually matches the
// raw pty stdout the replay compares against. Stripping ANSI and joining the
// visible text (the old behavior) produced an anchor with mid-line color codes
// removed that was never a substring of the raw output, so a colored prompt made
// the generated spec fail on replay.
func stableLine(output []byte) string {
	// Replace ANSI/OSC sequences with a NUL so the plain text on either side stays
	// contiguous and verbatim; fold CR so a redraw does not merge lines.
	s := ansiPattern.ReplaceAllString(string(output), "\x00")
	s = strings.ReplaceAll(s, "\r", "\n")
	best := ""
	for _, line := range strings.Split(s, "\n") {
		if run := longestPlainRun(line); run != "" {
			best = run
		}
	}
	return best
}

// longestPlainRun returns the longest run of line that contains no C0 control
// byte (the NUL standing in for an ANSI sequence, a tab, or any other), trimmed
// of surrounding whitespace. Each such run existed verbatim in the raw output.
func longestPlainRun(line string) string {
	best := ""
	var cur strings.Builder
	flush := func() {
		if t := strings.TrimSpace(cur.String()); len(t) > len(best) {
			best = t
		}
		cur.Reset()
	}
	for _, r := range line {
		if r < 0x20 {
			flush()
			continue
		}
		cur.WriteRune(r)
	}
	flush()
	return best
}
