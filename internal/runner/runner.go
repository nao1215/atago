// Package runner defines the Runner interface and the Result it produces.
// Concrete runners live in subpackages (internal/runner/cmd, .../http).
package runner

import (
	"context"
	"net/http"
	"time"

	"github.com/nao1215/atago/internal/spec"
)

// Result is the externally observable outcome of a run step (spec.md §31.3).
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

	// Browser fields, set only by the browser/CDP runner. CDPValue is the value
	// captured by the last text/eval action (the document the `value` assertion
	// target and `store from.value` read): a text capture is the raw string, an
	// eval capture is JSON.
	IsCDP    bool
	CDPValue []byte
}

// Runner executes a run step within a scenario workdir and returns the observed
// Result. A non-nil error means the runner could not execute at all (an
// execution error, spec.md §34 code 4); a command that runs but exits non-zero
// is a successful Run with Result.ExitCode set.
type Runner interface {
	Run(ctx context.Context, run *spec.Run, workdir string) (*Result, error)
}
