package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/nao1215/atago/internal/explain"
	"github.com/nao1215/atago/internal/loader"
)

// explainCmd implements `atago explain`: describe what one or more
// specs do without executing them.
func explainCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("atago explain", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprint(stderr, "Usage: atago explain <path | dir>...  (directories are searched recursively; default \".\")\n")
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
		targets = []string{"."} // parity with run/lint/doc
	}
	paths, err := collectSpecFiles(targets)
	if err != nil {
		fmt.Fprintf(stderr, "atago explain: %v\n", err)
		return ExitConfig
	}
	if len(paths) == 0 {
		fmt.Fprintln(stderr, "atago explain: no *.atago.yaml files found")
		return ExitConfig
	}

	exit := ExitOK
	for _, p := range paths {
		s, lerr := loader.Load(p)
		if lerr != nil {
			fmt.Fprintf(stderr, "%v\n", lerr)
			exit = worseExit(exit, ExitParse)
			continue
		}
		if err := explain.Explain(stdout, s, p); err != nil {
			fmt.Fprintf(stderr, "atago explain: %v\n", err)
			return worseExit(exit, ExitInternal)
		}
	}
	return exit
}
