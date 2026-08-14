package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nao1215/atago/internal/fskind"
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
// while `/etc/passwd` is not. `../` traversal is rejected either way, and so is a
// path that leaves root through a symlinked directory. The returned path is
// cleaned and ready to hand to the filesystem.
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
	if target, escapes := escapingAncestor(root, dest); escapes {
		return "", fmt.Errorf("%s %q resolves through a symlink to %q, which escapes the %s", field, p, target, rootLabel)
	}
	return dest, nil
}

// escapingAncestor reports where a directory symlink above dest's leaf leads,
// when that is outside root.
//
// WithinRoot is lexical — it compares path components — so it cannot see that
// `<root>/escape/secret.txt` names a host file once the program under test has
// made `escape` a link to a host directory. Every path-taking feature inherited
// that blind spot from this one resolver: a file assertion read the host file
// into the report, and a fixture, a run.stdout_to, an http.body_to wrote into
// the host directory, all while the run stayed green. Resolving the ancestors
// closes it for every caller at once.
//
// Only the ancestors are resolved. A link AT the leaf belongs to the read and
// write helpers, which refuse it by name (ReadFileNoFollow, StatNoFollow,
// WriteConfinedFile); resolving it here would take that refusal away from them
// and would turn a merely absent path into an error, which `exists: false` asks
// about legitimately.
func escapingAncestor(root, dest string) (string, bool) {
	for dir := filepath.Dir(dest); WithinRoot(root, dir); dir = filepath.Dir(dir) {
		if resolved, err := filepath.EvalSymlinks(dir); err == nil {
			if WithinResolvedRoot(root, resolved) {
				return "", false
			}
			return resolved, true
		}
		// dir does not resolve: it is missing, or it is a link whose target is.
		// A dangling link still declares where it points, and the program under
		// test can create that target at any moment, so judging it by its declared
		// target keeps containment from depending on that timing. Anything else
		// unresolvable is an absent directory — ordinary, and the caller's own
		// error to report — so keep walking up to a component that does resolve.
		if target, escapes := LinkChainEscapes(root, dir); escapes {
			return target, true
		}
	}
	return "", false
}

// maxLinkHops bounds how far a chain of declared symlink targets is followed. A
// chain this long is a cycle in practice, and the kernel bounds its own
// resolution the same way (ELOOP after SYMLOOP_MAX hops).
const maxLinkHops = 40

// LinkChainEscapes reports where the symlink at p leads, when following its
// declared targets leaves root at any hop. A path that is not a symlink, or
// whose chain stays inside root the whole way, does not escape.
//
// The chain matters because only its far end says where a path really goes:
// `escape -> alias` with `alias -> /outside` looks in-root for one hop, and a
// check that stops there accepts a path the program under test completes into a
// host directory whenever it likes. Every hop is judged, not only the last one —
// a chain that steps outside and back names a location outside root that the
// program controls, and the kernel resolves through it.
//
// This works on declared targets rather than EvalSymlinks so it answers for a
// DANGLING chain too, which is the case that matters: whether a target exists
// yet is a race the program under test wins, so containment must not depend on
// it. A cycle runs out of hops and reports no escape, which is correct — a
// cyclic link resolves to nothing on any OS, so it can never reach outside.
func LinkChainEscapes(root, p string) (string, bool) {
	target, ok := linkTarget(p)
	if !ok {
		return "", false
	}
	for range maxLinkHops {
		if !WithinResolvedRoot(root, target) {
			return target, true
		}
		next, ok := linkTarget(target)
		if !ok {
			return "", false
		}
		target = next
	}
	return "", false
}

// linkTarget reports where a symlink points, as a cleaned path. A relative
// target resolves against the directory holding the link, the way the kernel
// resolves it. Anything that is not a symlink reports false.
func linkTarget(p string) (string, bool) {
	target, err := os.Readlink(p)
	if err != nil {
		return "", false
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(p), target)
	}
	return filepath.Clean(target), true
}

// confinedFileMode is the mode of every file atago writes on a spec's behalf
// inside a containment root — run.stdout_to/stderr_to, http.body_to, fixtures,
// snapshots, CDP screenshots. These land in a private scenario workdir (a
// per-run temp directory) or next to the spec, so nothing is gained by making
// them group- or world-readable; one mode across every site is one fewer thing
// to diverge. run.stdout_to used to write 0o644 for no documented reason (#349).
const confinedFileMode = 0o600

// openConfined opens an os.Root at root and returns it with dest expressed as a
// root-relative path. Binding an operation to the returned *os.Root is what makes
// containment hold at the moment of the I/O rather than at an earlier pathname
// check: os.Root resolves every component through the root's own descriptor and
// refuses one that leaves the root, so an ancestor the program under test swaps
// for a symlink between the lexical check (resolveInRoot) and the read or write
// cannot redirect it out of the root (issue #430). The lexical check stays as a
// fast, spelling-preserving rejection; this is the enforcement.
//
// dest was produced by resolveInRoot against this same root, so it shares root's
// spelling and filepath.Rel yields a clean relative path with no leading "..".
func openConfined(root, dest string) (*os.Root, string, error) {
	rel, err := filepath.Rel(root, dest)
	if err != nil {
		return nil, "", err
	}
	r, err := os.OpenRoot(root)
	if err != nil {
		return nil, "", err
	}
	return r, rel, nil
}

// WriteConfinedFile creates dest's parent directories and writes data there,
// through an os.Root bound to root so no directory component can carry the write
// out of root, and without following a symlink planted at the leaf. dest must
// already be containment-checked by ResolveWorkdirPath or ResolveSpecPath, and
// must lie under root — this helper enforces the rest of the policy (create
// parents, never follow the leaf, one file mode) that every confined write used
// to spell out for itself.
//
// Parent creation means a spec can name a nested destination (stdout_to:
// logs/out.txt) without a prior mkdir step; the parent is created through the
// root, so only root-local directories are ever made.
func WriteConfinedFile(root, dest string, data []byte) error {
	r, rel, err := openConfined(root, dest)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()
	return writeConfined(r, rel, dest, data, confinedFileMode)
}

// WriteWorkdirFile resolves rel against the scenario workdir and writes data
// there, confined to the workdir. field names the spec field so both the
// containment error and the write error say which key produced them. It returns
// the absolute destination, which callers put in the Result or log.
func WriteWorkdirFile(field, workdir, rel string, data []byte) (string, error) {
	dest, err := ResolveWorkdirPath(field, workdir, rel)
	if err != nil {
		return "", err
	}
	if err := WriteConfinedFile(workdir, dest, data); err != nil {
		return "", fmt.Errorf("%s: %w", field, err)
	}
	return dest, nil
}

// ReadFileNoFollow reads dest within root, binding the read to an os.Root so no
// directory component can redirect it out of root, and refusing a symlink at the
// leaf. resolveInRoot proves dest is inside root, but the untrusted program under
// test can swap an ancestor for a symlink after that check — os.Root refuses the
// escape at the moment of the read instead (issue #430). A symlink AT the leaf is
// still refused outright: os.Root would follow an in-root one, but atago's rule
// is that a link the program planted where an output was expected is never read
// through (issue #16). A non-regular leaf (a FIFO whose open would block the
// unbounded assert phase) is refused by kind.
func ReadFileNoFollow(root, dest string) ([]byte, error) {
	r, rel, err := openConfined(root, dest)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()
	if fi, lerr := r.Lstat(rel); lerr == nil {
		switch {
		case fi.Mode()&os.ModeSymlink != 0:
			return nil, fmt.Errorf("refusing to read through the symlink %q (it escapes the scenario root)", dest)
		case !fi.IsDir() && !fskind.Openable(fi.Mode()):
			// Opening a named pipe for reading blocks until another process
			// writes to it, and nothing bounds the assertion phase, so a program
			// under test that leaves a pipe where a file was expected used to
			// hang the whole run instead of failing it. A directory is left to
			// r.ReadFile, whose "is a directory" error already says it.
			return nil, fmt.Errorf("%q is a %s, not a regular file", dest, fskind.Name(fi.Mode()))
		}
	}
	return r.ReadFile(rel)
}

// StatNoFollow reports dest's metadata within root without following a symlink at
// the leaf. It is the stat half of ReadFileNoFollow's rule, bound to the same
// os.Root so an ancestor swapped for an escaping symlink is refused at the stat
// itself (issue #430) rather than disclosing the metadata of a host file (issue
// #16). A leaf symlink is rejected outright rather than resolved, matching the
// read and write paths.
func StatNoFollow(root, dest string) (os.FileInfo, error) {
	r, rel, err := openConfined(root, dest)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()
	fi, err := r.Lstat(rel)
	if err != nil {
		return nil, err
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("refusing to stat through the symlink %q (it escapes the scenario root)", dest)
	}
	return fi, nil
}

// writeLocks serializes confined writes that target the same path. Several
// parallel scenarios can share one golden file (e.g. matrix rows with an
// identical snapshot under --update-snapshots); without serialization their
// writes race in the filesystem — a non-atomic remove/create window on POSIX, or
// a MoveFileEx sharing violation between concurrent renames on Windows. Keying
// the lock on the destination path lets writes to different files still proceed
// concurrently. Entries are never deleted: the count is bounded by the distinct
// output paths a run touches, and each holds only a mutex pointer (#250).
var writeLocks sync.Map // dest path -> *sync.Mutex

// writeConfined writes data to rel through r (an os.Root at the containment root)
// without following a symlink planted at rel. Every filesystem operation goes
// through the root's descriptor, so a directory component the program under test
// turns into a symlink cannot carry the write out of the root (TOCTOU, issue
// #430), and an existing symlink at the leaf is refused outright (issue #16).
//
// The payload is written to a fresh temp file (created O_EXCL, so a planted link
// is never written through) in rel's own directory and then atomically renamed
// over rel. os.Root.Rename replaces the destination name without following a link
// that may sit there, and the rename is a single filesystem operation, so a
// reader never observes a torn file. Concurrent writers targeting the same path
// are additionally serialized on a per-path lock (see writeLocks) so identical
// writes all succeed and the last one wins deterministically on every OS, instead
// of colliding in the create window (POSIX) or a rename sharing violation
// (Windows) (#250). dest is the absolute path, used only for the lock key and
// error text.
func writeConfined(r *os.Root, rel, dest string, data []byte, perm os.FileMode) error {
	muAny, _ := writeLocks.LoadOrStore(dest, &sync.Mutex{})
	mu := muAny.(*sync.Mutex) //nolint:errcheck // only *sync.Mutex is ever stored
	mu.Lock()
	defer mu.Unlock()

	if dir := filepath.Dir(rel); dir != "." {
		if err := r.MkdirAll(dir, 0o750); err != nil {
			return err
		}
	}
	if fi, err := r.Lstat(rel); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to write through the existing symlink %q (it escapes the scenario root)", dest)
	}
	// Create the temp file in rel's directory so the rename stays on one
	// filesystem (a cross-device rename is not atomic and errors). The random
	// suffix keeps concurrent writers to different destinations from colliding on
	// the temp name, and O_EXCL makes a collision an error rather than a reuse.
	tmpRel, f, err := createTempInRoot(r, rel)
	if err != nil {
		return err
	}
	// Best-effort cleanup: harmless if the rename already consumed the temp.
	defer func() { _ = r.Remove(tmpRel) }()
	if err := f.Chmod(perm); err != nil {
		_ = f.Close()
		return err
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return r.Rename(tmpRel, rel)
}

// createTempInRoot creates a uniquely named, O_EXCL temp file next to rel within
// r, returning its root-relative name and the open file. It mirrors os.CreateTemp
// (random suffix, retry on collision) but through the root descriptor so the temp
// is created inside the containment root like the final file.
func createTempInRoot(r *os.Root, rel string) (string, *os.File, error) {
	dir, base := filepath.Dir(rel), filepath.Base(rel)
	for range 10000 {
		var buf [8]byte
		if _, err := rand.Read(buf[:]); err != nil {
			return "", nil, err
		}
		name := "." + base + ".tmp-" + hex.EncodeToString(buf[:])
		tmpRel := name
		if dir != "." {
			tmpRel = filepath.Join(dir, name)
		}
		f, err := r.OpenFile(tmpRel, os.O_RDWR|os.O_CREATE|os.O_EXCL, confinedFileMode)
		if err == nil {
			return tmpRel, f, nil
		}
		if !os.IsExist(err) {
			return "", nil, err
		}
	}
	return "", nil, fmt.Errorf("could not create a temp file for %q after many attempts", rel)
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

// WithinResolvedRoot reports whether p stays inside root, accepting either
// spelling of a root that itself sits behind a symlink.
//
// The root can be reached through a link — macOS puts /tmp behind /private/tmp
// and /var behind /private/var, which is where CI's scenario workdirs live — and
// p arrives in either spelling: a path EvalSymlinks resolved takes the resolved
// form, while a dangling link's declared target keeps the form the link was
// written with. Both spellings name the same directory, so containment holds if
// either does; a path genuinely outside the root is inside neither.
func WithinResolvedRoot(root, p string) bool {
	if WithinRoot(root, p) {
		return true
	}
	resolved, err := filepath.EvalSymlinks(root)
	return err == nil && WithinRoot(resolved, p)
}
