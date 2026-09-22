//go:build windows && amd64

package conpty

import _ "embed"

const runtimeArch = "amd64"

//go:embed openconsole/amd64/conpty.dll
var openConsoleDLL []byte

//go:embed openconsole/amd64/OpenConsole.exe
var openConsoleExe []byte
