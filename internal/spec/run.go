package spec

// Fixture materializes an input file in the scenario workdir.
// Exactly one source is set: inline Content, inline Base64, or From (copy an
// existing file — e.g. committed binary testdata — resolved relative to the
// spec file's directory).
type Fixture struct {
	File    string `yaml:"file"`
	Content string `yaml:"content,omitempty"`
	Base64  string `yaml:"base64,omitempty"`
	From    string `yaml:"from,omitempty"`
	// Symlink, when set, makes File a symbolic link to this target instead of a
	// regular file (mutually exclusive with content/base64/from). The target is
	// written verbatim, so use ${workdir} for an absolute in-workdir target.
	Symlink string `yaml:"symlink,omitempty"`
	// Mode, when set, is an octal file mode (e.g. "0444") applied after writing.
	Mode string `yaml:"mode,omitempty"`
	// Mtime, when set, is an RFC3339 modification time applied after writing, so
	// specs can pin timestamps for content-vs-mtime change detection.
	Mtime string `yaml:"mtime,omitempty"`
}

// Run executes a command.
type Run struct {
	Command string `yaml:"command"`
	Runner  string `yaml:"runner,omitempty"`
	// Shell is a pointer so an authored `shell: false` is distinguishable from
	// "unset" when `defaults.run.shell` is layered in (an authored value always
	// wins over a default).
	Shell   *bool  `yaml:"shell,omitempty"`
	Cwd     string `yaml:"cwd,omitempty"`
	Timeout string `yaml:"timeout,omitempty"`
	// TimeoutSource names the level that supplied the effective Timeout
	// (run.timeout / runner.timeout / defaults.run.timeout / suite.timeout /
	// built-in default). It is set by the engine's precedence resolver (#17),
	// never authored, and lets a timeout-kill failure hint say which level to
	// adjust.
	TimeoutSource string            `yaml:"-"`
	Env           map[string]string `yaml:"env,omitempty"`
	// ClearEnv starts the child from an empty environment instead of inheriting
	// the host environment (#16), so host vars (LANG, GIT_*, proxies, ...) cannot
	// silently change the behavior under test. A pointer so an authored
	// `clear_env: false` is distinguishable from "unset" when
	// `defaults.run.clear_env` is layered in.
	ClearEnv *bool `yaml:"clear_env,omitempty"`
	// PassEnv copies the listed host variables into the cleared environment
	// (#16). Only meaningful with ClearEnv (the loader rejects it otherwise);
	// unset host variables are skipped, not an error.
	PassEnv []string `yaml:"pass_env,omitempty"`
	// SandboxHome points the child's home directory and per-OS config/cache/
	// data/state locations at a fresh `${workdir}/.atago-home` so a CLI that
	// reads or writes ~/.config, ~/.cache, or %APPDATA% runs hermetically. The
	// sandbox variables sit between pass_env and the step's own env in
	// precedence, and compose with clear_env (injected after the clear). A
	// pointer so an authored `sandbox_home: false` beats a defaulted true.
	SandboxHome *bool `yaml:"sandbox_home,omitempty"`
	// Stdin feeds the command's standard input: a scalar string (inline
	// text), {file: path} (workdir-relative, expanded and confined), or
	// {base64: data} for binary bytes (#18).
	Stdin Stdin `yaml:"stdin,omitempty"`
	// StdoutTo / StderrTo redirect the command's captured stdout / stderr to a
	// workdir-relative file (create/truncate), so a `shell: false` step can write
	// output to a file without borrowing the shell's `>` operator. The streams are
	// still captured internally, so stdout/stderr assertions on the same step keep
	// working. Paths follow the same workdir-confinement rule as assertion paths.
	StdoutTo string `yaml:"stdout_to,omitempty"`
	StderrTo string `yaml:"stderr_to,omitempty"`
	// Retry, when set, re-runs the command until the Until assertion passes,
	// polling declaratively for async behavior.
	Retry *Retry `yaml:"retry,omitempty"`
}

// Bool returns a pointer to v — sugar for authoring optional booleans (Shell)
// in Go literals.
func Bool(v bool) *bool { return &v }

// ShellEnabled reports whether the step opts into shell execution.
func (r *Run) ShellEnabled() bool { return r.Shell != nil && *r.Shell }

// ClearEnvEnabled reports whether the step opts into a cleared environment (#16).
func (r *Run) ClearEnvEnabled() bool { return r.ClearEnv != nil && *r.ClearEnv }

// SandboxHomeEnabled reports whether the run step opts into an isolated home (#71).
func (r *Run) SandboxHomeEnabled() bool { return r.SandboxHome != nil && *r.SandboxHome }

// Retry re-runs a command until Until passes or the attempt budget is exhausted.
// The last attempt's result is what subsequent steps observe.
type Retry struct {
	// Times is the maximum number of attempts (>= 1).
	Times int `yaml:"times"`
	// Interval is the wait between attempts as a Go duration (e.g. "200ms").
	// Empty means no wait.
	Interval string `yaml:"interval,omitempty"`
	// Until is a single assertion polled after each attempt; the loop stops as
	// soon as it passes. When it never passes, the run step fails.
	Until *Assert `yaml:"until"`
}
