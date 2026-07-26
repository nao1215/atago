package spec

import (
	"fmt"
	"strings"

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

// PTYAction is one expect/send/expect_screen entry in a pty session (#8).
type PTYAction struct {
	// Expect waits until the transcript matches this regexp. A never-matching
	// expect fails the step (reported like an assertion) when the session
	// timeout elapses.
	Expect string `yaml:"expect,omitempty"`
	// Send writes to the terminal: a scalar string verbatim (the empty string
	// sends EOF/^D; ${name} expansion applies) or {key: <name>} for a named
	// key (#26) — enter, tab, esc, arrows, f1-f12, ctrl-a..ctrl-z, and common
	// control-key aliases like ctrl-space / ctrl-[ / ctrl-_ — so sessions stay
	// readable instead of embedding \x1b escapes.
	Send *PTYSend `yaml:"send,omitempty"`
	// ExpectScreen waits until the CURRENT rendered screen (the transcript
	// replayed through the same vt10x emulator as a top-level `screen:` assert)
	// satisfies the matcher. `stable_for` requires the matcher to stay true
	// continuously for that long; `timeout` optionally bounds only this wait,
	// within the pty step's wider session timeout.
	ExpectScreen *PTYExpectScreen `yaml:"expect_screen,omitempty"`
}

// PTYSend is the polymorphic pty send payload (#26): exactly one of Text
// (scalar form) or Key (mapping form) is set.
type PTYSend struct {
	// Text is sent verbatim; the empty string transmits EOF (^D).
	Text *string
	// Key is a named key, normalized to lower case.
	Key string
}

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
		if k != "key" {
			return fail("send: unknown key %q (accepted: key)", k)
		}
		str, ok := v.(string)
		if !ok {
			return fail("send.key must be a string")
		}
		p.Key = strings.ToLower(strings.TrimSpace(str))
	}
	if p.Key == "" {
		return fail("send: {key: <name>} requires a key name (e.g. enter, tab, ctrl-c)")
	}
	return nil
}

// MarshalYAML emits the same shape UnmarshalYAML accepts — the scalar text form,
// or the {key: <name>} mapping — so a loaded send round-trips back to a loadable
// spec. Without it the default struct marshal writes a `text:` key the custom
// unmarshaler rejects.
func (p PTYSend) MarshalYAML() (any, error) {
	if p.Text != nil {
		return *p.Text, nil
	}
	return map[string]string{"key": p.Key}, nil
}

// ptyKeySequences maps each named key (#26) to the xterm byte sequence it
// transmits. Documented bytes: enter=\r, tab=\t, esc=\x1b, space=" ",
// backspace=\x7f (DEL, the modern erase), delete=\x1b[3~, arrows
// up/down/right/left=\x1b[A/B/C/D, home=\x1b[H, end=\x1b[F,
// pageup=\x1b[5~, pagedown=\x1b[6~, f1-f4=\x1bOP..\x1bOS,
// f5..f12=\x1b[15~,[17~..[21~,[23~,[24~, ctrl-a..ctrl-z=0x01..0x1a,
// plus the punctuation aliases terminals conventionally expose for the
// remaining C0 controls: ctrl-space/ctrl-@=NUL, ctrl-[=ESC, ctrl-\=FS,
// ctrl-]=GS, ctrl-^=RS, ctrl-_/ctrl-hyphen/ctrl-minus=US. ctrl-d therefore
// stays the readable alias for the empty-send EOF rule.
var ptyKeySequences = func() map[string]string {
	m := map[string]string{
		"enter":     "\r",
		"tab":       "\t",
		"esc":       "\x1b",
		"space":     " ",
		"backspace": "\x7f",
		"delete":    "\x1b[3~",
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
	}
	m["ctrl-space"] = "\x00"
	m["ctrl-@"] = "\x00"
	m["ctrl-["] = "\x1b"
	m["ctrl-\\"] = "\x1c"
	m["ctrl-]"] = "\x1d"
	m["ctrl-^"] = "\x1e"
	m["ctrl-_"] = "\x1f"
	m["ctrl-hyphen"] = "\x1f"
	m["ctrl-minus"] = "\x1f"
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
	}
	m["\x1c"] = "ctrl-\\"
	m["\x1d"] = "ctrl-]"
	m["\x1e"] = "ctrl-^"
	m["\x1f"] = "ctrl-_"
	for _, name := range []string{
		"enter", "tab", "esc", "space", "backspace", "delete",
		"up", "down", "right", "left", "home", "end",
		"pageup", "pagedown",
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
	return "enter, tab, esc, space, backspace, delete, up, down, left, right, home, end, pageup, pagedown, f1-f12, ctrl-a..ctrl-z, ctrl-space/ctrl-@, ctrl-[, ctrl-\\\\, ctrl-], ctrl-^, ctrl-_/ctrl-hyphen/ctrl-minus"
}

// Bytes resolves the send payload to the bytes written to the terminal: the
// named key's xterm sequence, the verbatim text, or 0x04 (VEOF, ^D) for the
// historical empty-string EOF rule.
func (p *PTYSend) Bytes() []byte {
	if p.Key != "" {
		return []byte(ptyKeySequences[p.Key])
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
