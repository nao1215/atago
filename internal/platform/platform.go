// Package platform reports the host OS for skip/only gating.
package platform

import "runtime"

// currentOS is a package var so tests can override it.
var currentOS = runtime.GOOS

// OS returns the host operating system as runtime.GOOS names it: linux,
// darwin, windows, or one of the BSDs (freebsd, openbsd, netbsd), which a
// gate names one at a time rather than as a family.
func OS() string { return currentOS }

// Matches reports whether the given condition OS equals the host OS. An empty
// condition OS matches nothing meaningful and returns false.
func Matches(os string) bool {
	return os != "" && os == currentOS
}
