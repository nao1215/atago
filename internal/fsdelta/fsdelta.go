// Package fsdelta computes the content-based difference of a directory tree
// between two points in time: which regular files were created, modified, or
// deleted. It backs the `changes:` assertion target (#70), which pins exactly
// what a run/pty step touched in the scenario workdir.
package fsdelta

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// Snapshot maps a regular file's forward-slash path (relative to the scanned
// root) to a hex-encoded SHA-256 of its content. Directories and symlinks are
// not tracked: a rename is delete+create in v1, and an empty directory is not a
// "file" the assertion reasons about.
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

// Scan walks root and hashes every regular file, keyed by its forward-slash
// path relative to root. It is best-effort about individual files: one that
// cannot be opened or read is skipped rather than failing the whole scan, so a
// transient permission quirk never turns a delta assertion into an engine
// error. A nil root scan (root missing) returns an empty snapshot.
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
		if d.IsDir() || !d.Type().IsRegular() {
			return nil
		}
		rel, rerr := filepath.Rel(root, path)
		if rerr != nil {
			return nil //nolint:nilerr // an un-relativizable path is skipped, not fatal
		}
		sum, herr := hashFile(path)
		if herr != nil {
			if os.IsNotExist(herr) {
				return nil //nolint:nilerr // raced away between walk and open; genuinely absent
			}
			// Exists but unreadable (e.g. mode 000): record a sentinel rather than
			// dropping it, so a created/deleted unreadable file is still reported.
			snap[filepath.ToSlash(rel)] = unreadableSentinel
			return nil //nolint:nilerr
		}
		snap[filepath.ToSlash(rel)] = sum
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return snap, err
	}
	return snap, nil
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
