// Package ptyrun runs one command inside a real pseudo-terminal and drives it
// with a declarative expect/send session (#8). It exists for CLIs that branch
// on TTY-ness (readline prompts, spinners, interactive shells) — the one
// surface a piped `run` step cannot reach. The captured transcript (terminal
// echo included) becomes the step's stdout, so the ordinary stream matchers,
// snapshots, and `store from.stdout` all work unchanged.
package ptyrun

import (
	"regexp"
	"time"

	"github.com/nao1215/atago/internal/spec"
)

// defaultSessionTimeout bounds a pty session when the spec sets none: an
// interactive program that never produces the expected prompt (or never
// exits) must fail loudly instead of hanging the run.
const defaultSessionTimeout = 30 * time.Second

// defaultRows / defaultCols are the terminal size when the spec sets none.
const (
	defaultRows = 24
	defaultCols = 80
)

// ExpectFailure describes the first session `expect` that never matched within
// the session budget. The engine reports it like a failed assertion (the
// scenario fails; it is not an execution error).
type ExpectFailure struct {
	Pattern    string // the regexp that never matched
	Transcript string // the transcript at the moment the budget elapsed
}

// compileSession validates and compiles the expect patterns up front so a bad
// regexp is one error for the whole step. The loader already validated shape;
// this is the runtime backstop.
func compileSession(actions []spec.PTYAction) ([]*regexp.Regexp, error) {
	res := make([]*regexp.Regexp, len(actions))
	for i, a := range actions {
		if a.Expect == "" {
			continue
		}
		re, err := regexp.Compile(a.Expect)
		if err != nil {
			return nil, err
		}
		res[i] = re
	}
	return res, nil
}

// sessionTimeout resolves the whole-session budget.
func sessionTimeout(p *spec.PTY) time.Duration {
	if p.Timeout == "" {
		return defaultSessionTimeout
	}
	d, err := time.ParseDuration(p.Timeout)
	if err != nil || d <= 0 {
		return defaultSessionTimeout
	}
	return d
}
