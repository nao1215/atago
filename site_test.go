package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/nao1215/atago/internal/sitegen"
)

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
		hasPrefix(target, "http://") || hasPrefix(target, "https://") || hasPrefix(target, "mailto:"))
}

func hasPrefix(s, p string) bool {
	return len(s) >= len(p) && s[:len(p)] == p
}
