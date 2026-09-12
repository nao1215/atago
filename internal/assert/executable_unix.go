//go:build !windows

package assert

import (
	"fmt"
	"io/fs"
)

// isExecutable reports whether the file can be run. On POSIX that is a property
// of the mode bits, and any of the three execute bits counts — a program the
// suite's own user cannot run may still be the artifact the spec is checking.
func isExecutable(info fs.FileInfo, _ string) bool {
	return info.Mode().Perm()&0o111 != 0
}

// executableActual describes what was observed, for the failure message. The
// mode is what an author reaches for `chmod` about, so it is quoted verbatim.
func executableActual(info fs.FileInfo, _ string) string {
	return fmt.Sprintf("executable=%t (mode %s)", isExecutable(info, ""), info.Mode().Perm())
}

// executableHint adds nothing on POSIX: the mode in executableActual already
// says everything there is to say.
func executableHint(string) string { return "" }
