// Package docgen renders Markdown documentation from specs
// using github.com/nao1215/markdown. Each spec becomes a section with
// Given / When / Then subsections per scenario.
package docgen

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
	"github.com/nao1215/markdown"
)

// Source pairs a spec with the path it was loaded from.
type Source struct {
	Path string
	Spec *spec.Spec
}

// Generate writes Markdown documentation for all sources to w. The document
// opens with a summary and a table of contents (#66) so large generated docs
// stay navigable, then renders one section per suite.
//
// Output is normalized to LF line endings on every platform: the underlying
// Markdown writer emits CRLF on Windows, which would otherwise make `atago doc`
// output — and the committed behavior docs — differ by OS and fail the drift
// tests on Windows CI.
func Generate(w io.Writer, sources []Source) error {
	return GenerateTo(w, sources, "")
}

// GenerateTo is Generate with the directory the document will be written to, used
// to build relative links to committed golden files (e.g. an image baseline) so
// they render inline when the doc is viewed (#67). Pass "" when writing to
// stdout or when relative links are not wanted; golden images then stay as text
// references instead of embeds.
func GenerateTo(w io.Writer, sources []Source, outputDir string) error {
	var buf bytes.Buffer
	md := markdown.NewMarkdown(&buf)
	md.H1("atago Behavior Specs")
	writeHeader(md, sources)

	for _, src := range sources {
		writeSuite(md, src, outputDir)
	}
	if err := md.Build(); err != nil {
		return err
	}
	normalized := bytes.ReplaceAll(buf.Bytes(), []byte("\r\n"), []byte("\n"))
	_, err := w.Write(normalized)
	return err
}

// writeHeader emits the top-of-file summary block and table of contents (#66).
// The summary reports suite/scenario counts and a tag breakdown; the TOC links
// to every suite and scenario. It is deterministic so committed docs stay
// byte-stable and reviewable.
func writeHeader(md *markdown.Markdown, sources []Source) {
	sum := computeSummary(sources)

	md.H2("Summary")
	md.PlainTextf("%s · %s", pluralize(sum.suites, "suite"), pluralize(sum.scenarios, "scenario"))
	if tags := sum.tagLine(); tags != "" {
		md.PlainTextf("Tags: %s", tags)
	}

	md.H2("Contents")
	toc, _ := tableOfContents(sources)
	md.PlainText(strings.TrimRight(toc, "\n"))
}

func writeSuite(md *markdown.Markdown, src Source, outputDir string) {
	s := src.Spec
	md.H2f("%s", s.Suite.Name)
	// Always render the source path with forward slashes so the generated docs are
	// byte-identical across platforms (Windows filepath.Clean uses backslashes).
	md.PlainTextf("Source: `%s`", filepath.ToSlash(src.Path))

	// Golden files (snapshots, image baselines) are resolved relative to the spec
	// file's directory, so the doc can inline/embed the committed expected result
	// (#67). outputDir anchors relative links to embedded images.
	specDir := filepath.Dir(src.Path)
	for i := range s.Scenarios {
		writeScenario(md, &s.Scenarios[i], specDir, outputDir)
	}
}

func writeScenario(md *markdown.Markdown, sc *spec.Scenario, specDir, outputDir string) {
	md.H3f("Scenario: %s", sc.Name)
	if meta := scenarioMeta(sc); meta != "" {
		md.PlainText(meta)
	}

	// A matrix instance's name already shows the row's concrete values; render
	// its commands and assertions with the same values so the reader sees
	// `git checkout v9.9.9`, not the template's ${ref}.
	expand := matrixExpander(sc)

	if given := givenBullets(sc, expand); len(given) > 0 {
		md.H4("Given")
		md.BulletList(given...)
	}

	if inputs := inputPreviews(sc); len(inputs) > 0 {
		md.H4("Inputs")
		writePreviews(md, inputs)
	}

	if cmds := commands(sc.Steps, expand); len(cmds) > 0 {
		md.H4("When")
		md.CodeBlocks(markdown.SyntaxHighlightShell, strings.Join(cmds, "\n"))
	}

	writeThen(md, sc, expand)

	// Teardown always runs — pass, fail, error, or interrupt — so document the
	// cleanup a scenario performs against external systems.
	if td := commands(sc.Teardown, expand); len(td) > 0 {
		md.H4("Finally (teardown, always runs)")
		md.CodeBlocks(markdown.SyntaxHighlightShell, strings.Join(td, "\n"))
	}

	if exact := exactPreviews(sc, specDir, outputDir); len(exact) > 0 {
		md.H4("Expected output")
		writePreviews(md, exact)
	}

	if gen := generatedArtifacts(sc); len(gen) > 0 {
		md.H4("Generated artifacts")
		md.BulletList(gen...)
	}
}

// matrixExpander returns a display-only ${name} expander seeded with the
// scenario's matrix-row variables. Runtime-captured variables (store,
// ready.store) stay as ${name} — their values exist only at run time, and the
// unresolved reference is exactly what the reader should see.
func matrixExpander(sc *spec.Scenario) func(string) string {
	if len(sc.Vars) == 0 {
		return func(s string) string { return s }
	}
	st := store.New()
	for k, v := range sc.Vars {
		st.Set(k, v)
	}
	return st.Expand
}

// writePreviews renders a list of preview blocks (#67) as labeled fenced code
// blocks. A block with an empty body (e.g. a snapshot reference) renders just its
// label so the reader still sees that the payload is authored elsewhere. An image
// block renders a Markdown image embed so a committed golden renders inline.
func writePreviews(md *markdown.Markdown, blocks []previewBlock) {
	for _, b := range blocks {
		md.PlainTextf("_%s:_", b.label)
		lang := b.lang
		if lang == "" {
			// "text" keeps the generated docs markdownlint-clean (MD040)
			// without guessing the payload's real format.
			lang = "text"
		}
		switch {
		case b.image && b.body != "":
			md.PlainTextf("![%s](%s)", b.label, b.body)
		case b.body != "":
			md.CodeBlocks(markdown.SyntaxHighlight(lang), b.body)
		}
	}
}

func scenarioMeta(sc *spec.Scenario) string {
	var parts []string
	if len(sc.Tags) > 0 {
		parts = append(parts, "tags: "+strings.Join(sc.Tags, ", "))
	}
	if sc.Only != nil && sc.Only.OS != "" {
		parts = append(parts, "only on "+sc.Only.OS)
	}
	if sc.Skip != nil && sc.Skip.OS != "" {
		parts = append(parts, "skipped on "+sc.Skip.OS)
	}
	if len(parts) == 0 {
		return ""
	}
	return "_" + strings.Join(parts, " · ") + "_"
}

func givenBullets(sc *spec.Scenario, expand func(string) string) []string {
	var out []string
	// Background services are part of the given world: they are started before
	// the scenario's steps run, so document them up front (#41).
	for i := range sc.Services {
		svc := &sc.Services[i]
		out = append(out, fmt.Sprintf("Background service `%s` is started: `%s`.", svc.Name, expand(svc.Command)))
	}
	for i := range sc.MockServers {
		ms := &sc.MockServers[i]
		out = append(out, fmt.Sprintf("Stub HTTP server `%s` serves %d canned route(s) at `${%s.url}` and records every request (#24).", ms.Name, len(ms.Routes), ms.Name))
	}
	for i := range sc.Steps {
		step := &sc.Steps[i]
		switch step.Kind() {
		case spec.StepFixture:
			out = append(out, fmt.Sprintf("Fixture file `%s` is created.", step.Fixture.File))
		case spec.StepRun:
			if len(step.Run.Env) > 0 {
				out = append(out, "Environment variables are set: "+strings.Join(sortedEnvKeys(step.Run.Env), ", ")+".")
			}
			if step.Run.ClearEnvEnabled() {
				out = append(out, clearedEnvBullet(step.Run.PassEnv))
			}
		case spec.StepPTY:
			if step.PTY.ClearEnvEnabled() {
				out = append(out, clearedEnvBullet(step.PTY.PassEnv))
			}
		}
	}
	return out
}

// clearedEnvBullet renders the hermetic-environment Given bullet shared by run
// and pty steps (#16).
func clearedEnvBullet(passEnv []string) string {
	bullet := "The command runs with a cleared environment"
	if len(passEnv) > 0 {
		bullet += " (passing through: " + strings.Join(passEnv, ", ") + ")"
	}
	return bullet + "."
}

// commands renders the "When" narrative. It covers every action step kind, not
// just run steps, so HTTP/query/gRPC/CDP interactions are documented too (#41),
// and store steps appear as comments so a later ${name} reference is explained
// where it is born instead of appearing out of nowhere.
func commands(steps []spec.Step, expand func(string) string) []string {
	var out []string
	for i := range steps {
		step := &steps[i]
		switch step.Kind() {
		case spec.StepRun:
			out = append(out, expand(step.Run.Command))
		case spec.StepHTTP:
			if step.HTTP != nil {
				out = append(out, fmt.Sprintf("# HTTP %s %s", step.HTTP.Method, expand(step.HTTP.Path)))
			}
		case spec.StepQuery:
			if step.Query != nil {
				out = append(out, fmt.Sprintf("# SQL via %s: %s", step.Query.Runner, expand(step.Query.SQL)))
			}
		case spec.StepGRPC:
			if step.GRPC != nil {
				out = append(out, fmt.Sprintf("# gRPC %s via %s", step.GRPC.Method, step.GRPC.Runner))
			}
		case spec.StepPTY:
			if step.PTY != nil {
				out = append(out, fmt.Sprintf("# interactive (pty): %s", expand(step.PTY.Command)))
			}
		case spec.StepCDP:
			if step.CDP != nil {
				out = append(out, "# CDP via "+step.CDP.Runner+": "+cdpActions(step.CDP))
			}
		case spec.StepStore:
			if step.Store != nil {
				out = append(out, fmt.Sprintf("# capture ${%s} from %s", step.Store.Name, storeSourceLabel(step.Store)))
			}
		case spec.StepSignal:
			if step.Signal != nil {
				line := fmt.Sprintf("# send SIG%s to service %s", spec.NormalizeSignalName(step.Signal.Signal), expand(step.Signal.Service))
				if step.Signal.Wait != nil {
					timeout := step.Signal.Wait.Timeout
					if timeout == "" {
						timeout = "5s"
					}
					line += fmt.Sprintf(" and wait up to %s for exit", timeout)
				}
				out = append(out, line)
			}
		}
	}
	return out
}

// storeSourceLabel names where a store step reads its value from.
func storeSourceLabel(st *spec.Store) string {
	f := st.From
	switch {
	case f == nil:
		return "the last result"
	case f.Stdout != nil:
		return "stdout"
	case f.Body != nil:
		return "the response body"
	case f.File != nil:
		return "file " + f.File.Path
	case f.Header != "":
		return "response header " + f.Header
	case f.Rows != nil:
		return "the result rows"
	case f.Message != nil:
		return "the response message"
	case f.Value != nil:
		return "the captured value"
	default:
		return "the last result"
	}
}

// cdpActions renders a cdp step's action list as a single readable line, reusing
// the shared per-action labels so doc stays aligned with explain/manifest (#50).
func cdpActions(c *spec.CDP) string {
	acts := make([]string, 0, len(c.Actions))
	for _, a := range c.Actions {
		acts = append(acts, spec.CDPActionLabel(a))
	}
	return strings.Join(acts, " → ")
}

// thenGroup ties the assert bullets that follow one action step back to that
// action. Without the grouping, a scenario with several commands pools every
// assertion into one list and the reader has to line up "exit code is 0 /
// exit code is not 0" with the right command by hand.
type thenGroup struct {
	actionIdx int    // step index of the action these bullets check (-1: none yet)
	action    string // short label of that action
	bullets   []string
}

// thenGroups walks the steps and groups each assert under the most recent
// action step (run/http/query/grpc/cdp). Fixture and store steps never break a
// group: they observe nothing.
func thenGroups(sc *spec.Scenario, expand func(string) string) []thenGroup {
	var groups []thenGroup
	actionIdx, actionLbl := -1, ""
	for i := range sc.Steps {
		step := &sc.Steps[i]
		switch step.Kind() {
		case spec.StepRun, spec.StepHTTP, spec.StepQuery, spec.StepGRPC, spec.StepCDP, spec.StepSignal:
			actionIdx, actionLbl = i, actionLabel(step, expand)
		case spec.StepAssert:
			if len(groups) == 0 || groups[len(groups)-1].actionIdx != actionIdx {
				groups = append(groups, thenGroup{actionIdx: actionIdx, action: actionLbl})
			}
			g := &groups[len(groups)-1]
			for _, b := range describeAsserts(step.Assert) {
				g.bullets = append(g.bullets, expand(b))
			}
		}
	}
	return groups
}

// actionLabel is the short inline label a Then group uses to name the action it
// checks.
func actionLabel(step *spec.Step, expand func(string) string) string {
	switch step.Kind() {
	case spec.StepRun:
		return expand(step.Run.Command)
	case spec.StepHTTP:
		return fmt.Sprintf("HTTP %s %s", step.HTTP.Method, expand(step.HTTP.Path))
	case spec.StepQuery:
		return expand(step.Query.SQL)
	case spec.StepGRPC:
		return "gRPC " + step.GRPC.Method
	case spec.StepCDP:
		return "the browser flow"
	case spec.StepSignal:
		return "SIG" + spec.NormalizeSignalName(step.Signal.Signal) + " to " + expand(step.Signal.Service)
	default:
		return ""
	}
}

// writeThen renders the Then section. A scenario with at most one action keeps
// the flat bullet list; with several, each group opens with "after
// `<command>`:" so every assertion reads against its command — even when only
// the last command is asserted on.
func writeThen(md *markdown.Markdown, sc *spec.Scenario, expand func(string) string) {
	groups := thenGroups(sc, expand)
	if len(groups) == 0 {
		return
	}
	// Keep this action set in sync with thenGroups: a drift flattens grouped
	// bullets (or vice versa) for scenarios mixing the two step kinds.
	actions := 0
	for i := range sc.Steps {
		switch sc.Steps[i].Kind() {
		case spec.StepRun, spec.StepHTTP, spec.StepQuery, spec.StepGRPC, spec.StepCDP, spec.StepSignal:
			actions++
		}
	}
	md.H4("Then")
	if actions <= 1 {
		var flat []string
		for _, g := range groups {
			flat = append(flat, g.bullets...)
		}
		md.BulletList(flat...)
		return
	}
	var b strings.Builder
	for _, g := range groups {
		if g.action == "" {
			for _, bl := range g.bullets {
				fmt.Fprintf(&b, "- %s\n", bl)
			}
			continue
		}
		fmt.Fprintf(&b, "- after %s:\n", markdown.Code(g.action))
		for _, bl := range g.bullets {
			fmt.Fprintf(&b, "  - %s\n", bl)
		}
	}
	md.PlainText(strings.TrimRight(b.String(), "\n"))
}

// generatedArtifacts lists the files this scenario produces, reusing the shared
// spec model so image outputs and browser screenshots are documented too — not
// just file exists:true assertions (#56).
func generatedArtifacts(sc *spec.Scenario) []string {
	paths := spec.GeneratedArtifacts(sc)
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		out = append(out, markdown.Code(p))
	}
	return out
}
