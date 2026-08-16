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

	b.WriteString("# Error codes\n\n")
	b.WriteString("Every error atago reports carries a code, so a failure can be searched for, linked to, and branched on without matching on prose that is free to improve. Assertion failures are not errors in this sense and carry no code: a spec that fails is a spec doing its job, and its message already says what was expected and what happened.\n\n")
	b.WriteString("A code is `ATG` followed by four digits, and the first of those digits is the exit status the run produced. Reading `ATG2103` tells you the process exited 2 before it tells you anything else.\n\n")

	b.WriteString("| Codes | Exit | Meaning |\n|---|---|---|\n")
	for _, f := range families {
		fmt.Fprintf(&b, "| `%s` | %d | %s |\n", f.label, f.exit, f.meaning)
	}
	b.WriteString("\n")
	b.WriteString("`ATG1xxx` is never assigned. Exit 1 means one or more scenarios failed, which is a result rather than an error.\n\n")
	b.WriteString("Codes are grouped by what you have to fix rather than by where in atago the error was raised, so one code can be reported from several places when the answer is the same in all of them.\n\n")
	b.WriteString("Look a code up from the terminal with `atago explain ATG2201`, which prints this same text without a browser.\n\n")

	all := All()
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
