package main

import (
	"bytes"
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
