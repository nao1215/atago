// Package buildinfo reports the binary's version. Release archives get it
// injected at link time; `go install`ed binaries fall back to the module
// version recorded by the Go toolchain, so `atago version` is meaningful on
// every install path instead of printing "dev".
package buildinfo

import "runtime/debug"

// Version is replaced at link time for tagged release builds.
var Version = "dev"

// Get returns the effective version: the ldflags-injected release version when
// present, otherwise the module version from the embedded build info (set by
// `go install module@version`), otherwise "dev" (a source build in a checkout).
func Get() string {
	if Version != "dev" {
		return Version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return Version
}
