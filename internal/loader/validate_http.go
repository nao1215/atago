package loader

import (
	"strings"

	"github.com/nao1215/atago/internal/spec"
)

// validateHTTPPayload enforces "a request has one payload": json, body,
// body_file, and form/files are mutually exclusive families (form and files
// combine into one multipart request, so they count as a single family).
func validateHTTPPayload(add func(string, ...any), where string, h *spec.HTTP) {
	var set []string
	if h.JSON != nil {
		set = append(set, "json")
	}
	if h.Body != "" {
		set = append(set, "body")
	}
	if h.BodyFile != "" {
		set = append(set, "body_file")
	}
	if len(h.Form) > 0 || len(h.Files) > 0 {
		set = append(set, "form/files")
	}
	if len(set) > 1 {
		add("%s.http sets %s; a request has one payload — use json for a structured value, body for raw text, body_file for a file's raw content, or form (+ files) for a form submission", where, strings.Join(set, " and "))
	}
}
