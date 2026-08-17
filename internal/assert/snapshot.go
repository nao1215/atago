package assert

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"

	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/snapshot"
)

// SnapshotWrites records what an `--update-snapshots` run has written to each
// snapshot path, so a second scenario writing DIFFERENT content to a path
// another already claimed is refused instead of silently overwriting it.
//
// Without it the update run reports every scenario green and the very next
// verify run is red — the write→verify invariant a snapshot workflow rests on —
// and under --parallel the two writers race for the file. Two scenarios
// asserting the same output against one golden is a legitimate pattern (the
// `-h` and `--help` case), so identical content is accepted; only a conflicting
// rewrite is an error.
type SnapshotWrites struct {
	mu      sync.Mutex
	written map[string][sha256.Size]byte
}

// NewSnapshotWrites returns an empty recorder for one run.
func NewSnapshotWrites() *SnapshotWrites {
	return &SnapshotWrites{written: map[string][sha256.Size]byte{}}
}

// claim records that path is being written with data. It reports false when the
// path was already written in this run with different content. A nil recorder
// claims everything: direct API use and the retry `until` context have no run to
// scope the writes to.
func (w *SnapshotWrites) claim(path string, data []byte) bool {
	if w == nil {
		return true
	}
	sum := sha256.Sum256(data)
	w.mu.Lock()
	defer w.mu.Unlock()
	prev, seen := w.written[path]
	if seen && prev != sum {
		return false
	}
	w.written[path] = sum
	return true
}

// checkSnapshot compares (or, in update mode, writes) a snapshot. The snapshot
// path is resolved relative to the spec file's directory, since snapshots are
// committed next to the spec — not generated into the scenario workdir — and may
// not escape that directory.
func checkSnapshot(desc, label, snapPath string, data []byte, env Env) *CheckResult {
	path, err := security.ResolveSpecPath("assert.snapshot", env.SpecDir, snapPath)
	if err != nil {
		return &CheckResult{Desc: desc, Hint: err.Error()}
	}
	opt := snapshot.Options{Workdir: env.Workdir, Secrets: env.Secrets, Scrub: env.Scrub}

	if env.UpdateSnapshots {
		// Claim the path with the bytes that would land on disk, so a second
		// scenario writing different content to it is refused rather than
		// overwriting the first — which the next verify run would then fail on.
		if !env.SnapshotWrites.claim(path, snapshot.Normalize(data, opt)) {
			return &CheckResult{
				Desc: desc,
				Hint: fmt.Sprintf("snapshot %q was already written in this run with different content; two scenarios cannot share one snapshot path unless they produce identical output, or every update run leaves the next verify run failing", snapPath),
			}
		}
		if err := snapshot.Update(env.SpecDir, path, data, opt); err != nil {
			return &CheckResult{Desc: desc, Hint: fmt.Sprintf("could not write snapshot %q: %v", snapPath, err)}
		}
		updated := pass(desc + " (updated)")
		updated.SnapshotUpdated = true
		return updated
	}

	ok, expected, actual, err := snapshot.Compare(env.SpecDir, path, data, opt)
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
