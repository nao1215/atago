//go:build !windows

package fsdelta

import (
	"fmt"
	"io/fs"
)

// permString renders the POSIX permission bits a delta compares, masked to
// Perm() so that the kind bits (which the fingerprint already carries) cannot
// leak in and make an unchanged entry look modified.
func permString(mode fs.FileMode) string {
	return fmt.Sprintf("mode=%04o", mode.Perm())
}
