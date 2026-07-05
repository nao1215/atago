package buildinfo

import "regexp"

// releaseTag matches a clean release tag (v1.2.3). A pseudo-version
// (v0.3.5-0.2026...-abcdef) or "dev" deliberately does not match: those refer to
// commits GitHub raw cannot resolve by name, so they fall back to "main".
var releaseTag = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

// SchemaRef is the git ref the emitted schema URL pins to: this build's release
// tag when it is one, otherwise "main". Pinning a released binary to its own tag
// keeps a generated spec's completion matched to the atago that scaffolded it,
// while dev builds still resolve against the latest schema on main (#121).
func SchemaRef() string {
	if v := Get(); releaseTag.MatchString(v) {
		return v
	}
	return "main"
}

// SchemaURL is the absolute, resolvable URL of the spec-file JSON schema for
// this build.
func SchemaURL() string {
	return "https://raw.githubusercontent.com/nao1215/atago/" + SchemaRef() + "/schema/atago.schema.json"
}

// SchemaHeader is the `# yaml-language-server: $schema=<url>` comment line that
// steers editor completion to the spec DSL, terminated with a newline. Emitted
// as the first line of every scaffolded/recorded spec so completion — the only
// delivery path for the DSL reference — is the default, not a thing users must
// discover and wire up by hand (#121).
func SchemaHeader() string {
	return "# yaml-language-server: $schema=" + SchemaURL() + "\n"
}
