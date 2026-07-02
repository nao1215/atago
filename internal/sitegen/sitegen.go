// Package sitegen generates a browsable, repo-local documentation site under
// site/ from repository sources (#72). The output is deterministic — sorted
// listings, zeroed durations, fixed sample bytes — so a committed site can be
// drift-guarded by a Go test in the normal unit-test run, and every link it
// emits points at a file that actually exists in the repository.
//
// It is a repo-local Markdown site (rendered by GitHub), not a deployed website;
// nothing here claims a hosted URL or an unpublished release.
package sitegen

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/engine"
	"github.com/nao1215/atago/internal/report"
	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// Files returns the full set of site files (relative path → content) that the
// site consists of, generated from the repository rooted at root. Callers write
// them (Generate) or compare them against the committed copies (the drift test),
// so both share one source of truth.
func Files(root string) (map[string][]byte, error) {
	files := map[string][]byte{}

	index, err := buildIndex(root)
	if err != nil {
		return nil, err
	}
	files["site/README.md"] = index

	// Deterministic sample report outputs from a fixed, zero-duration result set.
	results := sampleResults()
	for _, f := range []struct {
		name   string
		format report.Format
	}{
		{"site/samples/report.json", report.FormatJSON},
		{"site/samples/report.junit.xml", report.FormatJUnit},
		{"site/samples/report.tap", report.FormatTAP},
	} {
		var buf bytes.Buffer
		if err := report.Render(&buf, f.format, results); err != nil {
			return nil, fmt.Errorf("render %s: %w", f.name, err)
		}
		files[f.name] = buf.Bytes()
	}

	files["site/samples/sample.pdf"] = []byte(samplePDF)

	// A real, tiny image-diff example: baseline vs actual (one changed pixel) and
	// a deterministic per-pixel difference heatmap.
	base := solidImage(color.RGBA{R: 0xC0, G: 0x20, B: 0x20, A: 0xFF})
	actual := solidImage(color.RGBA{R: 0xC0, G: 0x20, B: 0x20, A: 0xFF})
	actual.Set(1, 1, color.RGBA{R: 0x20, G: 0x40, B: 0xC0, A: 0xFF})
	diff := diffImage(base, actual)
	for name, img := range map[string]*image.RGBA{
		"site/samples/imagediff/baseline.png": base,
		"site/samples/imagediff/actual.png":   actual,
		"site/samples/imagediff/diff.png":     diff,
	} {
		enc, err := encodePNG(img)
		if err != nil {
			return nil, err
		}
		files[name] = enc
	}

	files["site/samples/README.md"] = []byte(samplesReadme)

	return files, nil
}

// Generate writes the site into root, creating directories as needed.
func Generate(root string) error {
	files, err := Files(root)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		dest := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(dest), 0o750); err != nil {
			return err
		}
		if err := os.WriteFile(dest, files[name], 0o600); err != nil {
			return err
		}
	}
	return nil
}

// buildIndex renders the browsable site index, linking only files that exist
// under root. Directory listings are sorted so the output is deterministic.
func buildIndex(root string) ([]byte, error) {
	var b strings.Builder
	b.WriteString("# atago documentation\n\n")
	b.WriteString("A browsable, repo-local index of atago's documentation, generated from repository sources by `make site` (see `internal/sitegen`). Every link below points at a file committed in this repository; it is rendered by GitHub and is not a hosted website.\n\n")
	b.WriteString("> Regenerate with `make site`. A drift test (`TestSite_InSync`) keeps this in sync with the sources.\n\n")

	b.WriteString("## Start here\n\n")
	b.WriteString("- [Project README](../README.md)\n")
	b.WriteString("- [Format & CLI specification (`spec.md`)](../spec.md)\n\n")

	if err := writeLinkSection(&b, root, "## Behavior docs (generated from executable specs)", "doc/e2e", ".md", "../doc/e2e/"); err != nil {
		return nil, err
	}

	b.WriteString("## Schemas\n\n")
	for _, s := range []struct{ label, path string }{
		{"Spec file schema", "schema/atago.schema.json"},
		{"Manifest output schema", "schema/manifest.schema.json"},
		{"Report output schema", "schema/report.schema.json"},
		{"Manifest example", "schema/examples/manifest.example.json"},
		{"Report example", "schema/examples/report.example.json"},
	} {
		if exists(root, s.path) {
			fmt.Fprintf(&b, "- [%s](../%s)\n", s.label, s.path)
		}
	}
	b.WriteString("\n")

	b.WriteString("## Samples gallery\n\n")
	b.WriteString("Deterministic artifacts generated from a fixture run (see [samples/README.md](samples/README.md)):\n\n")
	b.WriteString("- Reports: [JSON](samples/report.json) · [JUnit XML](samples/report.junit.xml) · [TAP](samples/report.tap)\n")
	b.WriteString("- Generated PDF: [sample.pdf](samples/sample.pdf)\n")
	b.WriteString("- Image diff: [baseline](samples/imagediff/baseline.png) · [actual](samples/imagediff/actual.png) · [diff](samples/imagediff/diff.png)\n\n")

	b.WriteString("## Demos\n\n")
	for _, img := range []struct{ label, path string }{
		{"Run demo", "doc/img/demo.gif"},
		{"Review demo", "doc/img/review.gif"},
		{"Snapshot demo", "doc/img/snapshot.gif"},
	} {
		if exists(root, img.path) {
			fmt.Fprintf(&b, "![%s](../%s)\n\n", img.label, img.path)
		}
	}

	return []byte(b.String()), nil
}

// writeLinkSection appends a heading and a sorted bullet list of links to every
// file with the given suffix under dir (relative to root). The section is
// omitted entirely when the directory holds no matching files.
func writeLinkSection(b *strings.Builder, root, heading, dir, suffix, linkPrefix string) error {
	entries, err := os.ReadDir(filepath.Join(root, dir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), suffix) {
			names = append(names, e.Name())
		}
	}
	if len(names) == 0 {
		return nil
	}
	sort.Strings(names)
	b.WriteString(heading + "\n\n")
	for _, n := range names {
		fmt.Fprintf(b, "- [%s](%s%s)\n", n, linkPrefix, n)
	}
	b.WriteString("\n")
	return nil
}

func exists(root, rel string) bool {
	_, err := os.Stat(filepath.Join(root, rel))
	return err == nil
}

// sampleResults builds a fixed, illustrative result set with all durations
// zeroed, so the rendered reports are byte-stable across runs.
func sampleResults() []*engine.SuiteResult {
	pass := engine.ScenarioResult{
		Name:   "echo greets the world",
		Status: engine.StatusPassed,
		Steps: []engine.StepResult{
			{Index: 0, Kind: spec.StepRun, Run: &runner.Result{Command: "echo hello atago", ExitCode: 0, Stdout: []byte("hello atago\n")}},
			{Index: 1, Kind: spec.StepAssert, Checks: []*assert.CheckResult{{OK: true, Desc: `assert stdout contains "atago"`}}},
		},
	}
	fail := engine.ScenarioResult{
		Name:   "detects a wrong greeting",
		Status: engine.StatusFailed,
		Steps: []engine.StepResult{
			{Index: 0, Kind: spec.StepRun, Run: &runner.Result{Command: "echo Bob", ExitCode: 0, Stdout: []byte("Bob\n")}},
			{Index: 1, Kind: spec.StepAssert, Checks: []*assert.CheckResult{{
				OK:       false,
				Desc:     `assert stdout contains "Alice"`,
				Expected: `stdout contains "Alice"`,
				Actual:   "Bob",
				Hint:     `the substring "Alice" was not present in stdout`,
			}}},
		},
	}
	return []*engine.SuiteResult{{
		Suite:     "demo",
		SpecPath:  "demo.atago.yaml",
		Status:    engine.StatusFailed,
		Scenarios: []engine.ScenarioResult{pass, fail},
	}}
}

// solidImage returns a 4x4 image filled with c.
func solidImage(c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

// diffImage returns a grayscale-ish heatmap of the per-pixel absolute difference
// between a and b: unchanged pixels are black, changed pixels are red-scaled by
// magnitude. It is deterministic and needs no external library.
func diffImage(a, b *image.RGBA) *image.RGBA {
	bounds := a.Bounds()
	out := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			ar, ag, ab, _ := a.At(x, y).RGBA()
			br, bg, bb, _ := b.At(x, y).RGBA()
			d := (absdiff(ar, br) + absdiff(ag, bg) + absdiff(ab, bb)) / 3
			// d is bounded by 0xFFFF (average of per-channel diffs), so d>>8 fits a
			// byte; mask to make that invariant explicit for the overflow linter.
			v := uint8((d >> 8) & 0xFF)
			out.Set(x, y, color.RGBA{R: v, G: 0, B: 0, A: 0xFF})
		}
	}
	return out
}

func absdiff(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}

func encodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	if err := enc.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// samplePDF is a tiny, uncompressed 1-page PDF with Info metadata and a text
// stream — the same shape the pdf assertion E2E uses. Committed as a real
// generated-artifact example.
const samplePDF = `%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>
endobj
4 0 obj
<< /Length 44 >>
stream
BT /F1 24 Tf 72 700 Td (Sample atago report) Tj ET
endstream
endobj
6 0 obj
<< /Title (atago sample) /Author (atago) >>
endobj
trailer
<< /Root 1 0 R /Info 6 0 R >>
%%EOF
`

const samplesReadme = `# Samples

These artifacts are generated deterministically by ` + "`internal/sitegen`" + ` (run
` + "`make site`" + `) and drift-guarded by ` + "`TestSite_InSync`" + `.

- ` + "`report.json` / `report.junit.xml` / `report.tap`" + ` — sample outputs of
  ` + "`atago run --report <format>`" + `. They are built from a fixed result set with
  **all durations set to zero** so the committed files are byte-stable (a real run
  would report real durations).
- ` + "`sample.pdf`" + ` — a tiny generated PDF, the kind the ` + "`pdf`" + ` assertion inspects.
- ` + "`imagediff/`" + ` — a baseline image, a one-pixel-changed actual image, and a
  per-pixel difference heatmap, the kind produced for an image ` + "`similar_to`" + `
  failure.
`
