//go:build windows

package record

import (
	"os"
	"strings"
	"testing"
)

// TestCapturePTY_WindowsUnsupported proves interactive pty recording reports a
// clear POSIX-only error on Windows rather than failing obscurely (#69).
func TestCapturePTY_WindowsUnsupported(t *testing.T) {
	_, err := CapturePTY("echo hi", false, os.Stdin, os.Stdout)
	if err == nil {
		t.Fatal("CapturePTY on Windows should return an error")
	}
	if !strings.Contains(err.Error(), "not supported on Windows") {
		t.Errorf("error = %v, want a Windows-unsupported message", err)
	}
}
