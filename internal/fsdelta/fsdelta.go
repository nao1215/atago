// Package fsdelta computes the difference of a directory tree between two
// points in time: which entries were created, modified, or deleted, by content,
// kind, and permission bits. It backs the `changes:` assertion target (#70),
// which pins exactly what a run/pty step touched in the scenario workdir.
package fsdelta

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/nao1215/atago/internal/fskind"
)

// Snapshot maps a tracked path (forward-slash, relative to the scanned root) to
// a fingerprint of everything about the entry a step can observably change: a
// hex-encoded SHA-256 for a regular file, a symlinkPrefix-tagged target for a
// symlink, a kindPrefix-tagged name for anything else (FIFO, socket, device),
// each carrying the POSIX permission bits where the platform has them.
// Directories are not tracked — an empty directory is not a "file" the
// assertion reasons about, and the files inside one are tracked individually —
// and a rename is delete+create.
type Snapshot map[string]string

// unreadableSentinel marks a regular file that exists but could not be read
// (e.g. mode 000). It is not a valid 64-char hex SHA-256, so it never collides
// with a real content hash. Recording it keeps the file visible to created/
// deleted so a step that plants an unreadable file cannot slip past
// `created: []`. Two snapshots that both find the file unreadable share the
// sentinel and report no modification; a file readable in one snapshot and
// unreadable in the other reports as modified, because content equality cannot
// be established across that readability boundary.
const unreadableSentinel = "unreadable"

// symlinkPrefix tags a symlink's fingerprint, which is the link target rather
// than the content behind it. Tracking symlinks keeps `created: []` honest: a
// step that plants a link — an installer wiring bin/tool, a tool that swaps a
// "current" pointer — used to slip past an exhaustive changes assertion because
// only regular files were scanned. Comparing targets (not the resolved content)
// makes a retargeted link a modification, keeps a dangling link visible, and
// avoids following a link out of the workdir or around a cycle. The prefix ends
// in ':' so it can never collide with a 64-char hex digest.
const symlinkPrefix = "symlink:"

// kindPrefix tags an entry that is neither a regular file nor a symlink — a
// FIFO, a socket, a device node — whose fingerprint is its kind rather than any
// content. Such an entry is never opened: reading a FIFO blocks until a writer
// appears, and a device node can block or hand back the host's data, so a scan
// that opened one would hang the run rather than fail it. Recording the kind
// keeps `created: []` honest about a step that plants one, which is the shape a
// server or an IPC-using tool leaves behind. The prefix ends in ':' so it can
// never collide with a 64-char hex digest.
const kindPrefix = "kind:"

// Scan walks root and fingerprints every entry except directories, keyed by its
// forward-slash path relative to root. It is best-effort about individual
// files: one that cannot be opened or read is recorded by sentinel rather than
// failing the whole scan, so a transient permission quirk never turns a delta
// assertion into an engine error. A nil root scan (root missing) returns an
// empty snapshot.
func Scan(root string) (Snapshot, error) {
	snap := Snapshot{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			// A directory we cannot descend is skipped, not fatal: the scan
			// still reports every file it could reach. Returning nil here is
			// deliberate — a per-entry error must not abort the whole scan.
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		rel, rerr := filepath.Rel(root, path)
		if rerr != nil {
			return nil //nolint:nilerr // an un-relativizable path is skipped, not fatal
		}
		if fp, ok := fingerprint(path, d); ok {
			snap[filepath.ToSlash(rel)] = fp
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return snap, err
	}
	return snap, nil
}

// fingerprint describes one entry by everything a step can observably change
// about it: what it is, what it holds, and its permission bits. It reports false
// for an entry that raced away between the walk and the read, which is genuinely
// absent rather than unreadable.
//
// The kind comes from the DirEntry the walk already carries, so a non-regular
// entry is classified without being opened.
func fingerprint(path string, d fs.DirEntry) (string, bool) {
	switch {
	case d.Type()&fs.ModeSymlink != 0:
		// The link's own permission bits are not meaningful (POSIX ignores them
		// and there is no portable lchmod), so a symlink is its target alone.
		target, err := os.Readlink(path)
		if err != nil {
			if os.IsNotExist(err) {
				return "", false
			}
			// The link exists but its target cannot be read: keep it visible to
			// created/deleted rather than dropping it.
			return symlinkPrefix + unreadableSentinel, true
		}
		return symlinkPrefix + filepath.ToSlash(target), true
	case !d.Type().IsRegular():
		return withMode(kindPrefix+fskind.Token(d.Type()), d), true
	}
	sum, err := hashFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false
		}
		// Exists but unreadable (e.g. mode 000): record a sentinel rather than
		// dropping it, so a created/deleted unreadable file is still reported.
		return withMode(unreadableSentinel, d), true
	}
	return withMode(sum, d), true
}

// withMode appends the entry's POSIX permission bits to its content
// fingerprint, so a step that only widens a config or drops the execute bit off
// a released binary is reported as modifying it — the same change atago already
// treats as observable in file: {executable: ...}.
//
// The bits are masked to Perm() and left out entirely on Windows (permString
// answers ""), because Go synthesizes a mode there from the read-only attribute
// and a mode-aware delta would otherwise report a different result per OS for
// one spec. A stat that fails leaves the content fingerprint alone rather than
// inventing a mode.
func withMode(content string, d fs.DirEntry) string {
	info, err := d.Info()
	if err != nil {
		// Raced or unstattable: the content fingerprint alone still keeps the
		// entry visible to created/deleted.
		return content
	}
	if perm := permString(info.Mode()); perm != "" {
		return content + " " + perm
	}
	return content
}

// hashFile returns the hex SHA-256 of a file's content.
func hashFile(path string) (string, error) {
	f, err := os.Open(path) //nolint:gosec // scanning the scenario workdir is the purpose
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Delta is the content-based difference between a pre and post Snapshot. Each
// list is sorted for deterministic reporting.
type Delta struct {
	Created  []string
	Modified []string
	Deleted  []string
}

// Diff compares pre against post: a path in post but not pre is Created, a path
// in both whose hash changed is Modified, and a path in pre but not post is
// Deleted. Paths whose hash is unchanged are untouched and reported nowhere.
func Diff(pre, post Snapshot) Delta {
	var d Delta
	for p, postHash := range post {
		if preHash, ok := pre[p]; !ok {
			d.Created = append(d.Created, p)
		} else if preHash != postHash {
			d.Modified = append(d.Modified, p)
		}
	}
	for p := range pre {
		if _, ok := post[p]; !ok {
			d.Deleted = append(d.Deleted, p)
		}
	}
	sort.Strings(d.Created)
	sort.Strings(d.Modified)
	sort.Strings(d.Deleted)
	return d
}
