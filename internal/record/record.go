// Package record generates a ready-to-edit spec skeleton from one observed
// command run (#30) — the answer to the blank-YAML problem. It is
// deliberately a skeleton generator with conservative matchers the author
// then tightens, not a brittle exact-golden recorder.
package record

import (
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/goccy/go-yaml"
	"github.com/nao1215/atago/internal/buildinfo"
	"github.com/nao1215/atago/internal/loader"
	"github.com/nao1215/atago/internal/store"
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
	// Workdir is the record-time scratch directory the command ran in. Any
	// occurrence of it in the generated stdout anchor is rewritten to the
	// built-in ${workdir} reference, so a command that prints an absolute path
	// under its workdir (e.g. `pwd`) still replays green: the anchor expands to
	// the replay's own isolated workdir instead of pinning the dead scratch path
	// the record run happened to use.
	Workdir string
}

// Generate renders the spec skeleton and proves it loads cleanly — a
// generated spec that fails validation is an internal bug, never the user's
// problem.
func Generate(obs Observation, opts Options) ([]byte, error) {
	var b strings.Builder
	b.WriteString(buildinfo.SchemaHeader())
	b.WriteString("version: \"1\"\n\n")
	b.WriteString("# Recorded by `atago record` — a starting point, not a verdict:\n")
	b.WriteString("# tighten the matchers to pin the behavior you actually care about.\n")
	fmt.Fprintf(&b, "suite:\n  name: %s\n\n", yamlScalar(opts.SuiteName))
	b.WriteString("scenarios:\n")
	fmt.Fprintf(&b, "  - name: %s # TODO: describe the behavior\n", yamlScalar(scenarioLabel(obs.Command)))
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
		anchor := maskWorkdir(escapeVarRefs(firstLine(obs.Stdout)), opts.Workdir)
		fmt.Fprintf(&b, "            contains: %s # first non-empty line, trimmed\n", yamlScalar(anchor))
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

// scenarioLabel turns an observed command into a single-line scenario name. The
// loader rejects a control character in a name (it corrupts the list table and
// doc headings), so a recorded multi-line shell command or a command carrying a
// tab would otherwise generate a spec that cannot load — the round-trip law (#30)
// broken by construction. Replace every control byte with a space and collapse
// the result; the exact command is preserved verbatim in run.command below.
func scenarioLabel(command string) string {
	var b strings.Builder
	prevSpace := false
	for _, r := range command {
		if r == '\t' || r == '\n' || r == '\r' || r < 0x20 || r == 0x7f {
			r = ' '
		}
		if r == ' ' {
			if prevSpace {
				continue
			}
			prevSpace = true
		} else {
			prevSpace = false
		}
		b.WriteRune(r)
	}
	label := strings.TrimSpace(b.String())
	if label == "" {
		// An all-control-byte command still needs a name the loader accepts.
		return "recorded command"
	}
	return label
}

// escapeVarRefs escapes the variable references in observed text so the
// generated spec replays it verbatim. Recorded commands, output anchors, file
// paths, and typed pty input are RAW text — any ${...} in them is literal — but
// the engine expands ${name} in run.command, assert values, and pty sends at
// replay (and rejects an unresolved reference in a no-shell command), so an
// unescaped reference would make the generated spec diverge from, or fail to
// replay, the recorded run. store.Escape is the exact inverse of the expander:
// it escapes only what the expander would act on, so a ${ not followed by a
// valid name (e.g. a tool that prints ${1}) is left alone rather than turned
// into a $${1} that the expander never restores and that could never match.
func escapeVarRefs(s string) string {
	return store.Escape(s)
}

// maskWorkdir rewrites the record-time scratch directory to the built-in
// ${workdir} reference so the stdout anchor replays green. A command that
// prints an absolute path under its workdir (`pwd`, a tool echoing an output
// path) would otherwise pin the dead scratch dir the record run used, which the
// replay's own isolated workdir can never match — the round-trip law (#30)
// broken by construction. It runs after escapeVarRefs so the injected reference
// is a live one the expander restores, not a $${...} literal. The scratch path
// is a plain filesystem path with no ${...} of its own, so escaping never
// touches it. Snapshot mode masks the same path via snapshot.Normalize already;
// this brings the default contains anchor to parity.
func maskWorkdir(s, workdir string) string {
	if workdir == "" {
		return s
	}
	return strings.ReplaceAll(s, workdir, "${workdir}")
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
// exactly. A value carrying invalid UTF-8 (binary output, or Latin-1 /
// Shift-JIS text) cannot survive any string scalar at all — see yamlBinary —
// so it takes the !!binary path first.
func yamlScalar(s string) string {
	if !utf8.ValidString(s) {
		return yamlBinary(s)
	}
	if hasControlByte(s) {
		return yamlDoubleQuoted(s)
	}
	out, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Sprintf("%q", s)
	}
	scalar := strings.TrimRight(string(out), "\n")
	// yaml.Marshal leaves a value that begins with the explicit-key indicator
	// "?" unquoted (e.g. a recorded stdout line of just "?"), which reparses as a
	// mapping-key start rather than a scalar and makes the generated spec invalid.
	// Force double-quoting when the marshaled scalar came back bare but the value
	// starts with a YAML indicator that cannot open a plain scalar.
	if len(scalar) > 0 && scalar[0] != '"' && scalar[0] != '\'' && startsWithYAMLIndicator(s) {
		return yamlDoubleQuoted(s)
	}
	return scalar
}

// startsWithYAMLIndicator reports whether s begins with a character that cannot
// open a plain (unquoted) YAML scalar. Most are already quoted by yaml.Marshal,
// but "?" is not, so this backstops the scalar emitter against an unquoted
// indicator slipping into the generated spec.
func startsWithYAMLIndicator(s string) bool {
	if s == "" {
		return false
	}
	switch s[0] {
	case '?', ':', '-', ',', '[', ']', '{', '}', '#', '&', '*', '!', '|', '>', '\'', '"', '%', '@', '`', ' ', '\t':
		return true
	}
	return false
}

// yamlBinary renders s as a YAML `!!binary` (base64) scalar so a value carrying
// invalid UTF-8 bytes round-trips byte-for-byte through the loader. No string
// scalar can do this: a raw invalid byte in a plain, single-, or double-quoted
// scalar is lossily replaced with U+FFFD (ef bf bd) on reparse, and a `\xNN`
// double-quoted escape decodes to the Unicode code point U+00NN — re-encoded as
// two UTF-8 bytes for NN >= 0x80 — not the raw byte. `!!binary` is YAML's
// canonical arbitrary-byte representation, and go-yaml decodes its base64
// straight back into the destination string field as the exact original bytes,
// so the recorded contains anchor still matches the real (raw-byte) stdout on
// replay. The base64 alphabet ([A-Za-z0-9+/=]) contains nothing that breaks the
// inline flow or starts a trailing ` # comment`, so it needs no quoting.
func yamlBinary(s string) string {
	return "!!binary " + base64.StdEncoding.EncodeToString([]byte(s))
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
