//go:build windows

package record

import (
	"errors"
	"os"
)

// CapturePTY reports that interactive pty recording is POSIX-only, mirroring the
// pty runner's Windows behavior (#69). ConPTY support can lift this later.
func CapturePTY(_ string, _ bool, _, _ *os.File) (PTYRecording, error) {
	return PTYRecording{}, errors.New("record --pty is not supported on Windows yet (POSIX-only; the generated pty steps are POSIX-only too)")
}
