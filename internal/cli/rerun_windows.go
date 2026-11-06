//go:build windows

package cli

import (
	"errors"

	"golang.org/x/sys/windows"
)

// Windows may reject opens or replacements while another process has the
// destination open or is renaming it. Inspect wrapped OS errors, not messages.
func isTransientRerunIOError(err error) bool {
	return errors.Is(err, windows.ERROR_SHARING_VIOLATION) || errors.Is(err, windows.ERROR_ACCESS_DENIED)
}
