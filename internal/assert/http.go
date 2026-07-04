package assert

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// checkStatus evaluates an HTTP `status` assertion against the response code
// captured by the http runner.
func checkStatus(want *int, res *runner.Result) *CheckResult {
	if res == nil || !res.IsHTTP {
		return &CheckResult{Desc: "assert status", Hint: "no HTTP request has run in this scenario yet"}
	}
	desc := fmt.Sprintf("assert status is %d", *want)
	if res.StatusCode == *want {
		return pass(desc)
	}
	return &CheckResult{
		Desc:     desc,
		Expected: fmt.Sprintf("status %d", *want),
		Actual:   fmt.Sprintf("status %d", res.StatusCode),
		Hint:     fmt.Sprintf("expected HTTP status %d but the response was %d", *want, res.StatusCode),
	}
}

// checkHeader evaluates an HTTP `header` assertion against a response header
// . The header name is matched case-insensitively per RFC 7230.
func checkHeader(h *spec.HeaderMatch, res *runner.Result) *CheckResult {
	if res == nil || !res.IsHTTP {
		return &CheckResult{Desc: "assert header", Hint: "no HTTP request has run in this scenario yet"}
	}
	return checkHeaderValue(h, res.Header.Get(h.Name), "response")
}

// checkHeaderValue matches one header value; kind names the header's origin
// ("response" for the http target, "recorded request" for the mock target,
// #24) so failure hints read naturally in both contexts.
func checkHeaderValue(h *spec.HeaderMatch, got, kind string) *CheckResult {
	switch {
	case h.Equals != nil:
		desc := fmt.Sprintf("assert header %q equals %q", h.Name, *h.Equals)
		if got == *h.Equals {
			return pass(desc)
		}
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("%s: %s", h.Name, *h.Equals),
			Actual:   fmt.Sprintf("%s: %s", h.Name, got),
			Hint:     fmt.Sprintf("%s header %q did not equal the expected value", kind, h.Name),
		}
	case h.Contains != nil:
		desc := fmt.Sprintf("assert header %q contains %q", h.Name, *h.Contains)
		if strings.Contains(got, *h.Contains) {
			return pass(desc)
		}
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("%s containing %q", h.Name, *h.Contains),
			Actual:   fmt.Sprintf("%s: %s", h.Name, got),
			Hint:     fmt.Sprintf("%s header %q did not contain the substring", kind, h.Name),
		}
	case h.Matches != nil:
		desc := fmt.Sprintf("assert header %q matches /%s/", h.Name, *h.Matches)
		re, err := regexp.Compile(*h.Matches)
		if err != nil {
			return &CheckResult{Desc: desc, Hint: fmt.Sprintf("invalid regexp: %v", err)}
		}
		if re.MatchString(got) {
			return pass(desc)
		}
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("%s matching /%s/", h.Name, *h.Matches),
			Actual:   fmt.Sprintf("%s: %s", h.Name, got),
			Hint:     fmt.Sprintf("%s header %q did not match the pattern", kind, h.Name),
		}
	default:
		return &CheckResult{Desc: "assert header", Hint: "header assertion must set contains, equals, or matches"}
	}
}

// httpBody returns the captured HTTP response body, or nil when no HTTP request
// has run. A body assertion reuses the stream matchers (contains/equals/json/…).
func httpBody(res *runner.Result) []byte {
	if res == nil {
		return nil
	}
	return res.Body
}

// dbRows returns the captured query result rows as JSON, or nil when no query has
// run. A rows assertion reuses the stream matchers (json path/length, contains…).
func dbRows(res *runner.Result) []byte {
	if res == nil {
		return nil
	}
	return res.RowsJSON
}

// grpcMessage returns the captured gRPC response message as JSON, or nil when no
// call has run. A message assertion reuses the stream matchers.
func grpcMessage(res *runner.Result) []byte {
	if res == nil {
		return nil
	}
	return res.MessageJSON
}

// cdpValue returns the value captured by the last browser text/eval action, or
// nil when no cdp step has run. A value assertion reuses the stream matchers.
func cdpValue(res *runner.Result) []byte {
	if res == nil {
		return nil
	}
	return res.CDPValue
}

// checkGRPCStatus evaluates a `grpc_status` assertion against the numeric status
// code captured by the grpc runner.
func checkGRPCStatus(want *int, res *runner.Result) *CheckResult {
	if res == nil || !res.IsGRPC {
		return &CheckResult{Desc: "assert grpc_status", Hint: "no gRPC call has run in this scenario yet"}
	}
	desc := fmt.Sprintf("assert grpc_status is %d", *want)
	if res.GRPCStatus == *want {
		return pass(desc)
	}
	return &CheckResult{
		Desc:     desc,
		Expected: fmt.Sprintf("grpc status %d", *want),
		Actual:   fmt.Sprintf("grpc status %d", res.GRPCStatus),
		Hint:     fmt.Sprintf("expected gRPC status code %d but got %d", *want, res.GRPCStatus),
	}
}
