package cli

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nao1215/atago/internal/diag"
)

// specTargets resolves a subcommand's positional arguments into the spec files
// it should act on, and reports the failures in the caller's own voice. Every
// subcommand that takes spec paths — run, list, explain, doc, manifest — needs
// the same three steps: default to the current directory, expand directories,
// and refuse an empty result. Keeping them here is what makes a new
// spec-reading subcommand behave identically without copying the messages, and
// stops the "no spec files found" wording from drifting between commands.
//
// ok is false when the caller should return the accompanying exit code.
func specTargets(label string, args []string, stderr io.Writer) (paths []string, exit int, ok bool) {
	targets := args
	if len(targets) == 0 {
		targets = []string{"."}
	}
	paths, err := collectSpecFiles(targets)
	if err != nil {
		fmt.Fprintf(stderr, "%s: %s\n", label, diag.TargetNotFound.Annotate(err.Error()))
		return nil, ExitConfig, false
	}
	if len(paths) == 0 {
		fmt.Fprintf(stderr, "%s: %s\n", label,
			diag.NoSpecFiles.Annotate(fmt.Sprintf("no *.atago.yaml (or *.atago.yml) files found in %s; run `atago init` to scaffold one", quotedList(targets))))
		return nil, ExitConfig, false
	}
	return paths, ExitOK, true
}

// statReason strips the syscall wrapper off a stat error so a message that has
// already named the path does not repeat it: "no such file or directory"
// rather than "stat specs/: no such file or directory".
func statReason(err error) error {
	var pe *fs.PathError
	if errors.As(err, &pe) {
		return pe.Err
	}
	return err
}

// quotedList renders the targets a subcommand was given, for a message about
// all of them.
func quotedList(targets []string) string {
	parts := make([]string, len(targets))
	for i, t := range targets {
		parts[i] = strconv.Quote(t)
	}
	return strings.Join(parts, ", ")
}

// collectSpecFiles resolves run targets into a deduplicated list of spec files.
// A target may be a spec file or a directory; a directory is always searched
// recursively for *.atago.yaml files. atago targets every kind of CLI, so it
// avoids the Go-specific "dir/..." glob — a plain directory is enough. A trailing
// "..." is tolerated for convenience but is no longer required.
func collectSpecFiles(targets []string) ([]string, error) {
	seen := make(map[string]bool)
	var out []string
	add := func(p string) {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}

	for _, t := range targets {
		t = strings.TrimSuffix(t, "...")
		t = strings.TrimSuffix(t, string(os.PathSeparator))
		if t == "" {
			t = "."
		}

		info, err := os.Stat(t)
		if err != nil {
			// os.Stat's error repeats the path and prefixes the syscall
			// ("stat x: no such file or directory"), which reads as noise next
			// to the path atago already named. Keep the reason only.
			return nil, fmt.Errorf("cannot access %q: %w", t, statReason(err))
		}
		if !info.IsDir() {
			add(filepath.Clean(t))
			continue
		}
		if err := walkSpecDir(t, add); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// walkSpecDir recursively collects every *.atago.yaml file under dir.
func walkSpecDir(dir string, add func(string)) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Say what atago was doing when it hit this. A bare
			// "open /etc/credstore: permission denied" reads as if atago
			// wanted that file, when it was searching a directory the user
			// named for spec files.
			return fmt.Errorf("cannot search %q for spec files: %w", dir, err)
		}
		if d.IsDir() {
			return nil
		}
		if isSpecFile(path) {
			add(filepath.Clean(path))
		}
		return nil
	})
}

func isSpecFile(p string) bool {
	return strings.HasSuffix(p, ".atago.yaml") || strings.HasSuffix(p, ".atago.yml")
}
