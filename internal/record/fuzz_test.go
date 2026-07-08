package record

import (
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/loader"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

// fuzzWorkdir is a fixed realistic record-time scratch path (the CLI always
// passes an os.MkdirTemp path, which never contains $ or control bytes).
const fuzzWorkdir = "/tmp/atago-record-fz123"

// posixQuote models cli.posixJoin's quoting for one argv token: every non-shell
// recorded command is built from tokens quoted this way (posixJoin leaves plain
// words bare, but quoting a plain word is also a valid posixJoin output shape).
func posixQuote(a string) string {
	return "'" + strings.ReplaceAll(a, "'", `'\''`) + "'"
}

// FuzzGenerateRoundTrip drives `atago record`'s spec generator with arbitrary
// observed command lines, exit codes, raw stream bytes, and created-file names.
// Invariants (the round-trip law, #30):
//   - Generate never returns an error for inputs the record CLI can actually
//     produce ("generated spec does not validate" is an internal bug by its own
//     contract), and never panics;
//   - the generated YAML reparses, and expanding the parsed run.command with an
//     empty store restores the observed command byte-for-byte;
//   - the parsed exit_code equals the observed one;
//   - the stdout contains-anchor, expanded with workdir bound to the record-time
//     scratch dir, restores the first non-empty trimmed stdout line;
//   - each generated file.path expands back to the observed created file.
func FuzzGenerateRoundTrip(f *testing.F) {
	f.Add("echo", "hi", false, 0, []byte("hi\n"), []byte(""), "out.txt")
	f.Add("printf", "%s\t${1}", false, 1, []byte("\ta b\n"), []byte("err"), "a/b.txt")
	f.Add("cat", "", true, 0, []byte{0xff, 0xfe, '\n'}, []byte{}, "")
	f.Add("echo ${workdir} | tee x", "", true, 0, []byte(fuzzWorkdir+"/x here\n"), []byte(""), "weird\nname")
	f.Add("go", "version # comment", false, 0, []byte("go version go1.24 linux/amd64 # x\n"), []byte{}, "läuft.txt")
	f.Fuzz(func(t *testing.T, arg0, arg1 string, shell bool, exitCode int, stdout, stderr []byte, file string) {
		// Model the CLI: shell mode joins argv verbatim; no-shell mode quotes each
		// token (posixJoin). argv tokens cannot contain NUL (execve forbids it).
		if strings.ContainsRune(arg0, 0) || strings.ContainsRune(arg1, 0) {
			t.Skip()
		}
		var command string
		if shell {
			command = strings.Join([]string{arg0, arg1}, " ")
		} else {
			command = posixQuote(arg0) + " " + posixQuote(arg1)
		}
		if command == "" {
			t.Skip() // probed separately: empty shell command is a known repro
		}

		// Model listFiles output: workdir-relative, /-separated, no empty/dot-dot
		// components, no NUL (impossible in a walked tree).
		var files []string
		if file != "" && validWalkedPath(file) {
			files = []string{file}
		}

		obs := Observation{
			Command:      command,
			Shell:        shell,
			ExitCode:     exitCode,
			Stdout:       stdout,
			Stderr:       stderr,
			CreatedFiles: files,
		}
		out, err := Generate(obs, Options{SuiteName: "fz", Workdir: fuzzWorkdir})
		if err != nil {
			t.Fatalf("Generate failed for a CLI-producible observation: %v\ncommand=%q shell=%v stdout=%q files=%q", err, command, shell, stdout, files)
		}
		s, err := loader.LoadBytes("recorded.atago.yaml", out)
		if err != nil {
			t.Fatalf("generated spec does not reparse: %v\n%s", err, out)
		}

		// Round-trip: run.command.
		run := findRun(s)
		if run == nil {
			t.Fatalf("generated spec has no run step:\n%s", out)
		}
		if got := store.New().Expand(run.Command); got != command {
			t.Fatalf("command round-trip broke: recorded %q, replayed %q\nspec:\n%s", command, got, out)
		}
		if run.ShellEnabled() != shell {
			t.Fatalf("shell flag round-trip broke: recorded %v, got %v", shell, run.ShellEnabled())
		}

		// Round-trip: exit_code, stdout anchor, file paths.
		st := store.New()
		st.Set("workdir", fuzzWorkdir)
		wantAnchor := firstLine(stdout)
		var sawExit, sawAnchor bool
		wantFiles := map[string]bool{}
		for _, sc := range s.Scenarios {
			for _, step := range sc.Steps {
				if step.Assert == nil {
					continue
				}
				a := step.Assert
				if a.ExitCode != nil && a.ExitCode.Equals != nil {
					sawExit = true
					if *a.ExitCode.Equals != exitCode {
						t.Fatalf("exit_code round-trip broke: recorded %d, got %d", exitCode, *a.ExitCode.Equals)
					}
				}
				if a.Stdout != nil && len(a.Stdout.Contains) > 0 {
					sawAnchor = true
					if got := st.Expand(a.Stdout.Contains[0]); got != wantAnchor {
						t.Fatalf("stdout anchor round-trip broke: recorded first line %q, replayed anchor %q\nspec:\n%s", wantAnchor, got, out)
					}
				}
				if a.File != nil {
					wantFiles[st.Expand(a.File.Path)] = true
				}
			}
		}
		if !sawExit {
			t.Fatalf("generated spec asserts no exit_code:\n%s", out)
		}
		if wantAnchor != "" && !sawAnchor {
			t.Fatalf("stdout had a non-empty first line %q but no contains anchor was generated:\n%s", wantAnchor, out)
		}
		for _, want := range files {
			if !wantFiles[want] {
				t.Fatalf("created file %q lost in round-trip (got %v)\nspec:\n%s", want, wantFiles, out)
			}
		}
	})
}

// validWalkedPath reports whether p could be produced by walking a workdir:
// relative, /-separated, and with no empty, ".", or ".." components.
func validWalkedPath(p string) bool {
	if strings.ContainsRune(p, 0) || strings.HasPrefix(p, "/") {
		return false
	}
	for _, c := range strings.Split(p, "/") {
		if c == "" || c == "." || c == ".." {
			return false
		}
	}
	return true
}

// findRun returns the first run step of the spec.
func findRun(s *spec.Spec) *spec.Run {
	for _, sc := range s.Scenarios {
		for _, step := range sc.Steps {
			if step.Run != nil {
				return step.Run
			}
		}
	}
	return nil
}
