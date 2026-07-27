package assertdesc

import (
	"fmt"
	"sort"
	"strings"

	"github.com/nao1215/atago/internal/plural"
	"github.com/nao1215/atago/internal/spec"
)

type HeaderStyle struct {
	Name  func(string) string
	Value func(string) string
	Regex func(string) string
	Bare  func(string) string
}

func DescribeHeader(h *spec.HeaderMatch, style HeaderStyle) string {
	switch {
	case h.Contains != nil:
		return style.Name(h.Name) + " contains " + style.Value(*h.Contains)
	case h.Equals != nil:
		return style.Name(h.Name) + " equals " + style.Value(*h.Equals)
	case h.Matches != nil:
		return style.Name(h.Name) + " matches " + style.Regex(*h.Matches)
	default:
		return style.Bare(h.Name)
	}
}

type JSONStyle struct {
	Prefix  func(string) string
	Equals  func(any) string
	Matches func(string) string
	Length  func(int) string
	Compare func(string, any) string
	Default string
}

// WithPrefix returns a copy of the style that renders the path differently —
// the only thing that separates a YAML check from a JSON one. Copying the whole
// struct field by field to change one function meant a field added later was
// silently dropped from the copy.
func (s JSONStyle) WithPrefix(prefix func(string) string) JSONStyle {
	s.Prefix = prefix
	return s
}

// JSONValueText renders a json/yaml matcher's expected value for prose. Go
// prints a nil as "<nil>", which since #309 is a value a spec can legitimately
// assert (`equals: null`), and a Go-ism in a sentence about someone else's
// document. The failure messages already spell a JSON null the way the document
// does; the generated docs and `explain` now agree with them.
func JSONValueText(v any) string {
	if v == nil {
		return "null"
	}
	return fmt.Sprint(v)
}

func DescribeJSONChecks(list spec.JSONChecks, style JSONStyle) string {
	parts := make([]string, len(list))
	for i := range list {
		parts[i] = strings.TrimSpace(style.Prefix(list[i].Path) + " " + DescribeJSONMatcher(&list[i], style))
	}
	return strings.Join(parts, "; ")
}

func DescribeJSONMatcher(j *spec.JSONAssert, style JSONStyle) string {
	switch {
	case j.HasEquals():
		return style.Equals(j.Equals)
	case j.Matches != nil:
		return style.Matches(*j.Matches)
	case j.Length != nil:
		return style.Length(*j.Length)
	case j.Gt != nil:
		return style.Compare(">", *j.Gt)
	case j.Gte != nil:
		return style.Compare(">=", *j.Gte)
	case j.Lt != nil:
		return style.Compare("<", *j.Lt)
	case j.Lte != nil:
		return style.Compare("<=", *j.Lte)
	default:
		return style.Default
	}
}

type StreamStyle struct {
	List      func(spec.StringList) string
	Regex     func(string) string
	Equals    string
	NotEquals string
	JSON      JSONStyle
	YAML      JSONStyle
	Snapshot  func(string) string
	Line      func(int) string
	NoMatcher string
}

// DescribeStream renders every matcher a stream assertion sets, not just the
// first one: the text matchers compose (contains + not_contains + matches +
// not_matches all have to hold), so describing one of them would publish a
// weaker contract than the run enforces. A `line: N` selector is rendered as a
// prefix because it narrows what every matcher sees.
func DescribeStream(s *spec.StreamAssert, style StreamStyle) string {
	var parts []string
	if s.Empty != nil {
		if *s.Empty {
			parts = append(parts, "is empty")
		} else {
			parts = append(parts, "is not empty")
		}
	}
	if s.Contains != nil {
		parts = append(parts, "contains "+style.List(s.Contains))
	}
	if s.NotContains != nil {
		parts = append(parts, "does not contain "+style.List(s.NotContains))
	}
	if s.Matches != nil {
		parts = append(parts, "matches "+style.Regex(*s.Matches))
	}
	if s.NotMatches != nil {
		parts = append(parts, "does not match "+style.Regex(*s.NotMatches))
	}
	if s.Equals != nil {
		parts = append(parts, style.Equals)
	}
	if s.NotEquals != nil {
		parts = append(parts, style.NotEquals)
	}
	if len(s.JSON) > 0 {
		parts = append(parts, DescribeJSONChecks(s.JSON, style.JSON))
	}
	if len(s.YAML) > 0 {
		parts = append(parts, DescribeJSONChecks(s.YAML, style.YAML))
	}
	if s.Snapshot != "" {
		parts = append(parts, "matches snapshot "+style.Snapshot(s.Snapshot))
	}
	desc := style.NoMatcher
	if len(parts) > 0 {
		desc = strings.Join(parts, ", ")
	}
	// The line scope is part of what the assertion looks at, so it is named even
	// when no matcher survived — the reader still learns which line is in play.
	if s.Line != nil && style.Line != nil {
		return style.Line(*s.Line) + " " + desc
	}
	return desc
}

type FileStyle struct {
	Path       func(string) string
	List       func(spec.StringList) string
	JSON       JSONStyle
	Snapshot   func(string) string
	Checked    func(string) string
	ExactBytes string
}

// DescribeFile renders the single matcher a file assertion sets. Every matcher
// the loader accepts has a case here: one that fell through to Checked would
// publish "the file is checked" for an assertion that actually pins content or
// permissions.
func DescribeFile(f *spec.FileAssert, style FileStyle) string {
	switch {
	case f.Exists != nil:
		if *f.Exists {
			return style.Path(f.Path) + " exists"
		}
		return style.Path(f.Path) + " does not exist"
	case f.Contains != nil:
		return style.Path(f.Path) + " contains " + style.List(f.Contains)
	case f.NotContains != nil:
		return style.Path(f.Path) + " does not contain " + style.List(f.NotContains)
	case f.Executable != nil:
		if *f.Executable {
			return style.Path(f.Path) + " is executable"
		}
		return style.Path(f.Path) + " is not executable"
	case f.Equals != nil:
		return style.Path(f.Path) + " " + style.ExactBytes
	case f.EqualsFile != nil:
		return style.Path(f.Path) + " is byte-identical to " + style.Path(*f.EqualsFile)
	case len(f.JSON) > 0:
		return style.Path(f.Path) + " " + DescribeJSONChecks(f.JSON, style.JSON)
	case f.Snapshot != "":
		return style.Path(f.Path) + " matches snapshot " + style.Snapshot(f.Snapshot)
	default:
		return style.Checked(f.Path)
	}
}

type ImageStyle struct {
	Path      func(string) string
	Format    func(string) string
	SimilarTo func(string) string
	Checked   func(string) string
}

func DescribeImage(im *spec.ImageAssert, style ImageStyle) string {
	var parts []string
	if im.Format != "" {
		parts = append(parts, "is "+style.Format(im.Format))
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
		parts = append(parts, "similar to "+style.SimilarTo(im.SimilarTo))
	}
	if len(parts) == 0 {
		return style.Checked(im.Path)
	}
	return style.Path(im.Path) + " " + strings.Join(parts, ", ")
}

type DirStyle struct {
	Path    func(string) string
	Item    func(string) string
	Token   func(string) string
	Checked func(string) string
}

func DescribeDir(d *spec.DirAssert, style DirStyle) string {
	var parts []string
	if d.Exists != nil {
		if *d.Exists {
			parts = append(parts, "exists")
		} else {
			parts = append(parts, "does not exist")
		}
	}
	for _, c := range d.Contains {
		parts = append(parts, "contains "+style.Item(c))
	}
	for _, c := range d.NotContains {
		parts = append(parts, "does not contain "+style.Item(c))
	}
	if d.Count != nil {
		parts = append(parts, "has "+plural.Count(*d.Count, "entry", "entries"))
	}
	if d.MinCount != nil {
		parts = append(parts, "has >= "+plural.Count(*d.MinCount, "entry", "entries"))
	}
	if d.MaxCount != nil {
		parts = append(parts, "has <= "+plural.Count(*d.MaxCount, "entry", "entries"))
	}
	if d.Glob != "" {
		parts = append(parts, "matches glob "+style.Token(d.Glob))
	}
	if d.Snapshot != "" {
		parts = append(parts, "tree matches snapshot "+style.Token(d.Snapshot))
	}
	if d.Recursive {
		parts = append(parts, "(recursive)")
	}
	if len(d.Ignore) > 0 {
		parts = append(parts, "ignoring "+strings.Join(d.Ignore, ", "))
	}
	if len(parts) == 0 {
		return style.Checked(d.Path)
	}
	return style.Path(d.Path) + " " + strings.Join(parts, ", ")
}

type PDFStyle struct {
	Path    func(string) string
	Value   func(string) string
	Stream  func(*spec.StreamAssert) string
	Checked func(string) string
}

func DescribePDF(p *spec.PDFAssert, style PDFStyle) string {
	var parts []string
	if p.Pages != nil {
		parts = append(parts, plural.Count(*p.Pages, "page", "pages"))
	}
	if p.MinPages != nil {
		parts = append(parts, ">= "+plural.Count(*p.MinPages, "page", "pages"))
	}
	if p.MaxPages != nil {
		parts = append(parts, "<= "+plural.Count(*p.MaxPages, "page", "pages"))
	}
	keys := make([]string, 0, len(p.Metadata))
	for k := range p.Metadata {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s contains %s", k, style.Value(p.Metadata[k])))
	}
	if p.Text != nil {
		parts = append(parts, "text "+style.Stream(p.Text))
	}
	if len(parts) == 0 {
		return style.Checked(p.Path)
	}
	return style.Path(p.Path) + " " + strings.Join(parts, ", ")
}

type ChangesStyle struct {
	Entry func(string) string
	Join  string
}

func DescribeChanges(c *spec.ChangesAssert, style ChangesStyle) string {
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
		formatted := make([]string, len(*cat.entries))
		for i, entry := range *cat.entries {
			formatted[i] = style.Entry(entry)
		}
		parts = append(parts, cat.name+" "+strings.Join(formatted, ", "))
	}
	if len(parts) == 0 {
		return "nothing"
	}
	out := strings.Join(parts, style.Join)
	if len(c.Ignore) > 0 {
		// Named explicitly: a generated page that showed only the categories would
		// advertise an exhaustive claim the assertion no longer makes over those
		// paths.
		formatted := make([]string, len(c.Ignore))
		for i, pat := range c.Ignore {
			formatted[i] = style.Entry(pat)
		}
		out += style.Join + "ignoring " + strings.Join(formatted, ", ")
	}
	return out
}

type MockStyle struct {
	Name  func(string) string
	Route func(string) string
	Count func(int) string
}

func DescribeMock(m *spec.MockAssert, style MockStyle) string {
	desc := "mock " + style.Name(m.Name)
	if m.Method != "" || m.Path != "" {
		desc += " received " + style.Route(strings.TrimSpace(strings.ToUpper(m.Method)+" "+m.Path))
	} else {
		desc += " received a request"
	}
	if m.Count != nil {
		desc += style.Count(*m.Count)
	}
	return desc
}
