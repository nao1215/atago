package engine

import (
	"fmt"
	"os"
	"regexp"
	"strings"

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
		switch {
		case from.File.Text != nil:
			// Capture the whole file verbatim (#158); no json path needed for an
			// opaque blob.
			return string(data), nil
		case len(from.File.JSON) > 0:
			return jsonValue(data, from.File.JSON[0].Path)
		default:
			return "", fmt.Errorf("store %q: a file source needs a json or text selector", sp.Name)
		}
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
	case len(s.JSON) > 0:
		return jsonValue(data, s.JSON[0].Path)
	case s.Matches != nil:
		return regexValue(data, *s.Matches)
	case s.Trim != nil:
		// Capture the whole stream (#158). trim: true strips surrounding
		// whitespace (the common "grab the whole token, drop the trailing
		// newline" case); trim: false keeps the bytes verbatim.
		out := string(data)
		if *s.Trim {
			out = strings.TrimSpace(out)
		}
		return out, nil
	default:
		return "", fmt.Errorf("a stdout store source needs a json, matches, or trim selector")
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
	// Why: a JSON null selects exactly one node whose Go value is nil, and
	// fmt.Sprint(nil) yields the literal string "<nil>". Storing that would leak
	// a Go-ism into a user-visible variable and silently mask "the field was
	// null" as a captured value. Surface it as a clean error instead.
	if nodes[0] == nil {
		return "", fmt.Errorf("JSON path %q selected a null value", path)
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
