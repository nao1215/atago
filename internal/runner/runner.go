// Package runner defines the Runner interface and the Result it produces.
// Concrete runners live in subpackages (internal/runner/cmd, .../http).
package runner

import (
	"context"
	"net/http"
	"time"

	"github.com/nao1215/atago/internal/fsdelta"
	"github.com/nao1215/atago/internal/spec"
)

// Result is the externally observable outcome of a run step.
//
// A Result describes either a process run (the cmd runner: Command, ExitCode,
// Stdout, Stderr) or an HTTP exchange (the http runner: IsHTTP, StatusCode,
// Header, Body). The engine tracks the most recent Result as the scenario's
// "current" observation, and assertions/stores read from whichever family of
// fields applies.
type Result struct {
	Command  string
	ExitCode int
	Stdout   []byte
	Stderr   []byte
	Duration time.Duration
	Workdir  string
	TimedOut bool
	// TimeoutSource names the level that supplied the timeout that killed the
	// command (run.timeout / runner.timeout / defaults.run.timeout /
	// suite.timeout / built-in default), so the failure hint can say which
	// knob to adjust (#17). Empty when TimedOut is false.
	TimeoutSource string

	// EarlyEOF records, per stream name ("stdout" / "stderr"), how far ahead of
	// the command's own exit that stream reached end-of-file.
	//
	// atago creates the capture pipes itself and closes the parent's copy of each
	// write end immediately after Start, so from that moment the child process is
	// their only holder: EOF on the read end cannot arrive before the child exits
	// — unless the child closed its own stdout, or never had atago's pipe at all.
	// That distance is the one piece of evidence separating "the command printed
	// nothing" from "the command's output went somewhere else", which #339 spent a
	// day being unreadable for.
	//
	// An entry exists only when the gap is large enough to be meaningful (see the
	// cmd runner's margin), so an ordinary command carries none and the map is
	// nil. Closing stdout early is legal — a daemonizing tool does it — so this is
	// only ever a hint on a failure that already happened, never a failure of its
	// own (#344).
	EarlyEOF map[string]time.Duration

	// HTTP fields, set only by the http runner (IsHTTP reports which family is
	// populated, since a zero StatusCode is indistinguishable from "no response").
	IsHTTP     bool
	StatusCode int
	Header     http.Header
	Body       []byte

	// DB fields, set only by the db runner. RowsJSON is the result rows encoded as
	// a JSON array (the document the `rows` assertion target and `store from.rows`
	// read); RowsAffected is set for non-row statements (INSERT/UPDATE/DDL).
	IsDB         bool
	RowsJSON     []byte
	RowsAffected int64

	// gRPC fields, set only by the grpc runner. GRPCStatus is the numeric status
	// code; MessageJSON is the response message encoded as JSON (the document the
	// `message` assertion target and `store from.message` read).
	IsGRPC      bool
	GRPCStatus  int
	MessageJSON []byte

	// PTY fields, set only by the pty runner. Screen is the final rendered
	// terminal screen (vt10x emulation, plain text) for the `screen`
	// assertion target (#27); Stdout keeps the raw transcript.
	IsPTY  bool
	Screen []byte

	// Browser fields, set only by the browser/CDP runner. CDPValue is the value
	// captured by the last text/eval action (the document the `value` assertion
	// target and `store from.value` read): a text capture is the raw string, an
	// eval capture is JSON.
	IsCDP    bool
	CDPValue []byte

	// Changes is the workdir delta observed around a run/pty step (#70), set by
	// the engine only when a `changes:` assert immediately follows the step. Nil
	// means no delta was recorded (the assertion then reports that), so scenarios
	// that never use `changes:` pay for no workdir scan.
	Changes *fsdelta.Delta
}

// Runner executes a run step within a scenario workdir and returns the observed
// Result. A non-nil error means the runner could not execute at all (an
// execution error, exit code 4); a command that runs but exits non-zero
// is a successful Run with Result.ExitCode set.
type Runner interface {
	Run(ctx context.Context, run *spec.Run, workdir string) (*Result, error)
}
