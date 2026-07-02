package assert

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

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

func checkDirExists(d *spec.DirAssert, info os.FileInfo, statErr error) *CheckResult {
	desc := fmt.Sprintf("assert dir %q exists: %t", d.Path, *d.Exists)
	if statErr != nil && !os.IsNotExist(statErr) {
		return &CheckResult{Desc: desc, Actual: statErr.Error(), Hint: fmt.Sprintf("could not stat %q: %v", d.Path, statErr)}
	}
	exists := statErr == nil && info.IsDir()
	if exists == *d.Exists {
		return nil
	}
	return &CheckResult{
		Desc:     desc,
		Expected: fmt.Sprintf("directory %q exists=%t", d.Path, *d.Exists),
		Actual:   fmt.Sprintf("exists=%t", exists),
		Hint:     fmt.Sprintf("expected directory %q to %s", d.Path, existence(*d.Exists)),
	}
}

func checkDirChildren(d *spec.DirAssert, dirPath string) *CheckResult {
	for _, child := range d.Contains {
		childPath, err := security.ResolveWorkdirPath("assert.dir.contains", dirPath, child)
		if err != nil {
			return &CheckResult{Desc: fmt.Sprintf("assert dir %q contains %q", d.Path, child), Hint: err.Error()}
		}
		if _, err := os.Stat(childPath); err != nil {
			return &CheckResult{
				Desc:     fmt.Sprintf("assert dir %q contains %q", d.Path, child),
				Expected: fmt.Sprintf("child %q present", child),
				Actual:   "missing",
				Hint:     fmt.Sprintf("expected %q to exist under %q", child, d.Path),
			}
		}
	}
	for _, child := range d.NotContains {
		childPath, err := security.ResolveWorkdirPath("assert.dir.not_contains", dirPath, child)
		if err != nil {
			return &CheckResult{Desc: fmt.Sprintf("assert dir %q does not contain %q", d.Path, child), Hint: err.Error()}
		}
		if _, err := os.Stat(childPath); err == nil {
			return &CheckResult{
				Desc:     fmt.Sprintf("assert dir %q does not contain %q", d.Path, child),
				Expected: fmt.Sprintf("child %q absent", child),
				Actual:   "present",
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
		return dirCountFailure(d, n, fmt.Sprintf("exactly %d entries", *d.Count))
	}
	if d.MinCount != nil && n < *d.MinCount {
		return dirCountFailure(d, n, fmt.Sprintf("at least %d entries", *d.MinCount))
	}
	if d.MaxCount != nil && n > *d.MaxCount {
		return dirCountFailure(d, n, fmt.Sprintf("at most %d entries", *d.MaxCount))
	}
	return nil
}

func dirCountFailure(d *spec.DirAssert, got int, want string) *CheckResult {
	return &CheckResult{
		Desc:     fmt.Sprintf("assert dir %q entry count", d.Path),
		Expected: want,
		Actual:   fmt.Sprintf("%d entries", got),
		Hint:     fmt.Sprintf("directory %q has %d entries, expected %s", d.Path, got, want),
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
		Actual:   fmt.Sprintf("no match among %d entries", len(names)),
		Hint:     fmt.Sprintf("no direct entry of %q matched glob %q", d.Path, d.Glob),
	}
}

func dirStatActual(info os.FileInfo, err error) string {
	switch {
	case err != nil:
		return err.Error()
	case !info.IsDir():
		return "not a directory"
	default:
		return "directory"
	}
}
