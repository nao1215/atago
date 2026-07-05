package loader

import (
	"fmt"
	"path"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/nao1215/atago/internal/spec"
)

var validImageFormat = map[string]bool{
	"png": true, "jpeg": true, "gif": true, "webp": true,
	"bmp": true, "tiff": true, "avif": true, "svg": true,
}

func validateImage(add func(string, ...any), where string, im *spec.ImageAssert) {
	if im.Path == "" {
		add("%s.path is required", where)
	}
	n := 0
	if im.Format != "" {
		n++
		if !validImageFormat[im.Format] {
			add("%s.format %q is invalid (want png/jpeg/gif/webp/bmp/tiff/avif/svg)", where, im.Format)
		}
	}
	for _, d := range []*int{im.Width, im.Height, im.MinWidth, im.MaxWidth, im.MinHeight, im.MaxHeight} {
		if d != nil {
			n++
			if *d < 0 {
				add("%s: dimensions must be >= 0 (got %d)", where, *d)
			}
		}
	}
	if im.MinWidth != nil && im.MaxWidth != nil && *im.MinWidth > *im.MaxWidth {
		add("%s: min_width %d exceeds max_width %d", where, *im.MinWidth, *im.MaxWidth)
	}
	if im.MinHeight != nil && im.MaxHeight != nil && *im.MinHeight > *im.MaxHeight {
		add("%s: min_height %d exceeds max_height %d", where, *im.MinHeight, *im.MaxHeight)
	}
	if im.Alpha != nil {
		n++
	}
	if im.SimilarTo != "" {
		n++
	}
	if im.MaxDiff != nil {
		if im.SimilarTo == "" {
			add("%s.max_diff requires similar_to", where)
		}
		if *im.MaxDiff < 0 || *im.MaxDiff > 1 {
			add("%s.max_diff must be between 0 and 1 (got %g)", where, *im.MaxDiff)
		}
	}
	if n == 0 {
		add("%s: must set at least one of format/width/height/min_width/max_width/min_height/max_height/alpha/similar_to", where)
	}
	// AVIF and SVG cannot be decoded in pure Go, so only their format can be
	// asserted; reject measurement constraints up front for a clear error.
	if im.Format == "avif" || im.Format == "svg" {
		measures := im.Width != nil || im.Height != nil ||
			im.MinWidth != nil || im.MaxWidth != nil ||
			im.MinHeight != nil || im.MaxHeight != nil ||
			im.Alpha != nil || im.SimilarTo != ""
		if measures {
			add("%s: format %q cannot be measured (only format may be asserted for avif/svg)", where, im.Format)
		}
	}
}

// validatePDF checks a PDF assertion (#73): a path plus at least one constraint,
// sane page bounds, known metadata fields, and a well-formed text matcher.
func validatePDF(add func(string, ...any), where string, p *spec.PDFAssert) {
	if p.Path == "" {
		add("%s.path is required", where)
	}
	n := 0
	for _, c := range []*int{p.Pages, p.MinPages, p.MaxPages} {
		if c != nil {
			n++
			if *c < 0 {
				add("%s: page counts must be >= 0 (got %d)", where, *c)
			}
		}
	}
	if p.MinPages != nil && p.MaxPages != nil && *p.MinPages > *p.MaxPages {
		add("%s: min_pages %d exceeds max_pages %d", where, *p.MinPages, *p.MaxPages)
	}
	if len(p.Metadata) > 0 {
		n++
		for k := range p.Metadata {
			if !validPDFMetaField[strings.ToLower(k)] {
				add("%s.metadata: unknown field %q (want title/author/subject/keywords/creator/producer)", where, k)
			}
		}
	}
	if p.Text != nil {
		n++
		validateStream(add, where+".text", p.Text)
	}
	if n == 0 {
		add("%s: must set at least one of pages/min_pages/max_pages/metadata/text", where)
	}
}

var validPDFMetaField = map[string]bool{
	"title": true, "author": true, "subject": true,
	"keywords": true, "creator": true, "producer": true,
}

// validateDir checks a directory/tree assertion (#74): a path plus at least one
// constraint, with sane count bounds. Every set field is an independent
// constraint (like image), so no one-of rule applies beyond requiring at least
// one.
func validateDir(add func(string, ...any), where string, d *spec.DirAssert) {
	if d.Path == "" {
		add("%s.path is required", where)
	}
	n := 0
	if d.Exists != nil {
		n++
	}
	if len(d.Contains) > 0 {
		n++
	}
	if len(d.NotContains) > 0 {
		n++
	}
	for _, c := range []*int{d.Count, d.MinCount, d.MaxCount} {
		if c != nil {
			n++
			if *c < 0 {
				add("%s: counts must be >= 0 (got %d)", where, *c)
			}
		}
	}
	if d.Glob != "" {
		n++
	}
	if d.MinCount != nil && d.MaxCount != nil && *d.MinCount > *d.MaxCount {
		add("%s: min_count %d exceeds max_count %d", where, *d.MinCount, *d.MaxCount)
	}
	// Tree snapshot rules (#25): the golden manifest IS the whole assertion,
	// so it composes only with ignore; the matcher family needs recursive or
	// the historical single-level semantics.
	if d.Snapshot != "" {
		if n > 0 || d.Exists != nil {
			add("%s: snapshot cannot be combined with the matcher family (exists/contains/not_contains/count/glob) — the manifest already pins the whole tree", where)
		}
		if d.Recursive {
			add("%s: recursive is implied by snapshot; drop it", where)
		}
	} else {
		if d.Recursive && n == 0 {
			add("%s: recursive needs at least one of contains/not_contains/count/min_count/max_count/glob", where)
		}
		if len(d.Ignore) > 0 && !d.Recursive {
			add("%s: ignore only applies to recursive or snapshot assertions", where)
		}
	}
	for _, pat := range d.Ignore {
		trimmed := strings.TrimSuffix(pat, "/**")
		if _, err := path.Match(trimmed, "probe"); err != nil {
			add("%s.ignore %q is not a valid glob: %v", where, pat, err)
		}
	}
	if n == 0 && d.Snapshot == "" {
		add("%s: must set at least one of exists/contains/not_contains/count/min_count/max_count/glob/snapshot", where)
	}
}

// validateChanges checks a workdir-delta assertion (#70): at least one of
// created/modified/deleted set, and every entry a workdir-relative, confined
// path or a valid /-glob. Entries are compared with forward slashes at check
// time, so confinement is validated in the same /-separated space.
func validateChanges(add func(string, ...any), where string, c *spec.ChangesAssert) {
	if c.Created == nil && c.Modified == nil && c.Deleted == nil {
		add("%s: set at least one of created/modified/deleted (use [] to assert a category changed nothing)", where)
		return
	}
	for _, cat := range []struct {
		name    string
		entries *spec.StringList
	}{
		{"created", c.Created},
		{"modified", c.Modified},
		{"deleted", c.Deleted},
	} {
		if cat.entries == nil {
			continue
		}
		for _, entry := range []string(*cat.entries) {
			field := fmt.Sprintf("%s.%s %q", where, cat.name, entry)
			switch {
			case entry == "":
				add("%s must be a non-empty workdir-relative path or glob", field)
			case strings.HasPrefix(entry, "/"):
				add("%s must be workdir-relative, not absolute", field)
			case pathEscapesWorkdir(entry):
				add("%s escapes the scenario workdir (no ../ traversal)", field)
			default:
				// The entry doubles as a doublestar glob at check time (single
				// `*` stays single-level; `**` crosses `/`); reject a malformed
				// pattern here rather than silently matching nothing.
				if !doublestar.ValidatePattern(entry) {
					add("%s is not a valid glob: bad pattern syntax", field)
				}
			}
		}
	}
}

// pathEscapesWorkdir reports whether a /-separated relative entry would escape
// the workdir root via ../ traversal, using the same containment rule as
// security.ResolveWorkdirPath but in the loader's forward-slash space (#70).
func pathEscapesWorkdir(entry string) bool {
	cleaned := path.Clean(entry)
	return cleaned == ".." || strings.HasPrefix(cleaned, "../")
}
