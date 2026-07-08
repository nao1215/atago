package loader

import (
	"fmt"
	"regexp"

	"github.com/nao1215/atago/internal/spec"
)

// validatePTY checks a pty step (#8): a command, sane duration/size values,
// and a session whose entries each set exactly one of expect/send with
// compilable expect regexps.
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
		switch {
		case hasExpect && hasSend:
			add("%s: set exactly one of expect/send (got both)", aw)
		case !hasExpect && !hasSend:
			add("%s: set exactly one of expect/send (an empty send: \"\" transmits EOF)", aw)
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
		}
	}
}
