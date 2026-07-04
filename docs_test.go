package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/docgen"
	"github.com/nao1215/atago/internal/loader"
)

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

// e2eDocSuites maps a committed generated doc under doc/e2e/ to the spec
// directory it is rendered from. Keep this in sync with the `atago doc`
// invocations recorded in doc/e2e/README.md.
var e2eDocSuites = map[string]string{
	"doc/e2e/atago.md":       "test/e2e/atago",
	"doc/e2e/git.md":         "test/e2e/thirdparty/git",
	"doc/e2e/caddy.md":       "test/e2e/thirdparty/caddy",
	"doc/e2e/pushgateway.md": "test/e2e/thirdparty/pushgateway",
	"doc/e2e/webhook.md":     "test/e2e/thirdparty/webhook",
	"doc/e2e/gitea.md":       "test/e2e/thirdparty/gitea",
	"doc/e2e/minio.md":       "test/e2e/thirdparty/minio",
	"doc/e2e/prometheus.md":  "test/e2e/thirdparty/prometheus",
	"doc/e2e/rclone.md":      "test/e2e/thirdparty/rclone",
	"doc/e2e/restic.md":      "test/e2e/thirdparty/restic",
	"doc/e2e/coredns.md":     "test/e2e/thirdparty/coredns",
	"doc/e2e/nats.md":        "test/e2e/thirdparty/nats",
	"doc/e2e/mailpit.md":     "test/e2e/thirdparty/mailpit",
	"doc/e2e/ntfy.md":        "test/e2e/thirdparty/ntfy",
	"doc/e2e/transfersh.md":  "test/e2e/thirdparty/transfersh",
	"doc/e2e/gotify.md":      "test/e2e/thirdparty/gotify",
	"doc/e2e/grafana.md":     "test/e2e/thirdparty/grafana",
	"doc/e2e/gup.md":         "test/e2e/tools/gup",
	"doc/e2e/sqly.md":        "test/e2e/tools/sqly",
	"doc/e2e/truss.md":       "test/e2e/tools/truss",
	"doc/e2e/iso8583tool.md": "test/e2e/tools/iso8583tool",
	"doc/e2e/jose.md":        "test/e2e/tools/jose",
	"doc/e2e/career.md":      "test/e2e/tools/career",
	"doc/e2e/mimixbox.md":    "test/e2e/tools/mimixbox",
	"doc/e2e/mobilepkg.md":   "test/e2e/tools/mobilepkg",
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
	for docPath, specDir := range e2eDocSuites {
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
