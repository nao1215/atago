package assert

import (
	"image/color"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

func drawn(t *testing.T, sizes ...[2]int) []runner.TerminalImage {
	t.Helper()
	out := make([]runner.TerminalImage, len(sizes))
	for i, s := range sizes {
		out[i] = runner.TerminalImage{Width: s[0], Height: s[1], PNG: makePNG(t, s[0], s[1], color.White)}
	}
	return out
}

func checkImages(si *spec.ScreenImages, images []runner.TerminalImage, env Env) *CheckResult {
	return Check(&spec.Assert{Screen: &spec.ScreenAssert{Images: si}}, &runner.Result{IsPTY: true, Images: images}, env)
}

func TestCheckScreenImages_Counts(t *testing.T) {
	t.Parallel()
	two := drawn(t, [2]int{4, 2}, [2]int{8, 8})
	for _, tc := range []struct {
		name string
		si   spec.ScreenImages
		ok   bool
	}{
		{"exact count", spec.ScreenImages{Count: intp(2)}, true},
		{"wrong count", spec.ScreenImages{Count: intp(3)}, false},
		{"zero is a claim", spec.ScreenImages{Count: intp(0)}, false},
		{"min_count met", spec.ScreenImages{MinCount: intp(2)}, true},
		{"min_count missed", spec.ScreenImages{MinCount: intp(3)}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := checkImages(&tc.si, two, Env{})
			if got.OK != tc.ok {
				t.Fatalf("OK = %v, want %v (%s / %s)", got.OK, tc.ok, got.Actual, got.Hint)
			}
			if !got.OK && got.Actual != "2 image(s) on screen (1: 4x2, 2: 8x8)" {
				t.Fatalf("Actual = %q", got.Actual)
			}
		})
	}
}

// TestCheckScreenImages_ContainsMatchesAnyImage pins the order-free semantics:
// each entry is met by some drawn image, and every constraint of an entry must
// hold on the SAME image.
func TestCheckScreenImages_ContainsMatchesAnyImage(t *testing.T) {
	t.Parallel()
	images := drawn(t, [2]int{4, 2}, [2]int{8, 8})
	pass := spec.ScreenImages{Contains: []spec.ScreenImage{{Width: intp(8)}, {Width: intp(4), Height: intp(2)}}}
	if got := checkImages(&pass, images, Env{}); !got.OK {
		t.Fatalf("expected pass: %s", got.Hint)
	}
	// width 4 exists and height 8 exists, but not on one image.
	split := spec.ScreenImages{Contains: []spec.ScreenImage{{Width: intp(4), Height: intp(8)}}}
	got := checkImages(&split, images, Env{})
	if got.OK {
		t.Fatal("constraints spread over two images must not match")
	}
	if !strings.Contains(got.Desc, "width 4px, height 8px") {
		t.Fatalf("Desc = %q", got.Desc)
	}
	if len(got.ArtifactBlobs) != 2 || got.ArtifactBlobs[0].Role != "drawn-1" {
		t.Fatalf("every drawn image should be attached, got %+v", got.ArtifactBlobs)
	}
}

func TestCheckScreenImages_SimilarToUsesTheImageBaseline(t *testing.T) {
	t.Parallel()
	specDir := t.TempDir()
	writeImage(t, specDir, "white.png", makePNG(t, 4, 2, color.White))
	writeImage(t, specDir, "black.png", makePNG(t, 4, 2, color.Black))
	env := Env{SpecDir: specDir, Workdir: t.TempDir()}
	images := drawn(t, [2]int{4, 2})
	ok := spec.ScreenImages{Contains: []spec.ScreenImage{{SimilarTo: "white.png"}}}
	if got := checkImages(&ok, images, env); !got.OK {
		t.Fatalf("expected pass: %s", got.Hint)
	}
	bad := spec.ScreenImages{Contains: []spec.ScreenImage{{SimilarTo: "black.png"}}}
	got := checkImages(&bad, images, env)
	if got.OK || !strings.Contains(got.Hint, "differs from baseline") {
		t.Fatalf("expected a pixel-diff failure, got OK=%v hint=%q", got.OK, got.Hint)
	}
}

func TestCheckScreenImages_NoImagesExplainsGraphics(t *testing.T) {
	t.Parallel()
	got := checkImages(&spec.ScreenImages{MinCount: intp(1)}, nil, Env{})
	if got.OK || got.Actual != "no images on screen" || !strings.Contains(got.Hint, "graphics: kitty") {
		t.Fatalf("got OK=%v actual=%q hint=%q", got.OK, got.Actual, got.Hint)
	}
}

// TestCheckScreen_ImagesAlongsideText pins that images compose with the text
// matchers: both must hold.
func TestCheckScreen_ImagesAlongsideText(t *testing.T) {
	t.Parallel()
	res := &runner.Result{IsPTY: true, Screen: []byte("hello"), Images: drawn(t, [2]int{1, 1})}
	sa := &spec.ScreenAssert{StreamAssert: spec.StreamAssert{Contains: spec.StringList{"bye"}}, Images: &spec.ScreenImages{Count: intp(1)}}
	if got := Check(&spec.Assert{Screen: sa}, res, Env{}); got.OK {
		t.Fatal("a failing text matcher must fail the assert even when images hold")
	}
	sa.Contains = spec.StringList{"hello"}
	if got := Check(&spec.Assert{Screen: sa}, res, Env{}); !got.OK {
		t.Fatalf("expected pass: %s", got.Hint)
	}
}
