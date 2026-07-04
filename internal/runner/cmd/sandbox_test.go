package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestSandboxHomeVars_PerOS pins the environment family each OS gets and proves
// every path lands under the isolated home (#71).
func TestSandboxHomeVars_PerOS(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		home     string
		want     map[string]string
		wantDirs []string
	}{
		{
			name: "unix",
			goos: "linux",
			home: "/work/.atago-home",
			want: map[string]string{
				"HOME":            "/work/.atago-home",
				"XDG_CONFIG_HOME": "/work/.atago-home/.config",
				"XDG_CACHE_HOME":  "/work/.atago-home/.cache",
				"XDG_DATA_HOME":   "/work/.atago-home/.local/share",
				"XDG_STATE_HOME":  "/work/.atago-home/.local/state",
			},
			wantDirs: []string{
				"/work/.atago-home",
				"/work/.atago-home/.config",
				"/work/.atago-home/.cache",
				"/work/.atago-home/.local/share",
				"/work/.atago-home/.local/state",
			},
		},
		{
			name: "darwin",
			goos: "darwin",
			home: "/work/.atago-home",
			want: map[string]string{
				"HOME":            "/work/.atago-home",
				"XDG_CONFIG_HOME": "/work/.atago-home/.config",
				"XDG_CACHE_HOME":  "/work/.atago-home/.cache",
				"XDG_DATA_HOME":   "/work/.atago-home/.local/share",
				"XDG_STATE_HOME":  "/work/.atago-home/.local/state",
			},
			wantDirs: []string{
				"/work/.atago-home",
				"/work/.atago-home/.config",
				"/work/.atago-home/.cache",
				"/work/.atago-home/.local/share",
				"/work/.atago-home/.local/state",
			},
		},
		{
			name: "windows",
			goos: "windows",
			home: `C:\work\.atago-home`,
			want: map[string]string{
				"USERPROFILE":  `C:\work\.atago-home`,
				"APPDATA":      `C:\work\.atago-home\AppData\Roaming`,
				"LOCALAPPDATA": `C:\work\.atago-home\AppData\Local`,
				"HOMEDRIVE":    "C:",
				"HOMEPATH":     `\work\.atago-home`,
			},
			wantDirs: []string{
				`C:\work\.atago-home`,
				`C:\work\.atago-home\AppData\Roaming`,
				`C:\work\.atago-home\AppData\Local`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars, dirs := SandboxHomeVars(tt.goos, tt.home)
			if len(vars) != len(tt.want) {
				t.Fatalf("vars = %v, want %v", vars, tt.want)
			}
			for k, want := range tt.want {
				if got := vars[k]; got != want {
					t.Errorf("vars[%q] = %q, want %q", k, got, want)
				}
			}
			if len(dirs) != len(tt.wantDirs) {
				t.Fatalf("dirs = %v, want %v", dirs, tt.wantDirs)
			}
			for i, want := range tt.wantDirs {
				if dirs[i] != want {
					t.Errorf("dirs[%d] = %q, want %q", i, dirs[i], want)
				}
			}
		})
	}
}

// TestEnsureSandboxHome_CreatesDirs proves the isolated home tree is created
// under the workdir and that a second call is idempotent (reuse across steps).
func TestEnsureSandboxHome_CreatesDirs(t *testing.T) {
	workdir := t.TempDir()
	vars, err := EnsureSandboxHome(workdir)
	if err != nil {
		t.Fatalf("EnsureSandboxHome: %v", err)
	}
	home := filepath.Join(workdir, SandboxHomeDirName)
	if runtime.GOOS == "windows" {
		if vars["USERPROFILE"] != home {
			t.Errorf("USERPROFILE = %q, want %q", vars["USERPROFILE"], home)
		}
	} else if vars["HOME"] != home {
		t.Errorf("HOME = %q, want %q", vars["HOME"], home)
	}
	if fi, err := os.Stat(home); err != nil || !fi.IsDir() {
		t.Fatalf("home dir not created: %v", err)
	}
	// Idempotent: a second call must not error even though the dirs exist.
	if _, err := EnsureSandboxHome(workdir); err != nil {
		t.Fatalf("second EnsureSandboxHome: %v", err)
	}
}

// TestBuildEnv_SandboxPrecedence proves the precedence chain
// step env > sandbox > pass_env > host holds under clear_env (#71): pass_env
// cannot leak the host home past the sandbox, but an explicit env wins.
func TestBuildEnv_SandboxPrecedence(t *testing.T) {
	t.Setenv("HOME", "/host/home")
	sandbox := map[string]string{"HOME": "/sandbox/home", "XDG_CONFIG_HOME": "/sandbox/home/.config"}

	// clear_env + pass_env [HOME] + sandbox: the sandbox wins over pass_env.
	env := envMap(BuildEnv(nil, true, []string{"HOME"}, sandbox))
	if got := env["HOME"]; got != "/sandbox/home" {
		t.Errorf("HOME = %q, want /sandbox/home (sandbox beats pass_env)", got)
	}
	if got := env["XDG_CONFIG_HOME"]; got != "/sandbox/home/.config" {
		t.Errorf("XDG_CONFIG_HOME = %q, want /sandbox/home/.config", got)
	}

	// An explicit step env still wins over the sandbox.
	env = envMap(BuildEnv(map[string]string{"HOME": "/explicit"}, true, []string{"HOME"}, sandbox))
	if got := env["HOME"]; got != "/explicit" {
		t.Errorf("HOME = %q, want /explicit (step env beats sandbox)", got)
	}

	// Without clear_env the sandbox still overrides the inherited host HOME.
	env = envMap(BuildEnv(nil, false, nil, sandbox))
	if got := env["HOME"]; got != "/sandbox/home" {
		t.Errorf("HOME = %q, want /sandbox/home (sandbox overrides inherited host)", got)
	}
}
