//go:build windows && arm64

package conpty

import _ "embed"

const runtimeArch = "arm64"

//go:embed openconsole/arm64/conpty.dll
var openConsoleDLL []byte

//go:embed openconsole/arm64/OpenConsole.exe
var openConsoleExe []byte
