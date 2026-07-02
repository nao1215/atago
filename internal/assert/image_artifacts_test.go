package assert

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/nao1215/atago/internal/spec"
)

// blobByRole returns the artifact blob with the given role, or fails.
func blobByRole(t *testing.T, cr *CheckResult, role string) ArtifactBlob {
	t.Helper()
	for _, b := range cr.ArtifactBlobs {
		if b.Role == role {
			return b
		}
	}
	t.Fatalf("no %q artifact blob (have %d blobs)", role, len(cr.ArtifactBlobs))
	return ArtifactBlob{}
}

func hasBlob(cr *CheckResult, role string) bool {
	for _, b := range cr.ArtifactBlobs {
		if b.Role == role {
			return true
		}
	}
	return false
}

func metaOf(t *testing.T, cr *CheckResult) map[string]any {
	t.Helper()
	b := blobByRole(t, cr, "metadata")
	var m map[string]any
	if err := json.Unmarshal(b.Data, &m); err != nil {
		t.Fatalf("metadata not valid JSON: %v\n%s", err, b.Data)
	}
	return m
}

// TestImageArtifacts_DiffOnPixelDifference verifies a same-size pixel-difference
// failure emits actual/baseline/diff images plus metadata, and the diff heatmap
// is a deterministic decodable PNG (#52).
func TestImageArtifacts_DiffOnPixelDifference(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	specDir := t.TempDir()
	base := makePNG(t, 16, 16, color.RGBA{100, 100, 100, 255})
	writeImage(t, specDir, "baseline.png", base)
	writeImage(t, dir, "actual.png", makePNG(t, 16, 16, color.RGBA{255, 0, 0, 255}))

	cr := Check(&spec.Assert{Image: &spec.ImageAssert{Path: "actual.png", SimilarTo: "baseline.png", MaxDiff: floatp(0.1)}},
		nil, Env{Workdir: dir, SpecDir: specDir})
	if cr.OK {
		t.Fatal("expected similar_to failure")
	}
	if cr.ArtifactKind != "image" {
		t.Errorf("ArtifactKind = %q, want image", cr.ArtifactKind)
	}
	for _, role := range []string{"actual", "baseline", "diff", "metadata"} {
		if !hasBlob(cr, role) {
			t.Errorf("missing %q blob", role)
		}
	}
	// The diff heatmap is a decodable PNG the same size as the inputs.
	diff := blobByRole(t, cr, "diff")
	if diff.Ext != "png" {
		t.Errorf("diff ext = %q", diff.Ext)
	}
	img, _, err := image.Decode(bytes.NewReader(diff.Data))
	if err != nil {
		t.Fatalf("diff heatmap not decodable: %v", err)
	}
	if b := img.Bounds(); b.Dx() != 16 || b.Dy() != 16 {
		t.Errorf("diff dims = %dx%d, want 16x16", b.Dx(), b.Dy())
	}
	m := metaOf(t, cr)
	if m["diff_generated"] != true {
		t.Errorf("metadata diff_generated = %v, want true", m["diff_generated"])
	}
	if m["reason"] == "" {
		t.Errorf("metadata missing reason")
	}
}

// TestImageArtifacts_DeterministicDiff confirms the heatmap is byte-stable across
// runs (#52).
func TestImageArtifacts_DeterministicDiff(t *testing.T) {
	t.Parallel()
	a := image.NewRGBA(image.Rect(0, 0, 8, 8))
	b := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			a.Set(x, y, color.RGBA{uint8(x * 8), 0, 0, 255})
			b.Set(x, y, color.RGBA{0, uint8(y * 8), 0, 255})
		}
	}
	h1 := renderDiffHeatmap(a, b)
	h2 := renderDiffHeatmap(a, b)
	if !bytes.Equal(h1, h2) {
		t.Errorf("diff heatmap not deterministic")
	}
	if _, err := png.Decode(bytes.NewReader(h1)); err != nil {
		t.Errorf("heatmap not a valid PNG: %v", err)
	}
}

// TestImageArtifacts_SizeMismatchMetadataNoDiff verifies a dimension mismatch
// preserves actual + baseline and metadata but produces NO diff image (#52).
func TestImageArtifacts_SizeMismatchMetadataNoDiff(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	specDir := t.TempDir()
	writeImage(t, specDir, "baseline.png", makePNG(t, 20, 20, color.RGBA{100, 100, 100, 255}))
	writeImage(t, dir, "actual.png", makePNG(t, 10, 10, color.RGBA{100, 100, 100, 255}))

	cr := Check(&spec.Assert{Image: &spec.ImageAssert{Path: "actual.png", SimilarTo: "baseline.png"}},
		nil, Env{Workdir: dir, SpecDir: specDir})
	if cr.OK {
		t.Fatal("expected dimension-mismatch failure")
	}
	if !hasBlob(cr, "actual") || !hasBlob(cr, "baseline") {
		t.Errorf("expected actual+baseline preserved")
	}
	if hasBlob(cr, "diff") {
		t.Errorf("no diff should be produced for a dimension mismatch")
	}
	m := metaOf(t, cr)
	if m["diff_generated"] != false {
		t.Errorf("metadata diff_generated = %v, want false", m["diff_generated"])
	}
	if m["actual_dimensions"] != "10x10" || m["baseline_dimensions"] != "20x20" {
		t.Errorf("metadata dims = %v / %v", m["actual_dimensions"], m["baseline_dimensions"])
	}
}

// TestImageArtifacts_UndecodableActualNoRasterDiff is the regression from #52:
// an AVIF/SVG (or corrupt) actual image must fail cleanly with metadata and no
// misleading raster diff.
func TestImageArtifacts_UndecodableActualNoRasterDiff(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	specDir := t.TempDir()
	writeImage(t, specDir, "baseline.png", makePNG(t, 16, 16, color.RGBA{10, 10, 10, 255}))
	// An SVG cannot be decoded to raster in pure Go.
	svg := []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"></svg>`)
	writeImage(t, dir, "actual.svg", svg)

	cr := Check(&spec.Assert{Image: &spec.ImageAssert{Path: "actual.svg", SimilarTo: "baseline.png"}},
		nil, Env{Workdir: dir, SpecDir: specDir})
	if cr.OK {
		t.Fatal("expected decode failure")
	}
	if hasBlob(cr, "diff") {
		t.Errorf("no raster diff should be produced for an undecodable image")
	}
	m := metaOf(t, cr)
	if m["diff_generated"] != false {
		t.Errorf("diff_generated = %v, want false", m["diff_generated"])
	}
	reason, _ := m["reason"].(string)
	if reason == "" {
		t.Errorf("metadata must explain why no diff was produced")
	}
	// The actual bytes are still preserved with a non-raster extension.
	a := blobByRole(t, cr, "actual")
	if a.Ext == "png" {
		t.Errorf("actual ext = %q, an svg must not be labeled png", a.Ext)
	}
}
