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
