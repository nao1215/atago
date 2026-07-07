package engine

import (
	"runtime"
	"testing"
)

// TestEngine_SandboxHome_ReuseAcrossSteps proves the isolated home is created
// once and reused: a CLI writes config under the sandboxed home in step 1 and
// reads it back in step 2, and the file is visible under ${workdir}/.atago-home
// to an ordinary file: assert (#71).
func TestEngine_SandboxHome_ReuseAcrossSteps(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("XDG variable family is POSIX-only")
	}
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: sandbox
scenarios:
  - name: writes then reads its config
    steps:
      - run:
          shell: true
          sandbox_home: true
          command: 'mkdir -p "$XDG_CONFIG_HOME/app" && echo vim > "$XDG_CONFIG_HOME/app/editor"'
      - run:
          shell: true
          sandbox_home: true
          command: 'cat "$XDG_CONFIG_HOME/app/editor"'
      - assert:
          stdout:
            equals: "vim\n"
      - assert:
          file:
            path: .atago-home/.config/app/editor
            contains: vim
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios[0].Steps)
	}
}

// TestEngine_SandboxHome_ClearEnvComposition proves that with clear_env +
// sandbox_home the child sees the sandbox HOME rather than the host's, and that
// pass_env: [HOME] cannot leak the host home past the sandbox (#71).
func TestEngine_SandboxHome_ClearEnvComposition(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("HOME/XDG family is POSIX-only")
	}
	t.Setenv("HOME", "/definitely/not/the/sandbox")
	res := runSpec(t, `
version: "1"
suite:
  name: sandbox
scenarios:
  - name: sandbox home wins over pass_env
    steps:
      - run:
          shell: true
          clear_env: true
          pass_env: [HOME]
          sandbox_home: true
          command: 'printf "%s" "$HOME"'
      - assert:
          stdout:
            matches: '/\.atago-home$'
      - assert:
          stdout:
            not_contains: /definitely/not/the/sandbox
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios[0].Steps)
	}
}

// TestEngine_SandboxHome_DefaultsRunGovernsPTY proves defaults.run's sandbox
// family (clear_env + sandbox_home) layers onto a pty step, so a shared
// hermetic default governs interactive steps too (#77): the pty child sees the
// sandbox HOME under ${workdir}/.atago-home, never the host home, even though
// only defaults.run — not the pty step — declares the family. Mirrors
// TestEngine_SandboxHome_ClearEnvComposition for the pty path.
func TestEngine_SandboxHome_DefaultsRunGovernsPTY(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the HOME/XDG sandbox-home family is POSIX-only")
	}
	t.Setenv("HOME", "/definitely/not/the/sandbox")
	res := runSpec(t, `
version: "1"
suite:
  name: sandbox
defaults:
  run:
    clear_env: true
    pass_env: [PATH, HOME]
    sandbox_home: true
scenarios:
  - name: a shared default sandboxes a pty child's home
    steps:
      - pty:
          shell: true
          command: 'printf "HOME=[%s]\n" "$HOME"'
          session:
            - expect: 'HOME='
      - assert:
          stdout:
            contains: /.atago-home]
      - assert:
          stdout:
            not_contains: /definitely/not/the/sandbox
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios[0].Steps)
	}
}

// TestEngine_SandboxHome_IsolatedBetweenScenarios proves each scenario gets its
// own workdir and therefore its own isolated home: a file written in one
// scenario is not visible in the next (#71).
func TestEngine_SandboxHome_IsolatedBetweenScenarios(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("XDG variable family is POSIX-only")
	}
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: sandbox
scenarios:
  - name: first writes a marker
    steps:
      - run:
          shell: true
          sandbox_home: true
          command: 'echo one > "$XDG_DATA_HOME/marker"'
      - assert:
          file:
            path: .atago-home/.local/share/marker
            exists: true
  - name: second starts clean
    steps:
      - run:
          shell: true
          sandbox_home: true
          command: 'test ! -e "$XDG_DATA_HOME/marker"; echo $?'
      - assert:
          stdout:
            equals: "0\n"
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}
