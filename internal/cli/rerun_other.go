//go:build !windows

package cli

func isTransientRerunIOError(error) bool { return false }
