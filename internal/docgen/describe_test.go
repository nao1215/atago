package docgen

import (
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/spectest"
)

// TestDescribeTarget_CoversEveryAssertTarget walks spec.AllAssertTargets and
// proves the doc renderer has a case for each one. The switch's default returns
// the bare target name, which reads as a heading with no sentence — a new target
// would silently publish that instead of describing what the scenario
// guarantees. The Assert values are minimal (the target's field allocated, no
// matcher set), so this covers the dispatch, not the phrasing.
func TestDescribeTarget_CoversEveryAssertTarget(t *testing.T) {
	t.Parallel()
	for _, target := range spec.AllAssertTargets() {
		got := describeTarget(spectest.AssertForTarget(target), target)
		if got == "" {
			t.Errorf("target %q renders as an empty bullet", target)
		}
		if got == string(target) {
			t.Errorf("target %q fell through to the default branch (rendered as the bare name)", target)
		}
	}
}

// TestCode_QuotesInvisibleValues proves a spec-supplied value the reader cannot
// see is rendered as a Go-quoted literal rather than raw. `not_contains: "\r"`
// is a legitimate assertion — it is how a spec guards a golden against a stray
// carriage return — and rendering it raw published an inline-code span that read
// as empty and put a bare CR into the committed Markdown. Printable values,
// including non-ASCII, must keep their exact authored form, so the vast majority
// of the generated docs are untouched by this rule.
func TestCode_QuotesInvisibleValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "carriage return", in: "\r", want: "`\"\\r\"`"},
		{name: "tab inside text", in: "a\tb", want: "`\"a\\tb\"`"},
		{name: "embedded newline", in: "one\ntwo", want: "`\"one\\ntwo\"`"},
		{name: "escape byte", in: "\x1b[0m", want: "`\"\\x1b[0m\"`"},
		{name: "plain ascii", in: "PASSED", want: "`PASSED`"},
		{name: "non-ascii stays literal", in: "日本語", want: "`日本語`"},
		{name: "punctuation stays literal", in: "127.0.0.1:<port>", want: "`127.0.0.1:<port>`"},
		{name: "empty", in: "", want: "``"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := code(tt.in); got != tt.want {
				t.Errorf("code(%q) = %s, want %s", tt.in, got, tt.want)
			}
		})
	}
}

// TestInlineCode_SurvivesEmbeddedBackticks proves a value carrying backticks is
// published as a span the reader sees whole. Asserted output quotes commands —
// a bats failure reports `false' failed — and wrapping that in single ticks
// ended the span at the value's own tick, spilling the rest of the assertion
// into the page as prose. The delimiter now grows past the longest run inside
// the value, and the space padding CommonMark strips keeps a leading or
// trailing tick inside the span.
func TestInlineCode_SurvivesEmbeddedBackticks(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "no backtick keeps the single tick", in: "ok 1 passes", want: "`ok 1 passes`"},
		{name: "one backtick inside", in: "`false' failed", want: "`` `false' failed ``"},
		{name: "backtick at both ends", in: "`cmd`", want: "`` `cmd` ``"},
		{name: "longest run wins", in: "a ``b`` c", want: "``` a ``b`` c ```"},
		{name: "backtick only", in: "`", want: "`` ` ``"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := inlineCode(tt.in); got != tt.want {
				t.Errorf("inlineCode(%q) = %s, want %s", tt.in, got, tt.want)
			}
			if got := code(tt.in); got != tt.want {
				t.Errorf("code(%q) = %s, want %s", tt.in, got, tt.want)
			}
		})
	}
}

// TestCodeList_QuotesInvisibleValues carries the same rule through the
// contains/not_contains renderer, the path an assertion on a control character
// actually takes.
func TestCodeList_QuotesInvisibleValues(t *testing.T) {
	t.Parallel()
	got := codeList(spec.StringList{"ok", "\r"})
	want := "`ok`, `\"\\r\"`"
	if got != want {
		t.Errorf("codeList = %s, want %s", got, want)
	}
}

// TestCodeRegex_KeepsThePatternReadable proves a regex pattern keeps its
// authored backslashes and gets its invisible characters escaped. A pattern with
// a literal newline used to break its inline-code span across two lines, and
// quoting the whole pattern the way a plain value is quoted would double every
// backslash — `\.` published as `\\.` no longer reads as the regex the spec
// wrote.
func TestCodeRegex_KeepsThePatternReadable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "newline is escaped", in: "^[0-9a-f]{32}\n?$", want: "`/^[0-9a-f]{32}\\n?$/`"},
		{name: "backslash is not doubled", in: `^.+/src/a\.txt\n?$`, want: "`/^.+/src/a\\.txt\\n?$/`"},
		{name: "plain pattern is untouched", in: "^v[0-9]+", want: "`/^v[0-9]+/`"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := codeRegex(tt.in); got != tt.want {
				t.Errorf("codeRegex(%q) = %s, want %s", tt.in, got, tt.want)
			}
		})
	}
	if strings.Contains(codeRegex("a\nb"), "\n") {
		t.Error("codeRegex left a raw newline in the inline-code span")
	}
}
