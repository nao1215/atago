package loader

import (
	"fmt"
	"regexp"
	"time"

	"github.com/nao1215/atago/internal/diag"
	"github.com/nao1215/atago/internal/spec"
)

// validatePTY checks a pty step (#8): a command, sane duration/size values,
// and a session whose entries each set exactly one of expect/send/expect_screen.
func validatePTY(add addFunc, where string, p *spec.PTY) {
	if p.Command == "" {
		add(diag.RequiredKey, "%s.pty.command is required", where)
	}
	positiveDuration(add, where+".pty.timeout", p.Timeout, "30s", "30s")
	workdirRelativeDir(add, where+".pty.cwd", p.Cwd)
	validateHermeticEnv(add, where+".pty", p.ClearEnv, p.PassEnv)
	// A pty size is a uint16 on the wire; reject values the terminal cannot
	// represent instead of silently truncating.
	if p.Rows < 0 || p.Cols < 0 || p.Rows > 65535 || p.Cols > 65535 {
		add(diag.OutOfRange, "%s.pty: rows/cols must be between 0 and 65535", where)
	}
	for i, a := range p.Session {
		aw := fmt.Sprintf("%s.pty.session[%d]", where, i)
		hasExpect := a.Expect != ""
		hasSend := a.Send != nil
		hasExpectScreen := a.ExpectScreen != nil
		hasResize := a.Resize != nil
		hasExec := a.Exec != nil
		switch {
		case countBools(hasExpect, hasSend, hasExpectScreen, hasResize, hasExec) > 1:
			add(diag.ChooseExactlyOne, "%s: set exactly one of expect/send/expect_screen/resize/exec (got more than one)", aw)
		case !hasExpect && !hasSend && !hasExpectScreen && !hasResize && !hasExec:
			add(diag.ChooseExactlyOne, "%s: set exactly one of expect/send/expect_screen/resize/exec (an empty send: \"\" transmits EOF)", aw)
		case hasExpect:
			if _, err := regexp.Compile(a.Expect); err != nil {
				add(diag.BadRegexp, "%s.expect %q is not a valid regexp: %v", aw, a.Expect, err)
			}
		case hasSend:
			// A named key must be in the vocabulary (#26); a typo'd key would
			// otherwise silently send nothing.
			if a.Send.Key != "" && !spec.ValidPTYKey(a.Send.Key) {
				add(diag.NotAllowedValue, "%s.send.key %q is not a supported key (supported: %s)", aw, a.Send.Key, spec.PTYKeyNames())
			}
			if a.Send.Mouse != nil {
				validatePTYMouse(add, aw+".send.mouse", a.Send.Mouse)
			}
		case hasExpectScreen:
			validatePTYExpectScreen(add, aw+".expect_screen", a.ExpectScreen)
		case hasResize:
			// Both dimensions are required: a resize that names one side would
			// have to invent the other, and a terminal has no natural "keep".
			if a.Resize.Rows < 1 || a.Resize.Cols < 1 {
				add(diag.RequiredKey, "%s.resize: rows and cols are both required and must be at least 1", aw)
			} else if a.Resize.Rows > 65535 || a.Resize.Cols > 65535 {
				add(diag.OutOfRange, "%s.resize: rows/cols must be between 1 and 65535", aw)
			}
		case hasExec:
			if a.Exec.Command == "" {
				add(diag.RequiredKey, "%s.exec.command is required", aw)
			}
			positiveDuration(add, aw+".exec.timeout", a.Exec.Timeout, "10s", "10s")
		}
	}
}

// validatePTYMouse checks a mouse event (#381). Row and column are 1-based
// screen cells with no upper bound checked here: a mid-session resize can change
// the screen size, so the only honest claim at load time is that they name a
// cell at all.
func validatePTYMouse(add addFunc, where string, m *spec.PTYMouse) {
	if m.Row < 1 || m.Col < 1 {
		add(diag.OutOfRange, "%s: row and col are required 1-based screen cells (got row %d, col %d)", where, m.Row, m.Col)
	}
	if m.Button != "" && !spec.ValidPTYMouseButton(m.Button) {
		add(diag.NotAllowedValue, "%s.button %q is not a supported button (supported: %s)", where, m.Button, spec.PTYMouseButtonNames())
	}
	if m.Action != "" && !spec.ValidPTYMouseAction(m.Action) {
		add(diag.NotAllowedValue, "%s.action %q is not a supported action (supported: %s)", where, m.Action, spec.PTYMouseActionNames())
	}
	// A wheel notch is a single event in the SGR encoding — there is nothing to
	// release — so asking for one is a mistake worth naming rather than a no-op.
	if m.IsWheel() && m.Action == "release" {
		add(diag.ExclusiveKeys, "%s: a wheel button has no release event; use action: press (the default click sends one notch)", where)
	}
	seen := map[string]bool{}
	for _, mod := range m.Mods {
		if !spec.ValidPTYMouseMod(mod) {
			add(diag.NotAllowedValue, "%s.mods %q is not a supported modifier (supported: %s)", where, mod, spec.PTYMouseModNames())
			continue
		}
		// A modifier is a bit, so naming it twice says nothing new — and a spec
		// that repeats one is far more likely to have meant two different ones.
		if seen[mod] {
			add(diag.DuplicateEntry, "%s.mods lists %q twice", where, mod)
		}
		seen[mod] = true
	}
}

func validatePTYExpectScreen(add addFunc, where string, es *spec.PTYExpectScreen) {
	// Share the screen validator so a mid-session wait accepts exactly what the
	// post-step assert does — including an attrs-only wait (#382).
	validateScreen(add, where, &es.ScreenAssert)
	if es.Snapshot != "" {
		add(diag.KeyNotHere, "%s.snapshot is not supported in expect_screen; use a post-step assert screen snapshot or text matchers here", where)
	}
	if es.Trim != nil {
		add(diag.KeyNotHere, "%s.trim is not supported in expect_screen", where)
	}
	positiveDuration(add, where+".timeout", es.Timeout, "", "")
	positiveDuration(add, where+".stable_for", es.StableFor, "", "")
	if es.Timeout != "" && es.StableFor != "" {
		timeout, terr := time.ParseDuration(es.Timeout)
		stable, serr := time.ParseDuration(es.StableFor)
		if terr == nil && serr == nil && stable > timeout {
			add(diag.EmptyInterval, "%s.stable_for %q must not exceed %s.timeout %q", where, es.StableFor, where, es.Timeout)
		}
	}
}

func countBools(xs ...bool) int {
	n := 0
	for _, x := range xs {
		if x {
			n++
		}
	}
	return n
}
