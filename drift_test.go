package main

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/docgen"
	"github.com/nao1215/atago/internal/engine"
	"github.com/nao1215/atago/internal/loader"
	"github.com/nao1215/atago/internal/sitegen"
)

// This file holds the drift guards over committed generated artifacts: the
// per-suite behavior docs under doc/e2e/ (docgen), the browsable site under
// site/ (sitegen), and the examples/ specs the README references. Each keeps a
// committed artifact in lockstep with the source it is generated from, so a
// change to a spec or a generator cannot silently rot the published output.

// readDoc reads a committed doc file relative to the repo root, failing the test
// if it is missing.
func readDoc(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// TestDocs_NoStaleLintReferences is the regression from #55: user-facing docs and
// demo assets must not invoke the removed `atago lint` command.
func TestDocs_NoStaleLintReferences(t *testing.T) {
	for _, path := range []string{"README.md", "doc/vhs/review.tape"} {
		if strings.Contains(readDoc(t, path), "atago lint") {
			t.Errorf("%s still references the removed `atago lint` command", path)
		}
	}
}

// TestDocs_FixtureSourceKeyNamed guards #157: the README prose that introduces
// fixtures must name the real inline-source key `content:` — not just the word
// "text" — so a reader who skims the prose cannot guess a non-existent `text:`
// key and hit an "unknown field" loader error. The example file uses `content:`;
// this keeps the prose from drifting away from that schema.
func TestDocs_FixtureSourceKeyNamed(t *testing.T) {
	readme := readDoc(t, "README.md")
	if !strings.Contains(readme, "`content:`") {
		t.Errorf("README does not name the literal fixture source key `content:`; a reader could guess `text:` from prose and hit an unknown-field error")
	}
}

// docSkipDirs are spec directories that deliberately ship no committed doc under
// doc/e2e/. gup-offline is the fully offline variant of the gup suite, exercised
// only by `make dogfood-gup`; it intentionally has no generated behavior doc.
var docSkipDirs = map[string]bool{
	"gup-offline": true,
}

// dirHasSpecs reports whether dir contains at least one *.atago.yaml/.yml spec
// anywhere beneath it.
func dirHasSpecs(t *testing.T, dir string) bool {
	t.Helper()
	found := false
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && (strings.HasSuffix(path, ".atago.yaml") || strings.HasSuffix(path, ".atago.yml")) {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	return found
}

// e2eDocSuites derives the committed doc/e2e/*.md → spec-directory mapping from
// the filesystem, mirroring the `make docs` recipe: the self-hosted atago suite,
// every third-party suite under test/e2e/thirdparty/*, and every dogfood suite
// under test/e2e/tools/* (minus docSkipDirs). Deriving it — rather than hand-
// maintaining a map — means a new suite is drift-guarded the moment its specs
// land. A hand-maintained map is what let doc/e2e/{age,fzf,jq}.md silently rot:
// they were generated and committed by `make docs` but never registered here, so
// nothing caught that the committed docs had fallen behind their specs.
func e2eDocSuites(t *testing.T) map[string]string {
	t.Helper()
	suites := map[string]string{
		"doc/e2e/atago.md": "test/e2e/atago",
	}
	for _, parent := range []string{"test/e2e/thirdparty", "test/e2e/tools"} {
		entries, err := os.ReadDir(parent)
		if err != nil {
			t.Fatalf("read %s: %v", parent, err)
		}
		for _, e := range entries {
			if !e.IsDir() || docSkipDirs[e.Name()] {
				continue
			}
			dir := filepath.Join(parent, e.Name())
			if !dirHasSpecs(t, dir) {
				continue
			}
			doc := "doc/e2e/" + e.Name() + ".md"
			if existing, ok := suites[doc]; ok {
				t.Fatalf("doc name collision: %s maps to both %s and %s", doc, existing, filepath.ToSlash(dir))
			}
			suites[doc] = filepath.ToSlash(dir)
		}
	}
	return suites
}

var (
	// makeDocOutRe extracts each doc/e2e/<name>.md target from the Makefile
	// `docs` recipe's `atago doc --out` lines.
	makeDocOutRe = regexp.MustCompile(`doc/e2e/([a-zA-Z0-9_-]+)\.md`)
	// thirdpartyDirRe extracts each suite directory the thirdparty.yml CI matrix
	// runs (`dir: ./test/e2e/thirdparty/<name>`).
	thirdpartyDirRe = regexp.MustCompile(`dir:\s*\./test/e2e/thirdparty/([a-zA-Z0-9_-]+)`)
	// windowsSpecRe extracts each ./test/e2e spec path listed in the single-source
	// scripts/windows_portable_specs.sh.
	windowsSpecRe = regexp.MustCompile(`(\./test/e2e/\S+)`)
)

// collectSpecs mirrors cli.collectSpecFiles for a single directory target: every
// *.atago.yaml/.yml under dir in filepath.WalkDir's lexical order, cleaned. The
// cleaned path is what docgen prints as each scenario's `Source:`, so it must
// match the `atago doc` CLI byte-for-byte.
func collectSpecs(t *testing.T, dir string) []string {
	t.Helper()
	var specs []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".atago.yaml") || strings.HasSuffix(path, ".atago.yml") {
			specs = append(specs, filepath.Clean(path))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	return specs
}

// firstDiff returns a short, reviewable description of the first line that
// differs between want and got, with the surrounding bytes quoted (so invisible
// differences like CR/LF or trailing whitespace are visible). It makes a golden
// drift failure diagnosable — especially cross-platform ones seen only in CI.
func firstDiff(want, got []byte) string {
	wl := strings.Split(string(want), "\n")
	gl := strings.Split(string(got), "\n")
	n := min(len(wl), len(gl))
	for i := 0; i < n; i++ {
		if wl[i] != gl[i] {
			return fmt.Sprintf("first difference at line %d:\n  want: %q\n  got:  %q", i+1, wl[i], gl[i])
		}
	}
	if len(wl) != len(gl) {
		return fmt.Sprintf("files differ in length: want %d lines, got %d lines", len(wl), len(gl))
	}
	return "(differences are not line-based; check byte content)"
}

// TestDocs_E2EDocsInSync regenerates every doc/e2e/*.md from its spec directory
// and fails if the committed file drifts from docgen's output. This is the guard
// that a change to a spec — or to the doc generator itself — is reflected in the
// committed behavior documentation, so the generated docs cannot silently rot.
// Regenerate a stale file with:
//
//	atago doc --out doc/e2e/<tool>.md ./test/e2e/tools/<tool>
func TestDocs_E2EDocsInSync(t *testing.T) {
	for docPath, specDir := range e2eDocSuites(t) {
		t.Run(docPath, func(t *testing.T) {
			specs := collectSpecs(t, specDir)
			if len(specs) == 0 {
				t.Fatalf("no *.atago.yaml specs found under %s", specDir)
			}

			sources := make([]docgen.Source, 0, len(specs))
			for _, p := range specs {
				s, err := loader.Load(p)
				if err != nil {
					t.Fatalf("load %s: %v", p, err)
				}
				sources = append(sources, docgen.Source{Path: p, Spec: s})
			}

			// Match the CLI's `--out <docPath>` invocation so embedded golden-image
			// links resolve relative to the doc's directory (#67).
			var buf bytes.Buffer
			if err := docgen.GenerateTo(&buf, sources, filepath.Dir(docPath)); err != nil {
				t.Fatalf("generate %s: %v", docPath, err)
			}

			want, err := os.ReadFile(docPath)
			if err != nil {
				t.Fatalf("read %s: %v (regenerate with `atago doc --out %s ./%s`)", docPath, err, docPath, specDir)
			}
			if !bytes.Equal(buf.Bytes(), want) {
				t.Errorf("%s is out of date with %s; regenerate with `atago doc --out %s ./%s`\n%s",
					docPath, specDir, docPath, specDir, firstDiff(want, buf.Bytes()))
			}
		})
	}
}

// TestDocs_E2EDocSetMatchesCommitted asserts the derived suite set is exactly
// the committed doc/e2e/*.md set (minus the README index), so a doc whose spec
// directory was deleted — or a spec directory whose generated doc was never
// committed — fails the build instead of rotting unnoticed.
func TestDocs_E2EDocSetMatchesCommitted(t *testing.T) {
	suites := e2eDocSuites(t)
	committed := map[string]bool{}
	entries, err := os.ReadDir("doc/e2e")
	if err != nil {
		t.Fatalf("read doc/e2e: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || e.Name() == "README.md" {
			continue
		}
		committed["doc/e2e/"+e.Name()] = true
	}
	for doc := range suites {
		if !committed[doc] {
			t.Errorf("%s derives from a spec directory but is not committed; run `make docs`", doc)
		}
	}
	for doc := range committed {
		if _, ok := suites[doc]; !ok {
			t.Errorf("%s is committed but derives from no spec directory (stale doc, or a suite missing from docSkipDirs)", doc)
		}
	}
}

// TestDocs_MakefileGeneratesEverySuite asserts the `make docs` recipe emits a
// doc for exactly the derived suite set, so adding a suite directory forces a
// matching `atago doc --out` line and nothing generates an orphan doc. This is
// what keeps the one remaining hand-maintained doc list — the Makefile recipe —
// from drifting away from the specs on disk.
func TestDocs_MakefileGeneratesEverySuite(t *testing.T) {
	data, err := os.ReadFile("Makefile")
	if err != nil {
		t.Fatalf("read Makefile: %v", err)
	}
	generated := map[string]bool{}
	for _, m := range makeDocOutRe.FindAllSubmatch(data, -1) {
		generated["doc/e2e/"+string(m[1])+".md"] = true
	}
	suites := e2eDocSuites(t)
	for doc := range suites {
		if !generated[doc] {
			t.Errorf("Makefile `docs` target does not generate %s; add an `atago doc --out %s ...` line", doc, doc)
		}
	}
	for doc := range generated {
		if _, ok := suites[doc]; !ok {
			t.Errorf("Makefile `docs` generates %s but no spec directory derives it", doc)
		}
	}
}

// TestThirdParty_MatrixCoversEverySuite asserts the scheduled thirdparty.yml
// matrix runs every test/e2e/thirdparty/* suite, so a suite with specs but no
// matrix leg (which would silently never run in CI) fails the build. git is
// exempt: it is push-gated in e2e.yml rather than in the scheduled matrix.
func TestThirdParty_MatrixCoversEverySuite(t *testing.T) {
	data, err := os.ReadFile(".github/workflows/thirdparty.yml")
	if err != nil {
		t.Fatalf("read thirdparty.yml: %v", err)
	}
	inMatrix := map[string]bool{}
	for _, m := range thirdpartyDirRe.FindAllSubmatch(data, -1) {
		inMatrix[string(m[1])] = true
	}
	entries, err := os.ReadDir("test/e2e/thirdparty")
	if err != nil {
		t.Fatalf("read test/e2e/thirdparty: %v", err)
	}
	exempt := map[string]bool{"git": true}
	for _, e := range entries {
		if !e.IsDir() || exempt[e.Name()] {
			continue
		}
		if !dirHasSpecs(t, filepath.Join("test/e2e/thirdparty", e.Name())) {
			continue
		}
		if !inMatrix[e.Name()] {
			t.Errorf("test/e2e/thirdparty/%s has specs but no matrix leg in .github/workflows/thirdparty.yml", e.Name())
		}
	}
}

// TestWindowsPortableSubset_Exists asserts every spec path in the single-source
// scripts/windows_portable_specs.sh resolves to a real file or directory, so a
// spec rename fails here — a fast unit test — instead of only in the Windows CI
// legs that read the script (e2e.yml and e2e-cross.yml).
func TestWindowsPortableSubset_Exists(t *testing.T) {
	data, err := os.ReadFile("scripts/windows_portable_specs.sh")
	if err != nil {
		t.Fatalf("read scripts/windows_portable_specs.sh: %v", err)
	}
	var paths []string
	for _, m := range windowsSpecRe.FindAllSubmatch(data, -1) {
		paths = append(paths, string(m[1]))
	}
	if len(paths) == 0 {
		t.Fatal("no ./test/e2e spec paths found in scripts/windows_portable_specs.sh")
	}
	for _, p := range paths {
		if _, err := os.Stat(filepath.FromSlash(p)); err != nil {
			t.Errorf("scripts/windows_portable_specs.sh lists %q which does not resolve: %v", p, err)
		}
	}
}

// TestSite_InSync regenerates the browsable docs site (#72) from repository
// sources and fails if any committed file under site/ drifts from the generator
// output. This is the CI-verified site-generation smoke: it runs in the normal
// `go test ./...`, so no separate workflow is needed. Regenerate a stale site
// with `make site` (or `UPDATE_SITE=1 go test -run TestSite_InSync .`).
func TestSite_InSync(t *testing.T) {
	files, err := sitegen.Files(".")
	if err != nil {
		t.Fatalf("generate site: %v", err)
	}

	if os.Getenv("UPDATE_SITE") == "1" {
		if err := sitegen.Generate("."); err != nil {
			t.Fatalf("write site: %v", err)
		}
		return
	}

	for name, want := range files {
		got, err := os.ReadFile(name)
		if err != nil {
			t.Errorf("missing generated site file %s: %v (run `make site`)", name, err)
			continue
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s is out of date with the generator; regenerate with `make site`", name)
		}
	}
}

var siteLinkRe = regexp.MustCompile(`\]\(([^)]+)\)`)

// TestSite_AllReferencedAssetsExist parses the generated site index and asserts
// every relative link/image target resolves to a real file, so the site never
// points at a missing asset (#72).
func TestSite_AllReferencedAssetsExist(t *testing.T) {
	data, err := os.ReadFile("site/README.md")
	if err != nil {
		t.Fatalf("read site/README.md: %v (run `make site`)", err)
	}
	for _, m := range siteLinkRe.FindAllSubmatch(data, -1) {
		target := string(m[1])
		if isExternal(target) {
			continue
		}
		// Links are relative to the site/ directory.
		resolved := filepath.Clean(filepath.Join("site", target))
		if _, err := os.Stat(resolved); err != nil {
			t.Errorf("site/README.md links to %q which does not resolve to a file (%v)", target, err)
		}
	}
}

func isExternal(target string) bool {
	return len(target) > 0 && (target[0] == '#' ||
		strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") || strings.HasPrefix(target, "mailto:"))
}

// exampleSpecs categorizes every spec under examples/: hermetic examples run
// green with no external dependency and are executed here on every OS;
// non-hermetic ones (a live API, an SSH host, a gRPC server, a browser) are
// loaded and validated only. The README links to these files as the syntax
// reference, so this test is what keeps them from drifting away from the
// implementation.
var exampleSpecs = map[string]bool{ // path -> hermetic (run, not just validate)
	"examples/browser.atago.yaml":             false,
	"examples/changes.atago.yaml":             true,
	"examples/db.atago.yaml":                  true,
	"examples/defaults.atago.yaml":            true,
	"examples/dir_tree.atago.yaml":            true,
	"examples/duration.atago.yaml":            true,
	"examples/extend_host_env.atago.yaml":     true,
	"examples/files_and_fixtures.atago.yaml":  true,
	"examples/grpc.atago.yaml":                false,
	"examples/hermetic_env.atago.yaml":        true,
	"examples/http.atago.yaml":                false,
	"examples/image_and_pdf.atago.yaml":       true,
	"examples/json_and_yaml.atago.yaml":       true,
	"examples/matrix.atago.yaml":              true,
	"examples/mock_server.atago.yaml":         true,
	"examples/pty.atago.yaml":                 true,
	"examples/pty_portable.atago.yaml":        true,
	"examples/pty_screen.atago.yaml":          true,
	"examples/retry.atago.yaml":               true,
	"examples/run_and_assert.atago.yaml":      true,
	"examples/scrub.atago.yaml":               true,
	"examples/select_skip_only.atago.yaml":    true,
	"examples/services.atago.yaml":            true,
	"examples/shell_and_redirect.atago.yaml":  true,
	"examples/signal.atago.yaml":              true,
	"examples/snapshot.atago.yaml":            true,
	"examples/ssh.atago.yaml":                 false,
	"examples/stdin.atago.yaml":               true,
	"examples/store_and_variables.atago.yaml": true,
	"examples/suite_setup.atago.yaml":         false,
	"examples/teardown.atago.yaml":            true,
	"examples/timeouts.atago.yaml":            true,
}

// TestExamples_EveryFileCategorized fails when a spec is added to examples/
// without being registered above, so a new example cannot ship untested.
func TestExamples_EveryFileCategorized(t *testing.T) {
	t.Parallel()
	found := map[string]bool{}
	err := filepath.WalkDir("examples", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".atago.yaml") {
			found[filepath.ToSlash(path)] = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk examples: %v", err)
	}
	for path := range found {
		if _, ok := exampleSpecs[path]; !ok {
			t.Errorf("%s is not categorized in exampleSpecs; add it (hermetic or validate-only)", path)
		}
	}
	for path := range exampleSpecs {
		if !found[path] {
			t.Errorf("%s is categorized but does not exist", path)
		}
	}
}

// TestExamples_Valid loads and validates every example, hermetic or not.
func TestExamples_Valid(t *testing.T) {
	t.Parallel()
	for path := range exampleSpecs {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			if _, err := loader.Load(path); err != nil {
				t.Errorf("example does not load/validate: %v", err)
			}
		})
	}
}

// TestExamples_HermeticRunGreen executes every hermetic example through the
// real engine. OS-gated scenarios (skip/only) may be skipped, but nothing may
// fail or error: an example the README points at must actually work.
func TestExamples_HermeticRunGreen(t *testing.T) {
	t.Parallel()
	for path, hermetic := range exampleSpecs {
		if !hermetic {
			continue
		}
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			s, err := loader.Load(path)
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			res := engine.New().Run(context.Background(), s, path)
			if res.Status != engine.StatusPassed && res.Status != engine.StatusSkipped {
				t.Errorf("status = %s, want passed (or skipped by an OS gate): %+v", res.Status, res.Scenarios)
			}
		})
	}
}
