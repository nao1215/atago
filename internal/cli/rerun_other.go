//go:build !windows

package cli

// isTransientRerunIOError reports whether a rerun-state read or rename should
// be retried. Only Windows has transient sharing conflicts, so it is always
// false here.
func isTransientRerunIOError(error) bool { return false }
