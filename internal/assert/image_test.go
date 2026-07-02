package assert

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"

	"github.com/nao1215/atago/internal/spec"
)

func floatp(f float64) *float64 { return &f }

// writePNG encodes a w×h image filled with c and returns the bytes.
func makePNG(t *testing.T, w, h int, c color.Color) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png encode: %v", err)
	}
	return buf.Bytes()
}

// makeNRGBAPNG encodes a non-premultiplied image so the RGB of transparent
// pixels survives the PNG round-trip (unlike image.NewRGBA, which premultiplies).
func makeNRGBAPNG(t *testing.T, w, h int, c color.NRGBA) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png encode: %v", err)
	}
	return buf.Bytes()
}

func makeJPEG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewYCbCr(image.Rect(0, 0, w, h), image.YCbCrSubsampleRatio420)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("jpeg encode: %v", err)
	}
	return buf.Bytes()
}

// writeFile writes data to dir/name and returns the absolute path.
func writeImage(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return p
}

func TestCheckImage_Format(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeImage(t, dir, "p.png", makePNG(t, 4, 4, color.RGBA{1, 2, 3, 255}))
	writeImage(t, dir, "j.jpg", makeJPEG(t, 8, 8))

	tests := []struct {
		name   string
		im     *spec.ImageAssert
		wantOK bool
	}{
		{"png match", &spec.ImageAssert{Path: "p.png", Format: "png"}, true},
		{"png mismatch", &spec.ImageAssert{Path: "p.png", Format: "jpeg"}, false},
		{"jpeg match", &spec.ImageAssert{Path: "j.jpg", Format: "jpeg"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Check(&spec.Assert{Image: tt.im}, nil, Env{Workdir: dir})
			if got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

func TestCheckImage_Dimensions(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeImage(t, dir, "img.png", makePNG(t, 100, 50, color.RGBA{10, 20, 30, 255}))

	tests := []struct {
		name   string
		im     *spec.ImageAssert
		wantOK bool
	}{
		{"exact width/height", &spec.ImageAssert{Path: "img.png", Width: intp(100), Height: intp(50)}, true},
		{"wrong width", &spec.ImageAssert{Path: "img.png", Width: intp(99)}, false},
		{"wrong height", &spec.ImageAssert{Path: "img.png", Height: intp(51)}, false},
		{"min ok", &spec.ImageAssert{Path: "img.png", MinWidth: intp(100), MinHeight: intp(50)}, true},
		{"min too high", &spec.ImageAssert{Path: "img.png", MinWidth: intp(101)}, false},
		{"max ok", &spec.ImageAssert{Path: "img.png", MaxWidth: intp(100), MaxHeight: intp(50)}, true},
		{"max too low", &spec.ImageAssert{Path: "img.png", MaxHeight: intp(49)}, false},
		{"range ok", &spec.ImageAssert{Path: "img.png", MinWidth: intp(50), MaxWidth: intp(200)}, true},
		{"min height too high", &spec.ImageAssert{Path: "img.png", MinHeight: intp(60)}, false},
		{"max width too low", &spec.ImageAssert{Path: "img.png", MaxWidth: intp(80)}, false},
		{"format + dims together", &spec.ImageAssert{Path: "img.png", Format: "png", Width: intp(100)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Check(&spec.Assert{Image: tt.im}, nil, Env{Workdir: dir})
			if got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

func TestCheckImage_Alpha(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeImage(t, dir, "rgba.png", makePNG(t, 4, 4, color.RGBA{0, 0, 0, 128}))
	writeImage(t, dir, "jpg.jpg", makeJPEG(t, 4, 4)) // YCbCr, no alpha

	if got := Check(&spec.Assert{Image: &spec.ImageAssert{Path: "rgba.png", Alpha: boolp(true)}}, nil, Env{Workdir: dir}); !got.OK {
		t.Errorf("rgba alpha=true should pass: %s", got.Hint)
	}
	if got := Check(&spec.Assert{Image: &spec.ImageAssert{Path: "rgba.png", Alpha: boolp(false)}}, nil, Env{Workdir: dir}); got.OK {
		t.Error("rgba alpha=false should fail")
	}
	if got := Check(&spec.Assert{Image: &spec.ImageAssert{Path: "jpg.jpg", Alpha: boolp(false)}}, nil, Env{Workdir: dir}); !got.OK {
		t.Errorf("jpeg alpha=false should pass: %s", got.Hint)
	}
}

// Regression for issue #13: an opaque truecolor PNG (and 24-bit BMP) is decoded
// as *image.RGBA, whose color model is alpha-bearing, so the old model-based
// hasAlpha false-positived alpha=true. A pixel scan reports alpha=false.
func TestCheckImage_Alpha_OpaqueTruecolor(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Opaque RGB PNG: Go's encoder emits color-type 2 (no alpha channel).
	writeImage(t, dir, "opaque.png", makePNG(t, 4, 4, color.RGBA{10, 20, 30, 255}))
	// Opaque 24-bit-style BMP.
	bmpBuf := &bytes.Buffer{}
	opaqueRGBA := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			opaqueRGBA.Set(x, y, color.RGBA{10, 20, 30, 255})
		}
	}
	if err := bmp.Encode(bmpBuf, opaqueRGBA); err != nil {
		t.Fatalf("bmp encode: %v", err)
	}
	writeImage(t, dir, "opaque.bmp", bmpBuf.Bytes())

	for _, name := range []string{"opaque.png", "opaque.bmp"} {
		if got := Check(&spec.Assert{Image: &spec.ImageAssert{Path: name, Alpha: boolp(false)}}, nil, Env{Workdir: dir}); !got.OK {
			t.Errorf("%s alpha=false should pass (opaque image has no transparency): %s", name, got.Hint)
		}
		if got := Check(&spec.Assert{Image: &spec.ImageAssert{Path: name, Alpha: boolp(true)}}, nil, Env{Workdir: dir}); got.OK {
			t.Errorf("%s alpha=true should fail (opaque image)", name)
		}
	}
}

func TestCheckImage_SimilarTo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	specDir := t.TempDir()

	base := makePNG(t, 20, 20, color.RGBA{100, 100, 100, 255})
	writeImage(t, specDir, "baseline.png", base)
	writeImage(t, dir, "same.png", base)
	writeImage(t, dir, "slightly.png", makePNG(t, 20, 20, color.RGBA{101, 100, 100, 255}))
	writeImage(t, dir, "different.png", makePNG(t, 20, 20, color.RGBA{255, 0, 0, 255}))
	writeImage(t, dir, "smaller.png", makePNG(t, 10, 10, color.RGBA{100, 100, 100, 255}))

	env := Env{Workdir: dir, SpecDir: specDir}

	tests := []struct {
		name   string
		im     *spec.ImageAssert
		wantOK bool
	}{
		{"identical exact", &spec.ImageAssert{Path: "same.png", SimilarTo: "baseline.png"}, true},
		{"slight within tolerance", &spec.ImageAssert{Path: "slightly.png", SimilarTo: "baseline.png", MaxDiff: floatp(0.01)}, true},
		{"slight exceeds zero tolerance", &spec.ImageAssert{Path: "slightly.png", SimilarTo: "baseline.png"}, false},
		{"different fails", &spec.ImageAssert{Path: "different.png", SimilarTo: "baseline.png", MaxDiff: floatp(0.1)}, false},
		{"dimension mismatch", &spec.ImageAssert{Path: "smaller.png", SimilarTo: "baseline.png"}, false},
		{"missing baseline", &spec.ImageAssert{Path: "same.png", SimilarTo: "nope.png"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Check(&spec.Assert{Image: tt.im}, nil, env)
			if got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

// TestCheckImage_Conjunctive verifies that constraints combine: when similar_to
// is set alongside format/dimension constraints, all must hold.
func TestCheckImage_Conjunctive(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	specDir := t.TempDir()
	base := makePNG(t, 12, 12, color.RGBA{50, 60, 70, 255})
	writeImage(t, specDir, "baseline.png", base)
	writeImage(t, dir, "out.png", base)

	env := Env{Workdir: dir, SpecDir: specDir}

	// All constraints satisfied → pass.
	ok := Check(&spec.Assert{Image: &spec.ImageAssert{
		Path: "out.png", Format: "png", Width: intp(12), Height: intp(12), SimilarTo: "baseline.png",
	}}, nil, env)
	if !ok.OK {
		t.Errorf("all-satisfied conjunction should pass: %s", ok.Hint)
	}

	// A wrong dimension must fail even though format and similar_to would pass —
	// the dimension check must not be skipped.
	bad := Check(&spec.Assert{Image: &spec.ImageAssert{
		Path: "out.png", Format: "png", Width: intp(99), SimilarTo: "baseline.png",
	}}, nil, env)
	if bad.OK {
		t.Error("wrong width combined with similar_to should fail")
	}

	// A wrong format must fail before similar_to is even consulted.
	badFmt := Check(&spec.Assert{Image: &spec.ImageAssert{
		Path: "out.png", Format: "jpeg", SimilarTo: "baseline.png",
	}}, nil, env)
	if badFmt.OK {
		t.Error("wrong format combined with similar_to should fail")
	}
}

func TestCheckImage_SimilarToAbsolutePath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	base := makePNG(t, 8, 8, color.RGBA{5, 5, 5, 255})
	writeImage(t, dir, "out.png", base)
	abs := writeImage(t, dir, "ref.png", base)

	got := Check(&spec.Assert{Image: &spec.ImageAssert{Path: "out.png", SimilarTo: abs}}, nil, Env{Workdir: dir})
	if !got.OK {
		t.Errorf("absolute baseline path should resolve: %s", got.Hint)
	}
}

func TestCheckImage_Errors(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeImage(t, dir, "notimage.png", []byte("this is not an image"))

	tests := []struct {
		name string
		im   *spec.ImageAssert
	}{
		{"missing file", &spec.ImageAssert{Path: "missing.png", Format: "png"}},
		{"undecodable for dims", &spec.ImageAssert{Path: "notimage.png", Width: intp(10)}},
		{"undecodable for similar", &spec.ImageAssert{Path: "notimage.png", SimilarTo: "notimage.png"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Check(&spec.Assert{Image: tt.im}, nil, Env{Workdir: dir, SpecDir: dir})
			if got.OK {
				t.Error("expected failure")
			}
		})
	}
}

// TestCheckImage_SimilarTo_TransparentColorDiff guards against premultiplied-
// alpha masking: two images that are fully transparent but carry different RGB
// values must still be detected as different (non-premultiplied comparison).
func TestCheckImage_SimilarTo_TransparentColorDiff(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	specDir := t.TempDir()
	// Two semi-transparent (alpha 64) images with very different RGB. Under
	// premultiplied-alpha comparison the diff is attenuated to ~0.12 and would
	// wrongly pass max_diff:0.2; a correct non-premultiplied comparison sees ~0.47.
	writeImage(t, specDir, "baseline.png", makeNRGBAPNG(t, 8, 8, color.NRGBA{200, 200, 200, 64}))
	writeImage(t, dir, "out.png", makeNRGBAPNG(t, 8, 8, color.NRGBA{40, 40, 40, 64}))

	got := Check(&spec.Assert{Image: &spec.ImageAssert{Path: "out.png", SimilarTo: "baseline.png", MaxDiff: floatp(0.2)}},
		nil, Env{Workdir: dir, SpecDir: specDir})
	if got.OK {
		t.Error("semi-transparent images with different RGB must not be masked by alpha premultiplication")
	}
}

func TestIsBMP(t *testing.T) {
	t.Parallel()
	// A real BMP with a BITMAPINFOHEADER (size 40) at offset 14.
	good := make([]byte, 18)
	good[0], good[1] = 'B', 'M'
	good[14] = 40
	if !isBMP(good) {
		t.Error("valid BM + DIB size 40 should be detected as bmp")
	}
	// "BM" prefix but an implausible DIB header size is rejected.
	bad := make([]byte, 18)
	bad[0], bad[1] = 'B', 'M'
	bad[14] = 7
	if isBMP(bad) {
		t.Error("BM with bogus DIB header size should not be bmp")
	}
	if isBMP([]byte("BM")) {
		t.Error("too-short BM should not be bmp")
	}
}

func TestDetectImageFormat(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	gifBuf := &bytes.Buffer{}
	if err := gif.Encode(gifBuf, image.NewPaletted(image.Rect(0, 0, 2, 2), color.Palette{color.Black, color.White}), nil); err != nil {
		t.Fatalf("gif: %v", err)
	}
	bmpBuf := &bytes.Buffer{}
	if err := bmp.Encode(bmpBuf, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatalf("bmp: %v", err)
	}
	tiffBuf := &bytes.Buffer{}
	if err := tiff.Encode(tiffBuf, image.NewRGBA(image.Rect(0, 0, 2, 2)), nil); err != nil {
		t.Fatalf("tiff: %v", err)
	}

	cases := map[string][]byte{
		"png":  makePNG(t, 2, 2, color.White),
		"jpeg": makeJPEG(t, 8, 8),
		"gif":  gifBuf.Bytes(),
		"bmp":  bmpBuf.Bytes(),
		"tiff": tiffBuf.Bytes(),
		"webp": append([]byte("RIFF\x00\x00\x00\x00WEBP"), make([]byte, 8)...),
		"avif": append([]byte{0, 0, 0, 0x20}, append([]byte("ftypavif"), make([]byte, 16)...)...),
		"svg":  []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"></svg>`),
	}
	for want, data := range cases {
		if got := detectImageFormat(data); got != want {
			t.Errorf("detectImageFormat(%s) = %q, want %q", want, got, want)
		}
	}
	if got := detectImageFormat([]byte("random bytes here")); got != "" {
		t.Errorf("unknown format = %q, want empty", got)
	}

	// Format assertion against a detected-but-undecodable SVG passes without decoding.
	svg := writeImage(t, dir, "x.svg", cases["svg"])
	_ = svg
	got := Check(&spec.Assert{Image: &spec.ImageAssert{Path: "x.svg", Format: "svg"}}, nil, Env{Workdir: dir})
	if !got.OK {
		t.Errorf("svg format-only assert should pass: %s", got.Hint)
	}
}
