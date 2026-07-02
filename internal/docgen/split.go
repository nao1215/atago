package docgen

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nao1215/markdown"
)

// SplitDoc is one generated per-spec Markdown document: the file name to write it
// as (relative to the chosen output directory) and its rendered content.
type SplitDoc struct {
	Name    string
	Content []byte
}

// GenerateSplit renders one standalone Markdown document per source plus an index
// page that links them (#68). File names are derived deterministically from each
// spec's path and de-duplicated, and the sources' order is preserved, so the
// emitted set is stable across runs. The index is returned separately so the
// caller can name it (conventionally index.md).
// outputDir is the directory the per-spec files will be written to, so embedded
// golden images resolve to a correct relative link (#67).
func GenerateSplit(sources []Source, outputDir string) (index []byte, docs []SplitDoc, err error) {
	names := splitFilenames(sources)
	docs = make([]SplitDoc, 0, len(sources))
	for i, src := range sources {
		var buf bytes.Buffer
		if gerr := GenerateTo(&buf, []Source{src}, outputDir); gerr != nil {
			return nil, nil, fmt.Errorf("generate %s: %w", src.Path, gerr)
		}
		docs = append(docs, SplitDoc{Name: names[i], Content: buf.Bytes()})
	}

	var idx bytes.Buffer
	if err := generateIndex(&idx, sources, names); err != nil {
		return nil, nil, err
	}
	// Normalize to LF (the Markdown writer emits CRLF on Windows) so split output
	// is byte-identical across platforms, like Generate.
	index = bytes.ReplaceAll(idx.Bytes(), []byte("\r\n"), []byte("\n"))
	return index, docs, nil
}

// generateIndex renders the index page linking each per-spec document, with the
// document-wide summary at the top.
func generateIndex(w *bytes.Buffer, sources []Source, names []string) error {
	md := markdown.NewMarkdown(w)
	md.H1("atago Behavior Specs — Index")

	sum := computeSummary(sources)
	md.PlainTextf("%s · %s", pluralize(sum.suites, "suite"), pluralize(sum.scenarios, "scenario"))
	if tags := sum.tagLine(); tags != "" {
		md.PlainTextf("Tags: %s", tags)
	}

	md.H2("Documents")
	var b strings.Builder
	for i, src := range sources {
		fmt.Fprintf(&b, "- [%s](%s) — %s\n",
			mdEscape(src.Spec.Suite.Name), names[i], pluralize(len(src.Spec.Scenarios), "scenario"))
	}
	md.PlainText(strings.TrimRight(b.String(), "\n"))
	return md.Build()
}

// splitFilenames derives a deterministic, collision-free .md file name for each
// source from its spec path, preserving input order.
func splitFilenames(sources []Source) []string {
	used := map[string]int{}
	out := make([]string, len(sources))
	for i, src := range sources {
		base := splitBaseName(src.Path)
		name := base + ".md"
		if n := used[base]; n > 0 {
			name = fmt.Sprintf("%s-%d.md", base, n)
		}
		used[base]++
		out[i] = name
	}
	return out
}

// splitBaseName sanitizes a spec path into a file-name stem: the spec's base file
// name with the .atago.yaml/.atago.yml suffix removed and any character outside
// [A-Za-z0-9._-] replaced with a hyphen. It never returns an empty stem.
func splitBaseName(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, ".atago.yaml")
	base = strings.TrimSuffix(base, ".atago.yml")
	var b strings.Builder
	for _, r := range base {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '_', r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	name := strings.Trim(b.String(), "-")
	if name == "" {
		return "spec"
	}
	return name
}
