package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// A user-declared path in a spec (a file assertion target, a store source, a
// service ready file, a CDP screenshot output, a snapshot file) must stay inside
// the root it is scoped to. Without that guarantee a spec could read or write
// arbitrary locations via `../` traversal or an absolute path, which contradicts
// atago's scenario-isolation model. These helpers centralize the policy so every
// path-taking feature enforces the same containment instead of ad-hoc joins with
// differing safety properties.

// ResolveWorkdirPath resolves a workdir-scoped path and guarantees it stays
// inside the scenario workdir. field names the spec field for a clear error.
func ResolveWorkdirPath(field, workdir, p string) (string, error) {
	return resolveInRoot(field, "scenario workdir", workdir, p)
}

// ResolveSpecPath resolves a spec-scoped path (currently snapshot paths) and
// guarantees it stays inside the spec directory. field names the spec field for
// a clear error.
func ResolveSpecPath(field, specDir, p string) (string, error) {
	return resolveInRoot(field, "spec directory", specDir, p)
}

// resolveInRoot resolves p against root and rejects any result that would escape
// root. A relative path is joined onto root; an absolute path is taken as-is but
// must still land inside root — so an absolute `${workdir}/out.txt` is allowed
// while `/etc/passwd` is not. `../` traversal is rejected either way. The
// returned path is cleaned and ready to hand to the filesystem.
func resolveInRoot(field, rootLabel, root, p string) (string, error) {
	dest := p
	if filepath.IsAbs(dest) {
		dest = filepath.Clean(dest)
	} else {
		dest = filepath.Join(root, dest)
	}
	if !WithinRoot(root, dest) {
		return "", fmt.Errorf("%s %q escapes the %s", field, p, rootLabel)
	}
	return dest, nil
}

// ReadFileNoFollow reads path but refuses to follow a symlink planted at the
// leaf. Lexical containment (ResolveWorkdirPath/ResolveSpecPath) proves path is
// inside its root, but the untrusted program under test may have replaced the
// leaf with a link pointing outside the root — os.ReadFile would follow it and
// disclose an arbitrary host file (issue #16). An existing symlink is rejected
// outright, mirroring WriteFileNoFollow's guard on the write path so every
// path-taking feature enforces the same rule instead of a plain, link-following
// os.ReadFile.
func ReadFileNoFollow(path string) ([]byte, error) {
	if fi, err := os.Lstat(path); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("refusing to read through the symlink %q (it escapes the scenario root)", path)
	}
	return os.ReadFile(path) //nolint:gosec // path is containment-checked by the caller and Lstat-guarded against a leaf symlink
}

// WriteFileNoFollow writes data to dest without following a symlink planted at
// dest. Lexical containment proves dest is inside its root, but the untrusted
// program under test may have created a link there pointing outside the root;
// following it on write would escape containment and clobber a host file (TOCTOU,
// issue #16). An existing symlink is rejected outright.
//
// The payload is written to a fresh temp file (created O_EXCL, so a planted link
// is never written through) in dest's own directory and then atomically renamed
// over dest. os.Rename replaces the destination name without following a link
// that may sit there, and the rename is a single filesystem operation — so
// concurrent writers targeting one path (e.g. several parallel scenarios sharing
// one golden under --update-snapshots) can never observe a torn file or race in
// a Lstat→Remove→create window; identical-content writers all succeed and the
// last rename wins (#250). This is portable — unlike O_NOFOLLOW, which is not
// available on all platforms.
func WriteFileNoFollow(dest string, data []byte, perm os.FileMode) error {
	if fi, err := os.Lstat(dest); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to write through the existing symlink %q (it escapes the scenario root)", dest)
	}
	// Create the temp file in dest's directory so the rename stays on one
	// filesystem (a cross-device rename is not atomic and errors). CreateTemp's
	// name is unique per call, so concurrent writers never collide on the temp.
	tmp, err := os.CreateTemp(filepath.Dir(dest), "."+filepath.Base(dest)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	// Best-effort cleanup: harmless if the rename already consumed the temp.
	defer func() { _ = os.Remove(tmpName) }()
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, dest)
}

// WithinRoot reports whether resolved lies inside root (root itself counts).
// Callers that resolve a path with non-default semantics — a symlink target
// resolved against the link's own directory, say — can reuse this single
// containment test instead of re-deriving prefix logic. It uses filepath.Rel so
// a relative root such as "." (a spec loaded by a bare filename) is handled the
// same as an absolute one, comparing whole path components rather than raw string
// prefixes.
func WithinRoot(root, resolved string) bool {
	rel, err := filepath.Rel(filepath.Clean(root), resolved)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}
