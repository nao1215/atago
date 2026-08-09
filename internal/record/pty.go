package record

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

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
	// scenarioLabel, not the raw command: the loader rejects a control character
	// in a name, so a recorded command carrying a tab or a CR generated a spec
	// that failed its own validation and told the user to report an atago bug.
	fmt.Fprintf(&b, "  - name: %s # TODO: describe the behavior\n", yamlScalar(scenarioLabel(rec.Command)))
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
	// An SGR mouse report replays as the event it describes, so a mouse-driven
	// recording reads as clicks rather than as escape soup (#381).
	if m, ok := sgrMouseEvent(seg.Input); ok {
		return []string{fmt.Sprintf("            - send: {mouse: %s}\n", mouseFlowMapping(m))}
	}
	// A burst the terminal bracketed is a paste, and replaying it as typed text
	// would take the program's OTHER input path (#378).
	if inner, ok := bracketedPaste(seg.Input); ok {
		text := escapeVarRefs(literalSend(inner))
		if utf8.ValidString(text) {
			return []string{fmt.Sprintf("            - send: {paste: %s}\n", yamlDoubleQuoted(text))}
		}
		// Fall through: a paste that is not valid UTF-8 keeps the !!binary
		// escape hatch below, which preserves the bytes markers and all.
	}
	if key, ok := spec.PTYKeyForSequence(string(seg.Input)); ok {
		return []string{fmt.Sprintf("            - send: {key: %s}\n", key)}
	}
	// A held navigation key arrives as one burst of the same sequence over and
	// over (capture coalesces consecutive input reads), which would otherwise
	// record as an opaque wall of escapes. `times` says what happened (#377).
	if key, n, ok := keyRepeat(seg.Input); ok {
		return []string{fmt.Sprintf("            - send: {key: %s, times: %d}\n", key, n)}
	}
	// Typed text is raw: escape ${...} so the replay engine types the literal
	// bytes the user typed instead of expanding them (the secret placeholder
	// above is the one send that MUST stay a live reference).
	text := escapeVarRefs(literalSend(seg.Input))
	if !utf8.ValidString(text) {
		// A YAML scalar carries text, and rendering these bytes as one would
		// replace each invalid byte with U+FFFD — the replay would type three
		// different bytes than the recording captured. !!binary keeps them
		// exactly, the same escape hatch plain record uses for a capture that is
		// not valid UTF-8.
		return []string{
			"            # not valid UTF-8 (a paste in another encoding, a keyboard macro):\n",
			"            # base64 so the replay types the recorded bytes exactly.\n",
			fmt.Sprintf("            - send: !!binary \"%s\"\n", base64.StdEncoding.EncodeToString([]byte(text))),
		}
	}
	return []string{fmt.Sprintf("            - send: %s\n", yamlDoubleQuoted(text))}
}

// sgrMouseRe matches one xterm SGR (1006) mouse report: CSI < Cb ; col ; row
// followed by M for a press or m for a release.
var sgrMouseRe = regexp.MustCompile(`^\x1b\[<(\d+);(\d+);(\d+)([Mm])$`)

// sgrMouseEvent decodes a captured burst that is exactly one mouse report — or
// a press immediately followed by its own release, which is what a click
// delivers — into the event a spec would write (#381). Anything else (motion
// reports, a report mixed with typing, a modifier atago has no name for) is left
// to the literal-text path, where the recorded bytes survive untouched.
func sgrMouseEvent(input []byte) (*spec.PTYMouse, bool) {
	if m, ok := parseSGRMouse(input); ok {
		return m, true
	}
	// A click: the press and its release, back to back and identical apart from
	// the final byte.
	for split := 1; split < len(input); split++ {
		press, ok := parseSGRMouse(input[:split])
		if !ok || press.Action != "press" {
			continue
		}
		release, ok := parseSGRMouse(input[split:])
		if !ok || release.Action != "release" {
			continue
		}
		if press.Row != release.Row || press.Col != release.Col ||
			press.Button != release.Button || !slices.Equal(press.Mods, release.Mods) {
			continue
		}
		press.Action = "click"
		return press, true
	}
	return nil, false
}

// parseSGRMouse decodes a single report. The button code carries the modifier
// bits, so it only decodes when every bit maps to a name atago can write back.
func parseSGRMouse(b []byte) (*spec.PTYMouse, bool) {
	match := sgrMouseRe.FindSubmatch(b)
	if match == nil {
		return nil, false
	}
	cb, err := strconv.Atoi(string(match[1]))
	if err != nil {
		return nil, false
	}
	col, err := strconv.Atoi(string(match[2]))
	if err != nil {
		return nil, false
	}
	row, err := strconv.Atoi(string(match[3]))
	if err != nil {
		return nil, false
	}
	button, mods, ok := spec.DecodePTYMouseButton(cb)
	if !ok {
		return nil, false
	}
	action := "press"
	if match[4][0] == 'm' {
		action = "release"
	}
	return &spec.PTYMouse{Row: row, Col: col, Button: button, Action: action, Mods: mods}, true
}

// mouseFlowMapping renders the event as a one-line YAML flow mapping, keeping a
// recorded session as compact as the hand-written form.
func mouseFlowMapping(m *spec.PTYMouse) string {
	parts := []string{
		fmt.Sprintf("row: %d", m.Row),
		fmt.Sprintf("col: %d", m.Col),
		fmt.Sprintf("button: %s", m.Button),
		fmt.Sprintf("action: %s", m.Action),
	}
	if len(m.Mods) > 0 {
		parts = append(parts, fmt.Sprintf("mods: [%s]", strings.Join(m.Mods, ", ")))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// bracketedPaste unwraps a burst the terminal delivered as a paste — the
// PasteStart/PasteEnd markers around the pasted text (#378) — and reports the
// text inside. It requires the markers to bracket the WHOLE burst: a burst that
// merely contains them is something else (a program echoing them back, a paste
// interleaved with typing), and guessing there would rewrite input the replay
// then delivers differently.
func bracketedPaste(input []byte) ([]byte, bool) {
	start, end := []byte(spec.PasteStart), []byte(spec.PasteEnd)
	if len(input) < len(start)+len(end) {
		return nil, false
	}
	if !bytes.HasPrefix(input, start) || !bytes.HasSuffix(input, end) {
		return nil, false
	}
	inner := input[len(start) : len(input)-len(end)]
	if bytes.Contains(inner, start) || bytes.Contains(inner, end) {
		return nil, false
	}
	return inner, true
}

// keyRepeat reports whether input is exactly N (≥ 2) back-to-back copies of one
// named key's byte sequence, and names that key (#377). It scans candidate
// periods shortest-first so the answer is the most repetitions the bytes can be
// read as — "\r\r" is two enters, never one two-byte mystery — and the period
// must divide the input exactly, so a burst with anything else mixed in falls
// through to the literal-text path where nothing is lost.
func keyRepeat(input []byte) (string, int, bool) {
	if len(input) < 2 {
		return "", 0, false
	}
	for period := 1; period <= len(input)/2; period++ {
		if len(input)%period != 0 {
			continue
		}
		times := len(input) / period
		if times > spec.MaxPTYSendTimes {
			continue
		}
		unit := input[:period]
		if !bytes.Equal(input, bytes.Repeat(unit, times)) {
			continue
		}
		if key, ok := spec.PTYKeyForSequence(string(unit)); ok {
			return key, times, true
		}
	}
	return "", 0, false
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
// byte (the NUL standing in for an ANSI sequence, a tab, or any other) and no
// byte that is not valid UTF-8, trimmed of surrounding whitespace. Each such run
// existed verbatim in the raw output, which is what lets an expect built from it
// match the replayed transcript. A rune-by-rune scan silently turned an invalid
// byte into U+FFFD, so a Latin-1 prompt yielded an anchor that appears nowhere in
// the output and an expect that could never match — the session then hung until
// its timeout. Such a byte now breaks the run, and the anchor is the longest
// valid stretch around it.
func longestPlainRun(line string) string {
	best := ""
	var cur strings.Builder
	flush := func() {
		if t := strings.TrimSpace(cur.String()); len(t) > len(best) {
			best = t
		}
		cur.Reset()
	}
	for i := 0; i < len(line); {
		r, size := utf8.DecodeRuneInString(line[i:])
		i += size
		if r < 0x20 || (r == utf8.RuneError && size == 1) {
			flush()
			continue
		}
		cur.WriteRune(r)
	}
	flush()
	return best
}
