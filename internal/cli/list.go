package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/nao1215/atago/internal/loader"
	"github.com/nao1215/atago/internal/manifest"
)

// listCmd implements `atago list` (#63): read specs without executing them and
// surface the runnable scenarios — suite, scenario, tags, skip/only gates, and
// generated artifacts — for discovery-oriented workflows. It shares the same
// path semantics as run/doc/explain/manifest and reuses the manifest builder so
// the reported shape never drifts from `atago manifest`.
//
// The default output is a readable table; --json emits a deterministic document
// for tooling.
func listCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("atago list", flag.ContinueOnError)
	fs.SetOutput(stderr)
	asJSON := fs.Bool("json", false, "emit a stable JSON document instead of a table")
	printUsage := func(w io.Writer) {
		fmt.Fprint(w, "Usage: atago list [--json] <path | dir>...\n  (directories are searched recursively; default \".\")\n")
		fs.SetOutput(w)
		fs.PrintDefaults()
	}
	// Suppress the flag package's automatic usage print; usage is routed
	// explicitly below — to stdout for an explicit --help (so it can be piped),
	// to stderr for a genuine parse error.
	fs.Usage = func() {}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printUsage(stdout)
			return ExitOK
		}
		printUsage(stderr)
		return ExitConfig
	}

	targets := fs.Args()
	if len(targets) == 0 {
		targets = []string{"."}
	}
	paths, err := collectSpecFiles(targets)
	if err != nil {
		fmt.Fprintf(stderr, "atago list: %v\n", err)
		return ExitConfig
	}
	if len(paths) == 0 {
		fmt.Fprintln(stderr, "atago list: no *.atago.yaml files found")
		return ExitConfig
	}

	inputs := make([]manifest.Input, 0, len(paths))
	for _, p := range paths {
		s, lerr := loader.Load(p)
		if lerr != nil {
			fmt.Fprintf(stderr, "%v\n", lerr)
			return ExitParse
		}
		inputs = append(inputs, manifest.Input{Spec: s, Path: p})
	}
	doc := manifest.Build(inputs)

	if *asJSON {
		return writeListJSON(doc, stdout, stderr)
	}
	return writeListTable(doc, stdout, stderr)
}

// listRow is one scenario row in the JSON contract. It is a deliberately small,
// discovery-oriented projection of the manifest — enough to build filters
// (--filter/--tag/--skip-tag) and to preview generated artifacts without reading
// the full manifest or executing the suite.
type listRow struct {
	SpecPath  string   `json:"spec_path"`
	Suite     string   `json:"suite"`
	Scenario  string   `json:"scenario"`
	Tags      []string `json:"tags,omitempty"`
	Gates     []string `json:"gates,omitempty"`
	Artifacts []string `json:"artifacts,omitempty"`
}

type listDocument struct {
	SchemaVersion string    `json:"schema_version"`
	Scenarios     []listRow `json:"scenarios"`
}

// ListSchemaVersion versions the `atago list --json` contract independently of
// the manifest, since it is a distinct (smaller) projection.
const ListSchemaVersion = "1"

func listRows(doc manifest.Document) []listRow {
	var rows []listRow
	for _, sp := range doc.Specs {
		for _, sc := range sp.Scenarios {
			rows = append(rows, listRow{
				SpecPath:  sp.SpecPath,
				Suite:     sp.Suite,
				Scenario:  sc.Name,
				Tags:      sc.Tags,
				Gates:     scenarioGates(sc),
				Artifacts: sc.Generates,
			})
		}
	}
	return rows
}

// scenarioGates renders the skip/only conditions as short, sorted tokens such as
// "only:os=linux" or "skip:env=CI", so both the table and JSON expose the same
// gate summary deterministically.
func scenarioGates(sc manifest.Scenario) []string {
	var gates []string
	if g := gateTokens("only", sc.Only); len(g) > 0 {
		gates = append(gates, g...)
	}
	if g := gateTokens("skip", sc.Skip); len(g) > 0 {
		gates = append(gates, g...)
	}
	return gates
}

func gateTokens(kind string, c *manifest.Condition) []string {
	if c == nil {
		return nil
	}
	var out []string
	if c.OS != "" {
		out = append(out, fmt.Sprintf("%s:os=%s", kind, c.OS))
	}
	if c.Env != "" {
		out = append(out, fmt.Sprintf("%s:env=%s", kind, c.Env))
	}
	if c.Command != "" {
		out = append(out, fmt.Sprintf("%s:command", kind))
	}
	sort.Strings(out)
	return out
}

func writeListJSON(doc manifest.Document, stdout, stderr io.Writer) int {
	out := listDocument{SchemaVersion: ListSchemaVersion, Scenarios: listRows(doc)}
	if out.Scenarios == nil {
		out.Scenarios = []listRow{}
	}
	payload, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "atago list: %v\n", err)
		return ExitInternal
	}
	payload = append(payload, '\n')
	if _, err := stdout.Write(payload); err != nil {
		return ExitInternal
	}
	return ExitOK
}

func writeListTable(doc manifest.Document, stdout, stderr io.Writer) int {
	tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SUITE\tSCENARIO\tTAGS\tGATES\tARTIFACTS")
	for _, r := range listRows(doc) {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			r.Suite, r.Scenario, joinOrDash(r.Tags), joinOrDash(r.Gates), joinOrDash(r.Artifacts))
	}
	if err := tw.Flush(); err != nil {
		fmt.Fprintf(stderr, "atago list: %v\n", err)
		return ExitInternal
	}
	return ExitOK
}

func joinOrDash(items []string) string {
	if len(items) == 0 {
		return "-"
	}
	return strings.Join(items, ",")
}
