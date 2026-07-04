// Package fixture materializes input files declared by a spec into the scenario
// workdir.
package fixture

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
)

// Write materializes a fixture file inside workdir. The destination path is
// always resolved relative to workdir and may not escape it. A `from` source is
// read relative to specDir (the spec file's directory) when not absolute, so a
// spec can copy committed testdata (e.g. a binary .parquet) into the workdir.
// A `symlink` fixture creates a symbolic link instead of a regular file. Optional
// `mode` (octal) and `mtime` (RFC3339) are applied after writing.
func Write(f *spec.Fixture, workdir, specDir string) error {
	dest, err := safeJoin(workdir, f.File)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o750); err != nil {
		return fmt.Errorf("fixture %q: %w", f.File, err)
	}

	if f.Symlink != "" {
		// Validate the link target: it must resolve inside the workdir, so a
		// fixture cannot plant a link to /etc/... or ../../outside that a later
		// write would follow (issue #16). An in-workdir target — relative, or an
		// absolute ${workdir}/... — is allowed.
		if err := checkSymlinkTarget(workdir, dest, f.Symlink); err != nil {
			return fmt.Errorf("fixture %q: %w", f.File, err)
		}
		if err := os.Symlink(f.Symlink, dest); err != nil {
			return fmt.Errorf("fixture %q: symlink to %q: %w", f.File, f.Symlink, err)
		}
		return applyMtime(f, dest)
	}

	// A fixture that sets only mode/mtime (no content source) modifies an
	// existing file in place — e.g. chmod a file a previous step created — rather
	// than truncating it. This lets a spec make a tool-created file read-only.
	if f.Content == "" && f.Base64 == "" && f.From == "" && (f.Mode != "" || f.Mtime != "") {
		// dest is lexically inside the workdir (safeJoin), but the untrusted
		// program under test may have replaced it with a symlink pointing outside
		// — and chmod/chtimes FOLLOW symlinks, so applying mode/mtime here would
		// escape containment and re-permission a host file. Reject an existing
		// symlink, mirroring writeNoFollow's guard on the content path (issue #16).
		if fi, err := os.Lstat(dest); err == nil && fi.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("fixture %q: refusing to chmod/set-mtime through the existing symlink %q", f.File, dest)
		}
		if err := applyMode(f, dest); err != nil {
			return err
		}
		return applyMtime(f, dest)
	}

	var data []byte
	switch {
	case f.From != "":
		src := f.From
		if !filepath.IsAbs(src) {
			src = filepath.Join(specDir, src)
		}
		read, rerr := os.ReadFile(src) //nolint:gosec // src is the author-declared fixture source
		if rerr != nil {
			return fmt.Errorf("fixture %q: copy from %q: %w", f.File, f.From, rerr)
		}
		data = read
	case f.Base64 != "":
		decoded, derr := base64.StdEncoding.DecodeString(f.Base64)
		if derr != nil {
			return fmt.Errorf("fixture %q: invalid base64: %w", f.File, derr)
		}
		data = decoded
	default:
		data = []byte(f.Content)
	}

	// dest is contained within workdir by safeJoin above (cannot escape) — but the
	// untrusted program under test may have planted a symlink at dest pointing
	// outside the workdir; writing must not follow it (TOCTOU, issue #16).
	if err := writeNoFollow(dest, data, 0o600); err != nil {
		return fmt.Errorf("fixture %q: %w", f.File, err)
	}
	if err := applyMode(f, dest); err != nil {
		return err
	}
	return applyMtime(f, dest)
}

// applyMode chmods dest when the fixture declares an octal mode.
func applyMode(f *spec.Fixture, dest string) error {
	if f.Mode == "" {
		return nil
	}
	mode, perr := strconv.ParseUint(f.Mode, 8, 32)
	if perr != nil {
		return fmt.Errorf("fixture %q: invalid mode %q (want octal, e.g. 0444): %w", f.File, f.Mode, perr)
	}
	if err := os.Chmod(dest, os.FileMode(mode)); err != nil {
		return fmt.Errorf("fixture %q: chmod %s: %w", f.File, f.Mode, err)
	}
	return nil
}

// applyMtime sets the modification (and access) time of dest when the fixture
// declares an RFC3339 mtime. This lets specs pin timestamps so a tool's
// content-vs-mtime change detection can be exercised deterministically.
func applyMtime(f *spec.Fixture, dest string) error {
	if f.Mtime == "" {
		return nil
	}
	t, perr := time.Parse(time.RFC3339, f.Mtime)
	if perr != nil {
		return fmt.Errorf("fixture %q: invalid mtime %q (want RFC3339): %w", f.File, f.Mtime, perr)
	}
	if err := os.Chtimes(dest, t, t); err != nil {
		return fmt.Errorf("fixture %q: set mtime: %w", f.File, err)
	}
	return nil
}

// checkSymlinkTarget rejects a symlink whose target resolves outside workdir. A
// relative target resolves against the link's own directory (how the OS resolves
// it); an absolute target is checked as-is. This blocks `symlink: ../../etc/x`
// and `symlink: /etc/x` while allowing an in-workdir target (issue #16).
func checkSymlinkTarget(workdir, dest, target string) error {
	resolved := target
	// Treat a POSIX-absolute target (leading "/") as absolute on every host, not
	// only where filepath.IsAbs agrees, so `symlink: /etc/x` is rejected when the
	// spec is run on Windows too.
	if !filepath.IsAbs(resolved) && !strings.HasPrefix(resolved, "/") {
		resolved = filepath.Join(filepath.Dir(dest), target)
	}
	resolved = filepath.Clean(resolved)
	if !security.WithinRoot(workdir, resolved) {
		return fmt.Errorf("symlink target %q escapes the scenario workdir", target)
	}
	return nil
}

// writeNoFollow writes data to dest without following a symlink planted at dest.
// The untrusted program under test may have created a link there pointing outside
// the workdir; following it on write would escape containment (issue #16). An
// existing symlink is rejected outright; an existing regular file is removed and
// re-created with O_EXCL so a link planted in the race is never written through
// (O_EXCL fails atomically on any existing path). This is portable — unlike
// O_NOFOLLOW, which is not available on all platforms.
func writeNoFollow(dest string, data []byte, perm os.FileMode) error {
	if fi, err := os.Lstat(dest); err == nil {
		if fi.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing to write through the existing symlink %q", dest)
		}
		if err := os.Remove(dest); err != nil {
			return err
		}
	}
	f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm) //nolint:gosec // dest sanitized by safeJoin; O_EXCL guards against a planted link
	if err != nil {
		return err
	}
	_, werr := f.Write(data)
	cerr := f.Close()
	if werr != nil {
		return werr
	}
	return cerr
}

// safeJoin joins rel onto base and ensures the result stays within base, so a
// fixture cannot write outside the scenario workdir. It shares
// the workdir-containment policy used by file/store/service/browser/snapshot
// paths so every path-taking feature enforces the same rule.
func safeJoin(base, rel string) (string, error) {
	return security.ResolveWorkdirPath("fixture path", base, rel)
}
