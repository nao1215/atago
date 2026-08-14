package assert

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nao1215/atago/internal/fskind"
	"github.com/nao1215/atago/internal/plural"
	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
)

// checkDir evaluates a directory/tree assertion (#74). Every set field is a
// separate constraint and all must hold; the first failing one is reported. The
// directory path and every child path are confined to the scenario workdir, so a
// generator spec cannot assert over arbitrary filesystem locations.
func checkDir(d *spec.DirAssert, env Env) *CheckResult {
	dirPath, err := security.ResolveWorkdirPath("assert.dir.path", env.Workdir, d.Path)
	if err != nil {
		return &CheckResult{Desc: fmt.Sprintf("assert dir %q", d.Path), Hint: err.Error()}
	}
	dirPath, err = resolveDirRoot(env.Workdir, d.Path, dirPath)
	if err != nil {
		return &CheckResult{Desc: fmt.Sprintf("assert dir %q", d.Path), Hint: err.Error()}
	}

	// Existence is checked first: every other constraint needs the directory to be
	// present and readable. When exists:false is asserted, a missing directory is
	// the success and no other constraint is meaningful.
	info, statErr := os.Stat(dirPath)
	if d.Exists != nil {
		if cr := checkDirExists(d, info, statErr); cr != nil {
			return cr
		}
		if !*d.Exists {
			return pass(fmt.Sprintf("assert dir %q exists: false", d.Path))
		}
	}

	// Any remaining constraint requires a readable directory.
	if statErr != nil || !info.IsDir() {
		return &CheckResult{
			Desc:     fmt.Sprintf("assert dir %q", d.Path),
			Expected: fmt.Sprintf("directory %q", d.Path),
			Actual:   dirStatActual(info, statErr),
			Hint:     fmt.Sprintf("%q is not a readable directory", d.Path),
		}
	}

	// Tree snapshot (#25): the golden manifest covers the whole walk; the
	// loader guarantees it is not combined with the matcher family.
	if d.Snapshot != "" {
		return checkDirSnapshot(d, dirPath, env)
	}
	// Recursive mode (#25): the matcher family applies to the whole walk.
	if d.Recursive {
		if cr := checkDirRecursive(d, dirPath); cr != nil {
			return cr
		}
		return pass(fmt.Sprintf("assert dir %q (recursive)", d.Path))
	}

	if cr := checkDirChildren(d, dirPath); cr != nil {
		return cr
	}
	if cr := checkDirCounts(d, dirPath); cr != nil {
		return cr
	}
	if cr := checkDirGlob(d, dirPath); cr != nil {
		return cr
	}
	return pass(fmt.Sprintf("assert dir %q", d.Path))
}

// resolveDirRoot follows a symlinked directory path and puts the result through
// the same containment test the declared path faced.
//
// Two defects meet here. filepath.WalkDir Lstats its root, so a `path:` that is
// itself a symlink to a directory walked as a single non-directory entry:
// `recursive:` reported an empty tree for a populated directory, and `snapshot:`
// wrote an empty golden that then matched forever no matter what the directory
// held — while the non-recursive matchers, whose os.ReadDir does follow the
// link, read the real contents. One assertion, two answers. And the containment
// check is lexical, so it cannot see through a link the program under test
// planted: a `path:` naming a link out of the workdir passed it, and the
// directory outside was listed despite the confinement this file promises.
//
// Resolving the root settles both: every mode reads the same tree, and the tree
// it reads is inside the workdir. Only the ROOT is resolved — walkTree still
// records the links inside the tree without traversing them, which is what keeps
// a link cycle out of the walk.
func resolveDirRoot(workdir, declared, dirPath string) (string, error) {
	resolved, rerr := filepath.EvalSymlinks(dirPath)
	if rerr != nil {
		// Nothing resolvable: the path is missing, or it is a link whose target
		// is. A missing path is ordinary — the stat in checkDir reports it in the
		// assertion's own words, and exists:false expects exactly this. A DANGLING
		// link is different: it still declares where it points, and whether that
		// target exists is not a question a spec may put to the filesystem outside
		// the workdir. Judging it by its declared target keeps the containment
		// rule from depending on whether the target happens to exist.
		if target, ok := security.LinkTarget(dirPath); ok && !security.WithinResolvedRoot(workdir, target) {
			return "", escapesWorkdirError(declared, target)
		}
		return dirPath, nil
	}
	if resolved == dirPath {
		return dirPath, nil
	}
	if !security.WithinResolvedRoot(workdir, resolved) {
		return "", escapesWorkdirError(declared, resolved)
	}
	return resolved, nil
}

func escapesWorkdirError(declared, target string) error {
	return fmt.Errorf("assert.dir.path %q resolves through a symlink to %q, which escapes the scenario workdir", declared, target)
}

func checkDirExists(d *spec.DirAssert, info os.FileInfo, statErr error) *CheckResult {
	desc := fmt.Sprintf("assert dir %q exists: %t", d.Path, *d.Exists)
	if statErr != nil && !os.IsNotExist(statErr) {
		return &CheckResult{Desc: desc, Actual: statErr.Error(), Hint: fmt.Sprintf("could not stat %q: %v", d.Path, statErr)}
	}
	exists := statErr == nil && info.IsDir()
	if exists == *d.Exists {
		return nil
	}
	// A path that exists but is not a directory is the confusing case: a bare
	// "exists=false" reads as "nothing is there" and sends the author looking for
	// a missing output that is in fact sitting right where they asked, as a file.
	actual := fmt.Sprintf("exists=%t", exists)
	hint := fmt.Sprintf("expected directory %q to %s", d.Path, existence(*d.Exists))
	if statErr == nil && !info.IsDir() {
		actual = fmt.Sprintf("exists=%t (%q is a %s, not a directory)", exists, d.Path, fileKind(info))
		hint = fmt.Sprintf("%q exists but is a %s; use a file: assertion for it", d.Path, fileKind(info))
	}
	return &CheckResult{
		Desc:     desc,
		Expected: fmt.Sprintf("directory %q exists=%t", d.Path, *d.Exists),
		Actual:   actual,
		Hint:     hint,
	}
}

func checkDirChildren(d *spec.DirAssert, dirPath string) *CheckResult {
	for _, child := range d.Contains {
		childPath, err := security.ResolveWorkdirPath("assert.dir.contains", dirPath, child)
		if err != nil {
			return &CheckResult{Desc: fmt.Sprintf("assert dir %q contains %q", d.Path, child), Hint: err.Error()}
		}
		// Lstat, not Stat: membership is about the directory entry, not whether a
		// symlink target resolves. A dangling symlink is still a present entry.
		if _, err := os.Lstat(childPath); err != nil {
			return &CheckResult{
				Desc:     fmt.Sprintf("assert dir %q contains %q", d.Path, child),
				Expected: fmt.Sprintf("child %q present", child),
				Actual:   dirListing(dirPath),
				Hint:     fmt.Sprintf("expected %q to exist under %q", child, d.Path),
			}
		}
	}
	for _, child := range d.NotContains {
		childPath, err := security.ResolveWorkdirPath("assert.dir.not_contains", dirPath, child)
		if err != nil {
			return &CheckResult{Desc: fmt.Sprintf("assert dir %q does not contain %q", d.Path, child), Hint: err.Error()}
		}
		// Lstat, not Stat: a dangling symlink the step left behind is a present
		// entry, so not_contains must fail on it rather than follow the dead link.
		if _, err := os.Lstat(childPath); err == nil {
			return &CheckResult{
				Desc:     fmt.Sprintf("assert dir %q does not contain %q", d.Path, child),
				Expected: fmt.Sprintf("child %q absent", child),
				Actual:   dirListing(dirPath),
				Hint:     fmt.Sprintf("expected %q not to exist under %q", child, d.Path),
			}
		}
	}
	return nil
}

func checkDirCounts(d *spec.DirAssert, dirPath string) *CheckResult {
	if d.Count == nil && d.MinCount == nil && d.MaxCount == nil {
		return nil
	}
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return &CheckResult{Desc: fmt.Sprintf("assert dir %q count", d.Path), Hint: fmt.Sprintf("could not read %q: %v", d.Path, err)}
	}
	n := len(entries)
	if d.Count != nil && n != *d.Count {
		return dirCountFailure(d, n, "exactly "+plural.Count(*d.Count, "entry", "entries"), dirListing(dirPath))
	}
	if d.MinCount != nil && n < *d.MinCount {
		return dirCountFailure(d, n, "at least "+plural.Count(*d.MinCount, "entry", "entries"), dirListing(dirPath))
	}
	if d.MaxCount != nil && n > *d.MaxCount {
		return dirCountFailure(d, n, "at most "+plural.Count(*d.MaxCount, "entry", "entries"), dirListing(dirPath))
	}
	return nil
}

func dirCountFailure(d *spec.DirAssert, got int, want, listing string) *CheckResult {
	return &CheckResult{
		Desc:     fmt.Sprintf("assert dir %q entry count", d.Path),
		Expected: want,
		Actual:   listing,
		Hint:     fmt.Sprintf("directory %q has %s, expected %s", d.Path, plural.Count(got, "entry", "entries"), want),
	}
}

func checkDirGlob(d *spec.DirAssert, dirPath string) *CheckResult {
	if d.Glob == "" {
		return nil
	}
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return &CheckResult{Desc: fmt.Sprintf("assert dir %q glob %q", d.Path, d.Glob), Hint: fmt.Sprintf("could not read %q: %v", d.Path, err)}
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	for _, name := range names {
		ok, matchErr := filepath.Match(d.Glob, name)
		if matchErr != nil {
			return &CheckResult{Desc: fmt.Sprintf("assert dir %q glob %q", d.Path, d.Glob), Hint: fmt.Sprintf("invalid glob %q: %v", d.Glob, matchErr)}
		}
		if ok {
			return nil
		}
	}
	return &CheckResult{
		Desc:     fmt.Sprintf("assert dir %q glob %q", d.Path, d.Glob),
		Expected: fmt.Sprintf("at least one entry matching %q", d.Glob),
		Actual:   dirListing(dirPath),
		Hint:     fmt.Sprintf("no direct entry of %q matched glob %q", d.Path, d.Glob),
	}
}

// dirListingLimit caps how many entries a failure block lists. A generator that
// produced thousands of files would otherwise bury the rest of the report.
const dirListingLimit = 40

// dirListing renders a directory's direct entries for a failure block: one
// sorted name per line, directories marked with a trailing slash. A directory
// assertion that failed is asking "what did the step actually produce?", and
// answering with a count alone ("3 entries") makes the reader re-run the
// generator by hand. This mirrors what a file assertion does by showing the
// file's content.
//
// A read error becomes the listing text rather than an error return: the caller
// is already reporting a failure and the read problem is the useful detail.
func dirListing(dirPath string) string {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Sprintf("(could not read directory: %v)", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return renderListing(names, "entries")
}

// renderListing joins names one per line, capping the list at dirListingLimit
// and reporting how many were elided. noun labels what is being listed so the
// direct and recursive callers can say "entries" or "paths".
func renderListing(names []string, noun string) string {
	if len(names) == 0 {
		return fmt.Sprintf("(no %s)", noun)
	}
	shown := names
	suffix := ""
	if len(names) > dirListingLimit {
		shown = names[:dirListingLimit]
		suffix = fmt.Sprintf("\n... (%d more)", len(names)-dirListingLimit)
	}
	return strings.Join(shown, "\n") + suffix
}

func dirStatActual(info os.FileInfo, err error) string {
	switch {
	case err != nil:
		return err.Error()
	case !info.IsDir():
		return "not a directory (" + fileKind(info) + ")"
	default:
		return "directory"
	}
}

// fileKind names what a non-directory path actually is, so a failure can say
// "regular file" instead of leaving the author to guess why a directory
// assertion reports the path as absent. The vocabulary is shared with the tree
// manifest and the read refusals so one word means one thing everywhere.
func fileKind(info os.FileInfo) string {
	return fskind.Name(info.Mode())
}
