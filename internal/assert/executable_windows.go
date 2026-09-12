//go:build windows

package assert

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// defaultPATHEXT is what Windows uses when PATHEXT is unset or empty. It is the
// stock value of the variable, kept here so the assertion has an answer on a
// host with a stripped environment instead of calling every file unrunnable.
const defaultPATHEXT = ".COM;.EXE;.BAT;.CMD;.VBS;.VBE;.JS;.JSE;.WSF;.WSH;.MSC"

// isExecutable reports whether the file can be run. Windows has no execute bit
// — Go synthesizes 0666, or 0444 for a read-only file, from the read-only
// attribute — so reading the mode the way the POSIX build does answered "not
// executable" for every file on the platform, including a freshly built .exe.
// That made `executable: true` impossible to satisfy and `executable: false`
// pass vacuously on a real program, which is worse: an assertion that cannot
// fail is not checking anything.
//
// What decides whether Windows will run a file by name is its extension, per
// PATHEXT. That is the same rule atago's own subject builder applies when it
// appends .exe to an artifact, so the assertion and the builder now agree.
func isExecutable(_ fs.FileInfo, path string) bool {
	ext := filepath.Ext(path)
	if ext == "" {
		return false
	}
	for _, candidate := range strings.Split(pathExt(), ";") {
		if strings.EqualFold(strings.TrimSpace(candidate), ext) {
			return true
		}
	}
	return false
}

// pathExt returns the host's executable-extension list, falling back to the
// stock Windows value.
func pathExt() string {
	if v := os.Getenv("PATHEXT"); strings.TrimSpace(v) != "" {
		return v
	}
	return defaultPATHEXT
}

// executableActual describes what was observed. It names the extension rather
// than the mode: the mode is synthesized on Windows and quoting it would send
// an author looking for a permission problem that does not exist.
func executableActual(info fs.FileInfo, path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return fmt.Sprintf("executable=%t (no file extension)", isExecutable(info, path))
	}
	return fmt.Sprintf("executable=%t (extension %q)", isExecutable(info, path), ext)
}

// executableHint states the rule, because it is not the rule a spec author
// coming from POSIX expects — and because the fix is to name the file
// differently rather than to change its permissions.
func executableHint(string) string {
	return "on Windows an executable is decided by its extension (PATHEXT), not by a mode bit"
}
