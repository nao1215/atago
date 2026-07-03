package docgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
)

// Preview truncation budget. Previews surface the authored inputs and exact
// expected payloads that the command line alone hides (#67), but they must stay
// compact and deterministic so committed docs remain reviewable. A preview is
// capped at previewMaxLines lines and previewMaxBytes bytes; whatever is dropped
// is reported with a stable marker.
const (
	previewMaxLines = 20
	previewMaxBytes = 800
)

// previewBlock is one rendered preview: a short label and the (already
// truncated) body to show in a fenced code block. When image is set, body is a
// relative path and the block renders as a Markdown image embed instead.
type previewBlock struct {
	label string
	lang  string
	body  string
	image bool
}

// inputPreviews collects the authored inputs worth previewing for a scenario:
// inline fixture contents and run-step stdin payloads. These are the parts that
// the "When" command line does not reveal (#67).
func inputPreviews(sc *spec.Scenario) []previewBlock {
	var out []previewBlock
	for i := range sc.Steps {
		step := &sc.Steps[i]
		switch step.Kind() {
		case spec.StepFixture:
			f := step.Fixture
			if f.Content != "" {
				out = append(out, previewBlock{
					label: fmt.Sprintf("Fixture `%s`", f.File),
					lang:  "",
					body:  truncatePreview(f.Content),
				})
			}
		case spec.StepRun:
			switch s := step.Run.Stdin; {
			case s.Inline != "":
				out = append(out, previewBlock{
					label: fmt.Sprintf("stdin for `%s`", firstToken(step.Run.Command)),
					lang:  "",
					body:  truncatePreview(s.Inline),
				})
			case s.File != "":
				out = append(out, previewBlock{
					label: fmt.Sprintf("stdin for `%s`", firstToken(step.Run.Command)),
					lang:  "",
					body:  "(read from file " + s.File + ")",
				})
			case s.Base64 != "":
				out = append(out, previewBlock{
					label: fmt.Sprintf("stdin for `%s`", firstToken(step.Run.Command)),
					lang:  "",
					body:  fmt.Sprintf("(binary, %d base64 chars)", len(s.Base64)),
				})
			}
		}
	}
	return out
}

// exactPreviews collects previews of exact-value assertions and snapshots that
// are too large or multi-line to read inline in the "Then" bullets (#67). Short
// single-line exact values already render inline via describeAssert, so only the
// multi-line/long ones get a block here to avoid duplicate noise.
//
// A snapshot golden's committed content is inlined so a reader sees the expected
// output without opening the snapshot file (#67). specDir resolves the snapshot
// path (the same rule the runner uses); when the golden is not a committed file
// on disk (e.g. it is produced at runtime inside a fixture spec), the block
// falls back to just naming the snapshot reference.
func exactPreviews(sc *spec.Scenario, specDir, outputDir string) []previewBlock {
	var out []previewBlock
	for i := range sc.Steps {
		step := &sc.Steps[i]
		if step.Kind() != spec.StepAssert {
			continue
		}
		a := step.Assert
		for _, sa := range []struct {
			name   string
			stream *spec.StreamAssert
		}{
			{"stdout", a.Stdout}, {"stderr", a.Stderr}, {"body", a.Body},
			{"rows", a.Rows}, {"message", a.Message}, {"value", a.Value},
		} {
			if sa.stream == nil {
				continue
			}
			if sa.stream.Equals != nil && isPreviewable(*sa.stream.Equals) {
				out = append(out, previewBlock{
					label: fmt.Sprintf("expected %s", sa.name),
					body:  truncatePreview(*sa.stream.Equals),
				})
			}
			if sa.stream.Snapshot != "" {
				out = append(out, snapshotPreview(sa.name+" snapshot", sa.stream.Snapshot, specDir))
			}
		}
		if a.File != nil && a.File.Snapshot != "" {
			out = append(out, snapshotPreview(fmt.Sprintf("%s snapshot", a.File.Path), a.File.Snapshot, specDir))
		}
		if a.Image != nil && a.Image.SimilarTo != "" {
			if b, ok := imageBaselinePreview(a.Image.SimilarTo, specDir, outputDir); ok {
				out = append(out, b)
			}
		}
	}
	return out
}

// imageBaselinePreview builds an embed for an image `similar_to` baseline when it
// is a committed file on disk and outputDir is known, so the expected image
// renders inline in the doc. It returns ok=false for a runtime baseline (a
// ${workdir}-relative path), a missing file, or when no output dir is set (e.g.
// stdout) — the "Then" bullet still names the baseline in those cases.
func imageBaselinePreview(similarTo, specDir, outputDir string) (previewBlock, bool) {
	if outputDir == "" || strings.Contains(similarTo, "${") {
		return previewBlock{}, false
	}
	path, err := security.ResolveSpecPath("assert.image.similar_to", specDir, similarTo)
	if err != nil {
		return previewBlock{}, false
	}
	if info, statErr := os.Stat(path); statErr != nil || info.IsDir() {
		return previewBlock{}, false
	}
	rel, err := relPath(outputDir, path)
	if err != nil {
		return previewBlock{}, false
	}
	return previewBlock{label: fmt.Sprintf("expected image `%s`", similarTo), body: rel, image: true}, true
}

// relPath returns the slash-separated path of target relative to baseDir, using
// absolute forms so the result is independent of the current directory (both are
// anchored to the same cwd, which cancels out).
func relPath(baseDir, target string) (string, error) {
	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(baseAbs, targetAbs)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}

// snapshotPreview builds a preview block for a snapshot golden, inlining the
// committed file's content when it can be read, or falling back to a reference
// label (empty body) when it cannot (a runtime-produced snapshot).
func snapshotPreview(kind, snap, specDir string) previewBlock {
	label := fmt.Sprintf("%s `%s`", kind, snap)
	path, err := security.ResolveSpecPath("assert.snapshot", specDir, snap)
	if err != nil {
		return previewBlock{label: label}
	}
	data, err := os.ReadFile(path) //nolint:gosec // committed snapshot golden, path confined to the spec dir
	if err != nil {
		return previewBlock{label: label}
	}
	return previewBlock{label: label, body: truncatePreview(string(data))}
}

// isPreviewable reports whether an exact value is worth a dedicated preview
// block: multi-line, or longer than a short inline snippet. Short single-line
// values are already shown inline by describeAssert.
func isPreviewable(v string) bool {
	return strings.Contains(v, "\n") || len(v) > 60
}

// truncatePreview trims a payload to the preview budget, appending a stable,
// clearly-marked notice describing exactly what was dropped so truncation is
// deterministic and honest (#67).
func truncatePreview(s string) string {
	// Strip carriage returns so a preview is byte-identical regardless of the
	// source's line endings (e.g. a DOS-CRLF fixture like a dos2unix test input)
	// and regardless of the OS the docs are generated on. The preview is an
	// illustrative, human-readable approximation, so the invisible CR control
	// character carries no information worth keeping.
	s = strings.ReplaceAll(s, "\r", "")
	// Normalize trailing newline noise so the marker math is stable.
	s = strings.TrimRight(s, "\n")
	lines := strings.Split(s, "\n")
	droppedLines := 0
	if len(lines) > previewMaxLines {
		droppedLines = len(lines) - previewMaxLines
		lines = lines[:previewMaxLines]
	}
	body := strings.Join(lines, "\n")

	truncatedBytes := false
	if len(body) > previewMaxBytes {
		body = body[:previewMaxBytes]
		truncatedBytes = true
	}

	switch {
	case droppedLines > 0:
		plural := "lines"
		if droppedLines == 1 {
			plural = "line"
		}
		body += fmt.Sprintf("\n… (truncated, %d more %s)", droppedLines, plural)
	case truncatedBytes:
		body += "\n… (truncated)"
	}
	return body
}

// firstToken returns the first whitespace-delimited token of a command, used to
// label a stdin preview with the command it feeds.
func firstToken(cmd string) string {
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return cmd
	}
	return fields[0]
}
