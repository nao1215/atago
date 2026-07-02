// Package report renders SuiteResults in the supported output formats
// (spec.md §26): console, JSON, JUnit XML, and GitHub Actions annotations.
package report

// Format names a built-in report format.
type Format string

const (
	// FormatConsole is the default human-readable output.
	FormatConsole Format = "console"
	// FormatJSON is the machine-readable report (spec.md §26).
	FormatJSON Format = "json"
	// FormatJUnit is a JUnit XML report (spec.md §26).
	FormatJUnit Format = "junit"
	// FormatGHA emits GitHub Actions workflow-command annotations (spec.md §26).
	FormatGHA Format = "gha"
	// FormatTAP emits a Test Anything Protocol (TAP 13) stream (spec.md §26).
	FormatTAP Format = "tap"
)

// Valid reports whether f is a known report format. Render is the single entry
// point that dispatches on Format to the per-format build/write helpers.
func (f Format) Valid() bool {
	switch f {
	case FormatConsole, FormatJSON, FormatJUnit, FormatGHA, FormatTAP:
		return true
	default:
		return false
	}
}
