package spec

import (
	"fmt"
	"slices"
	"strings"
)

// PTYMouse is one mouse event (#381), encoded in xterm's SGR (1006) form.
type PTYMouse struct {
	// Row and Col are the 1-based screen cell, as a person would count them.
	Row int `yaml:"row"`
	Col int `yaml:"col"`
	// Button is left (default), middle, right, wheel-up, or wheel-down.
	Button string `yaml:"button,omitempty"`
	// Action is click (default: press and release in one write), press, or
	// release. A wheel button has no release, so it rejects one.
	Action string `yaml:"action,omitempty"`
	// Mods are shift, alt, and/or ctrl.
	Mods []string `yaml:"mods,omitempty"`
}

// Mouse button and action vocabularies, and the SGR base codes each button
// carries. Wheel buttons live at 64+ in xterm's encoding, which is why they are
// not simply the next numbers after right.
var (
	ptyMouseButtons = map[string]int{
		"left":       0,
		"middle":     1,
		"right":      2,
		"wheel-up":   64,
		"wheel-down": 65,
	}
	ptyMouseActions = map[string]bool{"click": true, "press": true, "release": true}
	// Modifier bits, as xterm adds them to the button code.
	ptyMouseMods = map[string]int{"shift": 4, "alt": 8, "ctrl": 16}
)

// PTYMouseButtonNames and friends spell the vocabularies for error messages.
func PTYMouseButtonNames() string { return "left, middle, right, wheel-up, wheel-down" }
func PTYMouseActionNames() string { return "click, press, release" }
func PTYMouseModNames() string    { return "shift, alt, ctrl" }

// ValidPTYMouseButton, ValidPTYMouseAction, and ValidPTYMouseMod report
// membership in each vocabulary, so the loader rejects a typo instead of
// sending an event the program cannot read.
func ValidPTYMouseButton(name string) bool { _, ok := ptyMouseButtons[name]; return ok }
func ValidPTYMouseAction(name string) bool { return ptyMouseActions[name] }
func ValidPTYMouseMod(name string) bool    { _, ok := ptyMouseMods[name]; return ok }

// IsWheel reports whether the button is a scroll wheel, which has no release
// event in the SGR encoding.
func (m *PTYMouse) IsWheel() bool {
	return m.Button == "wheel-up" || m.Button == "wheel-down"
}

// ButtonOrDefault and ActionOrDefault fill in the documented defaults, so the
// encoder and the validator agree about what an omitted field means.
func (m *PTYMouse) ButtonOrDefault() string {
	if m.Button == "" {
		return "left"
	}
	return m.Button
}

func (m *PTYMouse) ActionOrDefault() string {
	if m.Action == "" {
		return "click"
	}
	return m.Action
}

// Bytes encodes the event as xterm SGR (1006) mouse reports: press is
// "CSI < Cb ; COL ; ROW M" and release the same with a lowercase "m", where Cb
// is the button's base code plus the modifier bits. A click is a press
// immediately followed by its release, in one write, because that is what a
// real click delivers and splitting it would let a redraw land between the two.
func (m *PTYMouse) Bytes() []byte {
	cb := ptyMouseButtons[m.ButtonOrDefault()]
	for _, mod := range m.Mods {
		// OR, not add: the modifiers are bit flags, so repeating one must not
		// carry into the next bit. `mods: [ctrl, ctrl]` summed to 32, which is
		// the MOTION bit — a report atago does not send and a TUI would read as
		// something else entirely.
		cb |= ptyMouseMods[mod]
	}
	press := fmt.Sprintf("\x1b[<%d;%d;%dM", cb, m.Col, m.Row)
	release := fmt.Sprintf("\x1b[<%d;%d;%dm", cb, m.Col, m.Row)
	switch {
	case m.ActionOrDefault() == "press":
		return []byte(press)
	case m.ActionOrDefault() == "release":
		return []byte(release)
	case m.IsWheel():
		// A wheel notch is a single event: there is no release to pair with it,
		// and sending one would put a report on the wire that no real terminal
		// produces. `click` on a wheel therefore means one notch.
		return []byte(press)
	default:
		return []byte(press + release)
	}
}

// DecodePTYMouseButton reverses the SGR button code back into the button name
// and modifier names atago writes (#381), so `atago record --pty` can render a
// captured report as the event it describes. It reports false when the code
// carries a bit atago has no name for — motion reporting, a button beyond the
// three it names — because a recording must replay the bytes it captured, and
// naming only part of them would not.
func DecodePTYMouseButton(cb int) (button string, mods []string, ok bool) {
	// Peel the modifier bits off first; what remains has to be a button.
	remaining := cb
	for _, name := range []string{"ctrl", "alt", "shift"} { // high bit first
		if bit := ptyMouseMods[name]; remaining&bit != 0 {
			remaining -= bit
			mods = append(mods, name)
		}
	}
	// Report in the order the vocabulary documents, not the order peeled.
	slices.Reverse(mods)
	for name, base := range ptyMouseButtons {
		if base == remaining {
			return name, mods, true
		}
	}
	return "", nil, false
}

// Label renders the event for explain/doc (#381).
func (m *PTYMouse) Label() string {
	verb := m.ButtonOrDefault()
	if a := m.ActionOrDefault(); a != "click" || m.IsWheel() {
		verb += "-" + a
	}
	if len(m.Mods) > 0 {
		verb = strings.Join(m.Mods, "+") + "+" + verb
	}
	return fmt.Sprintf("%s at (%d,%d)", verb, m.Row, m.Col)
}

// decodeMouse fills a PTYMouse from the decoded mapping, normalizing the
// vocabularies to lower case. Membership is the loader's job; this only rejects
// shapes (a row that is not a number, a mods list that is not a list of
// strings), so every error still carries the offending node's [line:col].
func decodeMouse(raw map[string]any, m *PTYMouse, fail func(string, ...any) error) error {
	for k, v := range raw {
		switch k {
		case "row", "col":
			n, ok := asInt(v)
			if !ok {
				return fail("send.mouse.%s must be an integer", k)
			}
			if k == "row" {
				m.Row = n
			} else {
				m.Col = n
			}
		case "button", "action":
			str, ok := v.(string)
			if !ok {
				return fail("send.mouse.%s must be a string", k)
			}
			norm := strings.ToLower(strings.TrimSpace(str))
			if k == "button" {
				m.Button = norm
			} else {
				m.Action = norm
			}
		case "mods":
			list, ok := v.([]any)
			if !ok {
				return fail("send.mouse.mods must be a list (e.g. [ctrl])")
			}
			for _, item := range list {
				str, ok := item.(string)
				if !ok {
					return fail("send.mouse.mods entries must be strings (accepted: %s)", PTYMouseModNames())
				}
				m.Mods = append(m.Mods, strings.ToLower(strings.TrimSpace(str)))
			}
		default:
			return fail("send.mouse: unknown key %q (accepted: row, col, button, action, mods)", k)
		}
	}
	return nil
}
