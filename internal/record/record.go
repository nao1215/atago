// Package record generates a ready-to-edit spec skeleton from one observed
// command run (#30) — the answer to the blank-YAML problem. It is
// deliberately a skeleton generator with conservative matchers the author
// then tightens, not a brittle exact-golden recorder.
package record

import (
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/nao1215/atago/internal/loader"
)

// maxFileAsserts caps the generated file.exists asserts so a command that
// explodes a tree does not produce an unreadable spec; the rest is noted in a
// comment.
const maxFileAsserts = 10

// Observation is what one command run produced.
type Observation struct {
	// Command is the recorded command line, verbatim.
	Command string
	// Shell marks the run as shell-executed (`--shell`).
	Shell bool
	// ExitCode is the observed exit code.
	ExitCode int
	// Stdout / Stderr are the captured streams.
	Stdout []byte
	Stderr []byte
	// CreatedFiles lists workdir-relative /-separated files the command
	// created, sorted.
	CreatedFiles []string
}

// Options tunes generation.
type Options struct {
	// SuiteName names the suite (default: first command token's base name).
	SuiteName string
	// Snapshot switches the stdout assert to a snapshot matcher referencing
	// SnapshotPath (spec-relative).
	Snapshot     bool
	SnapshotPath string
}

// Generate renders the spec skeleton and proves it loads cleanly — a
// generated spec that fails validation is an internal bug, never the user's
// problem.
func Generate(obs Observation, opts Options) ([]byte, error) {
	var b strings.Builder
	b.WriteString("version: \"1\"\n\n")
	b.WriteString("# Recorded by `atago record` — a starting point, not a verdict:\n")
	b.WriteString("# tighten the matchers to pin the behavior you actually care about.\n")
	fmt.Fprintf(&b, "suite:\n  name: %s\n\n", yamlScalar(opts.SuiteName))
	b.WriteString("scenarios:\n")
	fmt.Fprintf(&b, "  - name: %s # TODO: describe the behavior\n", yamlScalar(obs.Command))
	b.WriteString("    steps:\n")
	b.WriteString("      - run:\n")
	if obs.Shell {
		b.WriteString("          shell: true\n")
	}
	fmt.Fprintf(&b, "          command: %s\n", yamlScalar(escapeVarRefs(obs.Command)))
	b.WriteString("      - assert:\n")
	fmt.Fprintf(&b, "          exit_code: %d\n", obs.ExitCode)

	switch {
	case opts.Snapshot:
		b.WriteString("      - assert:\n")
		b.WriteString("          stdout:\n")
		fmt.Fprintf(&b, "            snapshot: %s\n", yamlScalar(opts.SnapshotPath))
	case firstLine(obs.Stdout) != "":
		b.WriteString("      - assert:\n")
		b.WriteString("          stdout:\n")
		fmt.Fprintf(&b, "            contains: %s # first non-empty line, trimmed\n", yamlScalar(escapeVarRefs(firstLine(obs.Stdout))))
	}
	if len(obs.Stderr) == 0 {
		b.WriteString("      - assert:\n")
		b.WriteString("          stderr:\n")
		b.WriteString("            empty: true\n")
	}

	files := obs.CreatedFiles
	capped := false
	if len(files) > maxFileAsserts {
		files = files[:maxFileAsserts]
		capped = true
	}
	for _, f := range files {
		b.WriteString("      - assert:\n")
		b.WriteString("          file:\n")
		fmt.Fprintf(&b, "            path: %s\n", yamlScalar(escapeVarRefs(f)))
		b.WriteString("            exists: true\n")
	}
	if capped {
		fmt.Fprintf(&b, "      # ... and %d more created files not asserted here\n", len(obs.CreatedFiles)-maxFileAsserts)
	}

	out := []byte(b.String())
	if _, err := loader.LoadBytes("recorded.atago.yaml", out); err != nil {
		return nil, fmt.Errorf("generated spec does not validate (this is an atago bug, please report it): %w", err)
	}
	return out, nil
}

// escapeVarRefs rewrites every "${" in observed text as the documented "$${"
// literal escape. Recorded commands, output anchors, file paths, and typed
// pty input are RAW text — any ${...} in them is literal — but the engine
// expands ${name} in run.command, assert values, and pty sends at replay (and
// rejects an unresolved reference in a no-shell command), so an unescaped
// reference would make the generated spec diverge from, or fail to replay,
// the recorded run. The uniform rewrite round-trips: the expander turns
// "$${" back into the literal "${" the tool actually saw.
func escapeVarRefs(s string) string {
	return strings.ReplaceAll(s, "${", "$${")
}

// firstLine returns the first non-empty line, trimmed.
func firstLine(stream []byte) string {
	for _, l := range strings.Split(string(stream), "\n") {
		if t := strings.TrimSpace(l); t != "" {
			return t
		}
	}
	return ""
}

// yamlScalar renders an arbitrary string as one safe inline YAML scalar,
// delegating quoting decisions to the YAML library so recorded commands and
// output can never break the document structure.
//
// yaml.Marshal is trusted for punctuation-heavy scalars (`#`, `:`, `*`, a
// leading space, …), but NOT for control bytes: it leaves a raw tab in a plain
// scalar — which YAML re-parsing then silently strips, so a recorded
// tab-separated line would no longer match the real output — and it renders any
// multi-line string as a block scalar that cannot be spliced inline after
// `contains: ` / `command: ` (producing an invalid document that aborts
// `atago record`). For a value carrying any control byte, emit an explicit
// single-line double-quoted scalar that escapes it, so the value round-trips
// exactly.
func yamlScalar(s string) string {
	if hasControlByte(s) {
		return yamlDoubleQuoted(s)
	}
	out, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Sprintf("%q", s)
	}
	return strings.TrimRight(string(out), "\n")
}

// hasControlByte reports whether s contains a C0 control character (tab,
// newline, CR, …) — the bytes yaml.Marshal cannot safely inline and that
// yamlDoubleQuoted escapes.
func hasControlByte(s string) bool {
	for _, r := range s {
		if r < 0x20 {
			return true
		}
	}
	return false
}
