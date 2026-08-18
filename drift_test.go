package main

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/cli"
	"github.com/nao1215/atago/internal/diag"
	"github.com/nao1215/atago/internal/docgen"
	"github.com/nao1215/atago/internal/engine"
	"github.com/nao1215/atago/internal/loader"
	"github.com/nao1215/atago/internal/sitegen"
	"github.com/nao1215/atago/internal/spec"
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

// referenceSubcommandRowRe extracts the subcommand name (the first word after
// `atago `) from each Subcommands-table row in website/content/reference.md,
// e.g. "run" from "| `atago run` | run specs and report results |" and
// "snapshot" from "| `atago snapshot update` | ... |".
var referenceSubcommandRowRe = regexp.MustCompile("(?m)^\\| `atago ([a-z][a-z-]*)")

// TestDocs_ReferenceSubcommandsAreReal guards the website Reference page's
// Subcommands table against advertising a command that does not exist. The table
// once listed `atago rerun` — a subcommand that never existed: the real feature
// is the `--rerun-failed` flag on `atago run`, and `atago rerun` exits 3
// (unknown command). This checks every `atago <name>` table row against the real
// dispatch inventory that internal/cli exports, so a fictional or renamed
// subcommand fails the build instead of quietly misleading readers.
func TestDocs_ReferenceSubcommandsAreReal(t *testing.T) {
	ref := readDoc(t, "website/content/reference.md")
	known := map[string]bool{}
	for _, name := range cli.Subcommands() {
		known[name] = true
	}
	rows := referenceSubcommandRowRe.FindAllStringSubmatch(ref, -1)
	if len(rows) == 0 {
		t.Fatal("no `| `atago <name>`` Subcommands rows found in website/content/reference.md")
	}
	for _, m := range rows {
		if !known[m[1]] {
			t.Errorf("website/content/reference.md Subcommands table lists `atago %s`, which is not a known subcommand (known inventory: %v)", m[1], cli.Subcommands())
		}
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
	// thirdpartyRunsOnRe extracts the thirdparty workflow's Linux runner label.
	thirdpartyRunsOnRe = regexp.MustCompile(`(?m)^\s*runs-on:\s*(\S+)\s*$`)
)

func thirdpartyMatrixBlock(t *testing.T, yml, name string) string {
	t.Helper()
	marker := "          - name: " + name + "\n"
	start := strings.Index(yml, marker)
	if start < 0 {
		t.Fatalf("could not find thirdparty matrix block for %q", name)
	}
	rest := yml[start+len(marker):]
	next := strings.Index(rest, "\n          - name: ")
	if next < 0 {
		return yml[start:]
	}
	return yml[start : start+len(marker)+next]
}

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

// TestDocs_RealWorldIndexCoversEverySuite asserts the authoritative index
// doc/real-world.md links the generated doc of every suite that has one. The
// hand-maintained list in doc/e2e/README.md rotted precisely because nothing
// guarded it (19 of 35 third-party suites went unlisted); real-world.md now
// carries the complete table, so guard it against the same rot.
func TestDocs_RealWorldIndexCoversEverySuite(t *testing.T) {
	data, err := os.ReadFile("doc/real-world.md")
	if err != nil {
		t.Fatalf("read doc/real-world.md: %v", err)
	}
	index := string(data)
	for doc := range e2eDocSuites(t) {
		// The index links docs relative to doc/, so doc/e2e/git.md is "e2e/git.md".
		link := strings.TrimPrefix(doc, "doc/")
		if !strings.Contains(index, link) {
			t.Errorf("doc/real-world.md does not link %s; every suite with a generated doc must be indexed", link)
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

// TestThirdParty_InstallersArePinned guards the scheduled thirdparty matrix
// against reintroducing floating external-tool installs. The whole point of the
// matrix is to exercise fixed third-party contracts; `@latest`, a moving
// ubuntu-latest base image, or ad-hoc apt installs from today's repository all
// reintroduce drift.
func TestThirdParty_InstallersArePinned(t *testing.T) {
	data, err := os.ReadFile(".github/workflows/thirdparty.yml")
	if err != nil {
		t.Fatalf("read thirdparty.yml: %v", err)
	}
	yml := string(data)

	m := thirdpartyRunsOnRe.FindStringSubmatch(yml)
	if len(m) != 2 {
		t.Fatal("could not find runs-on in .github/workflows/thirdparty.yml")
	}
	if got := m[1]; got != "ubuntu-24.04" {
		t.Errorf("thirdparty.yml runs-on = %q; want ubuntu-24.04 so the pinned Ubuntu snapshot stays on a fixed distro", got)
	}
	if strings.Contains(yml, "@latest") {
		t.Error("thirdparty.yml still contains @latest; every third-party tool install must pin an explicit version or release tag")
	}
	if strings.Contains(yml, "sudo apt-get install -y") || strings.Contains(yml, "sudo apt install -y") {
		t.Error("thirdparty.yml still shells out to a raw apt install; Ubuntu-packaged tools must come through scripts/install_ubuntu_snapshot_packages.sh")
	}
	if strings.Contains(yml, "command -v ") {
		t.Error("thirdparty.yml still has command -v install guards; the third-party matrix must install fixed tool versions even if the runner image already ships one")
	}

	for _, leg := range []string{"awscli", "ecspresso"} {
		block := thirdpartyMatrixBlock(t, yml, leg)
		if !strings.Contains(block, "bash ./scripts/install_awscli_pinned.sh") {
			t.Errorf("thirdparty.yml %s block does not install the pinned AWS CLI bundle", leg)
		}
		if strings.Contains(block, "install_ubuntu_snapshot_packages.sh") {
			t.Errorf("thirdparty.yml %s block still installs awscli from the Ubuntu snapshot", leg)
		}
	}
}

// windowsSpecsTSV is the single source of truth for which self-hosted E2E
// targets run on Windows and under which shell. Both Windows CI legs (e2e.yml
// and e2e-cross.yml) expand it through scripts/windows_specs.sh.
const windowsSpecsTSV = "scripts/windows_specs.tsv"

// windowsBuckets are the classifications a row may carry. `none` is the only
// one that must be justified, because it is the only one that means a target
// goes untested on a supported OS.
var windowsBuckets = map[string]bool{"cmd": true, "bash": true, "none": true}

// windowsSpecRow is one classified target.
type windowsSpecRow struct {
	target string
	bucket string
	reason string
	line   int
}

// readWindowsSpecs parses the classification table, failing on any row that is
// not exactly target/bucket[/reason].
func readWindowsSpecs(t *testing.T) []windowsSpecRow {
	t.Helper()
	data, err := os.ReadFile(windowsSpecsTSV)
	if err != nil {
		t.Fatalf("read %s: %v", windowsSpecsTSV, err)
	}
	var rows []windowsSpecRow
	for i, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 2 || len(fields) > 3 {
			t.Fatalf("%s:%d: want target<TAB>bucket[<TAB>reason], got %q", windowsSpecsTSV, i+1, line)
		}
		row := windowsSpecRow{target: fields[0], bucket: fields[1], line: i + 1}
		if len(fields) == 3 {
			row.reason = fields[2]
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		t.Fatalf("%s classifies nothing", windowsSpecsTSV)
	}
	return rows
}

// TestWindowsSpecs_EveryTargetClassified is what keeps the Windows legs from
// being a subset nobody decided on. Every spec under test/e2e/atago must appear
// in the table exactly once, so a spec added without a thought about Windows
// fails a fast unit test rather than silently going untested on a supported OS
// for as long as nobody looks — which is how the subset came to hold under half
// of them.
func TestWindowsSpecs_EveryTargetClassified(t *testing.T) {
	t.Parallel()
	classified := map[string]int{}
	for _, row := range readWindowsSpecs(t) {
		classified[row.target]++
	}
	found := map[string]bool{}
	entries, err := os.ReadDir(filepath.Join("test", "e2e", "atago"))
	if err != nil {
		t.Fatalf("read the self-hosted spec directory: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".atago.yaml") {
			continue
		}
		found["./test/e2e/atago/"+e.Name()] = true
	}
	if len(found) == 0 {
		t.Fatal("no self-hosted specs found; the directory layout moved")
	}
	for target := range found {
		switch classified[target] {
		case 1:
		case 0:
			t.Errorf("%s is not classified in %s; add it as cmd, bash, or none with a reason", target, windowsSpecsTSV)
		default:
			t.Errorf("%s is classified %d times in %s", target, classified[target], windowsSpecsTSV)
		}
	}
}

// TestWindowsSpecs_RowsAreWellFormed checks the three properties a row can get
// wrong: an unknown bucket, a target that does not resolve, and an unjustified
// `none`. The last is the point of the file — a target excluded from Windows
// has to state a reason about the platform, not leave the exclusion implicit.
func TestWindowsSpecs_RowsAreWellFormed(t *testing.T) {
	t.Parallel()
	for _, row := range readWindowsSpecs(t) {
		if !windowsBuckets[row.bucket] {
			t.Errorf("%s:%d: unknown bucket %q (want cmd, bash, or none)", windowsSpecsTSV, row.line, row.bucket)
		}
		if _, err := os.Stat(filepath.FromSlash(row.target)); err != nil {
			t.Errorf("%s:%d: %q does not resolve: %v", windowsSpecsTSV, row.line, row.target, err)
		}
		switch row.bucket {
		case "none":
			if strings.TrimSpace(row.reason) == "" {
				t.Errorf("%s:%d: %q is excluded from Windows with no reason", windowsSpecsTSV, row.line, row.target)
			}
		case "cmd", "bash":
			if strings.TrimSpace(row.reason) != "" {
				t.Errorf("%s:%d: %q runs on Windows, so the reason column must be empty", windowsSpecsTSV, row.line, row.target)
			}
		}
	}
}

// TestWindowsSpecs_WorkflowsReadTheTable pins that both Windows CI legs expand
// the classification rather than carrying a list of their own. The file only
// prevents drift as long as it is what actually runs.
func TestWindowsSpecs_WorkflowsReadTheTable(t *testing.T) {
	t.Parallel()
	for _, wf := range []string{".github/workflows/e2e.yml", ".github/workflows/e2e-cross.yml"} {
		yml := readDoc(t, wf)
		for _, bucket := range []string{"cmd", "bash"} {
			if !strings.Contains(yml, "windows_specs.sh "+bucket) {
				t.Errorf("%s does not run the %q bucket via scripts/windows_specs.sh", wf, bucket)
			}
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

// TestDocs_ErrorReferenceInSync keeps the published error reference in lockstep
// with the diagnostic registry that produces the codes. The registry is the
// only place a code can be created, and this is what makes its documentation
// arrive with it: a code added, reworded, or retired without regenerating the
// page fails here. Regenerate with `make docs` (or `UPDATE_ERRORS=1 go test
// -run TestDocs_ErrorReferenceInSync .`).
func TestDocs_ErrorReferenceInSync(t *testing.T) {
	const path = "doc/errors.md"
	want := diag.Markdown()

	if os.Getenv("UPDATE_ERRORS") == "1" {
		if err := os.WriteFile(path, want, 0o600); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
		return
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v (run `make docs`)", path, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("%s is out of date with the diagnostic registry; regenerate with `make docs`", path)
	}
}

// TestDocs_EveryCodeIsDocumented is the reader's side of the same guarantee:
// whatever the generator does, the committed page has to name every code and
// say what to do about it.
func TestDocs_EveryCodeIsDocumented(t *testing.T) {
	page := readDoc(t, "doc/errors.md")
	for _, e := range diag.All() {
		if !strings.Contains(page, e.Code.String()) {
			t.Errorf("doc/errors.md does not mention %s (%s)", e.Code, e.Name)
		}
		if !strings.Contains(page, e.Fix) {
			t.Errorf("doc/errors.md does not tell the reader how to fix %s (%s)", e.Code, e.Name)
		}
	}
}

// TestE2E_EveryCodeIsProvoked is the coverage gate over the diagnostic
// registry: a published code must be produced by the real binary in a real
// run, not merely written down. Registering a code therefore obliges you to
// add the scenario that provokes it, in the same change.
//
// Documentation alone would let a code rot in place after the check that
// raised it was deleted or reworded past the point of reaching it. The
// scenarios live in test/e2e/atago/error_codes.atago.yaml; any spec in the
// directory counts, so a code that already has a home elsewhere needs no
// second one.
func TestE2E_EveryCodeIsProvoked(t *testing.T) {
	const dir = "test/e2e/atago"
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	provoked := map[diag.Code]bool{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") && !strings.HasSuffix(e.Name(), ".yml") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		s, lerr := loader.Load(path)
		if lerr != nil {
			t.Fatalf("load %s: %v", path, lerr)
		}
		// Only what a scenario ASSERTS counts. Scanning the file as text would
		// let a scenario title or a comment mentioning a code satisfy the gate,
		// which is the opposite of what it is for.
		for _, text := range assertedText(s) {
			for _, c := range diag.Codes(text) {
				provoked[c] = true
			}
		}
	}
	for _, e := range diag.All() {
		if provoked[e.Code] {
			if _, exempt := unprovokableFromASpec[e.Code]; exempt {
				t.Errorf("%s (%s) is listed as unprovokable from a spec, but a scenario in %s provokes it; remove the exemption", e.Code, e.Name, dir)
			}
			continue
		}
		if _, exempt := unprovokableFromASpec[e.Code]; exempt {
			continue
		}
		t.Errorf("%s (%s) is a published code that no scenario in %s provokes; add one to error_codes.atago.yaml", e.Code, e.Name, dir)
	}
}

// unprovokableFromASpec lists the codes no spec can reach, each with the reason.
// Everything else must be provoked by a scenario, and the gate above also fails
// when a code listed here turns out to be reachable after all — so the list
// shrinks as the ways to provoke them appear, and never quietly grows stale.
//
// This is the one place the coverage gate can be escaped, which is why it is a
// table of prose rather than a flag: adding a code here is a claim a reviewer
// can disagree with.
var unprovokableFromASpec = map[diag.Code]string{
	diag.RunInterrupted:        "needs a signal delivered to atago itself mid-run, which a scenario cannot arrange for its own runner",
	diag.MockServerFailed:      "needs the mock server's port to be taken, and a spec cannot pin the port it binds",
	diag.PTYFailed:             "needs the kernel to refuse a pseudo-terminal; every shape a spec can ask for is rejected by the loader first",
	diag.InputNotSupported:     "the key and signal vocabularies are validated while loading, so a spec never reaches the runtime check",
	diag.UnsupportedOnPlatform: "only raised on Windows, where the scenarios that would provoke it are the ones being refused",
	diag.BrowserActionFailed:   "needs a real browser, which the hermetic self-hosted suite deliberately does not start",
	diag.PayloadFailed:         "needs a value that will not marshal, which YAML cannot express",
	diag.ResponseUnreadable:    "needs a peer that answers with a malformed body, which the mock server will not produce",
	diag.InternalError:         "is unreachable by construction: every state it guards is refused while loading",
}

// assertedText collects the strings a spec's stream assertions compare
// against — the only place a diagnostic code can be claimed rather than merely
// mentioned.
func assertedText(s *spec.Spec) []string {
	var out []string
	collect := func(sa *spec.StreamAssert) {
		if sa == nil {
			return
		}
		out = append(out, []string(sa.Contains)...)
		for _, p := range []*string{sa.Matches, sa.Equals} {
			if p != nil {
				out = append(out, *p)
			}
		}
	}
	for i := range s.Scenarios {
		steps := append(append([]spec.Step{}, s.Scenarios[i].Steps...), s.Scenarios[i].Teardown...)
		for j := range steps {
			if a := steps[j].Assert; a != nil {
				collect(a.Stdout)
				collect(a.Stderr)
			}
		}
	}
	return out
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
	"examples/browser.atago.yaml":              false,
	"examples/changes.atago.yaml":              true,
	"examples/count_and_size.atago.yaml":       true,
	"examples/db.atago.yaml":                   true,
	"examples/defaults.atago.yaml":             true,
	"examples/deterministic.atago.yaml":        true,
	"examples/dir_tree.atago.yaml":             true,
	"examples/duration.atago.yaml":             true,
	"examples/extend_host_env.atago.yaml":      true,
	"examples/files_and_fixtures.atago.yaml":   true,
	"examples/grpc.atago.yaml":                 false,
	"examples/hermetic_env.atago.yaml":         true,
	"examples/http.atago.yaml":                 false,
	"examples/image_and_pdf.atago.yaml":        true,
	"examples/json_and_yaml.atago.yaml":        true,
	"examples/matrix.atago.yaml":               true,
	"examples/mock_server.atago.yaml":          true,
	"examples/pty.atago.yaml":                  true,
	"examples/pty_portable.atago.yaml":         true,
	"examples/pty_screen.atago.yaml":           true,
	"examples/pty_stdout_split.atago.yaml":     true,
	"examples/retry.atago.yaml":                true,
	"examples/run_and_assert.atago.yaml":       true,
	"examples/scrub.atago.yaml":                true,
	"examples/select_skip_only.atago.yaml":     true,
	"examples/services.atago.yaml":             true,
	"examples/shell_and_redirect.atago.yaml":   true,
	"examples/signal.atago.yaml":               true,
	"examples/snapshot.atago.yaml":             true,
	"examples/ssh.atago.yaml":                  false,
	"examples/stdin.atago.yaml":                true,
	"examples/store_and_variables.atago.yaml":  true,
	"examples/suite_env_from_setup.atago.yaml": true,
	"examples/suite_setup.atago.yaml":          false,
	"examples/teardown.atago.yaml":             true,
	"examples/timeouts.atago.yaml":             true,
	"examples/project_manifest.atago.yaml":     true,
	"examples/expect_fail.atago.yaml":          true,
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

// TestCookbook_SnippetsValid loads and validates every fenced YAML block in
// doc/cookbook.md through the real loader, so a recipe cannot drift from the
// schema it demonstrates. The recipes run placeholder commands (`mytool`), so
// they validate but do not execute.
func TestCookbook_SnippetsValid(t *testing.T) {
	t.Parallel()
	doc := readDoc(t, "doc/cookbook.md")
	// The prose preceding a block decides which loader validates it: the
	// cookbook shows two kinds of YAML, and a directory manifest is not a spec
	// (#392). Keying on the introducing text rather than on a special fence
	// keeps both kinds guarded — a manifest snippet with a typo would otherwise
	// have to be excluded from the drift test entirely to stop it failing as a
	// spec, which is how an unvalidated recipe gets in.
	blocks := regexp.MustCompile("(?s)```yaml\n(.*?)```").FindAllStringSubmatchIndex(doc, -1)
	if len(blocks) == 0 {
		t.Fatal("doc/cookbook.md contains no ```yaml blocks")
	}
	dir := t.TempDir()
	for i, loc := range blocks {
		body := doc[loc[2]:loc[3]]
		if isManifestSnippet(doc, loc[0]) {
			// A recipe names a fixtures_dir the reader will create; materialize
			// it so the snippet is validated for SHAPE, which is what a doc
			// drift guard is for, rather than failing on a directory that only
			// exists in the reader's repository.
			if err := os.MkdirAll(filepath.Join(dir, "testdata"), 0o750); err != nil {
				t.Fatalf("prepare snippet fixtures dir: %v", err)
			}
			path := filepath.Join(dir, fmt.Sprintf("snippet_%02d.project.yaml", i))
			if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
				t.Fatalf("write snippet %d: %v", i, err)
			}
			if _, err := loader.LoadProject(path); err != nil {
				t.Errorf("doc/cookbook.md manifest snippet %d does not load/validate: %v", i, err)
			}
			continue
		}
		path := filepath.Join(dir, fmt.Sprintf("snippet_%02d.atago.yaml", i))
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatalf("write snippet %d: %v", i, err)
		}
		if _, err := loader.Load(path); err != nil {
			t.Errorf("doc/cookbook.md snippet %d does not load/validate: %v", i, err)
		}
	}
}

// isManifestSnippet reports whether the fenced block starting at fenceStart is
// introduced as a directory manifest rather than a spec. It looks at the prose
// immediately above the fence, which is where the cookbook names the file.
func isManifestSnippet(doc string, fenceStart int) bool {
	const lookback = 200
	from := max(fenceStart-lookback, 0)
	return strings.Contains(doc[from:fenceStart], loader.ProjectFileName)
}

// TestExamplesIndex_InLockstep keeps doc/examples.md honest in both
// directions: its by-feature table must link every spec under examples/ (the
// table moved out of the README, where nothing guarded it), every
// cookbook.md#anchor in its by-task table must resolve to a real
// doc/cookbook.md heading, and every cookbook heading must be indexed. The
// anchors come from docgen's slugger — the one authority on how a heading
// becomes an anchor — so the guard cannot drift from the generated docs.
func TestExamplesIndex_InLockstep(t *testing.T) {
	t.Parallel()
	index := readDoc(t, "doc/examples.md")
	for path := range exampleSpecs {
		if !strings.Contains(index, "(../"+path+")") {
			t.Errorf("doc/examples.md by-feature table does not link %s", path)
		}
	}
	var headings []string
	for _, m := range regexp.MustCompile(`(?m)^## (.+)$`).FindAllStringSubmatch(readDoc(t, "doc/cookbook.md"), -1) {
		headings = append(headings, m[1])
	}
	anchors := map[string]bool{}
	for i, anchor := range docgen.Anchors(headings) {
		anchors[anchor] = true
		if !strings.Contains(index, "(cookbook.md#"+anchor+")") {
			t.Errorf("doc/cookbook.md heading %q is not indexed in doc/examples.md (expected a cookbook.md#%s link)", headings[i], anchor)
		}
	}
	for _, m := range regexp.MustCompile(`\(cookbook\.md#([^)]+)\)`).FindAllStringSubmatch(index, -1) {
		if !anchors[m[1]] {
			t.Errorf("doc/examples.md links cookbook.md#%s, which matches no doc/cookbook.md heading", m[1])
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

// examplesDeadOnWindows records every hermetic example whose scenarios ALL skip
// on Windows, with the reason. The README calls examples/ the syntax reference
// and says it is tested on Linux, macOS, and Windows; TestExamples_HermeticRunGreen
// accepts a skipped spec as green, so an example that executes nothing there
// looks exactly like one that passes. Fifteen of them did.
//
// The map is EXACT, not a floor: an example listed here that starts running a
// scenario on Windows fails the test, so the list shrinks as examples are made
// portable instead of quietly outliving its reasons. Everything in it is
// POSIX-shell scaffolding — the alternative is a second cmd.exe copy of every
// scenario in the file readers are pointed at to learn the syntax, which would
// cost more clarity than the coverage is worth.
var examplesDeadOnWindows = map[string]string{
	"examples/changes.atago.yaml":              "the steps that produce a delta are POSIX shell (rm, chmod, mkfifo)",
	"examples/deterministic.atago.yaml":        "cat, and a JSON document read back through it",
	"examples/extend_host_env.atago.yaml":      "POSIX PATH syntax: the : separator, sh, and command -v",
	"examples/hermetic_env.atago.yaml":         "env and printf, which have no cmd.exe builtin",
	"examples/mock_server.atago.yaml":          "curl, which the runner images do not guarantee on Windows",
	"examples/project_manifest.atago.yaml":     "cat, plus a ${specdir} interpolated into a shell command",
	"examples/pty.atago.yaml":                  "the inner programs are POSIX: [ -t 0 ], cat -v, and a SIGINT trap",
	"examples/pty_screen.atago.yaml":           "the inner program draws with printf escapes and a POSIX read loop",
	"examples/pty_stdout_split.atago.yaml":     "the inner program writes its UI to /dev/stderr from a POSIX shell",
	"examples/retry.atago.yaml":                "the marker-file poll is POSIX shell test/touch",
	"examples/scrub.atago.yaml":                "the volatile output is synthesized with $$ and $(date +%s)",
	"examples/services.atago.yaml":             "the stand-in service is a POSIX shell loop",
	"examples/signal.atago.yaml":               "signal steps are POSIX-only; Windows has no signals to deliver",
	"examples/suite_env_from_setup.atago.yaml": "the setup exports through a POSIX shell",
}

// TestExamples_RunSomethingOnThisOS fails when a hermetic example skips every
// scenario on the host OS without being recorded as dead there. On POSIX
// nothing may be dead at all: an example that runs nothing on the platform it
// was written for is a broken example, not a portability limit.
func TestExamples_RunSomethingOnThisOS(t *testing.T) {
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
			ran := 0
			for _, sc := range res.Scenarios {
				if sc.Status != engine.StatusSkipped {
					ran++
				}
			}
			reason, recorded := examplesDeadOnWindows[path]
			if runtime.GOOS != "windows" {
				if ran == 0 {
					t.Errorf("every scenario skipped on %s; an example must run somewhere", runtime.GOOS)
				}
				return
			}
			switch {
			case ran == 0 && !recorded:
				t.Errorf("every scenario skips on Windows and the example is not recorded in examplesDeadOnWindows; "+
					"make one scenario portable, or add it there with the reason (%d scenarios)", len(res.Scenarios))
			case ran > 0 && recorded:
				t.Errorf("%d of %d scenarios now run on Windows, so the examplesDeadOnWindows entry is stale "+
					"(recorded reason: %s); remove it", ran, len(res.Scenarios), reason)
			}
		})
	}
}

// TestExamples_DeadListNamesRealExamples keeps the record from outliving the
// file it describes: an entry for an example that was renamed or made
// non-hermetic would sit there forever claiming coverage nobody checks.
func TestExamples_DeadListNamesRealExamples(t *testing.T) {
	t.Parallel()
	for path, reason := range examplesDeadOnWindows {
		hermetic, known := exampleSpecs[path]
		switch {
		case !known:
			t.Errorf("examplesDeadOnWindows names %q, which is not a categorized example", path)
		case !hermetic:
			t.Errorf("examplesDeadOnWindows names %q, which is validate-only and never run", path)
		}
		if strings.TrimSpace(reason) == "" {
			t.Errorf("examplesDeadOnWindows[%q] has no reason", path)
		}
	}
}

// migrationFenceRe captures each fenced code block in the migration guide with
// its language tag, so every excerpt can be checked against the file family it
// claims to come from.
var migrationFenceRe = regexp.MustCompile("(?s)```([a-z]+)\n(.*?)```")

// migrationSourceFamilies maps a fence language in migrate.md to the committed
// parity files its excerpts must be verbatim substrings of: ```bash is the Bats
// suite, ```sh the ShellSpec suite, ```yaml the migrated atago specs. ```shell
// blocks are free-form command lines and are not checked.
func migrationSourceFamilies(t *testing.T) map[string][]string {
	t.Helper()
	families := map[string][]string{}
	for lang, glob := range map[string]string{
		"bash": "test/e2e/migration/bats/*.bats",
		"sh":   "test/e2e/migration/shellspec/spec/*_spec.sh",
		"yaml": "test/e2e/migration/*.atago.yaml",
	} {
		paths, err := filepath.Glob(glob)
		if err != nil || len(paths) == 0 {
			t.Fatalf("glob %s: %v (found %d files)", glob, err, len(paths))
		}
		families[lang] = paths
	}
	return families
}

// TestMigrationGuide_SnippetsAreExecutedExcerpts keeps the website's migration
// guide honest: every Bats, ShellSpec, and atago snippet it shows must be a
// verbatim excerpt of a committed file under test/e2e/migration/ — the files
// the MigrationParity workflow actually runs. A snippet edited only in the
// guide, or a parity file that drifts away from the guide, fails here instead
// of quietly turning the page into fiction.
func TestMigrationGuide_SnippetsAreExecutedExcerpts(t *testing.T) {
	t.Parallel()
	doc := readDoc(t, "website/content/migrate.md")
	families := migrationSourceFamilies(t)
	blocks := migrationFenceRe.FindAllStringSubmatch(doc, -1)
	if len(blocks) == 0 {
		t.Fatal("website/content/migrate.md contains no fenced code blocks")
	}
	checked := 0
	for i, m := range blocks {
		lang, body := m[1], m[2]
		paths, guarded := families[lang]
		if !guarded {
			continue
		}
		checked++
		found := false
		for _, path := range paths {
			if strings.Contains(readDoc(t, path), body) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("migrate.md block %d (```%s) is not a verbatim excerpt of any %s parity file:\n%s", i, lang, lang, body)
		}
	}
	if checked == 0 {
		t.Error("migrate.md has no bats/shellspec/atago snippets to guard; the fence languages may have drifted")
	}
}

// TestMigrationGuide_PinsMatchWorkflow keeps the versions the migration guide
// names in lockstep with what the MigrationParity workflow installs. The
// workflow fetches by immutable commit ID (a tag can move) and keeps the
// release label in a comment; the guide's claim "verified against Bats-core
// v1.14.0 and ShellSpec 0.28.1" is only true while all three stay together.
func TestMigrationGuide_PinsMatchWorkflow(t *testing.T) {
	t.Parallel()
	guide := readDoc(t, "website/content/migrate.md")
	workflow := readDoc(t, ".github/workflows/migration.yml")
	for _, pin := range []struct{ name, sha, label, guideRef string }{
		{"Bats", "eb7f42f8d608ac693d7a4b67474f6714ea68cfc5", "v1.14.0", "Bats-core v1.14.0"},
		{"ShellSpec", "90e48c950239f3b8a9fdfa3e869592872c77b981", "0.28.1", "ShellSpec 0.28.1"},
	} {
		if !strings.Contains(workflow, pin.sha) {
			t.Errorf("migration.yml no longer pins %s to commit %s; update the SHA, the label comment, and the guide together", pin.name, pin.sha)
		}
		if !strings.Contains(workflow, pin.label) {
			t.Errorf("migration.yml no longer labels the %s pin as %s; the release label must stay next to the commit ID", pin.name, pin.label)
		}
		if !strings.Contains(guide, pin.guideRef) {
			t.Errorf("migrate.md no longer names %q; the guide must state the exact release the parity workflow runs", pin.guideRef)
		}
	}
}

// TestMigration_HermeticRunGreen executes every migrated spec behind the
// migration guide through the real engine, on every platform the unit tests
// cover — the Linux-only parity workflow proves the Bats/ShellSpec side, this
// proves the atago side everywhere else. Each scenario is checked
// individually: gated scenarios may skip, but a suite-level "passed" that
// hides a flaky, xfail, or xpass scenario is not good enough for the suites a
// guide points at.
func TestMigration_HermeticRunGreen(t *testing.T) {
	t.Parallel()
	paths, err := filepath.Glob("test/e2e/migration/*.atago.yaml")
	if err != nil || len(paths) == 0 {
		t.Fatalf("glob migration specs: %v (found %d files)", err, len(paths))
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			s, err := loader.Load(path)
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			res := engine.New().Run(context.Background(), s, path)
			if res.Status != engine.StatusPassed && res.Status != engine.StatusSkipped {
				t.Errorf("status = %s, want passed (or skipped by a gate): %+v", res.Status, res.Scenarios)
			}
			if len(res.Scenarios) == 0 {
				t.Error("spec produced no scenario results")
			}
			for _, sc := range res.Scenarios {
				if sc.Status != engine.StatusPassed && sc.Status != engine.StatusSkipped {
					t.Errorf("scenario %q: status = %s, want passed or skipped", sc.Name, sc.Status)
				}
			}
		})
	}
}
