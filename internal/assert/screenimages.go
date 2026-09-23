package assert

import (
	"fmt"
	"strings"

	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// checkScreenImages evaluates `screen.images` against the images on a pty
// step's screen. Count bounds come first, then each `contains` entry must be
// met by at least one of them.
func checkScreenImages(si *spec.ScreenImages, images []runner.TerminalImage, env Env) *CheckResult {
	drawn := drawnImages(images)
	if si.Count != nil && len(images) != *si.Count {
		return &CheckResult{
			Desc:     fmt.Sprintf("assert screen shows %d image(s)", *si.Count),
			Expected: fmt.Sprintf("%d image(s)", *si.Count),
			Actual:   drawn,
			Hint:     noImagesHint(images),
		}
	}
	if si.MinCount != nil && len(images) < *si.MinCount {
		return &CheckResult{
			Desc:     fmt.Sprintf("assert screen shows at least %d image(s)", *si.MinCount),
			Expected: fmt.Sprintf("at least %d image(s)", *si.MinCount),
			Actual:   drawn,
			Hint:     noImagesHint(images),
		}
	}
	for i := range si.Contains {
		entry := &si.Contains[i]
		var last *CheckResult
		matched := false
		for j := range images {
			cr := checkImageData(entry.ImageAssert(fmt.Sprintf("image %d on screen", j+1)), images[j].PNG, env)
			if cr.OK {
				matched = true
				break
			}
			last = cr
		}
		if matched {
			continue
		}
		cr := &CheckResult{
			Desc:     "assert screen shows an image " + entry.Describe(),
			Expected: "at least one image on screen meeting every constraint",
			Actual:   drawn,
			Hint:     noImagesHint(images),
		}
		if last != nil {
			cr.Hint = "no image on screen meets it; the last one checked: " + last.Hint
		}
		// Every image on screen goes to --artifacts-dir, so a reviewer sees what
		// the program did draw instead of guessing from sizes.
		for j, im := range images {
			cr.ArtifactKind = "image"
			cr.ArtifactBlobs = append(cr.ArtifactBlobs, ArtifactBlob{Role: fmt.Sprintf("drawn-%d", j+1), Ext: "png", Data: im.PNG})
		}
		return cr
	}
	return pass("assert screen " + si.Describe())
}

// drawnImages lists the sizes of the images on screen for a failure message.
func drawnImages(images []runner.TerminalImage) string {
	if len(images) == 0 {
		return "no images on screen"
	}
	parts := make([]string, len(images))
	for i, im := range images {
		parts[i] = fmt.Sprintf("%d: %dx%d", i+1, im.Width, im.Height)
	}
	return fmt.Sprintf("%d image(s) on screen (%s)", len(images), strings.Join(parts, ", "))
}

func noImagesHint(images []runner.TerminalImage) string {
	if len(images) > 0 {
		return ""
	}
	return "images are recorded only by a pty step with graphics: kitty, only when the program transmits them directly (t=d), and only until it deletes them by id, number, id range, or all (a=d)"
}
