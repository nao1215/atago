// Package assert evaluates assertion steps against the current run result.
// Each Check returns a CheckResult carrying enough structured context
// (expected/actual/hint) for the console failure output and the JSON report.
package assert

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/runner/mock"
	"github.com/nao1215/atago/internal/spec"
)

// intList renders an accepted exit-code set as "[0, 2]" for descriptions and
// failure output (#19).
func intList(ns []int) string {
	parts := make([]string, len(ns))
	for i, n := range ns {
		parts[i] = strconv.Itoa(n)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// CheckResult is the structured outcome of one assertion.
type CheckResult struct {
	OK bool
	// Target is the assertion target family that produced this result (the
	// spec.AssertTarget string, e.g. "exit_code", "stdout", "file"). Reports
	// branch on it structurally — e.g. the console failure block appends the
	// captured stderr tail only for exit_code failures — instead of sniffing
	// the human-facing Desc.
	Target   string
	Desc     string // human label, e.g. `assert stdout contains "Alice"`
	Expected string
	Actual   string
	Hint     string
	// SnapshotUpdated reports that this check WROTE its snapshot instead of
	// comparing against it (`--update-snapshots`). Rewriting the committed
	// goldens is the one passing outcome a reviewer has to be told about, so
	// the reports can count it rather than reading it out of the description.
	SnapshotUpdated bool

	// ArtifactKind, ArtifactActual, and ArtifactExpected carry the full,
	// untruncated payloads a failed text assertion compared, for durable export
	// via --artifacts-dir (#48). Unlike Actual/Expected, which are excerpted for
	// display, these hold the complete bytes so a reviewer can inspect exactly
	// what atago matched against. They are set only for text-based assertions
	// (stdout/stderr/body/rows/message/value/file/snapshot); other checks leave
	// ArtifactKind empty, meaning "nothing to export". ArtifactExpected is nil
	// when the assertion has no meaningful expected payload (e.g. contains). The
	// engine masks these before writing, using the same masker as the display
	// fields.
	ArtifactKind     string
	ArtifactActual   []byte
	ArtifactExpected []byte

	// ArtifactBlobs are additional named binary payloads to persist for this
	// failed assertion (#52). Where ArtifactActual/ArtifactExpected are text, a
	// blob carries its own role and file extension, letting an image similar_to
	// failure emit the actual image, the baseline image, a deterministic visual
	// diff heatmap, and a metadata JSON as separate sidecar files.
	ArtifactBlobs []ArtifactBlob

	// ArtifactFiles lists the sidecar files the engine wrote for this failed
	// assertion when --artifacts-dir was set (#48). Paths are relative to the
	// artifacts dir root, in stable role order (actual before expected). It is
	// empty when no artifacts dir was configured or the assertion passed.
	ArtifactFiles []ArtifactFile
}

// ArtifactBlob is one named binary payload to persist for a failed assertion
// (#52). Role shapes the filename (e.g. "actual", "baseline", "diff",
// "metadata") and Ext is its file extension (e.g. "png", "json").
type ArtifactBlob struct {
	Role string
	Ext  string
	Data []byte
}

// ArtifactFile references one sidecar file written for a failed assertion when
// --artifacts-dir is set (#48).
type ArtifactFile struct {
	Role string // "actual" | "expected"
	Path string // relative to the artifacts dir root, slash-separated
}

// Env carries the resolution context an assertion needs: the scenario's working
// directory (for file paths and snapshot normalization), the spec file's
// directory (where committed snapshot files live), and whether snapshots should
// be written rather than compared.
type Env struct {
	Workdir         string
	SpecDir         string
	UpdateSnapshots bool
	// Secrets, when set, masks declared secret values in captured output before a
	// snapshot is written or compared, so a real credential is never committed to
	// a golden file (issue #11).
	Secrets func([]byte) []byte
	// Scrub, when set, applies the spec's declarative regex→placeholder rewrites
	// (#137) during snapshot normalization, so volatile output patterns the
	// built-in normalizers do not cover (auto-increment IDs, request identifiers)
	// are determinized before a snapshot is written or compared.
	Scrub func([]byte) []byte
	// MockRecords, when set, resolves a mock server's recorded requests by
	// name for the `mock:` assertion target (#24). Nil in contexts with no
	// mock servers (retry `until` asserts, direct API use).
	MockRecords func(name string) ([]mock.Record, bool)
	// SnapshotWrites, when set, is the run-scoped record of which snapshot
	// paths this `--update-snapshots` run has written and with what content, so
	// two scenarios cannot silently claim one path with different output. Nil
	// outside a run, where there is nothing to scope the writes to.
	SnapshotWrites *SnapshotWrites
	// Writer identifies whose write a snapshot claim is — a scenario, or a suite
	// lifecycle block. A path claimed twice by the SAME writer is a repeat
	// iteration or a retry attempt producing different output, which is a
	// different finding from two scenarios sharing one golden.
	Writer string
}

func pass(desc string) *CheckResult { return &CheckResult{OK: true, Desc: desc} }

// CheckAll evaluates every target set on an assert step and returns one
// CheckResult per target, in SetTargets order. An assert may set more than one
// target (exit_code + stdout + file …); each is an independent check and all
// must hold. res may be nil for targets that do not depend on a command (e.g.
// file assertions), in which case env.Workdir is still used to resolve paths.
// The returned slice always has at least one element.
func CheckAll(a *spec.Assert, res *runner.Result, env Env) []*CheckResult {
	targets := a.SetTargets()
	if len(targets) == 0 {
		return []*CheckResult{{Desc: "assert", Hint: "assertion must set at least one target"}}
	}
	out := make([]*CheckResult, 0, len(targets))
	for _, t := range targets {
		r := checkTarget(a, t, res, env)
		r.Target = string(t)
		out = append(out, r)
	}
	return out
}

// AllOK reports whether every check in the slice passed.
func AllOK(results []*CheckResult) bool {
	for _, r := range results {
		if r == nil || !r.OK {
			return false
		}
	}
	return true
}

// Check evaluates an assert step and returns a single verdict: the first failing
// target's result, or the first result when all pass. It is a convenience over
// CheckAll for callers (and tests) that only need one pass/fail outcome.
func Check(a *spec.Assert, res *runner.Result, env Env) *CheckResult {
	results := CheckAll(a, res, env)
	for _, r := range results {
		if r != nil && !r.OK {
			return r
		}
	}
	return results[0]
}

// checkTarget evaluates one assertion target family against the run result.
func checkTarget(a *spec.Assert, target spec.AssertTarget, res *runner.Result, env Env) *CheckResult {
	switch target {
	case spec.AssertExitCode:
		return checkExitCode(a.ExitCode, res)
	case spec.AssertStdout:
		// Two independent post-passes over a failing stream check, applied in the
		// order they read best: where the text actually is (#347), then why this
		// stream may be short (#344).
		return withEarlyEOFHint("stdout", res,
			withCounterpartHint("stdout", a.Stdout, res, env,
				checkStream("stdout", a.Stdout, streamBytes(res, "stdout"), res != nil, env)))
	case spec.AssertStderr:
		return withEarlyEOFHint("stderr", res,
			withCounterpartHint("stderr", a.Stderr, res, env,
				checkStream("stderr", a.Stderr, streamBytes(res, "stderr"), res != nil, env)))
	case spec.AssertFile:
		return checkFile(a.File, env)
	case spec.AssertStatus:
		return checkStatus(a.Status, res)
	case spec.AssertHeader:
		return checkHeader(a.Header, res)
	case spec.AssertBody:
		return checkStream("body", a.Body, httpBody(res), res != nil && res.IsHTTP, env)
	case spec.AssertRows:
		return checkStream("rows", a.Rows, dbRows(res), res != nil && res.IsDB, env)
	case spec.AssertGRPCStatus:
		return checkGRPCStatus(a.GRPCStatus, res)
	case spec.AssertMessage:
		return checkStream("message", a.Message, grpcMessage(res), res != nil && res.IsGRPC, env)
	case spec.AssertValue:
		return checkStream("value", a.Value, cdpValue(res), res != nil && res.IsCDP, env)
	case spec.AssertImage:
		return checkImage(a.Image, env)
	case spec.AssertDir:
		return checkDir(a.Dir, env)
	case spec.AssertPDF:
		return checkPDF(a.PDF, env)
	case spec.AssertMock:
		return checkMock(a.Mock, env)
	case spec.AssertScreen:
		return checkScreen(a.Screen, res, env)
	case spec.AssertDuration:
		return checkDuration(a.Duration, res)
	case spec.AssertChanges:
		return checkChanges(a.Changes, res, env)
	default:
		return &CheckResult{Desc: string(target), Hint: "assertion target not supported yet"}
	}
}

// withCounterpartHint appends a suggestion to a failed stdout/stderr check when
// the OTHER stream satisfies the very same assertion (#347). Asserting `stdout:`
// on text a CLI actually prints to `stderr` is the most common mistake in CLI
// end-to-end testing, and atago holds both streams while it reports the failure
// — the answer used to sit in `--verbose` output the reader had to go find,
// costing a round trip per occurrence in CI.
//
// It is a post-pass on the returned result, so checkStream stays unaware of
// streams it does not own. It never changes the verdict, and it exists only for
// the stdout/stderr pair: body/rows/message/value have no counterpart and the
// code must not pretend otherwise.
func withCounterpartHint(name string, s *spec.StreamAssert, res *runner.Result, env Env, out *CheckResult) *CheckResult {
	if out == nil || out.OK || s == nil || res == nil || !canSuggestCounterpart(s) {
		return out
	}
	other := "stderr"
	if name == "stderr" {
		other = "stdout"
	}
	// Re-run the SAME assertion through the SAME code path against the other
	// stream. Anything less would have to re-derive CRLF folding and the `line:`
	// selector, and a suggestion that is wrong on Windows — or wrong for a
	// line-scoped assert — is worse than no suggestion at all. A `line:` that the
	// counterpart does not have therefore simply fails to satisfy the assertion,
	// and no suggestion is made.
	if !checkStream(other, s, streamBytes(res, other), true, env).OK {
		return out
	}
	suggestion := fmt.Sprintf("%s satisfies this assertion (assert `%s:` instead?)", other, other)
	if out.Hint == "" {
		out.Hint = suggestion
	} else {
		out.Hint += " — but " + suggestion
	}
	return out
}

// withEarlyEOFHint appends the one fact a failing stream check cannot derive for
// itself: that the stream ended before the command did (#344).
//
// The cmd runner closes the parent's copy of each capture pipe's write end right
// after Start, so the child is its only holder and EOF cannot arrive first —
// unless the child closed its own stdout, or never had atago's pipe at all. A
// report that says only "the substring was not present in stdout" leaves those
// indistinguishable from a command that printed nothing, which is exactly what
// made #339 unreadable.
//
// It is a hint on an existing failure, never a failure of its own: closing stdout
// early is legal (a daemonizing tool does it), so a step that asserts nothing
// about the stream must stay green. Results with no observation — every ordinary
// command, and every non-process result family — grow no note.
func withEarlyEOFHint(name string, res *runner.Result, out *CheckResult) *CheckResult {
	if out == nil || out.OK || res == nil {
		return out
	}
	d, ok := res.EarlyEOF[name]
	if !ok {
		return out
	}
	note := fmt.Sprintf("%s ended %s before the command exited — its %s was closed, or was never connected to atago's pipe",
		name, d.Round(time.Millisecond), name)
	if out.Hint == "" {
		out.Hint = note
	} else {
		out.Hint += " — " + note
	}
	return out
}

// canSuggestCounterpart reports whether every matcher set on s is one whose
// success against the other stream is evidence the author named the wrong one.
//
// Only the positive text matchers qualify. A negative matcher proves nothing: a
// needle being absent from stderr too is not a reason to assert stderr. `empty`
// is trivially satisfied by any silent stream and would fire constantly. `json`
// and `yaml` compare a parsed document, where "the other stream also parses"
// says little. `snapshot` is excluded for a second reason as well: re-running it
// would compare against — and under --update-snapshots write — a golden file.
//
// Every set matcher must qualify, so a mixed assert (contains + not_contains)
// stays silent rather than suggesting on the strength of half of it.
func canSuggestCounterpart(s *spec.StreamAssert) bool {
	set := s.SetMatchers()
	if len(set) == 0 {
		return false
	}
	for _, m := range set {
		switch m {
		case "contains", "matches", "equals":
		default:
			return false
		}
	}
	return true
}

func streamBytes(res *runner.Result, which string) []byte {
	if res == nil {
		return nil
	}
	if which == "stdout" {
		return res.Stdout
	}
	return res.Stderr
}

func checkExitCode(e *spec.ExitCode, res *runner.Result) *CheckResult {
	if res == nil {
		return &CheckResult{Desc: "assert exit_code", Hint: "no command has run in this scenario yet"}
	}
	// A timed-out command was killed, not exited: say so instead of presenting
	// the synthetic -1 as a normal exit code.
	actual := fmt.Sprintf("exit code %d", res.ExitCode)
	if res.TimedOut {
		actual = fmt.Sprintf("exit code %d (the command timed out after %s and was killed)", res.ExitCode, res.Duration.Round(time.Millisecond))
	}
	// timeoutHint replaces the mismatch hint when the command was killed by a
	// timeout, naming the level that supplied it (step/runner/defaults/suite/
	// built-in, #17) so the user knows which knob to adjust.
	timeoutHint := func(fallback string) string {
		if !res.TimedOut {
			return fallback
		}
		source := res.TimeoutSource
		if source == "" {
			source = "run.timeout"
		}
		return fmt.Sprintf("the command hit its %s after %s and was killed before exiting", source, res.Duration.Round(time.Millisecond))
	}
	switch {
	case e.Equals != nil:
		desc := fmt.Sprintf("assert exit_code is %d", *e.Equals)
		if res.ExitCode == *e.Equals {
			return pass(desc)
		}
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("exit code %d", *e.Equals),
			Actual:   actual,
			Hint:     timeoutHint(fmt.Sprintf("expected exit code %d but the command exited with %d", *e.Equals, res.ExitCode)),
		}
	case len(e.In) > 0:
		set := intList(e.In)
		desc := fmt.Sprintf("assert exit_code in %s", set)
		if slices.Contains(e.In, res.ExitCode) {
			return pass(desc)
		}
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("exit code in %s", set),
			Actual:   actual,
			Hint:     timeoutHint(fmt.Sprintf("expected the exit code to be one of %s but the command exited with %d", set, res.ExitCode)),
		}
	case e.Not != nil:
		desc := fmt.Sprintf("assert exit_code is not %d", *e.Not)
		if res.ExitCode != *e.Not {
			return pass(desc)
		}
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("exit code != %d", *e.Not),
			Actual:   actual,
			Hint:     fmt.Sprintf("expected any exit code except %d", *e.Not),
		}
	default:
		return &CheckResult{Desc: "assert exit_code", Hint: "exit_code must be an int, {not: int}, or {in: [int, ...]}"}
	}
}
