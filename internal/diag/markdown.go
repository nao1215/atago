package diag

import (
	"fmt"
	"strings"
)

// families describes each code range on the generated reference page, in the
// order the page presents them.
var families = []struct {
	exit    int
	label   string
	meaning string
}{
	{2, "ATG2xxx", "spec error — the file could not be parsed, or does not describe a runnable suite"},
	{3, "ATG3xxx", "configuration error — the command line, its flags, or the spec files it selected"},
	{4, "ATG4xxx", "execution error — a step could not be carried out"},
	{5, "ATG5xxx", "internal error — a bug in atago"},
	{6, "ATG6xxx", "security policy violation — atago refused an operation on policy grounds"},
}

// Markdown renders the published error reference from the registry. The
// committed `doc/errors.md` is this output, and a drift test fails when the two
// disagree, so a code cannot be added, reworded, or removed without its
// documentation following.
func Markdown() []byte {
	var b strings.Builder

	all := All()

	b.WriteString("# Error codes\n\n")
	b.WriteString("A coded error carries a name that can be searched for, linked to, and branched on, so the message beside it stays free to improve without breaking anyone. Assertion failures are not errors in this sense and carry no code: a spec that fails is a spec doing its job, and its message already says what was expected and what happened.\n\n")
	b.WriteString("A code is `ATG` followed by four digits, and the first of those digits is the exit status the run produced. Reading `ATG2103` tells you the process exited 2 before it tells you anything else.\n\n")
	b.WriteString("Codes are being assigned one family at a time. The table says which families carry them today; an error from a family not yet covered still exits the same way and still says what went wrong, it just has no code to look up here yet.\n\n")

	b.WriteString("| Codes | Exit | Meaning | Assigned |\n|---|---|---|---|\n")
	for _, f := range families {
		assigned := "not yet"
		if n := len(entriesOf(all, f.exit)); n > 0 {
			assigned = fmt.Sprintf("%d codes", n)
		}
		fmt.Fprintf(&b, "| `%s` | %d | %s | %s |\n", f.label, f.exit, f.meaning, assigned)
	}
	b.WriteString("\n")
	b.WriteString("`ATG1xxx` is never assigned. Exit 1 means one or more scenarios failed, which is a result rather than an error.\n\n")
	b.WriteString("Codes are grouped by what you have to fix rather than by where in atago the error was raised, so one code can be reported from several places when the answer is the same in all of them.\n\n")
	b.WriteString("Look a code up from the terminal with `atago explain ATG2201`, which prints this same text without a browser.\n\n")

	b.WriteString("## Every code\n\n")
	b.WriteString("| Code | Meaning | Exit | Since |\n|---|---|---|---|\n")
	for _, e := range all {
		fmt.Fprintf(&b, "| [`%s`](#%s) | %s | %d | %s |\n", e.Code, anchor(e), e.Summary, e.Code.ExitCode(), e.Since)
	}
	b.WriteString("\n")

	for _, f := range families {
		section := entriesOf(all, f.exit)
		if len(section) == 0 {
			continue
		}
		fmt.Fprintf(&b, "## %s — exit %d\n\n", f.label, f.exit)
		for _, e := range section {
			fmt.Fprintf(&b, "### %s — %s\n\n", e.Code, e.Summary)
			b.WriteString(e.Detail)
			b.WriteString("\n\n")
			fmt.Fprintf(&b, "Fix: %s\n\n", e.Fix)
			fmt.Fprintf(&b, "Exits %d. Since %s.\n\n", e.Code.ExitCode(), e.Since)
		}
	}

	return []byte(b.String())
}

// Text renders one entry the way `atago explain` prints it: the same registry
// the published page is generated from, so a code cannot mean one thing in a
// terminal and another in a browser.
func Text(e Entry) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s  %s\n\n", e.Code, e.Summary)
	b.WriteString(wrap(e.Detail, 76))
	b.WriteString("\n\nFix\n")
	b.WriteString(wrap(e.Fix, 76))
	fmt.Fprintf(&b, "\n\nExits %d. Since %s.\n", e.Code.ExitCode(), e.Since)
	fmt.Fprintf(&b, "https://nao1215.github.io/atago/errors/#%s\n", anchor(e))
	return b.String()
}

// wrap breaks prose at width for a terminal. The registry stores each field as
// one paragraph, since that is what Markdown wants; a terminal wants lines.
func wrap(s string, width int) string {
	var b strings.Builder
	col := 0
	for i, word := range strings.Fields(s) {
		switch {
		case i == 0:
			col = len(word)
		case col+1+len(word) > width:
			b.WriteByte('\n')
			col = len(word)
		default:
			b.WriteByte(' ')
			col += 1 + len(word)
		}
		b.WriteString(word)
	}
	return b.String()
}

// entriesOf returns the entries belonging to one exit-code family, in order.
func entriesOf(all []Entry, exit int) []Entry {
	var out []Entry
	for _, e := range all {
		if e.Code.ExitCode() == exit {
			out = append(out, e)
		}
	}
	return out
}

// anchor is the heading anchor GitHub and Hugo both derive from an entry's
// heading: lowercased, non-alphanumerics dropped, spaces hyphenated.
func anchor(e Entry) string {
	heading := fmt.Sprintf("%s — %s", e.Code, e.Summary)
	var b strings.Builder
	for _, r := range strings.ToLower(heading) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			b.WriteRune(r)
		case r == ' ':
			b.WriteByte('-')
		}
	}
	return b.String()
}
