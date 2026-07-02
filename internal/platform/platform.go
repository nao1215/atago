// Package platform reports the host OS for skip/only gating (spec.md §19).
package platform

import "runtime"

// currentOS is a package var so tests can override it.
var currentOS = runtime.GOOS

// OS returns the normalized host operating system: linux, darwin, or windows.
func OS() string { return currentOS }

// Matches reports whether the given condition OS equals the host OS. An empty
// condition OS matches nothing meaningful and returns false.
func Matches(os string) bool {
	return os != "" && os == currentOS
}
