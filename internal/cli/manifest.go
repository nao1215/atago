package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/nao1215/atago/internal/loader"
	"github.com/nao1215/atago/internal/manifest"
)

// manifestCmd implements `atago manifest` (#49): read specs without executing
// them and emit a stable, machine-readable JSON summary of what they declare.
// Without --out it writes to stdout.
func manifestCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("atago manifest", flag.ContinueOnError)
	fs.SetOutput(stderr)
	out := fs.String("out", "", "write the JSON manifest to this file instead of stdout")
	fs.Usage = func() {
		fmt.Fprint(stderr, "Usage: atago manifest [--out file.json] <path | dir>...\n  (directories are searched recursively)\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitConfig
	}

	targets := fs.Args()
	if len(targets) == 0 {
		targets = []string{"."}
	}
	paths, err := collectSpecFiles(targets)
	if err != nil {
		fmt.Fprintf(stderr, "atago manifest: %v\n", err)
		return ExitConfig
	}
	if len(paths) == 0 {
		fmt.Fprintln(stderr, "atago manifest: no *.atago.yaml files found")
		return ExitConfig
	}

	inputs := make([]manifest.Input, 0, len(paths))
	for _, p := range paths {
		s, src, lerr := loader.LoadWithSource(p)
		if lerr != nil {
			fmt.Fprintf(stderr, "%v\n", lerr)
			return ExitParse
		}
		// src carries authored line/column positions so the manifest exposes source
		// locations for editor/review tooling (#80).
		inputs = append(inputs, manifest.Input{Spec: s, Path: p, Source: src})
	}

	doc := manifest.Build(inputs)
	payload, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "atago manifest: %v\n", err)
		return ExitInternal
	}
	payload = append(payload, '\n')

	if *out == "" {
		if _, err := stdout.Write(payload); err != nil {
			return ExitInternal
		}
		return ExitOK
	}
	if err := os.WriteFile(*out, payload, 0o600); err != nil {
		fmt.Fprintf(stderr, "atago manifest: %v\n", err)
		return ExitConfig
	}
	fmt.Fprintf(stdout, "Wrote %s\n", *out)
	return ExitOK
}
