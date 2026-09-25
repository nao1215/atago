package main

import (
	"bytes"
	"math"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/buildinfo"
	"github.com/nao1215/atago/internal/cli"
	"github.com/nao1215/atago/internal/loader"
)

// TestMainSmoke checks that the wiring from main → cli works for `version`.
// Thorough CLI behavior is covered by internal/cli tests and the self-hosted
// E2E specs under test/e2e/atago.
func TestMainSmoke(t *testing.T) {
	previous := buildinfo.Version
	buildinfo.Version = "test-version"
	t.Cleanup(func() { buildinfo.Version = previous })

	var stdout, stderr bytes.Buffer
	if got := cli.Main([]string{"version"}, &stdout, &stderr); got != cli.ExitOK {
		t.Fatalf("cli.Main(version) = %d, want %d", got, cli.ExitOK)
	}
	if got, want := stdout.String(), "atago test-version\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestMainUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if got := cli.Main([]string{"frobnicate"}, &stdout, &stderr); got != cli.ExitConfig {
		t.Fatalf("cli.Main(frobnicate) = %d, want %d", got, cli.ExitConfig)
	}
	if !strings.Contains(stderr.String(), "unknown command") {
		t.Fatalf("stderr = %q, want unknown command", stderr.String())
	}
}

// TestDogfood_SpecsLoad loads every committed spec and asserts it passes the
// loader's schema and semantic validation. This is the self-hosted acceptance
// check that the repo's own specs stay valid before a release rather than after
// users hit them. It reads the same corpus the schema conformance guard does,
// so the loader and the published schema are always asked about the same set of
// files — a spec that only one of them sees is a spec nobody guards.
func TestDogfood_SpecsLoad(t *testing.T) {
	for _, p := range shippedSpecPaths(t) {
		if _, err := loader.Load(p); err != nil {
			t.Errorf("%s: failed to load: %v", p, err)
		}
	}
}

// TestTuneGC pins that atago's collector settings apply only where the user
// left the variable unset.
func TestTuneGC(t *testing.T) {
	prevPercent := debug.SetGCPercent(100)
	prevLimit := debug.SetMemoryLimit(math.MaxInt64)
	t.Cleanup(func() {
		debug.SetGCPercent(prevPercent)
		debug.SetMemoryLimit(prevLimit)
	})

	tuneGC(func(string) string { return "" })
	if got := debug.SetGCPercent(100); got != gcPercent {
		t.Errorf("GC percent = %d, want %d", got, gcPercent)
	}
	if got := debug.SetMemoryLimit(math.MaxInt64); got != memoryLimit {
		t.Errorf("memory limit = %d, want %d", got, memoryLimit)
	}

	tuneGC(func(k string) string { return map[string]string{"GOGC": "50", "GOMEMLIMIT": "1GiB"}[k] })
	if got := debug.SetGCPercent(100); got != 100 {
		t.Errorf("GC percent with GOGC set = %d, want it left alone", got)
	}
	if got := debug.SetMemoryLimit(math.MaxInt64); got != math.MaxInt64 {
		t.Errorf("memory limit with GOMEMLIMIT set = %d, want it left alone", got)
	}
}
