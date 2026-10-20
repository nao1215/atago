package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/cli"
	"github.com/nao1215/atago/internal/diag"
	"github.com/nao1215/atago/internal/loader"
)

// The agent skill under skills/atago is read by coding agents, which copy its
// specs and commands verbatim. A skill that teaches a key the loader rejects,
// a flag that was renamed, or a failure report atago no longer prints teaches
// the agent to fail. This file runs the skill the way the agent would.
//
// A fenced block is checked according to the HTML comment on the line above it:
//
//	<!-- verify: spec NAME -->  a ```yaml block, written to NAME in a scratch
//	                            directory shared by the whole page; it must load
//	<!-- verify: spec NAME invalid -->
//	                            the same, for a page that shows a spec error;
//	                            it must fail to load
//	<!-- verify: run -->        a ```shell block whose `atago` lines are run in
//	<!-- verify: run exit=N --> order through cli.Main; each must exit N (0)
//	<!-- verify: output -->     a ```text block that must appear in the output
//	                            of the last command run above it
//
// Every ```yaml block must carry a spec marker, and every valid spec must be
// run by a later block, so no spec in the skill goes untested. An unmarked ```shell block is not run, but each `atago` line in it
// must name a real subcommand and only flags that subcommand accepts.

// skillPages lists the skill's Markdown pages.
func skillPages(t *testing.T) []string {
	t.Helper()
	refs, err := filepath.Glob("skills/atago/references/*.md")
	if err != nil {
		t.Fatalf("glob skill references: %v", err)
	}
	if len(refs) == 0 {
		t.Fatal("skills/atago/references has no pages; the glob or the layout is wrong")
	}
	return append([]string{"skills/atago/SKILL.md"}, refs...)
}

// skillBlockRe captures a fenced block with its language and the line above it.
var skillBlockRe = regexp.MustCompile("(?m)^(.*)\n```([a-z]*)\n((?s:.*?))^```$")

// skillVerifyRe reads a verify marker.
var skillVerifyRe = regexp.MustCompile(`^<!-- verify: ([a-z]+)(?: (.*?))? -->$`)

// skillDurationRe matches the run time in a summary line, which differs per run.
var skillDurationRe = regexp.MustCompile(`\((?:[0-9.]+(?:h|m|s|ms|µs|us|ns))+\)`)

type skillBlock struct {
	marker, arg, lang, body string
}

func skillBlocks(t *testing.T, page string) []skillBlock {
	t.Helper()
	var blocks []skillBlock
	for _, m := range skillBlockRe.FindAllStringSubmatch(readDoc(t, page), -1) {
		b := skillBlock{lang: m[2], body: m[3]}
		if v := skillVerifyRe.FindStringSubmatch(strings.TrimSpace(m[1])); v != nil {
			b.marker, b.arg = v[1], v[2]
		}
		blocks = append(blocks, b)
	}
	if len(blocks) == 0 {
		t.Fatalf("%s has no fenced blocks; the parser or the page is wrong", page)
	}
	return blocks
}

// skillCommands returns the argv (without "atago") of each `atago` line in a
// shell block. A trailing `# comment` is dropped.
func skillCommands(body string) [][]string {
	var cmds [][]string
	for line := range strings.SplitSeq(body, "\n") {
		line, _, _ = strings.Cut(line, " #")
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] != "atago" {
			continue
		}
		cmds = append(cmds, fields[1:])
	}
	return cmds
}

// normalizeSkillOutput removes what legitimately differs between the page and
// a run: the scratch directory, the run time, CRLF, and trailing blanks.
func normalizeSkillOutput(s, dir string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, dir+string(filepath.Separator), "")
	s = skillDurationRe.ReplaceAllString(s, "(<duration>)")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// TestSkill_BlocksRunAsWritten runs every verified block of every skill page.
func TestSkill_BlocksRunAsWritten(t *testing.T) {
	t.Parallel()
	schema := loadSchema(t)
	for _, page := range skillPages(t) {
		t.Run(page, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			var last string
			ran := 0
			unrun := map[string]bool{} // valid specs no command has run yet
			for i, b := range skillBlocks(t, page) {
				switch {
				case b.lang == "yaml" && b.marker != "spec":
					t.Errorf("block %d: a ```yaml block needs <!-- verify: spec NAME --> so it is loaded", i)
				case b.marker == "spec":
					name, mode, _ := strings.Cut(b.arg, " ")
					if b.lang != "yaml" || !strings.HasSuffix(name, ".atago.yaml") || (mode != "" && mode != "invalid") {
						t.Fatalf("block %d: verify: spec needs a ```yaml block and NAME.atago.yaml [invalid], got ```%s %q", i, b.lang, b.arg)
					}
					path := filepath.Join(dir, name)
					if err := os.WriteFile(path, []byte(b.body), 0o600); err != nil {
						t.Fatalf("write %s: %v", name, err)
					}
					_, err := loader.Load(path)
					switch {
					case mode == "" && err != nil:
						t.Errorf("block %d (%s) does not load: %v", i, name, err)
					case mode == "":
						unrun[name] = true
						// The editor validates against the JSON Schema, not the
						// loader, so a spec the agent copies must pass both.
						if err := schema.Validate(yamlToAny(t, []byte(b.body))); err != nil {
							t.Errorf("block %d (%s) does not conform to schema/atago.schema.json:\n%v", i, name, err)
						}
					case mode == "invalid" && err == nil:
						t.Errorf("block %d (%s) is marked invalid but loads", i, name)
					}
				case b.marker == "run":
					want := cli.ExitOK
					if code, ok := strings.CutPrefix(b.arg, "exit="); ok {
						n, err := strconv.Atoi(code)
						if err != nil {
							t.Fatalf("block %d: bad exit in %q", i, b.arg)
						}
						want = n
					}
					cmds := skillCommands(b.body)
					if len(cmds) == 0 {
						t.Fatalf("block %d: verify: run on a block with no atago command", i)
					}
					for _, args := range cmds {
						for _, a := range args {
							delete(unrun, a)
						}
						got, out := runSkillCommand(dir, args)
						if got != want {
							t.Errorf("block %d: `atago %s` exited %d, want %d\n%s", i, strings.Join(args, " "), got, want, out)
						}
						last = out
						ran++
					}
				case b.marker == "output":
					if ran == 0 {
						t.Fatalf("block %d: verify: output with no command run above it", i)
					}
					want := normalizeSkillOutput(b.body, dir)
					got := normalizeSkillOutput(last, dir)
					if !strings.Contains(got, want) {
						t.Errorf("block %d: the page shows output atago does not print\n--- page\n%s\n--- atago\n%s", i, want, got)
					}
				case b.marker != "":
					t.Errorf("block %d: unknown marker %q", i, b.marker)
				}
			}
			for name := range unrun {
				t.Errorf("%s is loaded but never run; add a <!-- verify: run --> block that runs it", name)
			}
		})
	}
}

// runSkillCommand runs one `atago` line with its spec paths moved into dir.
// Everything after `--` belongs to the recorded program and is left alone.
func runSkillCommand(dir string, args []string) (int, string) {
	args = slices.Clone(args)
	for i, a := range args {
		if a == "--" {
			break
		}
		if strings.HasSuffix(a, ".atago.yaml") {
			args[i] = filepath.Join(dir, a)
		}
	}
	var stdout, stderr bytes.Buffer
	code := cli.Main(args, &stdout, &stderr)
	return code, stdout.String() + stderr.String()
}

// skillFlagRe reads the flag names a subcommand's --help prints.
var skillFlagRe = regexp.MustCompile(`(?m)^  -([a-z][a-z-]*)`)

// TestSkill_CommandsAreReal checks the `atago` lines the skill shows but does
// not run: the subcommand must exist and every flag must be one it accepts.
func TestSkill_CommandsAreReal(t *testing.T) {
	t.Parallel()
	known := map[string]bool{}
	for _, name := range cli.Subcommands() {
		known[name] = true
	}
	flags := map[string]map[string]bool{}
	flagsOf := func(sub string) map[string]bool {
		if f, ok := flags[sub]; ok {
			return f
		}
		var stdout, stderr bytes.Buffer
		cli.Main([]string{sub, "--help"}, &stdout, &stderr)
		f := map[string]bool{}
		for _, m := range skillFlagRe.FindAllStringSubmatch(stdout.String()+stderr.String(), -1) {
			f[m[1]] = true
		}
		flags[sub] = f
		return f
	}
	inline := regexp.MustCompile("`atago ([^`]+)`")
	checked := 0
	for _, page := range skillPages(t) {
		var lines [][]string
		for _, b := range skillBlocks(t, page) {
			if b.lang == "shell" {
				lines = append(lines, skillCommands(b.body)...)
			}
		}
		// Commands quoted in the prose are copied just as often as blocks.
		for _, m := range inline.FindAllStringSubmatch(readDoc(t, page), -1) {
			lines = append(lines, strings.Fields(m[1]))
		}
		for _, args := range lines {
			if len(args) == 0 {
				continue
			}
			sub := args[0]
			if !known[sub] {
				t.Errorf("%s: `atago %s` names no subcommand (known: %v)", page, strings.Join(args, " "), cli.Subcommands())
				continue
			}
			for _, a := range args[1:] {
				if a == "--" {
					break
				}
				name, ok := strings.CutPrefix(a, "--")
				if !ok {
					continue
				}
				name, _, _ = strings.Cut(name, "=")
				if !flagsOf(sub)[name] {
					t.Errorf("%s: `atago %s` passes --%s, which `atago %s --help` does not list", page, strings.Join(args, " "), name, sub)
				}
			}
			checked++
		}
	}
	if checked == 0 {
		t.Fatal("the skill shows no atago command; the parser is wrong")
	}
}

// TestSkill_ReferencesExist checks the codes and example files the skill names.
func TestSkill_ReferencesExist(t *testing.T) {
	t.Parallel()
	codeRe := regexp.MustCompile(`ATG[0-9]{4}`)
	exampleRe := regexp.MustCompile(`examples/[a-z_]+\.atago\.yaml`)
	for _, page := range skillPages(t) {
		doc := readDoc(t, page)
		for _, c := range codeRe.FindAllString(doc, -1) {
			code, ok := diag.Parse(c)
			if !ok {
				t.Errorf("%s names %s, which is not a registered code", page, c)
				continue
			}
			if _, ok := diag.Lookup(code); !ok {
				t.Errorf("%s names %s, which has no entry", page, c)
			}
		}
		for _, p := range exampleRe.FindAllString(doc, -1) {
			if _, err := os.Stat(p); err != nil {
				t.Errorf("%s names %s, which does not exist", page, p)
			}
		}
	}
}
