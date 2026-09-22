//go:build windows && !amd64 && !arm64

package conpty

import "runtime"

// No console host is bundled for this architecture; StartOpenConsole reports
// ErrNoOpenConsole.
const runtimeArch = runtime.GOARCH

var openConsoleDLL, openConsoleExe []byte
