package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/nao1215/atago/internal/diag"
	"github.com/nao1215/atago/internal/explain"
	"github.com/nao1215/atago/internal/loader"
)

// codeArgRe matches an argument written like a diagnostic code, whether or not
// any such code is assigned. It decides which of the two things `atago explain`
// was asked about, so an unassigned number is reported as an unknown code
// rather than as a missing file.
var codeArgRe = regexp.MustCompile(`^(?i:atg)[0-9]{4}$`)

// explainCodes reports whether the operands name diagnostic codes rather than
// spec paths. Mixing the two is refused by returning false, which sends the
// arguments down the spec path and lets it complain about the code as a file —
// there is no sensible output for "explain this spec and also this code".
func explainCodes(operands []string) (string, bool) {
	if len(operands) != 1 || !codeArgRe.MatchString(strings.TrimSpace(operands[0])) {
		return "", false
	}
	return strings.TrimSpace(operands[0]), true
}

// explainCode prints one diagnostic's entry from the same registry the website
// is generated from. It exists so that a code found in a CI log can be looked
// up where the log is — in a container, over ssh, with no browser.
func explainCode(arg string, stdout, stderr io.Writer) int {
	code, ok := diag.Parse(arg)
	if !ok {
		msg := fmt.Sprintf("unknown diagnostic code %q", arg)
		if near, found := nearestCode(arg); found {
			msg += fmt.Sprintf(" (did you mean %s?)", near)
		}
		fmt.Fprintf(stderr, "atago explain: %s\n", diag.BadOptionValue.Annotate(msg))
		fmt.Fprintln(stderr, "Every code is listed at https://nao1215.github.io/atago/errors/")
		return ExitConfig
	}
	e, _ := diag.Lookup(code)
	fmt.Fprint(stdout, diag.Text(e))
	return ExitOK
}

// nearestCode finds the assigned code closest in number to an unassigned one,
// so a mistyped digit points somewhere useful instead of nowhere.
func nearestCode(arg string) (diag.Code, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(arg)[3:])
	if err != nil {
		return 0, false
	}
	best, bestDist := diag.Code(0), 0
	for _, e := range diag.All() {
		d := int(e.Code) - n
		if d < 0 {
			d = -d
		}
		// Only within the same family: ATG4999 is a mistyped exec code, and
		// pointing it at a spec code would be worse than saying nothing.
		if e.Code.ExitCode() != n/1000 {
			continue
		}
		if best == 0 || d < bestDist {
			best, bestDist = e.Code, d
		}
	}
	return best, best != 0
}

// explainCmd implements `atago explain`: describe what one or more
// specs do without executing them.
func explainCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("atago explain", flag.ContinueOnError)
	fs.SetOutput(stderr)
	printUsage := func(w io.Writer) {
		fmt.Fprint(w, "Usage: atago explain <path | dir>...  (directories are searched recursively; default \".\")\n")
		fmt.Fprint(w, "       atago explain ATG2201            (what a diagnostic code means)\n")
		fs.SetOutput(w)
		fs.PrintDefaults()
	}
	// Suppress the flag package's automatic usage print; usage is routed
	// explicitly below — to stdout for an explicit --help (so it can be piped),
	// to stderr for a genuine parse error.
	fs.Usage = func() {}
	operands, err := parseFlagsAnywhere(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printUsage(stdout)
			return ExitOK
		}
		reportFlagError("atago explain", err, stderr)
		printUsage(stderr)
		return ExitConfig
	}

	// A diagnostic code where a path would go looks up what the code means.
	// The two cannot be confused: a spec path never spells ATG followed by four
	// digits, and a code is never a file on disk.
	if code, ok := explainCodes(operands); ok {
		return explainCode(code, stdout, stderr)
	}

	paths, exitCode, ok := specTargets("atago explain", operands, stderr)
	if !ok {
		return exitCode
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
