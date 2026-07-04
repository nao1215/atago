package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// SandboxHomeDirName is the fixed directory created inside the scenario workdir
// to hold the isolated home (#71). Its path is deterministic
// (`${workdir}/.atago-home`), so a spec can inspect files the CLI wrote there
// with ordinary file: asserts — no new builtin variable is needed.
const SandboxHomeDirName = ".atago-home"

// SandboxHomeVars returns the environment variables that redirect a child
// process's home and per-OS config/cache/data/state directories at home, plus
// the concrete directories that must exist under it (#71). goos selects the
// variable family so the mapping is unit-testable off-platform; separators are
// chosen from goos too, which is correct on the host because goos == GOOS in
// real use.
func SandboxHomeVars(goos, home string) (vars map[string]string, dirs []string) {
	sep := "/"
	if goos == "windows" {
		sep = `\`
	}
	join := func(parts ...string) string { return strings.Join(parts, sep) }

	if goos == "windows" {
		appdata := join(home, "AppData", "Roaming")
		local := join(home, "AppData", "Local")
		vars = map[string]string{
			"USERPROFILE":  home,
			"APPDATA":      appdata,
			"LOCALAPPDATA": local,
		}
		// HOMEDRIVE/HOMEPATH split the home path the way cmd.exe expects. Derive
		// the drive from the path itself (not filepath.VolumeName, which is a
		// no-op off Windows) so the mapping is deterministic in tests.
		if len(home) >= 2 && home[1] == ':' {
			vars["HOMEDRIVE"] = home[:2]
			vars["HOMEPATH"] = home[2:]
		}
		return vars, []string{home, appdata, local}
	}

	// Unix and macOS: HOME plus the XDG base-directory family. macOS derives its
	// ~/Library paths from HOME, so setting HOME plus the XDG vars covers both.
	config := join(home, ".config")
	cache := join(home, ".cache")
	data := join(home, ".local", "share")
	state := join(home, ".local", "state")
	vars = map[string]string{
		"HOME":            home,
		"XDG_CONFIG_HOME": config,
		"XDG_CACHE_HOME":  cache,
		"XDG_DATA_HOME":   data,
		"XDG_STATE_HOME":  state,
	}
	return vars, []string{home, config, cache, data, state}
}

// EnsureSandboxHome creates `${workdir}/.atago-home` and its per-OS subdirectories
// and returns the environment overlay redirecting the child's home there (#71).
// The directory path is deterministic and reused across steps within a scenario:
// MkdirAll is idempotent, so a CLI can write config in one step and read it back
// in the next.
func EnsureSandboxHome(workdir string) (map[string]string, error) {
	home := filepath.Join(workdir, SandboxHomeDirName)
	vars, dirs := SandboxHomeVars(runtime.GOOS, home)
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil { //nolint:gosec // the sandbox home lives inside the scenario workdir
			return nil, fmt.Errorf("sandbox_home: %w", err)
		}
	}
	return vars, nil
}
