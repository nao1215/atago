// Package diag defines atago's diagnostic codes: the stable, searchable names
// for the errors atago itself reports.
//
// A message alone is a dead end. "unknown assert key" cannot be searched for,
// linked to, or branched on, and rewording it next release breaks every
// bookmark and every script that matched on the prose. A code survives
// rewording, so the message stays free to improve.
//
// A code is `ATG` followed by four digits whose thousands digit is the process
// exit code the diagnostic produces, making the code a refinement of the exit
// contract rather than a second contract beside it:
//
//	ATG2xxx  exit 2  spec error — YAML syntax, schema, semantic validation
//	ATG3xxx  exit 3  configuration error — CLI invocation, flags, target selection
//	ATG4xxx  exit 4  execution error — a step could not be carried out
//	ATG5xxx  exit 5  internal error — an atago bug
//	ATG6xxx  exit 6  security policy violation
//
// ATG1xxx is never assigned: exit 1 is an assertion failing, which is a spec
// doing its job rather than atago failing to do its own.
//
// Codes are grouped by what the reader has to fix, never one per call site.
// Two errors share a code when the fix is the same and differ when it differs,
// so the published reference reads as a list of problems a person can act on.
//
// # Why a code cannot go missing
//
// Codes are not constants that a registry then describes; they are what
// registering returns. [register] is the only way to obtain a [Code], and it
// refuses an entry that omits any of the text the published reference needs. A
// code with no documentation is therefore not something to test for — it is
// unconstructable.
package diag

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Code is one diagnostic's stable identity. The zero value is not a valid
// code; every Code comes from [register], which is unexported, so the set of
// codes is closed and each member is documented by construction.
type Code int

// String renders the code in its published form, e.g. "ATG2201".
func (c Code) String() string {
	return fmt.Sprintf("ATG%04d", int(c))
}

// ExitCode is the process exit status a diagnostic with this code produces,
// which is the code's own thousands digit.
func (c Code) ExitCode() int {
	return int(c) / 1000
}

// Annotate prefixes a message with the code, which is how every coded
// diagnostic reaches the user: the code is added to the sentence, never
// substituted for it.
func (c Code) Annotate(msg string) string {
	return c.String() + ": " + msg
}

// Errorf builds an error whose message carries the code, and is the form used
// where diagnostics are raised as errors rather than collected as strings. It
// forwards to [fmt.Errorf], so a `%w` in the format still wraps.
//
// Only the site that names the problem carries a code. A wrapper that adds
// context to someone else's error — "service %q: %w" — leaves the code to the
// error it wraps, so a message never accumulates two of them.
func (c Code) Errorf(format string, args ...any) error {
	return fmt.Errorf(c.String()+": "+format, args...)
}

// Entry is a code's published documentation.
type Entry struct {
	// Code is the diagnostic's identity.
	Code Code
	// Name is the Go identifier this package binds the code to. It lets a test
	// prove that every registered code is reachable from real code, and that
	// the identifier and the entry have not drifted apart.
	Name string
	// Summary is the one-line meaning, written as the problem rather than as
	// the message any single call site happens to print.
	Summary string
	// Detail explains what atago was doing and why this is refused.
	Detail string
	// Fix tells the reader what to change. A diagnostic that does not say how
	// to proceed gets worked around instead of fixed.
	Fix string
	// Since is the atago version the code first shipped in.
	Since string
}

// codeText matches a code in its published form, anywhere in a string. It is
// the same syntax [Parse] accepts and the form the E2E coverage gate looks for
// in the specs.
var codeText = regexp.MustCompile(`\bATG([0-9]{4})\b`)

// registry holds every code in registration order; [All] sorts it.
var registry []Entry

// byCode indexes the registry for lookup, and is what makes a duplicate
// number impossible to register.
var byCode = map[Code]Entry{}

// register records one diagnostic and returns its code. It is the only source
// of a [Code], so a code cannot exist without the documentation the published
// reference is generated from.
//
// It panics on a malformed registration rather than returning an error: every
// call is a package-level variable in this file, so a mistake is a programming
// error caught the first time anything imports diag, including the first test.
func register(number int, name string, e Entry) Code {
	c := Code(number)
	switch {
	case number < 2000 || number > 6999:
		panic(fmt.Sprintf("diag: code %d is outside the ATG2000-ATG6999 range", number))
	case number/1000 == 1:
		panic(fmt.Sprintf("diag: code %d is in the ATG1xxx range, which is reserved: exit 1 is an assertion failing, not an atago error", number))
	case name == "":
		panic(fmt.Sprintf("diag: %s has no identifier name", c))
	case e.Summary == "":
		panic(fmt.Sprintf("diag: %s has no summary", c))
	case e.Detail == "":
		panic(fmt.Sprintf("diag: %s has no detail", c))
	case e.Fix == "":
		panic(fmt.Sprintf("diag: %s has no fix", c))
	case e.Since == "":
		panic(fmt.Sprintf("diag: %s does not say which version it shipped in", c))
	}
	if prev, dup := byCode[c]; dup {
		panic(fmt.Sprintf("diag: %s is already registered as %s", c, prev.Name))
	}
	e.Code = c
	e.Name = name
	registry = append(registry, e)
	byCode[c] = e
	return c
}

// All returns every registered diagnostic, ordered by code. The published
// reference and `atago explain` both render this, so neither can drift from
// the other.
func All() []Entry {
	out := slices.Clone(registry)
	slices.SortFunc(out, func(a, b Entry) int { return int(a.Code) - int(b.Code) })
	return out
}

// Lookup returns the entry for a code.
func Lookup(c Code) (Entry, bool) {
	e, ok := byCode[c]
	return e, ok
}

// Parse reads a code written in its published form, e.g. "ATG2201". It accepts
// only registered codes: a well-formed number nobody assigned is not a code,
// and reporting it as one would send the reader to a page that does not exist.
// Case is ignored so a code copied out of prose still resolves.
func Parse(s string) (Code, bool) {
	s = strings.TrimSpace(s)
	if !strings.EqualFold(s[:min(len(s), 3)], "ATG") {
		return 0, false
	}
	n, err := strconv.Atoi(s[3:])
	if err != nil {
		return 0, false
	}
	c := Code(n)
	if _, ok := byCode[c]; !ok {
		return 0, false
	}
	return c, true
}

// Codes returns every registered code mentioned in text, in order of
// appearance and without repeats. The E2E coverage gate uses it to prove that
// each published code is provoked by a real scenario.
func Codes(text string) []Code {
	var out []Code
	for _, m := range codeText.FindAllStringSubmatch(text, -1) {
		n, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		c := Code(n)
		if _, ok := byCode[c]; !ok {
			continue
		}
		if !slices.Contains(out, c) {
			out = append(out, c)
		}
	}
	return out
}
