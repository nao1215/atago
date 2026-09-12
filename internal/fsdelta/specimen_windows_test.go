//go:build windows

package fsdelta

import (
	"os"
	"testing"
)

// Windows has neither of these kinds, and their specimens skip before reaching
// a planter. The stubs exist so the table compiles as one list on every OS
// rather than forking per platform.
func mkfifoAt(t *testing.T, _ string, _ os.FileMode) {
	t.Helper()
	t.Fatal("named pipes are not planted on Windows; the specimen should have skipped")
}

func socketAt(t *testing.T, _ string) {
	t.Helper()
	t.Fatal("unix sockets are not planted on Windows; the specimen should have skipped")
}
