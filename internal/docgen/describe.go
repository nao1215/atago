package docgen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/markdown"
)

// codeList renders a contains/not_contains matcher argument. A single element
// is rendered as one inline-code span (byte-identical to the pre-list format); a
// list joins its elements with ", " so the generated doc shows every required
// (or forbidden) substring.
func codeList(subs spec.StringList) string {
	parts := make([]string, len(subs))
	for i, s := range subs {
		parts[i] = markdown.Code(s)
	}
	return strings.Join(parts, ", ")
}

// describeAsserts renders an assertion as one Markdown "Then" bullet per target.
// An assert may set several targets (exit_code + stdout + …); each is its own
// independent check, so each gets its own bullet.
func describeAsserts(a *spec.Assert) []string {
	targets := a.SetTargets()
	if len(targets) == 0 {
		return []string{"_(invalid assertion)_"}
	}
	bullets := make([]string, 0, len(targets))
	for _, t := range targets {
		bullets = append(bullets, describeTarget(a, t))
	}
	return bullets
}

// describeTarget renders a single assertion target as a Markdown "Then" bullet.
func describeTarget(a *spec.Assert, target spec.AssertTarget) string {
	switch target {
	case spec.AssertExitCode:
		if a.ExitCode.Not != nil {
			return fmt.Sprintf("exit code is not %s", markdown.Code(fmt.Sprint(*a.ExitCode.Not)))
		}
		if len(a.ExitCode.In) > 0 {
			codes := make([]string, len(a.ExitCode.In))
			for i, n := range a.ExitCode.In {
				codes[i] = markdown.Code(fmt.Sprint(n))
			}
			return "exit code is one of " + strings.Join(codes, ", ")
		}
		if a.ExitCode.Equals != nil {
			return fmt.Sprintf("exit code is %s", markdown.Code(fmt.Sprint(*a.ExitCode.Equals)))
		}
		return "exit code is checked"
	case spec.AssertMock:
		m := a.Mock
		desc := "mock " + markdown.Code(m.Name) + " received"
		if m.Method != "" || m.Path != "" {
			desc += " " + markdown.Code(strings.TrimSpace(strings.ToUpper(m.Method)+" "+m.Path))
		} else {
			desc += " a request"
		}
		if m.Count != nil {
			desc += fmt.Sprintf(" exactly %d time(s)", *m.Count)
		}
		return desc
	case spec.AssertScreen:
		return "rendered screen " + describeStream(a.Screen)
	case spec.AssertDuration:
		return "completes " + a.Duration.DescribeDuration()
	case spec.AssertChanges:
		return "the step changed exactly " + describeChanges(a.Changes)
	case spec.AssertStdout:
		return "stdout " + describeStream(a.Stdout)
	case spec.AssertStderr:
		return "stderr " + describeStream(a.Stderr)
	case spec.AssertFile:
		return "file " + describeFile(a.File)
	case spec.AssertImage:
		return "image " + describeImage(a.Image)
	case spec.AssertDir:
		return "dir " + describeDir(a.Dir)
	case spec.AssertPDF:
		return "pdf " + describePDF(a.PDF)
	case spec.AssertStatus:
		if a.Status != nil {
			return "HTTP status is " + markdown.Code(fmt.Sprint(*a.Status))
		}
		return "HTTP status is checked"
	case spec.AssertHeader:
		if a.Header != nil {
			return "header " + describeHeader(a.Header)
		}
		return "header is checked"
	case spec.AssertBody:
		return "body " + describeStream(a.Body)
	case spec.AssertRows:
		return "rows " + describeStream(a.Rows)
	case spec.AssertGRPCStatus:
		if a.GRPCStatus != nil {
			return "gRPC status is " + markdown.Code(fmt.Sprint(*a.GRPCStatus))
		}
		return "gRPC status is checked"
	case spec.AssertMessage:
		return "message " + describeStream(a.Message)
	case spec.AssertValue:
		return "value " + describeStream(a.Value)
	default:
		return string(target)
	}
}

// describeChanges renders a workdir-delta assertion (#70) as a compact phrase
// listing each set category. `modified: []` renders as "modified nothing".
func describeChanges(c *spec.ChangesAssert) string {
	var parts []string
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
		if len(*cat.entries) == 0 {
			parts = append(parts, cat.name+" nothing")
			continue
		}
		parts = append(parts, cat.name+" "+codeList(*cat.entries))
	}
	if len(parts) == 0 {
		return "nothing"
	}
	return strings.Join(parts, ", ")
}

func describeHeader(h *spec.HeaderMatch) string {
	switch {
	case h.Contains != nil:
		return markdown.Code(h.Name) + " contains " + markdown.Code(*h.Contains)
	case h.Equals != nil:
		return markdown.Code(h.Name) + " equals " + markdown.Code(*h.Equals)
	// A header `matches` regexp is a real matcher (the natural shape for auth
	// headers like "^Bearer "); without this case it fell to the generic "is
	// checked" and the documented constraint vanished from the doc, so a reviewer
	// could not see it. Kept in step with explain.describeHeader.
	case h.Matches != nil:
		return markdown.Code(h.Name) + " matches " + markdown.Code("/"+*h.Matches+"/")
	default:
		return markdown.Code(h.Name) + " is checked"
	}
}

func describeImage(im *spec.ImageAssert) string {
	parts := imageConstraints(im)
	if len(parts) == 0 {
		return markdown.Code(im.Path) + " is checked"
	}
	return markdown.Code(im.Path) + " " + strings.Join(parts, ", ")
}

// imageConstraints renders each set image constraint as a short phrase.
func imageConstraints(im *spec.ImageAssert) []string {
	var parts []string
	if im.Format != "" {
		parts = append(parts, "is "+markdown.Code(im.Format))
	}
	if im.Width != nil {
		parts = append(parts, fmt.Sprintf("width %d", *im.Width))
	}
	if im.Height != nil {
		parts = append(parts, fmt.Sprintf("height %d", *im.Height))
	}
	if im.MinWidth != nil {
		parts = append(parts, fmt.Sprintf("width >= %d", *im.MinWidth))
	}
	if im.MaxWidth != nil {
		parts = append(parts, fmt.Sprintf("width <= %d", *im.MaxWidth))
	}
	if im.MinHeight != nil {
		parts = append(parts, fmt.Sprintf("height >= %d", *im.MinHeight))
	}
	if im.MaxHeight != nil {
		parts = append(parts, fmt.Sprintf("height <= %d", *im.MaxHeight))
	}
	if im.Alpha != nil {
		if *im.Alpha {
			parts = append(parts, "has alpha")
		} else {
			parts = append(parts, "has no alpha")
		}
	}
	if im.SimilarTo != "" {
		parts = append(parts, "similar to "+markdown.Code(im.SimilarTo))
	}
	return parts
}

// describeDir renders a directory/tree assertion (#74) as a compact phrase
// listing each set constraint.
func describeDir(d *spec.DirAssert) string {
	parts := dirConstraints(d)
	if len(parts) == 0 {
		return markdown.Code(d.Path) + " is checked"
	}
	return markdown.Code(d.Path) + " " + strings.Join(parts, ", ")
}

func dirConstraints(d *spec.DirAssert) []string {
	var parts []string
	if d.Exists != nil {
		if *d.Exists {
			parts = append(parts, "exists")
		} else {
			parts = append(parts, "does not exist")
		}
	}
	for _, c := range d.Contains {
		parts = append(parts, "contains "+markdown.Code(c))
	}
	for _, c := range d.NotContains {
		parts = append(parts, "does not contain "+markdown.Code(c))
	}
	if d.Count != nil {
		parts = append(parts, fmt.Sprintf("has %d entries", *d.Count))
	}
	if d.MinCount != nil {
		parts = append(parts, fmt.Sprintf("has >= %d entries", *d.MinCount))
	}
	if d.MaxCount != nil {
		parts = append(parts, fmt.Sprintf("has <= %d entries", *d.MaxCount))
	}
	if d.Glob != "" {
		parts = append(parts, "matches glob "+markdown.Code(d.Glob))
	}
	if d.Snapshot != "" {
		parts = append(parts, "tree matches snapshot "+markdown.Code(d.Snapshot))
	}
	if d.Recursive {
		parts = append(parts, "(recursive)")
	}
	if len(d.Ignore) > 0 {
		parts = append(parts, "ignoring "+strings.Join(d.Ignore, ", "))
	}
	return parts
}

// describePDF renders a PDF assertion (#73) as a compact phrase.
func describePDF(p *spec.PDFAssert) string {
	var parts []string
	if p.Pages != nil {
		parts = append(parts, fmt.Sprintf("%d pages", *p.Pages))
	}
	if p.MinPages != nil {
		parts = append(parts, fmt.Sprintf(">= %d pages", *p.MinPages))
	}
	if p.MaxPages != nil {
		parts = append(parts, fmt.Sprintf("<= %d pages", *p.MaxPages))
	}
	for _, k := range sortedMapKeys(p.Metadata) {
		parts = append(parts, fmt.Sprintf("%s contains %s", k, markdown.Code(p.Metadata[k])))
	}
	if p.Text != nil {
		parts = append(parts, "text "+describeStream(p.Text))
	}
	if len(parts) == 0 {
		return markdown.Code(p.Path) + " is checked"
	}
	return markdown.Code(p.Path) + " " + strings.Join(parts, ", ")
}

// sortedMapKeys returns a string map's keys in sorted order for deterministic
// rendering.
func sortedMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func describeStream(s *spec.StreamAssert) string {
	switch {
	case s.Empty != nil:
		if *s.Empty {
			return "is empty"
		}
		return "is not empty"
	case s.Contains != nil:
		return "contains " + codeList(s.Contains)
	case s.NotContains != nil:
		return "does not contain " + codeList(s.NotContains)
	case s.Matches != nil:
		return "matches " + markdown.Code("/"+*s.Matches+"/")
	case s.NotMatches != nil:
		return "does not match " + markdown.Code("/"+*s.NotMatches+"/")
	case s.Equals != nil:
		return "equals an exact value"
	case s.NotEquals != nil:
		return "does not equal an exact value"
	case s.JSON != nil:
		return "at " + markdown.Code(s.JSON.Path) + " " + jsonMatcher(s.JSON)
	case s.YAML != nil:
		return "YAML at " + markdown.Code(s.YAML.Path) + " " + jsonMatcher(s.YAML)
	case s.Snapshot != "":
		return "matches snapshot " + markdown.Code(s.Snapshot)
	default:
		return "is checked"
	}
}

func describeFile(f *spec.FileAssert) string {
	switch {
	case f.Exists != nil:
		if *f.Exists {
			return markdown.Code(f.Path) + " exists"
		}
		return markdown.Code(f.Path) + " does not exist"
	case f.Contains != nil:
		return markdown.Code(f.Path) + " contains " + codeList(f.Contains)
	case f.JSON != nil:
		return markdown.Code(f.Path) + " at " + markdown.Code(f.JSON.Path) + " " + jsonMatcher(f.JSON)
	case f.Snapshot != "":
		return markdown.Code(f.Path) + " matches snapshot " + markdown.Code(f.Snapshot)
	default:
		return markdown.Code(f.Path) + " is checked"
	}
}

func jsonMatcher(j *spec.JSONAssert) string {
	switch {
	case j.Equals != nil:
		return "equals " + markdown.Code(fmt.Sprint(j.Equals))
	case j.Matches != nil:
		return "matches " + markdown.Code("/"+*j.Matches+"/")
	case j.Length != nil:
		return fmt.Sprintf("has length %d", *j.Length)
	case j.Gt != nil:
		return "is " + markdown.Code(fmt.Sprintf("> %v", *j.Gt))
	case j.Gte != nil:
		return "is " + markdown.Code(fmt.Sprintf(">= %v", *j.Gte))
	case j.Lt != nil:
		return "is " + markdown.Code(fmt.Sprintf("< %v", *j.Lt))
	case j.Lte != nil:
		return "is " + markdown.Code(fmt.Sprintf("<= %v", *j.Lte))
	default:
		return "is checked"
	}
}

func sortedEnvKeys(env map[string]string) []string {
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
