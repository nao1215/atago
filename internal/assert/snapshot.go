package assert

import (
	"errors"
	"fmt"

	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/snapshot"
)

// checkSnapshot compares (or, in update mode, writes) a snapshot. The snapshot
// path is resolved relative to the spec file's directory, since snapshots are
// committed next to the spec — not generated into the scenario workdir — and may
// not escape that directory.
func checkSnapshot(desc, label, snapPath string, data []byte, env Env) *CheckResult {
	path, err := security.ResolveSpecPath("assert.snapshot", env.SpecDir, snapPath)
	if err != nil {
		return &CheckResult{Desc: desc, Hint: err.Error()}
	}
	opt := snapshot.Options{Workdir: env.Workdir, Secrets: env.Secrets}

	if env.UpdateSnapshots {
		if err := snapshot.Update(path, data, opt); err != nil {
			return &CheckResult{Desc: desc, Hint: fmt.Sprintf("could not write snapshot %q: %v", snapPath, err)}
		}
		return pass(desc + " (updated)")
	}

	ok, expected, actual, err := snapshot.Compare(path, data, opt)
	switch {
	case errors.Is(err, snapshot.ErrMissing):
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("snapshot file %q", snapPath),
			Actual:   "missing",
			Hint:     fmt.Sprintf("snapshot %q does not exist; create it with: atago run --update-snapshots", snapPath),
		}
	case err != nil:
		return &CheckResult{Desc: desc, Hint: fmt.Sprintf("could not read snapshot %q: %v", snapPath, err)}
	case ok:
		return pass(desc)
	default:
		// Persist the normalized actual (what atago compared) and the committed
		// snapshot, so --artifacts-dir lets a reviewer diff exactly what differed
		// against the golden file (#48).
		return &CheckResult{
			Desc:             desc,
			Expected:         excerpt(expected),
			Actual:           excerpt(actual),
			Hint:             fmt.Sprintf("%s did not match snapshot %q (update with --update-snapshots if intended)", label, snapPath),
			ArtifactKind:     "snapshot",
			ArtifactActual:   []byte(actual),
			ArtifactExpected: []byte(expected),
		}
	}
}
