package engine

import (
	"context"
	"fmt"

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
	return ptyrun.Run(ctx, &c, workdir, runnercmd.BuildEnv(c.Env, c.ClearEnvEnabled(), c.PassEnv))
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
