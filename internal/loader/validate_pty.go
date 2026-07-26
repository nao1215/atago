package loader

import (
	"fmt"
	"regexp"
	"time"

	"github.com/nao1215/atago/internal/spec"
)

// validatePTY checks a pty step (#8): a command, sane duration/size values,
// and a session whose entries each set exactly one of expect/send/expect_screen.
func validatePTY(add func(string, ...any), where string, p *spec.PTY) {
	if p.Command == "" {
		add("%s.pty.command is required", where)
	}
	positiveDuration(add, where+".pty.timeout", p.Timeout, "30s", "30s")
	validateHermeticEnv(add, where+".pty", p.ClearEnv, p.PassEnv)
	// A pty size is a uint16 on the wire; reject values the terminal cannot
	// represent instead of silently truncating.
	if p.Rows < 0 || p.Cols < 0 || p.Rows > 65535 || p.Cols > 65535 {
		add("%s.pty: rows/cols must be between 0 and 65535", where)
	}
	for i, a := range p.Session {
		aw := fmt.Sprintf("%s.pty.session[%d]", where, i)
		hasExpect := a.Expect != ""
		hasSend := a.Send != nil
		hasExpectScreen := a.ExpectScreen != nil
		switch {
		case countBools(hasExpect, hasSend, hasExpectScreen) > 1:
			add("%s: set exactly one of expect/send/expect_screen (got more than one)", aw)
		case !hasExpect && !hasSend && !hasExpectScreen:
			add("%s: set exactly one of expect/send/expect_screen (an empty send: \"\" transmits EOF)", aw)
		case hasExpect:
			if _, err := regexp.Compile(a.Expect); err != nil {
				add("%s.expect %q is not a valid regexp: %v", aw, a.Expect, err)
			}
		case hasSend:
			// A named key must be in the vocabulary (#26); a typo'd key would
			// otherwise silently send nothing.
			if a.Send.Key != "" && !spec.ValidPTYKey(a.Send.Key) {
				add("%s.send.key %q is not a supported key (supported: %s)", aw, a.Send.Key, spec.PTYKeyNames())
			}
		case hasExpectScreen:
			validatePTYExpectScreen(add, aw+".expect_screen", a.ExpectScreen)
		}
	}
}

func validatePTYExpectScreen(add func(string, ...any), where string, es *spec.PTYExpectScreen) {
	validateStream(add, where, &es.StreamAssert)
	if es.Snapshot != "" {
		add("%s.snapshot is not supported in expect_screen; use a post-step assert screen snapshot or text matchers here", where)
	}
	if es.Trim != nil {
		add("%s.trim is not supported in expect_screen", where)
	}
	positiveDuration(add, where+".timeout", es.Timeout, "", "")
	positiveDuration(add, where+".stable_for", es.StableFor, "", "")
	if es.Timeout != "" && es.StableFor != "" {
		timeout, terr := time.ParseDuration(es.Timeout)
		stable, serr := time.ParseDuration(es.StableFor)
		if terr == nil && serr == nil && stable > timeout {
			add("%s.stable_for %q must not exceed %s.timeout %q", where, es.StableFor, where, es.Timeout)
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
