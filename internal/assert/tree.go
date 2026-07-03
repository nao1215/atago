package assert

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nao1215/atago/internal/spec"
)

// treeEntry is one filesystem object in a walked tree (#25).
type treeEntry struct {
	rel    string // /-separated path relative to the asserted directory
	kind   string // "dir", "file", or "link"
	hash   string // sha256 of the raw bytes, files only
	target string // symlink target, links only
}

// manifestLine renders the entry's snapshot-manifest line.
func (e treeEntry) manifestLine() string {
	switch e.kind {
	case "file":
		return "file " + e.rel + " sha256:" + e.hash
	case "link":
		return "link " + e.rel + " -> " + e.target
	default:
		return "dir " + e.rel
	}
}

// walkTree walks dirPath depth-first and returns every entry (root excluded)
// with /-separated relative paths, sorted, so manifests are deterministic
// across platforms. Symlinks are recorded, never traversed; an ignored
// directory prunes its whole subtree.
func walkTree(dirPath string, ignore []string) ([]treeEntry, error) {
	var out []treeEntry
	err := filepath.WalkDir(dirPath, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == dirPath {
			return nil
		}
		rel := filepath.ToSlash(strings.TrimPrefix(p, dirPath+string(filepath.Separator)))
		if ignoredPath(rel, ignore) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		switch {
		case d.Type()&fs.ModeSymlink != 0:
			target, lerr := os.Readlink(p)
			if lerr != nil {
				return lerr
			}
			out = append(out, treeEntry{rel: rel, kind: "link", target: filepath.ToSlash(target)})
		case d.IsDir():
			out = append(out, treeEntry{rel: rel, kind: "dir"})
		default:
			h, herr := hashFile(p)
			if herr != nil {
				return herr
			}
			out = append(out, treeEntry{rel: rel, kind: "file", hash: h})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].rel < out[j].rel })
	return out, nil
}

// hashFile streams the file's raw bytes into sha256 — byte-exact by design,
// so CRLF differences ARE differences (documented on DirAssert.Snapshot).
func hashFile(p string) (string, error) {
	f, err := os.Open(p) //nolint:gosec // confined under the asserted workdir directory
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }() // read-only handle
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ignoredPath reports whether rel matches any ignore glob (#25): a "<dir>/**"
// pattern prunes the subtree, a pattern without "/" also matches basenames at
// any depth, and everything else is path.Match on the whole relative path.
// Invalid patterns are rejected at load time.
func ignoredPath(rel string, ignore []string) bool {
	for _, pat := range ignore {
		if prefix, found := strings.CutSuffix(pat, "/**"); found {
			if rel == prefix || strings.HasPrefix(rel, prefix+"/") {
				return true
			}
			continue
		}
		if ok, _ := path.Match(pat, rel); ok {
			return true
		}
		if !strings.Contains(pat, "/") {
			if ok, _ := path.Match(pat, path.Base(rel)); ok {
				return true
			}
		}
	}
	return false
}

// checkDirRecursive evaluates the recursive matcher family over the walked
// tree (#25): contains/not_contains against relative paths, counts over files
// only, glob against each relative path (or basename for /-less patterns).
func checkDirRecursive(d *spec.DirAssert, dirPath string) *CheckResult {
	entries, err := walkTree(dirPath, d.Ignore)
	if err != nil {
		return &CheckResult{Desc: fmt.Sprintf("assert dir %q recursive", d.Path), Hint: fmt.Sprintf("could not walk %q: %v", d.Path, err)}
	}
	present := make(map[string]bool, len(entries))
	files := 0
	for _, e := range entries {
		present[e.rel] = true
		if e.kind == "file" {
			files++
		}
	}

	for _, child := range d.Contains {
		if !present[path.Clean(filepath.ToSlash(child))] {
			return &CheckResult{
				Desc:     fmt.Sprintf("assert dir %q contains %q", d.Path, child),
				Expected: fmt.Sprintf("path %q present in the tree", child),
				Actual:   "missing",
				Hint:     fmt.Sprintf("expected %q to exist under %q (recursive)", child, d.Path),
			}
		}
	}
	for _, child := range d.NotContains {
		if present[path.Clean(filepath.ToSlash(child))] {
			return &CheckResult{
				Desc:     fmt.Sprintf("assert dir %q does not contain %q", d.Path, child),
				Expected: fmt.Sprintf("path %q absent from the tree", child),
				Actual:   "present",
				Hint:     fmt.Sprintf("expected %q not to exist under %q (recursive)", child, d.Path),
			}
		}
	}

	if d.Count != nil && files != *d.Count {
		return dirCountFailure(d, files, fmt.Sprintf("exactly %d files in the tree", *d.Count))
	}
	if d.MinCount != nil && files < *d.MinCount {
		return dirCountFailure(d, files, fmt.Sprintf("at least %d files in the tree", *d.MinCount))
	}
	if d.MaxCount != nil && files > *d.MaxCount {
		return dirCountFailure(d, files, fmt.Sprintf("at most %d files in the tree", *d.MaxCount))
	}

	if d.Glob != "" {
		matched := false
		for _, e := range entries {
			if ok, _ := path.Match(d.Glob, e.rel); ok {
				matched = true
				break
			}
			if !strings.Contains(d.Glob, "/") {
				if ok, _ := path.Match(d.Glob, path.Base(e.rel)); ok {
					matched = true
					break
				}
			}
		}
		if !matched {
			return &CheckResult{
				Desc:     fmt.Sprintf("assert dir %q glob %q", d.Path, d.Glob),
				Expected: fmt.Sprintf("at least one tree entry matching %q", d.Glob),
				Actual:   fmt.Sprintf("no match among %d entries", len(entries)),
				Hint:     fmt.Sprintf("no entry under %q matched glob %q (recursive)", d.Path, d.Glob),
			}
		}
	}
	return nil
}

// checkDirSnapshot compares (or updates) the tree's golden manifest (#25).
// On mismatch the failure shows an added/removed/changed summary instead of
// two full manifests; the full texts still flow to --artifacts-dir.
func checkDirSnapshot(d *spec.DirAssert, dirPath string, env Env) *CheckResult {
	entries, err := walkTree(dirPath, d.Ignore)
	if err != nil {
		return &CheckResult{Desc: fmt.Sprintf("assert dir %q snapshot", d.Path), Hint: fmt.Sprintf("could not walk %q: %v", d.Path, err)}
	}
	lines := make([]string, len(entries))
	for i, e := range entries {
		lines[i] = e.manifestLine()
	}
	manifest := strings.Join(lines, "\n")
	if manifest != "" {
		manifest += "\n"
	}

	desc := fmt.Sprintf("assert dir %q snapshot %q", d.Path, d.Snapshot)
	cr := checkSnapshot(desc, "tree of "+d.Path, d.Snapshot, []byte(manifest), env)
	if cr.OK || cr.ArtifactKind == "" {
		return cr
	}
	// Replace the two full-dump excerpts with a manifest diff: the names of
	// the added/removed/changed paths are the review-relevant signal.
	cr.Expected = fmt.Sprintf("tree matches snapshot %q", d.Snapshot)
	cr.Actual = manifestDiff(string(cr.ArtifactExpected), string(cr.ArtifactActual))
	return cr
}

// manifestDiff summarizes two tree manifests as added/removed/changed path
// lists. A path present in both with a different line (hash or kind) counts
// as changed.
func manifestDiff(expected, actual string) string {
	parse := func(text string) map[string]string {
		m := map[string]string{}
		for _, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
			if line == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			m[fields[1]] = line
		}
		return m
	}
	want, got := parse(expected), parse(actual)

	var added, removed, changed []string
	for p, line := range got {
		wline, ok := want[p]
		switch {
		case !ok:
			added = append(added, line)
		case wline != line:
			changed = append(changed, p)
		}
	}
	for p, line := range want {
		if _, ok := got[p]; !ok {
			removed = append(removed, line)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	sort.Strings(changed)

	var b strings.Builder
	writeGroup := func(label string, items []string) {
		for _, it := range items {
			fmt.Fprintf(&b, "%s %s\n", label, it)
		}
	}
	writeGroup("added:  ", added)
	writeGroup("removed:", removed)
	writeGroup("changed:", changed)
	if b.Len() == 0 {
		return "(manifests differ only in ordering or trailing whitespace)"
	}
	return strings.TrimRight(b.String(), "\n")
}
