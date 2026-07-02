package assert

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nao1215/atago/internal/spec"
)

// checkStream evaluates a stdout/stderr/body matcher against captured bytes.
// hasData reports whether a command actually ran (to distinguish "no output"
// from "no command").
func checkStream(name string, s *spec.StreamAssert, data []byte, hasData bool, env Env) (out *CheckResult) {
	matchers := s.SetMatchers()
	if len(matchers) != 1 {
		return &CheckResult{Desc: "assert " + name, Hint: "stream assertion must set exactly one matcher"}
	}
	if !hasData {
		return &CheckResult{Desc: "assert " + name, Hint: "no command has run in this scenario yet"}
	}
	got := string(data)

	// Attach the full compared payload to any failing result so --artifacts-dir
	// can persist it as a durable sidecar (#48). The snapshot matcher shapes its
	// own artifact (normalized actual + committed expected) via checkSnapshot.
	kind := name
	defer func() {
		if out != nil && !out.OK && out.ArtifactKind == "" {
			out.ArtifactKind = kind
			if out.ArtifactActual == nil {
				out.ArtifactActual = data
			}
		}
	}()

	// A line selector narrows the stream to a single 1-based line before the
	// matcher runs (spec.md §16.2). json/snapshot are excluded by the loader.
	if s.Line != nil {
		line, ok := selectLine(got, *s.Line)
		if !ok {
			return &CheckResult{
				Desc:     fmt.Sprintf("assert %s line %d", name, *s.Line),
				Expected: fmt.Sprintf("%s to have a line %d", name, *s.Line),
				Actual:   excerpt(got),
				Hint:     fmt.Sprintf("%s has only %d line(s); line %d is out of range", name, countLines(got), *s.Line),
			}
		}
		got = line
		name = fmt.Sprintf("%s line %d", name, *s.Line)
	}

	switch {
	case s.Empty != nil:
		want := *s.Empty
		isEmpty := len(strings.TrimSpace(got)) == 0
		desc := fmt.Sprintf("assert %s empty: %t", name, want)
		if isEmpty == want {
			return pass(desc)
		}
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("%s empty=%t", name, want),
			Actual:   excerpt(got),
			Hint:     fmt.Sprintf("expected %s to be %s", name, emptiness(want)),
		}

	case s.Contains != nil:
		return checkContainsAll(name, got, s.Contains)

	case s.NotContains != nil:
		return checkNotContainsAll(name, got, s.NotContains)

	case s.Matches != nil:
		re, err := regexp.Compile(*s.Matches)
		if err != nil {
			return &CheckResult{Desc: "assert " + name, Hint: fmt.Sprintf("invalid regexp %q: %v", *s.Matches, err)}
		}
		desc := fmt.Sprintf("assert %s matches %q", name, *s.Matches)
		if re.MatchString(got) {
			return pass(desc)
		}
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("%s matches /%s/", name, *s.Matches),
			Actual:   excerpt(got),
			Hint:     fmt.Sprintf("regexp /%s/ did not match %s", *s.Matches, name),
		}

	case s.Equals != nil:
		desc := fmt.Sprintf("assert %s equals exact text", name)
		if equalsNormalized(got, *s.Equals) {
			return pass(desc)
		}
		return &CheckResult{
			Desc:             desc,
			Expected:         excerpt(*s.Equals),
			Actual:           excerpt(got),
			Hint:             fmt.Sprintf("%s did not equal the expected text", name),
			ArtifactExpected: []byte(*s.Equals),
		}

	case s.NotEquals != nil:
		desc := fmt.Sprintf("assert %s does not equal exact text", name)
		if !equalsNormalized(got, *s.NotEquals) {
			return pass(desc)
		}
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("%s different from the given text", name),
			Actual:   excerpt(got),
			Hint:     fmt.Sprintf("%s unexpectedly equaled the given text", name),
		}

	case s.JSON != nil:
		return checkJSON("assert "+name+" json", name, data, s.JSON)

	case s.YAML != nil:
		return checkYAML("assert "+name+" yaml", name, data, s.YAML)

	case s.Snapshot != "":
		return checkSnapshot("assert "+name+" snapshot", name, s.Snapshot, data, env)

	default:
		return &CheckResult{Desc: "assert " + name, Hint: "matcher not supported yet"}
	}
}

// checkContainsAll verifies every substring in subs is present in got. The whole
// list counts as one matcher; on the first missing element it fails, naming that
// element. A single-element list produces failure text identical to the scalar
// `contains` form.
func checkContainsAll(name, got string, subs spec.StringList) *CheckResult {
	desc := containsDesc(name, subs, true)
	if sub, idx, missing := firstMissing(got, subs); missing {
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("%s contains %q", name, sub),
			Actual:   excerpt(got),
			Hint:     fmt.Sprintf("the substring %q%s was not present in %s", sub, elementLabel(idx, len(subs)), name),
		}
	}
	return pass(desc)
}

// checkNotContainsAll verifies every substring in subs is absent from got.
func checkNotContainsAll(name, got string, subs spec.StringList) *CheckResult {
	desc := containsDesc(name, subs, false)
	if sub, idx, present := firstPresent(got, subs); present {
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("%s without %q", name, sub),
			Actual:   excerpt(got),
			Hint:     fmt.Sprintf("the substring %q%s was unexpectedly present in %s", sub, elementLabel(idx, len(subs)), name),
		}
	}
	return pass(desc)
}

// containsDesc renders the assertion label. A single element keeps the original
// phrasing (`assert stdout contains "x"`); a list reads `... contains all of` /
// `... contains none of` followed by the quoted elements.
func containsDesc(name string, subs spec.StringList, want bool) string {
	if len(subs) == 1 {
		if want {
			return fmt.Sprintf("assert %s contains %q", name, subs[0])
		}
		return fmt.Sprintf("assert %s does not contain %q", name, subs[0])
	}
	if want {
		return fmt.Sprintf("assert %s contains all of %s", name, quoteList(subs))
	}
	return fmt.Sprintf("assert %s contains none of %s", name, quoteList(subs))
}

// equalsNormalized compares ignoring a single trailing newline difference, since
// most commands emit a trailing newline that YAML block scalars also carry.
// CRLF line endings are folded to LF first so a spec written with `equals:`
// passes against cmd.exe output on Windows exactly as it does against POSIX
// output — line endings are an OS artifact, not observable CLI behavior.
func equalsNormalized(got, want string) bool {
	got = strings.ReplaceAll(got, "\r\n", "\n")
	want = strings.ReplaceAll(want, "\r\n", "\n")
	return got == want || strings.TrimRight(got, "\n") == strings.TrimRight(want, "\n")
}

// selectLine returns the n-th (1-based) line of s, ignoring a single trailing
// newline so a normal command's final "\n" does not introduce a phantom empty
// line. A trailing "\r" (CRLF) is stripped from the returned line.
func selectLine(s string, n int) (string, bool) {
	lines := splitLines(s)
	if n < 1 || n > len(lines) {
		return "", false
	}
	return lines[n-1], true
}

// countLines reports how many lines selectLine can address.
func countLines(s string) int {
	return len(splitLines(s))
}

func splitLines(s string) []string {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		lines[i] = strings.TrimRight(ln, "\r")
	}
	return lines
}

func emptiness(want bool) string {
	if want {
		return "empty"
	}
	return "non-empty"
}

const excerptLimit = 2000

func excerpt(s string) string {
	if len(s) <= excerptLimit {
		return s
	}
	return s[:excerptLimit] + "\n... (truncated)"
}
