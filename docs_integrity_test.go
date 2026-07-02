package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// docRefPattern matches the local documentation paths this project cites from
// its code and prose — the repo-root spec.md (the design docs it once also
// guarded were removed from the repository). These are the references whose
// absence (issue #70) silently breaks reader trust, so the test asserts each
// referenced file actually exists on disk.
var docRefPattern = regexp.MustCompile(`\bspec\.md\b`)

// scanRootsForDocRefs are the trees and files whose comments/strings/prose may
// reference local docs. Kept small and deterministic: all *.go under the repo
// (excluding vendor/.git) plus the two Markdown files that cite design docs.
func collectDocReferences(t *testing.T) map[string][]string {
	t.Helper()

	// refs maps a referenced doc path to the source files that mention it, so a
	// failure can name where the dangling reference came from.
	refs := make(map[string][]string)
	record := func(source, content string) {
		for _, m := range docRefPattern.FindAllString(content, -1) {
			ref := filepath.ToSlash(filepath.Clean(m))
			refs[ref] = appendUnique(refs[ref], source)
		}
	}

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "vendor" || d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") {
			data, rerr := os.ReadFile(path)
			if rerr != nil {
				return rerr
			}
			record(filepath.ToSlash(path), string(data))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk repo: %v", err)
	}

	for _, doc := range []string{"CHANGELOG.md", "README.md"} {
		data, rerr := os.ReadFile(doc)
		if rerr != nil {
			t.Fatalf("read %s: %v", doc, rerr)
		}
		record(doc, string(data))
	}
	return refs
}

func appendUnique(xs []string, x string) []string {
	for _, existing := range xs {
		if existing == x {
			return xs
		}
	}
	return append(xs, x)
}

// TestDocsIntegrity_ReferencedLocalDocsExist is the regression guard for issue
// #70: the code and CHANGELOG referenced spec.md and doc/design/*.md files that
// did not exist. Every local doc path cited from the source, README, or
// CHANGELOG must resolve to a real file relative to the repo root.
func TestDocsIntegrity_ReferencedLocalDocsExist(t *testing.T) {
	refs := collectDocReferences(t)
	if len(refs) == 0 {
		t.Fatal("found no local doc references to check; the extractor is likely broken")
	}

	paths := make([]string, 0, len(refs))
	for p := range refs {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, p := range paths {
		if info, err := os.Stat(p); err != nil || info.IsDir() {
			sources := refs[p]
			sort.Strings(sources)
			t.Errorf("referenced local doc %q does not exist on disk (cited by: %s)",
				p, strings.Join(sources, ", "))
		}
	}
}
