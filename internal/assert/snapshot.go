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
// It is also the run's record of WHICH goldens were rewritten, which is what
// the reports count: the walk over scenario results they used instead could not
// see a teardown, a suite lifecycle block, or the iterations of a repeat/retry
// that are not the surviving one, and counted one rewrite per matrix row of a
// shared golden. The recorder sees every write exactly where it happens.
type SnapshotWrites struct {
	mu      sync.Mutex
	claimed map[string]snapshotClaim
	written map[string]bool
}

// snapshotClaim is what one path was claimed with: the normalized bytes that
// were about to be written, and who claimed it. The writer's identity is what
// tells a genuine cross-scenario collision apart from a scenario colliding with
// its own earlier attempt under --retry-failed or --repeat, which is not a
// shared path at all but output that changes between attempts.
type snapshotClaim struct {
	sum    [sha256.Size]byte
	writer string
	// shared marks a path a second writer legitimately claimed with identical
	// content (the `-h` and `--help` pattern). A later conflict on it is a
	// disagreement between scenarios however it is spelled, so the attempt
	// wording — which claims one scenario is arguing with itself — does not
	// apply.
	shared bool
}

// NewSnapshotWrites returns an empty recorder for one run.
func NewSnapshotWrites() *SnapshotWrites {
	return &SnapshotWrites{claimed: map[string]snapshotClaim{}, written: map[string]bool{}}
}

// claim records that path is being written with data by writer. ok is false
// when the path was already claimed in this run with different content, in
// which case sameWriter reports whether the earlier claim came from this same
// scenario. A nil recorder claims everything: direct API use and the retry
// `until` context have no run to scope the writes to.
func (w *SnapshotWrites) claim(path string, data []byte, writer string) (ok, sameWriter bool) {
	if w == nil {
		return true, false
	}
	sum := sha256.Sum256(data)
	w.mu.Lock()
	defer w.mu.Unlock()
	prev, seen := w.claimed[path]
	if seen && prev.sum != sum {
		// An unnamed writer is not evidence of anything: outside a run there is
		// no identity to compare, so the clash keeps the general wording rather
		// than claiming two unnamed writers are the same scenario.
		return false, writer != "" && !prev.shared && prev.writer == writer
	}
	// The first claimant keeps the path. An identical re-claim by someone else
	// marks it shared rather than taking it over, so a later conflict is read
	// against everyone who agreed on the content, not just the last of them.
	next := snapshotClaim{sum: sum, writer: writer}
	if seen {
		next.writer = prev.writer
		next.shared = prev.shared || prev.writer != writer
	}
	w.claimed[path] = next
	return true, false
}

// markWritten records that path's golden actually reached disk. Counting these
// rather than the claims keeps a failed write out of the total.
func (w *SnapshotWrites) markWritten(path string) {
	if w == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.written[path] = true
}

// Count reports how many distinct golden files this run rewrote. It names
// files, not checks: several matrix rows asserting one golden rewrite one file.
func (w *SnapshotWrites) Count() int {
	if w == nil {
		return 0
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.written)
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

	// A frozen golden is compared, never rewritten, even under
	// --update-snapshots. It also stays unclaimed: the run wrote nothing to the
	// path, so another scenario's legitimate write must not be refused as a
	// conflict with a write that never happened.
	if env.UpdateSnapshots && !env.KeepSnapshots {
		// Claim the path with the bytes that would land on disk, so a second
		// scenario writing different content to it is refused rather than
		// overwriting the first — which the next verify run would then fail on.
		if ok, sameWriter := env.SnapshotWrites.claim(path, snapshot.Normalize(data, opt), env.Writer); !ok {
			// A scenario colliding with its OWN earlier claim is not a shared path:
			// a --repeat iteration or a --retry-failed attempt produced different
			// output from the one before it, and pointing at a second scenario sends
			// the reader looking for one that does not exist.
			hint := fmt.Sprintf("snapshot %q was already written in this run with different content; two scenarios cannot share one snapshot path unless they produce identical output, or every update run leaves the next verify run failing", snapPath)
			if sameWriter {
				hint = fmt.Sprintf("snapshot %q changed between attempts of this scenario: an earlier iteration or retry wrote different content, so the snapshotted output is not deterministic for this input and no golden can pin it", snapPath)
			}
			return &CheckResult{Desc: desc, Hint: hint}
		}
		if err := snapshot.Update(env.SpecDir, path, data, opt); err != nil {
			return &CheckResult{Desc: desc, Hint: fmt.Sprintf("could not write snapshot %q: %v", snapPath, err)}
		}
		env.SnapshotWrites.markWritten(path)
		updated := pass(desc + " (updated)")
		updated.SnapshotUpdated = true
		return updated
	}

	// Under --update-snapshots a frozen golden's hints must not send the reader
	// back to the flag they already passed: it deliberately skipped this check.
	frozen := env.UpdateSnapshots && env.KeepSnapshots
	rerecord := "update with --update-snapshots if intended"
	create := "create it with: atago run --update-snapshots"
	if frozen {
		rerecord = "this scenario is expect_fail, so --update-snapshots keeps its golden; drop expect_fail: to re-record it"
		create = "this scenario is expect_fail, so --update-snapshots does not create its golden; write the output the fix should produce, or drop expect_fail:"
	}

	ok, expected, actual, err := snapshot.Compare(env.SpecDir, path, data, opt)
	switch {
	case errors.Is(err, snapshot.ErrMissing):
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("snapshot file %q", snapPath),
			Actual:   "missing",
			Hint:     fmt.Sprintf("snapshot %q does not exist; %s", snapPath, create),
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
			Hint:             fmt.Sprintf("%s did not match snapshot %q (%s)", label, snapPath, rerecord),
			ArtifactKind:     "snapshot",
			ArtifactActual:   []byte(actual),
			ArtifactExpected: []byte(expected),
		}
	}
}
