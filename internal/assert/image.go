package assert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	// Register the raster decoders used by checkImage. PNG/JPEG/GIF come from the
	// standard library; BMP/TIFF/WebP from golang.org/x/image (pure Go, decode
	// only). AVIF and SVG cannot be decoded in pure Go, so they are recognized by
	// content for format assertions but cannot be measured or compared.
	_ "image/gif"
	_ "image/jpeg"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	"github.com/nao1215/atago/internal/spec"
)

// checkImage evaluates an image assertion (ADR-0030). Every constraint set on
// the assertion must hold. Relative paths resolve against the scenario workdir;
// a relative similar_to baseline resolves against the spec file's directory.
func checkImage(im *spec.ImageAssert, env Env) *CheckResult {
	path := im.Path
	if !filepath.IsAbs(path) {
		path = filepath.Join(env.Workdir, im.Path)
	}

	data, err := os.ReadFile(path) //nolint:gosec // path is the user-declared assertion target
	if err != nil {
		return &CheckResult{
			Desc:     fmt.Sprintf("assert image %q", im.Path),
			Expected: fmt.Sprintf("readable image file %q", im.Path),
			Actual:   err.Error(),
			Hint:     fmt.Sprintf("could not read image %q", im.Path),
		}
	}

	// Every constraint that is set must hold (conjunctive), so each check returns
	// only on failure and the assertion passes only if all of them pass.
	if im.Format != "" {
		if cr := checkImageFormat(im, data); cr != nil {
			return cr
		}
	}

	needDecode := im.Width != nil || im.Height != nil ||
		im.MinWidth != nil || im.MaxWidth != nil ||
		im.MinHeight != nil || im.MaxHeight != nil || im.Alpha != nil
	if needDecode {
		if cr := checkImageProperties(im, data); !cr.OK {
			return cr
		}
	}

	if im.SimilarTo != "" {
		return checkImageSimilarTo(im, data, env)
	}

	return pass(fmt.Sprintf("assert image %q", im.Path))
}

// checkImageFormat verifies the encoded format detected from the file content.
func checkImageFormat(im *spec.ImageAssert, data []byte) *CheckResult {
	want := strings.ToLower(im.Format)
	got := detectImageFormat(data)
	desc := fmt.Sprintf("assert image %q format is %q", im.Path, want)
	if got == want {
		return nil
	}
	actual := got
	if actual == "" {
		actual = "unknown"
	}
	return &CheckResult{
		Desc:     desc,
		Expected: fmt.Sprintf("format %q", want),
		Actual:   fmt.Sprintf("format %q", actual),
		Hint:     fmt.Sprintf("image %q is encoded as %q, not %q", im.Path, actual, want),
	}
}

// checkImageProperties decodes the image and checks every dimension/alpha
// constraint that is set.
func checkImageProperties(im *spec.ImageAssert, data []byte) *CheckResult {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return &CheckResult{
			Desc:     fmt.Sprintf("assert image %q", im.Path),
			Expected: "a decodable raster image",
			Actual:   err.Error(),
			Hint:     fmt.Sprintf("could not decode image %q (avif/svg cannot be measured)", im.Path),
		}
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()

	if cr := checkExactDim(im.Path, "width", im.Width, w); cr != nil {
		return cr
	}
	if cr := checkExactDim(im.Path, "height", im.Height, h); cr != nil {
		return cr
	}
	if cr := checkMinDim(im.Path, "width", im.MinWidth, w); cr != nil {
		return cr
	}
	if cr := checkMaxDim(im.Path, "width", im.MaxWidth, w); cr != nil {
		return cr
	}
	if cr := checkMinDim(im.Path, "height", im.MinHeight, h); cr != nil {
		return cr
	}
	if cr := checkMaxDim(im.Path, "height", im.MaxHeight, h); cr != nil {
		return cr
	}
	if im.Alpha != nil {
		if cr := checkImageAlpha(im, img); cr != nil {
			return cr
		}
	}
	return pass(fmt.Sprintf("assert image %q properties", im.Path))
}

func checkExactDim(path, dim string, want *int, got int) *CheckResult {
	if want == nil || *want == got {
		return nil
	}
	desc := fmt.Sprintf("assert image %q %s is %d", path, dim, *want)
	return &CheckResult{
		Desc:     desc,
		Expected: fmt.Sprintf("%s %dpx", dim, *want),
		Actual:   fmt.Sprintf("%s %dpx", dim, got),
		Hint:     fmt.Sprintf("image %q has %s %dpx, expected %dpx", path, dim, got, *want),
	}
}

func checkMinDim(path, dim string, min *int, got int) *CheckResult {
	if min == nil || got >= *min {
		return nil
	}
	desc := fmt.Sprintf("assert image %q %s >= %d", path, dim, *min)
	return &CheckResult{
		Desc:     desc,
		Expected: fmt.Sprintf("%s >= %dpx", dim, *min),
		Actual:   fmt.Sprintf("%s %dpx", dim, got),
		Hint:     fmt.Sprintf("image %q %s %dpx is below the minimum %dpx", path, dim, got, *min),
	}
}

func checkMaxDim(path, dim string, max *int, got int) *CheckResult {
	if max == nil || got <= *max {
		return nil
	}
	desc := fmt.Sprintf("assert image %q %s <= %d", path, dim, *max)
	return &CheckResult{
		Desc:     desc,
		Expected: fmt.Sprintf("%s <= %dpx", dim, *max),
		Actual:   fmt.Sprintf("%s %dpx", dim, got),
		Hint:     fmt.Sprintf("image %q %s %dpx exceeds the maximum %dpx", path, dim, got, *max),
	}
}

func checkImageAlpha(im *spec.ImageAssert, img image.Image) *CheckResult {
	got := hasAlpha(img)
	desc := fmt.Sprintf("assert image %q alpha: %t", im.Path, *im.Alpha)
	if got == *im.Alpha {
		return nil
	}
	return &CheckResult{
		Desc:     desc,
		Expected: fmt.Sprintf("alpha=%t", *im.Alpha),
		Actual:   fmt.Sprintf("alpha=%t", got),
		Hint:     fmt.Sprintf("image %q %s an alpha channel", im.Path, presence(got)),
	}
}

// checkImageSimilarTo decodes both the asserted image and the baseline and
// compares their pixels. The normalized mean per-pixel difference must be at or
// below MaxDiff (default 0, an exact match). On failure it attaches review
// artifacts (the actual image, the baseline image, a deterministic diff heatmap
// when both are decodable and same-size, plus a metadata JSON) for durable
// inspection via --artifacts-dir (#52). The comparison semantics are unchanged.
func checkImageSimilarTo(im *spec.ImageAssert, data []byte, env Env) *CheckResult {
	desc := fmt.Sprintf("assert image %q similar to %q", im.Path, im.SimilarTo)
	maxDiff := 0.0
	if im.MaxDiff != nil {
		maxDiff = *im.MaxDiff
	}
	meta := imageDiffMeta{Actual: im.Path, Baseline: im.SimilarTo, MaxDiff: maxDiff}

	got, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		// AVIF/SVG and corrupt rasters cannot be decoded in pure Go; emit metadata
		// explaining why no visual diff was produced instead of a misleading one.
		meta.Reason = "actual image could not be decoded (avif/svg cannot be compared)"
		cr := &CheckResult{
			Desc:     desc,
			Expected: "a decodable raster image",
			Actual:   err.Error(),
			Hint:     fmt.Sprintf("could not decode image %q for comparison", im.Path),
		}
		return attachImageArtifacts(cr, meta, data, nil)
	}
	meta.ActualFormat = detectImageFormat(data)
	meta.ActualDimensions = dimStr(got)

	basePath := im.SimilarTo
	if !filepath.IsAbs(basePath) {
		basePath = filepath.Join(env.SpecDir, im.SimilarTo)
	}
	baseData, err := os.ReadFile(basePath) //nolint:gosec // path is the user-declared baseline
	if err != nil {
		meta.Reason = "baseline image could not be read"
		cr := &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("readable baseline image %q", im.SimilarTo),
			Actual:   err.Error(),
			Hint:     fmt.Sprintf("could not read baseline image %q", im.SimilarTo),
		}
		return attachImageArtifacts(cr, meta, data, nil)
	}
	base, _, err := image.Decode(bytes.NewReader(baseData))
	if err != nil {
		meta.Reason = "baseline image could not be decoded (avif/svg cannot be compared)"
		cr := &CheckResult{
			Desc:     desc,
			Expected: "a decodable baseline raster image",
			Actual:   err.Error(),
			Hint:     fmt.Sprintf("could not decode baseline image %q", im.SimilarTo),
		}
		return attachImageArtifacts(cr, meta, data, baseData)
	}
	meta.BaselineFormat = detectImageFormat(baseData)
	meta.BaselineDimensions = dimStr(base)

	gb, bb := got.Bounds(), base.Bounds()
	if gb.Dx() != bb.Dx() || gb.Dy() != bb.Dy() {
		meta.Reason = "actual and baseline dimensions differ; a pixel diff was not produced"
		cr := &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("dimensions %dx%d (baseline)", bb.Dx(), bb.Dy()),
			Actual:   fmt.Sprintf("dimensions %dx%d", gb.Dx(), gb.Dy()),
			Hint:     "images must share dimensions to compare pixels",
		}
		return attachImageArtifacts(cr, meta, data, baseData)
	}

	diff := meanPixelDiff(got, base)
	if diff <= maxDiff {
		return pass(desc)
	}
	meta.MeanDiff = &diff
	meta.Reason = "mean pixel difference exceeded max_diff"
	cr := &CheckResult{
		Desc:     desc,
		Expected: fmt.Sprintf("mean pixel difference <= %.4f", maxDiff),
		Actual:   fmt.Sprintf("mean pixel difference %.4f", diff),
		Hint:     fmt.Sprintf("image %q differs from baseline %q by %.4f (allowed %.4f)", im.Path, im.SimilarTo, diff, maxDiff),
	}
	// Both sides decoded and share dimensions: emit a deterministic heatmap.
	heatmap := renderDiffHeatmap(got, base)
	return attachImageArtifacts(cr, meta, data, baseData, heatmap)
}

// imageDiffMeta is the structured metadata written alongside an image similar_to
// failure, so tooling can present the comparison even when a visual diff could
// not be produced (#52).
type imageDiffMeta struct {
	Actual             string   `json:"actual"`
	Baseline           string   `json:"baseline"`
	MaxDiff            float64  `json:"max_diff"`
	MeanDiff           *float64 `json:"mean_diff,omitempty"`
	ActualFormat       string   `json:"actual_format,omitempty"`
	BaselineFormat     string   `json:"baseline_format,omitempty"`
	ActualDimensions   string   `json:"actual_dimensions,omitempty"`
	BaselineDimensions string   `json:"baseline_dimensions,omitempty"`
	DiffGenerated      bool     `json:"diff_generated"`
	Reason             string   `json:"reason"`
}

// attachImageArtifacts records the actual image, the baseline image (when read),
// an optional diff heatmap, and the metadata JSON on a failed image check for
// export via --artifacts-dir (#52). It is a no-op-safe builder: nil/empty inputs
// are skipped.
func attachImageArtifacts(cr *CheckResult, meta imageDiffMeta, actual, baseline []byte, heatmap ...[]byte) *CheckResult {
	cr.ArtifactKind = "image"
	if len(actual) > 0 {
		cr.ArtifactBlobs = append(cr.ArtifactBlobs, ArtifactBlob{Role: "actual", Ext: imageExt(actual), Data: actual})
	}
	if len(baseline) > 0 {
		cr.ArtifactBlobs = append(cr.ArtifactBlobs, ArtifactBlob{Role: "baseline", Ext: imageExt(baseline), Data: baseline})
	}
	if len(heatmap) > 0 && len(heatmap[0]) > 0 {
		meta.DiffGenerated = true
		cr.ArtifactBlobs = append(cr.ArtifactBlobs, ArtifactBlob{Role: "diff", Ext: "png", Data: heatmap[0]})
	}
	if payload, err := json.MarshalIndent(meta, "", "  "); err == nil {
		cr.ArtifactBlobs = append(cr.ArtifactBlobs, ArtifactBlob{Role: "metadata", Ext: "json", Data: append(payload, '\n')})
	}
	return cr
}

// imageExt returns a filename extension for the raw image bytes, falling back to
// "bin" for formats that are recognized but non-raster (avif/svg) or unknown, so
// the preserved actual/baseline keep a meaningful suffix.
func imageExt(data []byte) string {
	if f := detectImageFormat(data); f != "" {
		return f
	}
	return "bin"
}

func dimStr(img image.Image) string {
	b := img.Bounds()
	return fmt.Sprintf("%dx%d", b.Dx(), b.Dy())
}

// renderDiffHeatmap produces a deterministic red-on-black heatmap the same size
// as the inputs: each pixel's red intensity is its normalized per-pixel
// difference (0 = identical/black, 1 = maximally different/bright red). Encoding
// is the standard library's deterministic PNG encoder.
func renderDiffHeatmap(a, b image.Image) []byte {
	ab := a.Bounds()
	w, h := ab.Dx(), ab.Dy()
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	bb := b.Bounds()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			ac, _ := color.NRGBA64Model.Convert(a.At(ab.Min.X+x, ab.Min.Y+y)).(color.NRGBA64)
			bc, _ := color.NRGBA64Model.Convert(b.At(bb.Min.X+x, bb.Min.Y+y)).(color.NRGBA64)
			d := absDiff(uint32(ac.R), uint32(bc.R)) + absDiff(uint32(ac.G), uint32(bc.G)) +
				absDiff(uint32(ac.B), uint32(bc.B)) + absDiff(uint32(ac.A), uint32(bc.A))
			// Normalize the 4-channel sum (0..4*65535) to 0..255.
			intensity := uint8(d / (4 * 65535) * 255)
			out.Set(x, y, color.RGBA{R: intensity, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, out); err != nil {
		return nil
	}
	return buf.Bytes()
}

// meanPixelDiff returns the mean absolute per-channel difference between two
// equally sized images, normalized to 0..1 (0 = identical, 1 = maximally
// different). Channels are read as non-premultiplied 16-bit values (NRGBA64) so
// color differences in transparent or semi-transparent regions are not masked
// by alpha premultiplication.
func meanPixelDiff(a, b image.Image) float64 {
	ab, bb := a.Bounds(), b.Bounds()
	var total float64
	var count int64
	for y := 0; y < ab.Dy(); y++ {
		for x := 0; x < ab.Dx(); x++ {
			ac, _ := color.NRGBA64Model.Convert(a.At(ab.Min.X+x, ab.Min.Y+y)).(color.NRGBA64)
			bc, _ := color.NRGBA64Model.Convert(b.At(bb.Min.X+x, bb.Min.Y+y)).(color.NRGBA64)
			total += absDiff(uint32(ac.R), uint32(bc.R)) + absDiff(uint32(ac.G), uint32(bc.G)) +
				absDiff(uint32(ac.B), uint32(bc.B)) + absDiff(uint32(ac.A), uint32(bc.A))
			count += 4
		}
	}
	if count == 0 {
		return 0
	}
	return total / (float64(count) * 65535.0)
}

func absDiff(a, b uint32) float64 {
	if a > b {
		return float64(a - b)
	}
	return float64(b - a)
}

// hasAlpha reports whether the image actually carries transparency: any pixel
// whose alpha is less than fully opaque.
//
// It scans decoded pixels rather than trusting img.ColorModel(), because the Go
// decoders materialize an opaque truecolor PNG (color-type 2, no alpha channel)
// and a 24-bit BMP as *image.RGBA — whose color model *is* alpha-bearing — so a
// model-based check false-positives `alpha=true` on images with no alpha (issue
// #13). A pixel scan reflects the observable truth: an opaque image reports
// alpha=false regardless of how the decoder represents it in memory.
func hasAlpha(img image.Image) bool {
	// Fast path: an *image.Opaque-implementing type (YCbCr, CMYK, some RGBA) can
	// declare itself fully opaque, letting us skip the per-pixel scan.
	if o, ok := img.(interface{ Opaque() bool }); ok && o.Opaque() {
		return false
	}
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if _, _, _, a := img.At(x, y).RGBA(); a < 0xffff {
				return true
			}
		}
	}
	return false
}

func presence(has bool) string {
	if has {
		return "has"
	}
	return "does not have"
}

// detectImageFormat sniffs the encoded image format from the file content,
// returning a lowercase name (png, jpeg, gif, webp, bmp, tiff, avif, svg) or ""
// when unrecognized. Detection covers formats that cannot be decoded in pure Go
// (avif, svg) so format assertions still work for them.
func detectImageFormat(data []byte) string {
	switch {
	case len(data) >= 8 && bytes.Equal(data[:8], []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}):
		return "png"
	case len(data) >= 3 && data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff:
		return "jpeg"
	case len(data) >= 6 && (bytes.Equal(data[:6], []byte("GIF87a")) || bytes.Equal(data[:6], []byte("GIF89a"))):
		return "gif"
	case len(data) >= 12 && bytes.Equal(data[:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")):
		return "webp"
	case isBMP(data):
		return "bmp"
	case len(data) >= 4 && (bytes.Equal(data[:4], []byte{'I', 'I', 0x2a, 0x00}) || bytes.Equal(data[:4], []byte{'M', 'M', 0x00, 0x2a})):
		return "tiff"
	case isAVIF(data):
		return "avif"
	case isSVG(data):
		return "svg"
	default:
		return ""
	}
}

// knownBMPHeaderSizes are the DIB header sizes BMP variants use; checking the
// field at offset 14 guards against false positives on arbitrary "BM…" data.
var knownBMPHeaderSizes = map[uint32]bool{12: true, 40: true, 52: true, 56: true, 64: true, 108: true, 124: true}

// isBMP detects a BMP by its "BM" signature plus a plausible DIB header size.
func isBMP(data []byte) bool {
	if len(data) < 18 || data[0] != 'B' || data[1] != 'M' {
		return false
	}
	dibSize := uint32(data[14]) | uint32(data[15])<<8 | uint32(data[16])<<16 | uint32(data[17])<<24
	return knownBMPHeaderSizes[dibSize]
}

// isAVIF detects an ISOBMFF file whose major/compatible brand is avif/avis.
func isAVIF(data []byte) bool {
	if len(data) < 12 || !bytes.Equal(data[4:8], []byte("ftyp")) {
		return false
	}
	brands := data[8:min(len(data), 64)]
	return bytes.Contains(brands, []byte("avif")) || bytes.Contains(brands, []byte("avis"))
}

// isSVG detects an SVG document by scanning the leading bytes for an "<svg" tag,
// tolerating a leading XML declaration, byte-order mark, or whitespace.
func isSVG(data []byte) bool {
	head := data
	if len(head) > 1024 {
		head = head[:1024]
	}
	return bytes.Contains(head, []byte("<svg"))
}
