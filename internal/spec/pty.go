package spec

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// ClearEnvEnabled reports whether the pty step opts into a cleared environment (#16).
func (p *PTY) ClearEnvEnabled() bool { return p.ClearEnv != nil && *p.ClearEnv }

// SandboxHomeEnabled reports whether the pty step opts into an isolated home (#71).
func (p *PTY) SandboxHomeEnabled() bool { return p.SandboxHome != nil && *p.SandboxHome }

// PTY runs a command inside a pseudo-terminal (#8). The captured transcript
// (terminal echo included, ANSI intact) becomes the step's stdout, so every
// stream matcher, snapshot (with its ANSI normalization), and
// `store from.stdout` works unchanged.
type PTY struct {
	Command string `yaml:"command"`
	// Shell runs Command through the shell like run.shell.
	Shell *bool  `yaml:"shell,omitempty"`
	Cwd   string `yaml:"cwd,omitempty"`
	// Rows / Cols set the terminal size (default 24x80).
	Rows int `yaml:"rows,omitempty"`
	Cols int `yaml:"cols,omitempty"`
	// Timeout bounds the WHOLE session as a Go duration (default "30s"): a
	// prompt that never appears or a program that never exits fails loudly
	// instead of hanging the run.
	Timeout string            `yaml:"timeout,omitempty"`
	Env     map[string]string `yaml:"env,omitempty"`
	// ClearEnv starts the pty child from an empty environment instead of
	// inheriting the host environment (#16), mirroring run.clear_env.
	ClearEnv *bool `yaml:"clear_env,omitempty"`
	// PassEnv copies the listed host variables into the cleared environment
	// (#16). Only meaningful with ClearEnv; unset host variables are skipped.
	PassEnv []string `yaml:"pass_env,omitempty"`
	// SandboxHome isolates the pty child's home and per-OS config/cache/data/
	// state directories under `${workdir}/.atago-home`, mirroring run.sandbox_home.
	SandboxHome *bool `yaml:"sandbox_home,omitempty"`
	// Session is the ordered expect/send script. Each entry sets exactly one
	// of Expect (wait until the accumulated transcript matches the regexp),
	// Send (write the string to the terminal; an empty send transmits EOF,
	// i.e. ^D), or ExpectScreen (wait until the rendered terminal screen matches
	// a stream matcher, optionally stably for a duration). Deliberately no
	// branching — atago is not a scripting language.
	Session []PTYAction `yaml:"session,omitempty"`
}

// PTYAction is one expect/send/expect_screen/resize entry in a pty session (#8).
type PTYAction struct {
	// Expect waits until the transcript matches this regexp. A never-matching
	// expect fails the step (reported like an assertion) when the session
	// timeout elapses.
	Expect string `yaml:"expect,omitempty"`
	// Send writes to the terminal: a scalar string verbatim (the empty string
	// sends EOF/^D; ${name} expansion applies) or {key: <name>} for a named
	// key (#26) — enter, tab, shift-tab, esc, arrows, f1-f12, ctrl-a..ctrl-z,
	// alt-a..alt-z, modified arrows like ctrl-left, and common control-key
	// aliases like ctrl-space / ctrl-[ / ctrl-_ plus terminal key events like
	// ctrl-hyphen (#376) — so sessions stay readable instead of embedding \x1b
	// escapes.
	Send *PTYSend `yaml:"send,omitempty"`
	// Exec runs one command on the HOST while the program under test keeps
	// running (#380), so a session can test what a TUI does when the world
	// changes underneath it — a commit made outside lazygit, a file another
	// process creates, a log line appended to what a viewer is following.
	// Everything else in a session is caused by keystrokes; this is the one
	// action that is not. It blocks until the command exits, which is the
	// point: after it, the change exists, and the expect_screen that follows is
	// waiting for the program to notice.
	Exec *PTYExec `yaml:"exec,omitempty"`
	// Resize changes the terminal size mid-session (#379), delivering the size
	// change the way a real terminal does — SIGWINCH on POSIX, a ConPTY
	// notification on Windows — so a TUI's relayout is testable instead of
	// being fixed at whatever the step started with.
	Resize *PTYResize `yaml:"resize,omitempty"`
	// ExpectScreen waits until the CURRENT rendered screen (the transcript
	// replayed through the same vt10x emulator as a top-level `screen:` assert)
	// satisfies the matcher. `stable_for` requires the matcher to stay true
	// continuously for that long; `timeout` optionally bounds only this wait,
	// within the pty step's wider session timeout.
	ExpectScreen *PTYExpectScreen `yaml:"expect_screen,omitempty"`
}

// PTYExec is one host command run mid-session (#380). It accepts either a
// scalar string (argv-parsed like run.command) or a mapping carrying the shell
// and timeout knobs.
//
// The command runs in the scenario workdir with the same environment the pty
// child received, so `sandbox_home` / `clear_env` isolation is not quietly
// punctured by the helper. Its output never joins the transcript — the
// transcript is what the TERMINAL showed — and is kept only to explain a
// failure. A non-zero exit, a timeout, or a failure to start is a hard step
// error: the command is scaffolding, not the subject under test, so a broken
// one must stop the run rather than leave the following expect_screen waiting
// for a change that was never made.
//
// It is deliberately not a scripting language: a fixed command at a fixed point
// in the session, no branching, no captured output feeding later steps.
type PTYExec struct {
	// Command is the program to run. Required.
	Command string `yaml:"command"`
	// Shell runs Command through the shell, like run.shell.
	Shell *bool `yaml:"shell,omitempty"`
	// Timeout bounds this command as a Go duration (default 10s). The remaining
	// session budget bounds it too, whichever is shorter.
	Timeout string `yaml:"timeout,omitempty"`

	// mapped records that the author used the mapping form, so an empty mapping
	// ({}) is rejected rather than read as "no command".
	mapped bool
}

// ShellEnabled reports whether the exec command runs through the shell.
func (e *PTYExec) ShellEnabled() bool { return e.Shell != nil && *e.Shell }

// DefaultPTYExecTimeout bounds a mid-session host command when the spec does
// not (#380). Scaffolding a session waits on should be quick; a command that is
// not is one the author should have to bound on purpose.
const DefaultPTYExecTimeout = 10 * time.Second

// UnmarshalYAML decodes exec as a scalar command string or a mapping, rejecting
// unknown mapping keys (a custom unmarshaler bypasses the loader's strict
// decode). It decodes from the AST node so every shape error carries the
// offending value's [line:col].
func (e *PTYExec) UnmarshalYAML(node ast.Node) error {
	fail := func(format string, args ...any) error {
		return &yaml.SyntaxError{Message: fmt.Sprintf(format, args...), Token: node.GetToken()}
	}
	var one string
	if err := yaml.NodeToValue(node, &one); err == nil {
		e.Command = one
		return nil
	}
	var raw map[string]any
	if err := yaml.NodeToValue(node, &raw); err != nil {
		return fail("exec must be a string or {command: ..., shell: bool, timeout: duration}")
	}
	e.mapped = true
	for k, v := range raw {
		switch k {
		case "command":
			str, ok := v.(string)
			if !ok {
				return fail("exec.command must be a string")
			}
			e.Command = str
		case "shell":
			b, ok := v.(bool)
			if !ok {
				return fail("exec.shell must be true or false")
			}
			e.Shell = &b
		case "timeout":
			str, ok := v.(string)
			if !ok {
				return fail("exec.timeout must be a duration string (e.g. \"10s\")")
			}
			e.Timeout = str
		default:
			return fail("exec: unknown key %q (accepted: command, shell, timeout)", k)
		}
	}
	return nil
}

// MarshalYAML emits the shape UnmarshalYAML accepts, so a loaded exec
// round-trips: the scalar form when nothing else was set, the mapping otherwise.
func (e PTYExec) MarshalYAML() (any, error) {
	if !e.mapped && e.Shell == nil && e.Timeout == "" {
		return e.Command, nil
	}
	m := map[string]any{"command": e.Command}
	if e.Shell != nil {
		m["shell"] = *e.Shell
	}
	if e.Timeout != "" {
		m["timeout"] = e.Timeout
	}
	return m, nil
}

// Label renders the exec symbolically for explain/doc (#380).
func (e *PTYExec) Label() string { return fmt.Sprintf("exec %q", e.Command) }

// PTYResize is a mid-session terminal resize (#379). Both dimensions are
// required — there is no "keep the other one", because a spec that says what
// the window becomes reads better than one that says what it changes by.
//
// The size change reaches the child the way a real terminal delivers it, so a
// program that redraws on SIGWINCH redraws. The rendered screen follows: every
// later `expect_screen`, the post-step `screen:` assert, and the snapshot all
// see the transcript replayed at the sizes it was actually produced under.
//
// Authoring rule: settle the screen (an `expect` or `expect_screen`, ideally
// with `stable_for`) before and after a resize. Output already in flight when
// the resize lands is attributed to the old size, exactly as a real terminal
// would — waiting for a quiet screen is what makes the boundary unambiguous.
type PTYResize struct {
	Rows int `yaml:"rows"`
	Cols int `yaml:"cols"`
}

// PTYSend is the polymorphic pty send payload (#26): exactly one of Text
// (scalar form) or Key (mapping form) is set.
type PTYSend struct {
	// Text is sent verbatim; the empty string transmits EOF (^D).
	Text *string
	// Key is a named key, normalized to lower case.
	Key string
	// Times repeats a named key (#377), so sixteen presses of an arrow are one
	// readable action instead of sixteen identical session entries. Zero and one
	// both mean a single press; the repeats go out as ONE terminal write, which
	// is what a held key looks like to the program reading it. Only valid with
	// Key: repeating text is already expressible inline.
	Times int
	// Paste is text delivered as a BRACKETED PASTE (#378): the bytes go out
	// wrapped in the markers a terminal puts around pasted input, so a program
	// that distinguishes a paste from typing — a REPL that must not execute a
	// pasted block line by line, an editor that must not auto-indent it — takes
	// its paste path. The empty string is an empty paste, not the scalar form's
	// EOF. Mutually exclusive with Key and with the scalar Text form.
	Paste *string
}

// MaxPTYSendTimes bounds the repeat count (#377). Navigation never needs
// anywhere near it, and the cap keeps a typo'd `times: 100000000` from writing
// gigabytes into the terminal instead of failing at the mistake.
const MaxPTYSendTimes = 10000

// The xterm bracketed-paste markers, i.e. what a terminal wraps pasted input in
// once the application has turned DEC private mode 2004 on (#378). A terminal
// emits them ONLY after that request, which is why sending a paste is gated on
// having seen the program ask.
const (
	PasteStart = "\x1b[200~"
	PasteEnd   = "\x1b[201~"
)

// PTYExpectScreen is a session-local rendered-screen wait: the matcher runs on
// the live terminal screen during a pty session, not only after the program
// exits. It reuses the StreamAssert surface (line/contains/matches/equals/json/
// yaml, etc.) except snapshot/trim, which are validated out of this
// mid-session context.
type PTYExpectScreen struct {
	StreamAssert `yaml:",inline"`
	// Timeout bounds THIS wait only; when empty, the enclosing pty timeout
	// supplies the budget.
	Timeout string `yaml:"timeout,omitempty"`
	// StableFor requires the screen to keep matching continuously for at least
	// this duration before the action passes, absorbing redraw churn without a
	// blind sleep.
	StableFor string `yaml:"stable_for,omitempty"`
}

// SendText is sugar for authoring the scalar form in Go literals (tests).
func SendText(s string) *PTYSend { return &PTYSend{Text: &s} }

// UnmarshalYAML decodes send as a scalar string or a {key: name} mapping,
// rejecting unknown mapping keys (a custom unmarshaler bypasses the loader's
// strict decode). It decodes from the AST node so every shape error carries
// the offending value's [line:col] for the loader's excerpt formatter.
func (p *PTYSend) UnmarshalYAML(node ast.Node) error {
	fail := func(format string, args ...any) error {
		return &yaml.SyntaxError{Message: fmt.Sprintf(format, args...), Token: node.GetToken()}
	}
	var one string
	if err := yaml.NodeToValue(node, &one); err == nil {
		p.Text = &one
		return nil
	}
	var raw map[string]any
	if err := yaml.NodeToValue(node, &raw); err != nil {
		return fail("send must be a string or {key: <name>} (e.g. {key: enter})")
	}
	for k, v := range raw {
		switch k {
		case "key":
			str, ok := v.(string)
			if !ok {
				return fail("send.key must be a string")
			}
			p.Key = strings.ToLower(strings.TrimSpace(str))
		case "times":
			// The range check lives here rather than in the loader because this
			// is the only layer that can still tell an authored `times: 0` from
			// an omitted one — the decoded field is a plain int, so by the time
			// validation runs the two are the same value.
			n, ok := asInt(v)
			if !ok {
				return fail("send.times must be an integer (e.g. {key: left, times: 16})")
			}
			if n < 1 {
				return fail("send.times must be at least 1 (got %d); omit it for a single press", n)
			}
			if n > MaxPTYSendTimes {
				return fail("send.times must not exceed %d (got %d)", MaxPTYSendTimes, n)
			}
			p.Times = n
		case "paste":
			str, ok := v.(string)
			if !ok {
				return fail("send.paste must be a string")
			}
			p.Paste = &str
		default:
			return fail("send: unknown key %q (accepted: key, times, paste)", k)
		}
	}
	if p.Key != "" && p.Paste != nil {
		return fail("send: set either {key: <name>} or {paste: <text>}, not both")
	}
	if p.Paste != nil {
		if p.Times != 0 {
			return fail("send.times repeats a named key; a paste is delivered once")
		}
		return nil
	}
	if p.Key == "" {
		// This is also what rejects a lone `{times: N}`: repeating verbatim text
		// is already expressible inline, so times only ever qualifies a key.
		return fail("send: {key: <name>} requires a key name (e.g. enter, tab, ctrl-c), or use {paste: <text>}")
	}
	return nil
}

// asInt accepts the several numeric shapes a YAML decoder may hand back for an
// integer scalar, and rejects a float that is not whole (`times: 1.5` is a
// mistake, not a truncation).
func asInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case uint64:
		if n > math.MaxInt32 {
			return 0, false
		}
		return int(n), true
	case float64:
		if n != math.Trunc(n) {
			return 0, false
		}
		return int(n), true
	default:
		return 0, false
	}
}

// MarshalYAML emits the same shape UnmarshalYAML accepts — the scalar text form,
// or the {key: <name>} mapping — so a loaded send round-trips back to a loadable
// spec. Without it the default struct marshal writes a `text:` key the custom
// unmarshaler rejects.
func (p PTYSend) MarshalYAML() (any, error) {
	if p.Text != nil {
		return *p.Text, nil
	}
	if p.Paste != nil {
		return map[string]string{"paste": *p.Paste}, nil
	}
	if p.Times != 0 {
		return map[string]any{"key": p.Key, "times": p.Times}, nil
	}
	return map[string]string{"key": p.Key}, nil
}

// ptyKeySequences maps each named key (#26) to the xterm byte sequence it
// transmits. Documented bytes: enter=\r, tab=\t, esc=\x1b, space=" ",
// backspace=\x7f (DEL, the modern erase), delete=\x1b[3~, insert=\x1b[2~,
// arrows up/down/right/left=\x1b[A/B/C/D, home=\x1b[H, end=\x1b[F,
// pageup=\x1b[5~, pagedown=\x1b[6~, f1-f4=\x1bOP..\x1bOS,
// f5..f12=\x1b[15~,[17~..[21~,[23~,[24~, ctrl-a..ctrl-z=0x01..0x1a,
// plus the punctuation aliases terminals conventionally expose for the
// remaining C0 controls: ctrl-space/ctrl-@=NUL, ctrl-[=ESC, ctrl-\=FS,
// ctrl-]=GS, ctrl-^=RS, ctrl-_=US. For modifier combos whose physical key does
// not have a stable legacy C0 byte (e.g. Ctrl+-), use the terminal's CSI-u key
// event instead so modern TUIs like Yazi see the intended modified key. ctrl-d
// therefore stays the readable alias for the empty-send EOF rule.
//
// Three more families cover the chords real TUIs bind (#376), each the sequence
// an xterm-class terminal sends for that physical key:
//
//   - shift-tab=\x1b[Z (CSI Z, "backtab"), the reverse-focus key every
//     form-style TUI reads. `backtab` is an accepted alias for the same bytes.
//   - alt-<letter>=\x1b<letter>, the ESC prefix a terminal transmits for a Meta
//     chord (readline word operations, helix/emacs-style bindings), plus
//     alt-enter=\x1b\r and alt-backspace=\x1b\x7f. Letters only: a Meta digit
//     has no encoding stable enough to promise.
//   - Modified arrows, in xterm's CSI 1 ; <modifier> <final> form —
//     ctrl-<arrow>=\x1b[1;5A..D (word-wise cursor movement) and
//     shift-<arrow>=\x1b[1;2A..D (selection movement).
var ptyKeySequences = func() map[string]string {
	m := map[string]string{
		"enter":     "\r",
		"tab":       "\t",
		"esc":       "\x1b",
		"space":     " ",
		"backspace": "\x7f",
		"delete":    "\x1b[3~",
		"insert":    "\x1b[2~",
		"shift-tab": "\x1b[Z",
		"backtab":   "\x1b[Z",
		"up":        "\x1b[A",
		"down":      "\x1b[B",
		"right":     "\x1b[C",
		"left":      "\x1b[D",
		"home":      "\x1b[H",
		"end":       "\x1b[F",
		"pageup":    "\x1b[5~",
		"pagedown":  "\x1b[6~",
		"f1":        "\x1bOP",
		"f2":        "\x1bOQ",
		"f3":        "\x1bOR",
		"f4":        "\x1bOS",
		"f5":        "\x1b[15~",
		"f6":        "\x1b[17~",
		"f7":        "\x1b[18~",
		"f8":        "\x1b[19~",
		"f9":        "\x1b[20~",
		"f10":       "\x1b[21~",
		"f11":       "\x1b[23~",
		"f12":       "\x1b[24~",
	}
	for c := byte('a'); c <= 'z'; c++ {
		m["ctrl-"+string(c)] = string([]byte{c - 'a' + 1})
		m["alt-"+string(c)] = "\x1b" + string(c)
	}
	m["alt-enter"] = "\x1b\r"
	m["alt-backspace"] = "\x1b\x7f"
	// Modified arrows share one shape: CSI 1 ; <modifier> <final>, where the
	// modifier is 1+bitmask (shift 1, alt 2, ctrl 4) — so ctrl is 5 and shift 2.
	for i, final := range []string{"A", "B", "C", "D"} {
		name := []string{"up", "down", "right", "left"}[i]
		m["ctrl-"+name] = "\x1b[1;5" + final
		m["shift-"+name] = "\x1b[1;2" + final
	}
	m["ctrl-space"] = "\x00"
	m["ctrl-@"] = "\x00"
	m["ctrl-["] = "\x1b"
	m["ctrl-\\"] = "\x1c"
	m["ctrl-]"] = "\x1d"
	m["ctrl-^"] = "\x1e"
	m["ctrl-_"] = "\x1f"
	// xterm/kitty CSI-u for Ctrl+- (#286): raw 0x1f is Ctrl+_ and does not
	// trigger TUIs that bind the physical hyphen key as a distinct modified key.
	m["ctrl-hyphen"] = "\x1b[45;5u"
	m["ctrl-minus"] = "\x1b[45;5u"
	return m
}()

// ValidPTYKey reports whether name is in the named-key vocabulary (#26).
func ValidPTYKey(name string) bool {
	_, ok := ptyKeySequences[strings.ToLower(strings.TrimSpace(name))]
	return ok
}

// ptyKeyBySequence reverse-maps an xterm byte sequence to its friendly key name
// (#69), preferring the readable name over a ctrl-* alias when a byte is shared
// (e.g. \r is both enter and ctrl-m — enter wins). Built once at init: the
// ctrl-* aliases go in first, then the friendly names overwrite any collision.
var ptyKeyBySequence = func() map[string]string {
	m := make(map[string]string, len(ptyKeySequences))
	m["\x00"] = "ctrl-space"
	for c := byte('a'); c <= 'z'; c++ {
		m[string([]byte{c - 'a' + 1})] = "ctrl-" + string(c)
		// alt-<letter> is two bytes (ESC + the letter), so it never collides
		// with a single-byte ctrl chord or with a CSI sequence. A user who
		// pressed Esc and then typed the letter produces the same bytes, and
		// replaying {key: alt-x} reproduces them exactly either way.
		m["\x1b"+string(c)] = "alt-" + string(c)
	}
	m["\x1c"] = "ctrl-\\"
	m["\x1d"] = "ctrl-]"
	m["\x1e"] = "ctrl-^"
	m["\x1f"] = "ctrl-_"
	m["\x1b[45;5u"] = "ctrl-hyphen"
	// backtab is deliberately absent: it and shift-tab share \x1b[Z, and a
	// recording reads better naming the key a keyboard has.
	for _, name := range []string{
		"enter", "tab", "esc", "space", "backspace", "delete", "insert",
		"shift-tab",
		"up", "down", "right", "left", "home", "end",
		"pageup", "pagedown",
		"alt-enter", "alt-backspace",
		"ctrl-up", "ctrl-down", "ctrl-right", "ctrl-left",
		"shift-up", "shift-down", "shift-right", "shift-left",
		"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11", "f12",
	} {
		m[ptyKeySequences[name]] = name
	}
	return m
}()

// PTYKeyForSequence returns the friendly named key whose xterm sequence exactly
// equals seq (#69), so `atago record --pty` can render a lone control key as
// {key: <name>} instead of an opaque escape. It reports false when no named key
// matches the bytes exactly.
func PTYKeyForSequence(seq string) (string, bool) {
	name, ok := ptyKeyBySequence[seq]
	return name, ok
}

// PTYKeyNames lists the vocabulary for error messages, compactly.
func PTYKeyNames() string {
	return "enter, tab, esc, space, backspace, delete, insert, up, down, left, right, home, end, pageup, pagedown, f1-f12, " +
		"shift-tab (alias backtab), alt-a..alt-z, alt-enter, alt-backspace, " +
		"ctrl-up/ctrl-down/ctrl-left/ctrl-right, shift-up/shift-down/shift-left/shift-right, " +
		"ctrl-a..ctrl-z, ctrl-space/ctrl-@, ctrl-[, ctrl-\\, ctrl-], ctrl-^, ctrl-_, ctrl-hyphen/ctrl-minus"
}

// Bytes resolves the send payload to the bytes written to the terminal: the
// named key's xterm sequence, the verbatim text, or 0x04 (VEOF, ^D) for the
// historical empty-string EOF rule.
func (p *PTYSend) Bytes() []byte {
	if p.Paste != nil {
		b := make([]byte, 0, len(PasteStart)+len(*p.Paste)+len(PasteEnd))
		b = append(b, PasteStart...)
		b = append(b, *p.Paste...)
		return append(b, PasteEnd...)
	}
	if p.Key != "" {
		seq := []byte(ptyKeySequences[p.Key])
		if p.Times > 1 {
			// One write, not p.Times writes: a held key reaches the program as a
			// burst in a single read, and splitting it would let a redraw land
			// between presses and change what the program sees.
			return bytes.Repeat(seq, p.Times)
		}
		return seq
	}
	if p.Text != nil && *p.Text == "" {
		return []byte{0x04}
	}
	if p.Text != nil {
		return []byte(*p.Text)
	}
	return nil
}

// Label renders the send symbolically for explain/doc (#26): "press Enter"
// for keys, a quoted excerpt for text, "EOF (^D)" for the empty string.
func (p *PTYSend) Label() string {
	switch {
	case p.Paste != nil:
		return fmt.Sprintf("paste %q", *p.Paste)
	case p.Key != "" && p.Times > 1:
		return fmt.Sprintf("press %s x%d", p.Key, p.Times)
	case p.Key != "":
		return "press " + p.Key
	case p.Text != nil && *p.Text == "":
		return "send EOF (^D)"
	case p.Text != nil:
		return fmt.Sprintf("type %q", *p.Text)
	default:
		return "send"
	}
}
