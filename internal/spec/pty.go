package spec

// ClearEnvEnabled reports whether the pty step opts into a cleared environment (#16).
func (p *PTY) ClearEnvEnabled() bool { return p.ClearEnv != nil && *p.ClearEnv }

// SandboxHomeEnabled reports whether the pty step opts into an isolated home (#71).
func (p *PTY) SandboxHomeEnabled() bool { return p.SandboxHome != nil && *p.SandboxHome }

// PTY runs a command inside a pseudo-terminal (#8). The captured transcript
// (terminal echo included, ANSI intact) becomes the step's stdout, so every
// stream matcher, snapshot (with its ANSI normalization), and
// `store from.stdout` works unchanged.
type PTY struct {
	Command string `yaml:"command"`
	// Shell runs Command through the shell like run.shell.
	Shell *bool  `yaml:"shell,omitempty"`
	Cwd   string `yaml:"cwd,omitempty"`
	// Rows / Cols set the terminal size (default 24x80).
	Rows int `yaml:"rows,omitempty"`
	Cols int `yaml:"cols,omitempty"`
	// Timeout bounds the WHOLE session as a Go duration (default "30s"): a
	// prompt that never appears or a program that never exits fails loudly
	// instead of hanging the run.
	Timeout string            `yaml:"timeout,omitempty"`
	Env     map[string]string `yaml:"env,omitempty"`
	// ClearEnv starts the pty child from an empty environment instead of
	// inheriting the host environment (#16), mirroring run.clear_env.
	ClearEnv *bool `yaml:"clear_env,omitempty"`
	// PassEnv copies the listed host variables into the cleared environment
	// (#16). Only meaningful with ClearEnv; unset host variables are skipped.
	PassEnv []string `yaml:"pass_env,omitempty"`
	// SandboxHome isolates the pty child's home and per-OS config/cache/data/
	// state directories under `${workdir}/.atago-home`, mirroring run.sandbox_home.
	SandboxHome *bool `yaml:"sandbox_home,omitempty"`
	// Session is the ordered expect/send script. Each entry sets exactly one
	// of Expect (wait until the accumulated transcript matches the regexp),
	// Send (write the string to the terminal; an empty send transmits EOF,
	// i.e. ^D), or ExpectScreen (wait until the rendered terminal screen matches
	// a stream matcher, optionally stably for a duration). Deliberately no
	// branching — atago is not a scripting language.
	Session []PTYAction `yaml:"session,omitempty"`
}

// PTYAction is one expect/send/expect_screen/resize entry in a pty session (#8).
type PTYAction struct {
	// Expect waits until the transcript matches this regexp. A never-matching
	// expect fails the step (reported like an assertion) when the session
	// timeout elapses.
	Expect string `yaml:"expect,omitempty"`
	// Send writes to the terminal: a scalar string verbatim (the empty string
	// sends EOF/^D; ${name} expansion applies) or {key: <name>} for a named
	// key (#26) — enter, tab, shift-tab, esc, arrows, f1-f12, ctrl-a..ctrl-z,
	// alt-a..alt-z, modified arrows like ctrl-left, and common control-key
	// aliases like ctrl-space / ctrl-[ / ctrl-_ plus terminal key events like
	// ctrl-hyphen (#376) — so sessions stay readable instead of embedding \x1b
	// escapes.
	Send *PTYSend `yaml:"send,omitempty"`
	// Exec runs one command on the HOST while the program under test keeps
	// running (#380), so a session can test what a TUI does when the world
	// changes underneath it — a commit made outside lazygit, a file another
	// process creates, a log line appended to what a viewer is following.
	// Everything else in a session is caused by keystrokes; this is the one
	// action that is not. It blocks until the command exits, which is the
	// point: after it, the change exists, and the expect_screen that follows is
	// waiting for the program to notice.
	Exec *PTYExec `yaml:"exec,omitempty"`
	// Resize changes the terminal size mid-session (#379), delivering the size
	// change the way a real terminal does — SIGWINCH on POSIX, a ConPTY
	// notification on Windows — so a TUI's relayout is testable instead of
	// being fixed at whatever the step started with.
	Resize *PTYResize `yaml:"resize,omitempty"`
	// ExpectScreen waits until the CURRENT rendered screen (the transcript
	// replayed through the same vt10x emulator as a top-level `screen:` assert)
	// satisfies the matcher. `stable_for` requires the matcher to stay true
	// continuously for that long; `timeout` optionally bounds only this wait,
	// within the pty step's wider session timeout.
	ExpectScreen *PTYExpectScreen `yaml:"expect_screen,omitempty"`
}

// PTYResize is a mid-session terminal resize (#379). Both dimensions are
// required — there is no "keep the other one", because a spec that says what
// the window becomes reads better than one that says what it changes by.
//
// The size change reaches the child the way a real terminal delivers it, so a
// program that redraws on SIGWINCH redraws. The rendered screen follows: every
// later `expect_screen`, the post-step `screen:` assert, and the snapshot all
// see the transcript replayed at the sizes it was actually produced under.
//
// Authoring rule: settle the screen (an `expect` or `expect_screen`, ideally
// with `stable_for`) before and after a resize. Output already in flight when
// the resize lands is attributed to the old size, exactly as a real terminal
// would — waiting for a quiet screen is what makes the boundary unambiguous.
type PTYResize struct {
	Rows int `yaml:"rows"`
	Cols int `yaml:"cols"`
}

// PTYExpectScreen is a session-local rendered-screen wait: the matcher runs on
// the live terminal screen during a pty session, not only after the program
// exits. It reuses the StreamAssert surface (line/contains/matches/equals/json/
// yaml, etc.) except snapshot/trim, which are validated out of this
// mid-session context.
type PTYExpectScreen struct {
	// Embedding ScreenAssert rather than StreamAssert gives a mid-session wait
	// the attribute matchers too (#382): the live frame is the same frame the
	// post-step assert reads, so it deserves the same questions.
	ScreenAssert `yaml:",inline"`
	// Timeout bounds THIS wait only; when empty, the enclosing pty timeout
	// supplies the budget.
	Timeout string `yaml:"timeout,omitempty"`
	// StableFor requires the screen to keep matching continuously for at least
	// this duration before the action passes, absorbing redraw churn without a
	// blind sleep.
	StableFor string `yaml:"stable_for,omitempty"`
}
