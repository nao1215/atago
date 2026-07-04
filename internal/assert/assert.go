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
	OK       bool
	Desc     string // human label, e.g. `assert stdout contains "Alice"`
	Expected string
	Actual   string
	Hint     string

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
	// MockRecords, when set, resolves a mock server's recorded requests by
	// name for the `mock:` assertion target (#24). Nil in contexts with no
	// mock servers (retry `until` asserts, direct API use).
	MockRecords func(name string) ([]mock.Record, bool)
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
		out = append(out, checkTarget(a, t, res, env))
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
		return checkStream("stdout", a.Stdout, streamBytes(res, "stdout"), res != nil, env)
	case spec.AssertStderr:
		return checkStream("stderr", a.Stderr, streamBytes(res, "stderr"), res != nil, env)
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
		return checkChanges(a.Changes, res)
	default:
		return &CheckResult{Desc: string(target), Hint: "assertion target not supported yet"}
	}
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
