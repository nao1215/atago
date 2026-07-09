package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/engine"
	"github.com/nao1215/atago/internal/report"
)

func TestParseRunFlags_ReturnsValidatedOptions(t *testing.T) {
	t.Setenv("NO_COLOR", "")

	dir := t.TempDir()
	spec := writeSpec(t, dir, "ok.atago.yaml", passingSpec)
	arts := filepath.Join(dir, "artifacts")
	var out, errb bytes.Buffer

	opts, exit, done := parseRunFlags("atago run", []string{
		"--report", "json",
		"--update-snapshots",
		"--parallel", "3",
		"--fail-fast",
		"--filter", "alpha,beta",
		"--tag", "slow",
		"--skip-tag", "net",
		"--artifacts-dir", arts,
		"--rerun-failed",
		"--retry-failed", "2",
		"--verbose",
		"--ci",
		spec,
	}, &out, &errb)
	if done {
		t.Fatalf("done = true, want false (exit=%d, stderr=%s)", exit, errb.String())
	}
	if exit != ExitOK {
		t.Fatalf("exit = %d, want %d", exit, ExitOK)
	}
	if opts == nil {
		t.Fatal("opts = nil, want a populated runOptions")
	}
	if opts.label != "atago run" || opts.format != report.FormatJSON {
		t.Fatalf("opts = %+v, want label/report preserved", opts)
	}
	if !opts.updateSnapshots || !opts.failFast || !opts.rerunFailed || !opts.verbose || !opts.ci {
		t.Fatalf("opts = %+v, want boolean flags preserved", opts)
	}
	if opts.parallel != 3 || opts.retryFailed != 2 {
		t.Fatalf("opts = %+v, want parallel=3 retryFailed=2", opts)
	}
	if !reflect.DeepEqual([]string(opts.filter), []string{"alpha", "beta"}) {
		t.Fatalf("filter = %v, want [alpha beta]", opts.filter)
	}
	if !reflect.DeepEqual([]string(opts.tag), []string{"slow"}) {
		t.Fatalf("tag = %v, want [slow]", opts.tag)
	}
	if !reflect.DeepEqual([]string(opts.skipTag), []string{"net"}) {
		t.Fatalf("skipTag = %v, want [net]", opts.skipTag)
	}
	if !reflect.DeepEqual(opts.paths, []string{spec}) {
		t.Fatalf("paths = %v, want [%s]", opts.paths, spec)
	}
	if opts.artifactsDir != arts {
		t.Fatalf("artifactsDir = %q, want %q", opts.artifactsDir, arts)
	}
	if got := os.Getenv("NO_COLOR"); got != "1" {
		t.Fatalf("NO_COLOR = %q, want 1", got)
	}
	if info, err := os.Stat(arts); err != nil || !info.IsDir() {
		t.Fatalf("artifacts dir was not validated/created: stat err=%v isDir=%v", err, err == nil && info.IsDir())
	}
}

func TestParseRunFlags_RejectsUnusableArtifactsDir(t *testing.T) {
	dir := t.TempDir()
	spec := writeSpec(t, dir, "ok.atago.yaml", passingSpec)
	afile := filepath.Join(dir, "artifacts-file")
	if err := os.WriteFile(afile, []byte("not a dir"), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer

	opts, exit, done := parseRunFlags("atago run", []string{"--artifacts-dir", afile, spec}, &out, &errb)
	if !done {
		t.Fatalf("done = false, want true (opts=%+v)", opts)
	}
	if exit != ExitConfig {
		t.Fatalf("exit = %d, want %d", exit, ExitConfig)
	}
	if opts != nil {
		t.Fatalf("opts = %+v, want nil on config error", opts)
	}
	if !strings.Contains(errb.String(), "is not usable") {
		t.Fatalf("stderr = %q, want the artifacts-dir validation error", errb.String())
	}
}

func TestFinishRun_RerunMatchedNothingKeepsLedger(t *testing.T) {
	dir := t.TempDir()
	withWorkdir(t, dir, func() {
		if err := saveRerunState([]failedEntry{{SpecPath: "spec.atago.yaml", Scenario: "still failing"}}); err != nil {
			t.Fatal(err)
		}

		var out, errb bytes.Buffer
		opts := &runOptions{
			label:       "atago run",
			format:      report.FormatJSON,
			paths:       []string{"spec.atago.yaml"},
			rerunFailed: true,
			stdout:      &out,
			stderr:      &errb,
		}
		suiteResults := []*engine.SuiteResult{{
			Suite:    "sample",
			SpecPath: "spec.atago.yaml",
			Status:   engine.StatusPassed,
		}}

		if got := finishRun(opts, suiteResults, []error{nil}, nil, 5*time.Millisecond, context.Background()); got != ExitConfig {
			t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitConfig, errb.String())
		}
		if !strings.Contains(errb.String(), "no recorded failing scenarios matched the current specs") {
			t.Fatalf("stderr = %q, want the rerun-matched-nothing warning", errb.String())
		}

		st, err := loadRerunState()
		if err != nil {
			t.Fatal(err)
		}
		if len(st.Failed) != 1 || st.Failed[0].Scenario != "still failing" {
			t.Fatalf("ledger = %+v, want the preserved failing scenario", st.Failed)
		}
	})
}

func TestFinishRun_CIEmptySelectionExitsConfig(t *testing.T) {
	dir := t.TempDir()
	withWorkdir(t, dir, func() {
		var out, errb bytes.Buffer
		opts := &runOptions{
			label:  "atago run",
			format: report.FormatJSON,
			paths:  []string{"spec.atago.yaml"},
			filter: csvFlag{"missing"},
			ci:     true,
			stdout: &out,
			stderr: &errb,
		}
		suiteResults := []*engine.SuiteResult{{
			Suite:    "sample",
			SpecPath: "spec.atago.yaml",
			Status:   engine.StatusPassed,
		}}

		if got := finishRun(opts, suiteResults, []error{nil}, nil, 5*time.Millisecond, context.Background()); got != ExitConfig {
			t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitConfig, errb.String())
		}
		s := errb.String()
		if !strings.Contains(s, `no scenarios matched --filter "missing" under --ci`) {
			t.Fatalf("stderr = %q, want the empty-selection CI error", s)
		}
		if !strings.Contains(s, "case-sensitive substring") {
			t.Fatalf("stderr = %q, want the selector note", s)
		}
	})
}

func TestFinishRun_InterruptedWithoutResultsSkipsReport(t *testing.T) {
	var out, errb bytes.Buffer
	opts := &runOptions{
		label:  "atago run",
		format: report.FormatJSON,
		paths:  []string{"spec.atago.yaml"},
		stdout: &out,
		stderr: &errb,
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if got := finishRun(opts, []*engine.SuiteResult{nil}, []error{nil}, nil, 5*time.Millisecond, ctx); got != ExitExec {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitExec, errb.String())
	}
	if out.Len() != 0 {
		t.Fatalf("stdout = %q, want no report when every suite was skipped by the interrupt", out.String())
	}
	if !strings.Contains(errb.String(), "interrupted") {
		t.Fatalf("stderr = %q, want the interrupt diagnostic", errb.String())
	}
}

func TestFinishRun_IncompleteResultSlicesReturnInternal(t *testing.T) {
	var out, errb bytes.Buffer
	opts := &runOptions{
		label:  "atago run",
		format: report.FormatJSON,
		paths:  []string{"one.atago.yaml", "two.atago.yaml"},
		stdout: &out,
		stderr: &errb,
	}

	if got := finishRun(opts, []*engine.SuiteResult{{
		Suite:    "one",
		SpecPath: "one.atago.yaml",
		Status:   engine.StatusPassed,
	}}, []error{nil}, nil, 5*time.Millisecond, context.Background()); got != ExitInternal {
		t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitInternal, errb.String())
	}
	if out.Len() != 0 {
		t.Fatalf("stdout = %q, want no report for incomplete internal results", out.String())
	}
	if !strings.Contains(errb.String(), "internal error: incomplete run results") {
		t.Fatalf("stderr = %q, want the incomplete-results diagnostic", errb.String())
	}
}

func TestFinishRun_ReportWriteFailureReturnsInternal(t *testing.T) {
	dir := t.TempDir()
	withWorkdir(t, dir, func() {
		var errb bytes.Buffer
		opts := &runOptions{
			label:  "atago run",
			format: report.FormatJSON,
			paths:  []string{"spec.atago.yaml"},
			stdout: errWriter{},
			stderr: &errb,
		}
		suiteResults := []*engine.SuiteResult{{
			Suite:    "sample",
			SpecPath: "spec.atago.yaml",
			Status:   engine.StatusPassed,
			Scenarios: []engine.ScenarioResult{{
				Name:   "ok",
				Suite:  "sample",
				Status: engine.StatusPassed,
			}},
		}}

		if got := finishRun(opts, suiteResults, []error{nil}, nil, 5*time.Millisecond, context.Background()); got != ExitInternal {
			t.Fatalf("exit = %d, want %d (stderr=%s)", got, ExitInternal, errb.String())
		}
		if !strings.Contains(errb.String(), "failed to write report: boom") {
			t.Fatalf("stderr = %q, want the report write failure", errb.String())
		}
	})
}
