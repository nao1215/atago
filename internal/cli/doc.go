package cli

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nao1215/atago/internal/docgen"
	"github.com/nao1215/atago/internal/loader"
)

// docCmd implements `atago doc`: generate Markdown documentation
// from specs. Without --out it writes to stdout.
func docCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("atago doc", flag.ContinueOnError)
	fs.SetOutput(stderr)
	out := fs.String("out", "", "write Markdown to this file instead of stdout")
	outDir := fs.String("out-dir", "", "with --split-by-spec, write one file per spec plus index.md into this directory")
	split := fs.Bool("split-by-spec", false, "emit one Markdown file per spec and an index.md linking them (requires --out-dir)")
	fs.Usage = func() {
		fmt.Fprint(stderr, "Usage: atago doc [--out file.md | --split-by-spec --out-dir DIR] <path | dir>...\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitConfig
	}
	if *split && *outDir == "" {
		fmt.Fprintln(stderr, "atago doc: --split-by-spec requires --out-dir DIR")
		return ExitConfig
	}
	if *outDir != "" && !*split {
		fmt.Fprintln(stderr, "atago doc: --out-dir requires --split-by-spec")
		return ExitConfig
	}
	if *split && *out != "" {
		// The split branch writes into --out-dir and never honors --out; rejecting
		// the combination stops a silently-ignored --out (the file is never
		// written and no docs land where the user asked).
		fmt.Fprintln(stderr, "atago doc: --out and --split-by-spec are mutually exclusive (--split-by-spec writes one file per spec into --out-dir)")
		return ExitConfig
	}

	targets := fs.Args()
	if len(targets) == 0 {
		targets = []string{"."}
	}
	paths, err := collectSpecFiles(targets)
	if err != nil {
		fmt.Fprintf(stderr, "atago doc: %v\n", err)
		return ExitConfig
	}
	if len(paths) == 0 {
		fmt.Fprintln(stderr, "atago doc: no *.atago.yaml files found")
		return ExitConfig
	}

	sources := make([]docgen.Source, 0, len(paths))
	for _, p := range paths {
		s, lerr := loader.Load(p)
		if lerr != nil {
			fmt.Fprintf(stderr, "%v\n", lerr)
			return ExitParse
		}
		sources = append(sources, docgen.Source{Path: p, Spec: s})
	}

	if *split {
		return writeSplitDocs(sources, *outDir, stdout, stderr)
	}

	var buf bytes.Buffer
	if *out == "" {
		// Writing to stdout: no stable directory to anchor relative image links, so
		// golden images stay as text references.
		if err := docgen.Generate(&buf, sources); err != nil {
			fmt.Fprintf(stderr, "atago doc: %v\n", err)
			return ExitInternal
		}
		if _, err := stdout.Write(buf.Bytes()); err != nil {
			return ExitInternal
		}
		return ExitOK
	}
	// Anchor embedded golden images to the output file's directory.
	if err := docgen.GenerateTo(&buf, sources, filepath.Dir(*out)); err != nil {
		fmt.Fprintf(stderr, "atago doc: %v\n", err)
		return ExitInternal
	}
	if err := os.WriteFile(*out, buf.Bytes(), 0o600); err != nil {
		fmt.Fprintf(stderr, "atago doc: %v\n", err)
		return ExitConfig
	}
	fmt.Fprintf(stdout, "Wrote %s\n", *out)
	return ExitOK
}

// writeSplitDocs renders one Markdown file per spec plus an index.md into outDir
// (#68). The directory is created if needed; file names are deterministic and
// collision-free so the output is reproducible.
func writeSplitDocs(sources []docgen.Source, outDir string, stdout, stderr io.Writer) int {
	index, docs, err := docgen.GenerateSplit(sources, outDir)
	if err != nil {
		fmt.Fprintf(stderr, "atago doc: %v\n", err)
		return ExitInternal
	}
	if err := os.MkdirAll(outDir, 0o750); err != nil {
		fmt.Fprintf(stderr, "atago doc: %v\n", err)
		return ExitConfig
	}
	for _, d := range docs {
		p := filepath.Join(outDir, d.Name)
		if err := os.WriteFile(p, d.Content, 0o600); err != nil {
			fmt.Fprintf(stderr, "atago doc: %v\n", err)
			return ExitConfig
		}
	}
	indexPath := filepath.Join(outDir, "index.md")
	if err := os.WriteFile(indexPath, index, 0o600); err != nil {
		fmt.Fprintf(stderr, "atago doc: %v\n", err)
		return ExitConfig
	}
	fmt.Fprintf(stdout, "Wrote %d files and %s\n", len(docs), indexPath)
	return ExitOK
}
