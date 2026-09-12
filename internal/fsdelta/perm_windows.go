//go:build windows

package fsdelta

import "io/fs"

// permString reports no permission component on Windows. There are no POSIX
// mode bits there: Go synthesizes 0666 or 0444 from the read-only attribute, so
// including it would make one spec report a different delta per OS while saying
// nothing a Windows user can act on.
func permString(fs.FileMode) string { return "" }
