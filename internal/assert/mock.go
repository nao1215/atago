package assert

import (
	"fmt"
	"strings"

	"github.com/nao1215/atago/internal/runner/mock"
	"github.com/nao1215/atago/internal/spec"
)

// checkMock evaluates a `mock:` assertion (#24) against the requests a mock
// server recorded: filter by the optional path/method, pin the exact Count
// (or require at least one match without it), then apply the header/body
// matchers to the LAST matching request.
func checkMock(m *spec.MockAssert, env Env) *CheckResult {
	desc := fmt.Sprintf("assert mock %q", m.Name)
	if env.MockRecords == nil {
		return &CheckResult{Desc: desc, Hint: "mock assertions are not available in this context"}
	}
	records, ok := env.MockRecords(m.Name)
	if !ok {
		return &CheckResult{Desc: desc, Hint: fmt.Sprintf("no mock server named %q is running", m.Name)}
	}

	var matched []mock.Record
	for _, r := range records {
		if m.Path != "" && r.Path != m.Path {
			continue
		}
		if m.Method != "" && !strings.EqualFold(r.Method, m.Method) {
			continue
		}
		matched = append(matched, r)
	}

	filter := mockFilterLabel(m)
	if m.Count != nil {
		desc := fmt.Sprintf("assert mock %q received %d %s", m.Name, *m.Count, plural("request", *m.Count))
		if len(matched) != *m.Count {
			return &CheckResult{
				Desc:     desc,
				Expected: fmt.Sprintf("%d %s %s", *m.Count, plural("request", *m.Count), filter),
				Actual:   fmt.Sprintf("%d matching of %d recorded:\n%s", len(matched), len(records), summarizeRecords(records)),
				Hint:     "the CLI under test did not send the expected number of matching requests",
			}
		}
	} else if len(matched) == 0 {
		return &CheckResult{
			Desc:     desc,
			Expected: "at least one request " + filter,
			Actual:   fmt.Sprintf("0 matching of %d recorded:\n%s", len(records), summarizeRecords(records)),
			Hint:     "the CLI under test never sent a matching request",
		}
	}

	if m.Header != nil || m.Body != nil {
		if len(matched) == 0 {
			// Count: 0 with matchers is contradictory; the validator rejects it,
			// but stay safe for direct API users.
			return &CheckResult{Desc: desc, Hint: "header/body matchers need at least one matching request"}
		}
		last := matched[len(matched)-1]
		if m.Header != nil {
			if cr := checkHeaderValue(m.Header, last.Header.Get(m.Header.Name), "recorded request"); !cr.OK {
				return cr
			}
		}
		if m.Body != nil {
			if cr := checkStream("mock request body", m.Body, last.Body, true, env); !cr.OK {
				return cr
			}
		}
	}
	if m.Count != nil {
		return pass(fmt.Sprintf("assert mock %q received %d %s", m.Name, *m.Count, plural("request", *m.Count)))
	}
	return pass(desc + " received a matching request")
}

// mockFilterLabel names the path/method filter for Expected lines.
func mockFilterLabel(m *spec.MockAssert) string {
	switch {
	case m.Method != "" && m.Path != "":
		return fmt.Sprintf("for %s %s", strings.ToUpper(m.Method), m.Path)
	case m.Path != "":
		return "for " + m.Path
	case m.Method != "":
		return "for " + strings.ToUpper(m.Method)
	default:
		return "(any route)"
	}
}

// summarizeRecords renders the recorded requests one per line so a failing
// count shows what the CLI actually sent.
func summarizeRecords(records []mock.Record) string {
	if len(records) == 0 {
		return "  (no requests recorded)"
	}
	var b strings.Builder
	for i, r := range records {
		fmt.Fprintf(&b, "  %d: %s %s -> %d\n", i+1, r.Method, r.Path, r.Status)
	}
	return strings.TrimRight(b.String(), "\n")
}

func plural(word string, n int) string {
	if n == 1 {
		return word
	}
	return word + "s"
}
