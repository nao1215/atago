// Command atago is a YAML-based black-box behavior spec runner for CLIs, APIs,
// and generated artifacts. See the README for the format and CLI contract, and
// schema/atago.schema.json for the machine-readable spec file schema.
package main

import (
	"os"

	"github.com/nao1215/atago/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:], os.Stdout, os.Stderr))
}
