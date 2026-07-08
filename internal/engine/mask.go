package engine

import (
	"errors"
	"strings"

	"github.com/nao1215/atago/internal/artifact"
	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/runner"
	mockrunner "github.com/nao1215/atago/internal/runner/mock"
	servicerunner "github.com/nao1215/atago/internal/runner/service"
	"github.com/nao1215/atago/internal/scrub"
	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
)

// newScrubber compiles the spec's declarative snapshot scrub rules (#137). The
// loader already validated every pattern, so New cannot fail here; on the
// impossible error path it returns a nil scrubber (a no-op) rather than aborting
// the run — a broken scrub rule must never silently swallow a real assertion.
func newScrubber(s *spec.Spec) *scrub.Scrubber {
	sc, err := scrub.New(s.Scrub)
	if err != nil {
		return nil
	}
	return sc
}

// isPolicyViolation reports whether err is a network-allowlist denial, so the
// engine can flag a security-policy violation for grpc/ssh steps (issue #17),
// mirroring how the HTTP runner surfaces its own PolicyError.
func isPolicyViolation(err error) bool {
	var pe *security.PolicyError
	return errors.As(err, &pe)
}

// maskResult returns a copy of r with secret values masked in the command and
// captured output. The original (unmasked) result is kept for assertions; only
// this copy flows into reports.
func maskResult(m *security.Masker, r *runner.Result) *runner.Result {
	if m.Empty() {
		return r
	}
	c := *r
	c.Command = m.Mask(r.Command)
	c.Stdout = m.MaskBytes(r.Stdout)
	c.Stderr = m.MaskBytes(r.Stderr)
	c.Body = m.MaskBytes(r.Body)
	c.RowsJSON = m.MaskBytes(r.RowsJSON)
	c.MessageJSON = m.MaskBytes(r.MessageJSON)
	c.CDPValue = m.MaskBytes(r.CDPValue)
	return &c
}

// maskCheck masks secret values in the human-facing fields of a check result.
func maskCheck(m *security.Masker, cr *assert.CheckResult) {
	if m.Empty() || cr == nil {
		return
	}
	cr.Desc = m.Mask(cr.Desc)
	cr.Expected = m.Mask(cr.Expected)
	cr.Actual = m.Mask(cr.Actual)
	cr.Hint = m.Mask(cr.Hint)
	// Mask the full artifact payloads too, so secrets never reach a durable
	// sidecar file (#48). Guard on length so a nil "no expected payload" stays
	// nil rather than becoming an empty masked slice.
	if len(cr.ArtifactActual) > 0 {
		cr.ArtifactActual = m.MaskBytes(cr.ArtifactActual)
	}
	if len(cr.ArtifactExpected) > 0 {
		cr.ArtifactExpected = m.MaskBytes(cr.ArtifactExpected)
	}
}

// recordChecks masks each check and writes its failure artifacts, in place. It
// is the shared post-processing for both assert steps and retry `until` checks,
// which can each carry several targets (one CheckResult apiece).
func (e *Engine) recordChecks(m *security.Masker, checks []*assert.CheckResult, specPath, scenario string, scenarioIdx, stepIdx int) {
	for _, cr := range checks {
		maskCheck(m, cr)
		e.writeArtifacts(cr, specPath, scenario, scenarioIdx, stepIdx)
	}
}

// writeArtifacts persists a failed assertion's full compared payloads as durable
// sidecar files under the configured artifacts dir (#48), recording the written
// relative paths on the CheckResult so the report can reference them. It is a
// no-op when no artifacts dir is configured, the assertion passed, or the check
// carries no exportable payload. Masking has already happened via maskCheck, so
// the bytes written here never contain secrets.
func (e *Engine) writeArtifacts(cr *assert.CheckResult, specPath, scenario string, scenarioIdx, stepIdx int) {
	if e.Artifacts == nil || cr == nil || cr.OK || cr.ArtifactKind == "" {
		return
	}
	write := func(role string, content []byte) {
		if len(content) == 0 {
			return
		}
		rel := artifact.FailurePath(specPath, scenario, scenarioIdx, stepIdx, cr.ArtifactKind, role, "txt")
		p, err := e.Artifacts.Write(rel, content)
		if err != nil {
			return
		}
		cr.ArtifactFiles = append(cr.ArtifactFiles, assert.ArtifactFile{Role: role, Path: p})
	}
	write("actual", cr.ArtifactActual)
	write("expected", cr.ArtifactExpected)
	// Image similar_to failures carry richer binary blobs (actual/baseline/diff
	// images + metadata) with their own extensions (#52).
	for _, blob := range cr.ArtifactBlobs {
		if len(blob.Data) == 0 {
			continue
		}
		rel := artifact.FailurePath(specPath, scenario, scenarioIdx, stepIdx, cr.ArtifactKind, blob.Role, blob.Ext)
		p, err := e.Artifacts.Write(rel, blob.Data)
		if err != nil {
			continue
		}
		cr.ArtifactFiles = append(cr.ArtifactFiles, assert.ArtifactFile{Role: blob.Role, Path: p})
	}
}

// writeServiceLogs preserves each service's combined stdout/stderr as a durable
// log artifact under the configured artifacts dir (#51), recording the written
// paths on the ScenarioResult. It is a no-op when no artifacts dir is
// configured; a service that produced no output is skipped so a readiness
// failure with a silent service never creates an empty, noisy artifact. Secrets
// are masked before the log is written, keeping saved logs consistent with the
// rest of the report. Already-recorded services are not written twice.
func (e *Engine) writeServiceLogs(out *ScenarioResult, m *security.Masker, procs []*servicerunner.Proc, specPath, scenario string, scenarioIdx int) {
	if e.Artifacts == nil {
		return
	}
	for _, proc := range procs {
		if proc == nil || serviceLogged(out, proc.Name()) {
			continue
		}
		raw := proc.Output()
		if strings.TrimSpace(raw) == "" {
			continue // no output → no artifact (avoids empty/noisy files)
		}
		rel := artifact.ServiceLogPath(specPath, scenario, scenarioIdx, proc.Name())
		p, err := e.Artifacts.Write(rel, m.MaskBytes([]byte(raw)))
		if err != nil {
			continue
		}
		out.ServiceLogs = append(out.ServiceLogs, ServiceLog{Name: proc.Name(), Path: p})
	}
}

// writeMockLogs preserves each mock server's recorded requests as a durable
// artifact when a scenario fails, honoring RequestLog's contract ("the durable
// artifact written next to service logs when a scenario fails"). The request
// log is often the sharpest failure evidence a mock scenario has — a typo'd
// client path shows up as a recorded 404. Failure-gated and artifacts-dir-gated
// like service logs; a mock that recorded no request writes nothing.
func (e *Engine) writeMockLogs(out *ScenarioResult, m *security.Masker, mocks []*mockrunner.Server, specPath, scenario string, scenarioIdx int) {
	if e.Artifacts == nil {
		return
	}
	for _, srv := range mocks {
		if srv == nil {
			continue
		}
		name := srv.Name() + " (mock requests)"
		if serviceLogged(out, name) {
			continue
		}
		log := srv.RequestLog()
		if log == "" {
			continue // no recorded request -> no artifact
		}
		rel := artifact.MockLogPath(specPath, scenario, scenarioIdx, srv.Name())
		p, err := e.Artifacts.Write(rel, m.MaskBytes([]byte(log)))
		if err != nil {
			continue
		}
		out.ServiceLogs = append(out.ServiceLogs, ServiceLog{Name: name, Path: p})
	}
}

// serviceLogged reports whether a service's log artifact was already recorded,
// so the readiness-failure and post-loop capture paths do not duplicate it.
func serviceLogged(out *ScenarioResult, name string) bool {
	for _, sl := range out.ServiceLogs {
		if sl.Name == name {
			return true
		}
	}
	return false
}
