package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/runner"
	runnercmd "github.com/nao1215/atago/internal/runner/cmd"
	"github.com/nao1215/atago/internal/runner/ptyrun"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

// runPTY executes a pty step (#8): expand ${name} in the command and session,
// compose the process environment (scenario env layered under the step's own,
// both expanded), and drive the terminal session in the scenario workdir.
func (e *Engine) runPTY(ctx context.Context, p *spec.PTY, st *store.Store, scenarioEnv map[string]string, workdir string) (*runner.Result, *ptyrun.ExpectFailure, error) {
	// Guard every session entry against a live unresolved ${name}/${env:NAME}
	// before any terminal I/O (#78): otherwise st.Expand leaves the reference
	// verbatim and it would be TYPED into the program (or matched literally by an
	// expect), so a typo'd store name or a forgotten env: wiring feeds garbage to
	// the program under test instead of failing at the mistake. Checking all
	// entries up front guarantees a half-typed secret placeholder never reaches
	// the child. Escaped $${...} literals are exempt (Unresolved skips them).
	if err := checkPTYSessionResolved(p.Session, st); err != nil {
		return nil, nil, err
	}
	c := *p
	c.Command = st.Expand(p.Command)
	c.Cwd = st.Expand(p.Cwd)
	merged := make(map[string]string, len(scenarioEnv)+len(p.Env))
	for k, v := range st.ExpandMap(scenarioEnv) {
		merged[k] = v
	}
	for k, v := range st.ExpandMap(p.Env) { // step env overrides scenario env
		merged[k] = v
	}
	// A pty step drives a REAL terminal, and full-screen TUIs (less, vim, htop —
	// anything ncurses/termios) consult $TERM to decide whether to draw at all:
	// an unset or "dumb" TERM makes them refuse full-screen mode ("terminal is
	// not fully functional") or fall back to line mode, so a pty/screen assertion
	// can never see the real UI. atago renders the transcript through an
	// xterm-compatible vt10x emulator, so default TERM to xterm-256color unless
	// the spec set it — giving deterministic behavior regardless of the host's
	// own TERM (unset in CI, tmux/screen locally).
	if _, ok := merged["TERM"]; !ok {
		merged["TERM"] = "xterm-256color"
	}
	c.Env = merged
	if len(p.Session) > 0 {
		c.Session = make([]spec.PTYAction, len(p.Session))
		for i, a := range p.Session {
			na := spec.PTYAction{Expect: st.Expand(a.Expect)}
			if a.Send != nil {
				cs := *a.Send
				// Only verbatim text gets ${name} expansion; named keys are
				// fixed byte sequences (#26).
				if cs.Text != nil {
					txt := st.Expand(*cs.Text)
					cs.Text = &txt
				}
				na.Send = &cs
			}
			c.Session[i] = na
		}
	}
	// sandbox_home (#71) redirects the pty child's home under ${workdir}/.atago-home.
	var sandbox map[string]string
	if c.SandboxHomeEnabled() {
		s, err := runnercmd.EnsureSandboxHome(workdir)
		if err != nil {
			return nil, nil, err
		}
		sandbox = s
	}
	return ptyrun.Run(ctx, &c, workdir, runnercmd.BuildEnv(c.Env, c.ClearEnvEnabled(), c.PassEnv, sandbox))
}

// checkPTYSessionResolved reports the first session entry whose send text or
// expect pattern carries a live unresolved ${name}/${env:NAME} reference (#78).
// It mirrors the run.command guard's message — naming the entry, the reference,
// and the $${...} literal escape — so the fix-forward path is identical. Named
// keys and the empty-string EOF send have no text to resolve and are skipped.
func checkPTYSessionResolved(session []spec.PTYAction, st *store.Store) error {
	for i, a := range session {
		if err := unresolvedRefError(i, "expect", a.Expect, st); err != nil {
			return err
		}
		if a.Send != nil && a.Send.Text != nil {
			if err := unresolvedRefError(i, "send", *a.Send.Text, st); err != nil {
				return err
			}
		}
	}
	return nil
}

// unresolvedRefError returns an explained error for the first unresolved
// reference in text, or nil when every reference resolves.
func unresolvedRefError(idx int, field, text string, st *store.Store) error {
	names := st.Unresolved(text)
	if len(names) == 0 {
		return nil
	}
	name := names[0]
	if envName, isEnv := strings.CutPrefix(name, "env:"); isEnv {
		return fmt.Errorf(
			"pty session entry %[1]d (%[2]s) references ${env:%[3]s}, but the environment variable %[3]s is not set; set it or write $${env:%[3]s} for the literal text",
			idx, field, envName)
	}
	return fmt.Errorf(
		"pty session entry %[1]d (%[2]s) references ${%[3]s}, but no variable with that name is defined (builtins, matrix vars, store, ready.store, env:); define the variable or write $${%[3]s} for the literal text",
		idx, field, name)
}

// ptyExpectCheck converts a never-matched session expect into the structured
// check shape every other failure uses, so the console block and JSON report
// need no new machinery.
func ptyExpectCheck(ef *ptyrun.ExpectFailure) *assert.CheckResult {
	return &assert.CheckResult{
		Desc:           fmt.Sprintf("pty expect /%s/", ef.Pattern),
		Expected:       fmt.Sprintf("terminal transcript matches /%s/", ef.Pattern),
		Actual:         ef.Transcript,
		Hint:           "the expected pattern never appeared in the terminal transcript before the session timeout",
		ArtifactKind:   "pty",
		ArtifactActual: []byte(ef.Transcript),
	}
}
