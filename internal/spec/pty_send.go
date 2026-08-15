package spec

import (
	"bytes"
	"fmt"
	"math"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

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
	// Mouse is a mouse event delivered as an SGR (1006) report (#381), for the
	// TUIs that accept clicks and scrolling — lazygit, yazi, htop, fzf --mouse,
	// anything on bubbletea. It is also sometimes the only sane way to reach a
	// target: clicking a pane beats walking to it with twenty keystrokes.
	// Mutually exclusive with Key, Paste, and the scalar Text form.
	Mouse *PTYMouse `yaml:"mouse,omitempty"`
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
		if err := p.decodeSendField(k, v, fail); err != nil {
			return err
		}
	}
	return p.checkSendShape(fail)
}

// decodeSendField decodes one key of the send mapping, rejecting unknown keys.
func (p *PTYSend) decodeSendField(k string, v any, fail func(string, ...any) error) error {
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
	case "mouse":
		var mouse PTYMouse
		raw, ok := v.(map[string]any)
		if !ok {
			return fail("send.mouse must be a mapping (e.g. {mouse: {row: 5, col: 12}})")
		}
		if err := decodeMouse(raw, &mouse, fail); err != nil {
			return err
		}
		p.Mouse = &mouse
	default:
		return fail("send: unknown key %q (accepted: key, times, paste, mouse)", k)
	}
	return nil
}

// checkSendShape enforces the mapping form's one-of rule (key/paste/mouse) and
// that times only ever qualifies a named key.
func (p *PTYSend) checkSendShape(fail func(string, ...any) error) error {
	if countSet(p.Key != "", p.Paste != nil, p.Mouse != nil) > 1 {
		return fail("send: set exactly one of {key: <name>}, {paste: <text>}, or {mouse: {...}}")
	}
	if p.Mouse != nil {
		if p.Times != 0 {
			return fail("send.times repeats a named key; a mouse event is delivered once")
		}
		return nil
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

// countSet counts how many of the flags are true, for the one-of rules.
func countSet(flags ...bool) int {
	n := 0
	for _, f := range flags {
		if f {
			n++
		}
	}
	return n
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
	if p.Mouse != nil {
		mouse := map[string]any{"row": p.Mouse.Row, "col": p.Mouse.Col}
		if p.Mouse.Button != "" {
			mouse["button"] = p.Mouse.Button
		}
		if p.Mouse.Action != "" {
			mouse["action"] = p.Mouse.Action
		}
		if len(p.Mouse.Mods) > 0 {
			mouse["mods"] = p.Mouse.Mods
		}
		return map[string]any{"mouse": mouse}, nil
	}
	if p.Times != 0 {
		return map[string]any{"key": p.Key, "times": p.Times}, nil
	}
	return map[string]string{"key": p.Key}, nil
}

// Bytes resolves the send payload to the bytes written to the terminal: the
// named key's xterm sequence, the verbatim text, or 0x04 (VEOF, ^D) for the
// historical empty-string EOF rule.
func (p *PTYSend) Bytes() []byte {
	if p.Mouse != nil {
		return p.Mouse.Bytes()
	}
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
	case p.Mouse != nil:
		return p.Mouse.Label()
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
