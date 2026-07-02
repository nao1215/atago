package engine

import (
	"fmt"
	"os"
	"regexp"

	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"

	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
)

// extractValue resolves the value a store step captures. MVP
// supports extracting from stdout (via JSON path or regex) and from a generated
// file (via JSON path); `from.body` (HTTP) is post-MVP.
func extractValue(sp *spec.Store, current *runner.Result, workdir string) (string, error) {
	from := sp.From
	if from == nil {
		return "", fmt.Errorf("store %q: 'from' is required", sp.Name)
	}
	switch {
	case from.Stdout != nil:
		if current == nil {
			return "", fmt.Errorf("store %q: no command has run in this scenario yet", sp.Name)
		}
		return extractStream(from.Stdout, current.Stdout)
	case from.File != nil:
		path, err := security.ResolveWorkdirPath("store.from.file.path", workdir, from.File.Path)
		if err != nil {
			return "", fmt.Errorf("store %q: %w", sp.Name, err)
		}
		data, err := os.ReadFile(path) //nolint:gosec // path is the user-declared store source, confined to the workdir
		if err != nil {
			return "", fmt.Errorf("store %q: %w", sp.Name, err)
		}
		if from.File.JSON == nil {
			return "", fmt.Errorf("store %q: a file source needs a json selector", sp.Name)
		}
		return jsonValue(data, from.File.JSON.Path)
	case from.Body != nil:
		if current == nil || !current.IsHTTP {
			return "", fmt.Errorf("store %q: no HTTP request has run in this scenario yet", sp.Name)
		}
		return extractStream(from.Body, current.Body)
	case from.Header != "":
		if current == nil || !current.IsHTTP {
			return "", fmt.Errorf("store %q: no HTTP request has run in this scenario yet", sp.Name)
		}
		v := current.Header.Get(from.Header)
		if v == "" {
			return "", fmt.Errorf("store %q: response has no %q header", sp.Name, from.Header)
		}
		return v, nil
	case from.Rows != nil:
		if current == nil || !current.IsDB {
			return "", fmt.Errorf("store %q: no query has run in this scenario yet", sp.Name)
		}
		return extractStream(from.Rows, current.RowsJSON)
	case from.Message != nil:
		if current == nil || !current.IsGRPC {
			return "", fmt.Errorf("store %q: no gRPC call has run in this scenario yet", sp.Name)
		}
		return extractStream(from.Message, current.MessageJSON)
	case from.Value != nil:
		if current == nil || !current.IsCDP {
			return "", fmt.Errorf("store %q: no cdp step has run in this scenario yet", sp.Name)
		}
		return extractStream(from.Value, current.CDPValue)
	default:
		return "", fmt.Errorf("store %q: 'from' must set stdout, file, body, header, rows, message, or value", sp.Name)
	}
}

func extractStream(s *spec.StreamAssert, data []byte) (string, error) {
	switch {
	case s.JSON != nil:
		return jsonValue(data, s.JSON.Path)
	case s.Matches != nil:
		return regexValue(data, *s.Matches)
	default:
		return "", fmt.Errorf("a stdout store source needs a json or matches selector")
	}
}

// jsonValue selects exactly one value at path and returns it as a string.
func jsonValue(data []byte, path string) (string, error) {
	v, err := oj.Parse(data)
	if err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	expr, err := jp.ParseString(path)
	if err != nil {
		return "", fmt.Errorf("invalid JSON path %q: %w", path, err)
	}
	nodes := expr.Get(v)
	if len(nodes) != 1 {
		return "", fmt.Errorf("JSON path %q selected %d values, want exactly 1", path, len(nodes))
	}
	return fmt.Sprint(nodes[0]), nil
}

// regexValue returns the first capture group, or the whole match if the pattern
// has no capture group.
func regexValue(data []byte, pattern string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regexp %q: %w", pattern, err)
	}
	m := re.FindSubmatch(data)
	if m == nil {
		return "", fmt.Errorf("regexp %q did not match", pattern)
	}
	if len(m) > 1 {
		return string(m[1]), nil
	}
	return string(m[0]), nil
}
