package loader

import (
	"reflect"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/spec"
)

// emptyPolicy is the decision about what an EXPLICITLY empty value of a
// collection-valued spec field means — `contains: []`, `compare: []`, `skip: {}`
// — as opposed to the field being absent, which decodes to nil and every default
// reads as "unset".
//
// The two shapes look identical in Go for a plain slice or map, so an empty
// collection is easy to discard silently and run the spec with a default nobody
// wrote. `deterministic.compare: []` was exactly that: the author asked to
// compare nothing across runs and atago compared stdout and exit_code.
type emptyPolicy struct {
	// refuse is a spec whose only fault is that the field is written empty. When
	// set, the loader must reject it, and the message must name the field.
	refuse string
	// meaning states what an accepted empty value means. Required when there is
	// no refuse spec, so "nobody has thought about this one" cannot pass as a
	// decision.
	meaning string
}

// policySpec builds a minimal spec around one step.
func policySpec(step string) string {
	return "version: \"1\"\nsuite:\n  name: s\nscenarios:\n  - name: a\n    steps:\n      - " + step + "\n"
}

// emptyPolicies records the decision for every collection-valued field the spec
// exposes, keyed by "Type.Field". TestSpec_EveryCollectionFieldHasAnEmptyPolicy
// walks the spec by reflection and fails on a field with no entry, so a new list
// or map cannot ship without someone deciding what writing it empty means — the
// question nobody was asked for `compare`, `skip`, and `only`.
var emptyPolicies = map[string]emptyPolicy{
	// Refused: an empty value here is an authoring mistake, not a claim.
	"Spec.Scenarios":  {refuse: "version: \"1\"\nsuite:\n  name: s\nscenarios: []\n"},
	"Scenario.Steps":  {refuse: "version: \"1\"\nsuite:\n  name: s\nscenarios:\n  - name: a\n    steps: []\n"},
	"Scenario.Matrix": {refuse: "version: \"1\"\nsuite:\n  name: s\nscenarios:\n  - name: a\n    matrix: []\n    steps:\n      - run: {command: echo}\n"},
	"CDP.Actions": {refuse: "version: \"1\"\nsuite:\n  name: s\nrunners:\n  b: {type: browser}\nscenarios:\n" +
		"  - name: a\n    steps:\n      - cdp: {runner: b, actions: []}\n"},
	"Deterministic.Compare":    {refuse: policySpec("run: {command: echo, deterministic: {compare: []}}")},
	"StreamAssert.Contains":    {refuse: policySpec("run: {command: echo}") + "      - assert: {stdout: {contains: []}}\n"},
	"StreamAssert.NotContains": {refuse: policySpec("run: {command: echo}") + "      - assert: {stdout: {not_contains: []}}\n"},
	"FileAssert.Contains":      {refuse: policySpec("assert: {file: {path: out.txt, contains: []}}")},
	"FileAssert.NotContains":   {refuse: policySpec("assert: {file: {path: out.txt, not_contains: []}}")},

	// Accepted: an empty value is a claim the author makes on purpose.
	"ChangesAssert.Created":  {meaning: "created nothing — the exhaustive claim is the whole point of writing the category"},
	"ChangesAssert.Modified": {meaning: "modified nothing"},
	"ChangesAssert.Deleted":  {meaning: "deleted nothing"},

	// Accepted: empty and absent genuinely mean the same thing here, and every
	// renderer already shows them the same way.
	"Spec.Runners":          {meaning: "no declared runners"},
	"Spec.Secrets":          {meaning: "no masked values"},
	"Spec.Scrub":            {meaning: "no scrub rules, so snapshots are compared verbatim"},
	"Suite.Env":             {meaning: "no suite env"},
	"Suite.Setup":           {meaning: "no suite setup steps, same as omitting the block"},
	"Suite.Teardown":        {meaning: "no suite teardown steps"},
	"ScenarioDefaults.Env":  {meaning: "no default env"},
	"Scenario.Tags":         {meaning: "untagged; every renderer already shows it as a scenario with no tags"},
	"Scenario.Env":          {meaning: "no scenario env overrides"},
	"Scenario.Services":     {meaning: "no scenario-scoped services"},
	"Scenario.MockServers":  {meaning: "no scenario-scoped mock servers"},
	"Scenario.Teardown":     {meaning: "no teardown steps"},
	"Run.Env":               {meaning: "no per-step env overrides"},
	"Run.PassEnv":           {meaning: "no host variable survives clear_env, which is what clear_env alone already means"},
	"PTY.Env":               {meaning: "no per-step env overrides"},
	"PTY.PassEnv":           {meaning: "no host variable survives clear_env"},
	"PTY.Session":           {meaning: "a pty step with no expect/send session: run the command in a terminal and assert on the transcript"},
	"PTYMouse.Mods":         {meaning: "an unmodified mouse event"},
	"Service.Env":           {meaning: "no service env overrides"},
	"Service.PassEnv":       {meaning: "no host variable survives clear_env"},
	"MockServer.Routes":     {meaning: "a mock server with no routes; its own validation decides whether that is useful"},
	"MockRoute.Header":      {meaning: "no response headers"},
	"HTTP.Header":           {meaning: "no request headers"},
	"HTTP.Form":             {meaning: "no form fields"},
	"HTTP.Files":            {meaning: "no uploaded file parts"},
	"GRPC.Header":           {meaning: "no request metadata"},
	"Runner.BrowserArgs":    {meaning: "no extra Chrome launch flags"},
	"NetworkPolicy.Allow":   {meaning: "allow no host, which is how a spec declares that scenarios must reach nothing"},
	"StreamAssert.JSON":     {meaning: "no json checks; the other matchers of the same assert carry it"},
	"StreamAssert.YAML":     {meaning: "no yaml checks"},
	"FileAssert.JSON":       {meaning: "no json checks"},
	"DirAssert.Contains":    {meaning: "no required children; the other dir constraints carry the assert"},
	"DirAssert.NotContains": {meaning: "no forbidden children"},
	"DirAssert.Ignore":      {meaning: "ignore nothing"},
	"ChangesAssert.Ignore":  {meaning: "ignore nothing"},
	"PDFAssert.Metadata":    {meaning: "no metadata fields asserted"},
	"ScreenAssert.Attrs":    {meaning: "no styling claims; the text matchers carry the assert"},
}

// collectionKind reports whether a field is collection-valued for this purpose:
// a slice or a map, whose emptiness a spec author can write explicitly.
func collectionKind(t reflect.Type) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Kind() == reflect.Slice || t.Kind() == reflect.Map
}

// TestSpec_EveryCollectionFieldHasAnEmptyPolicy walks every type reachable from
// spec.Spec and requires a stated policy for each collection-valued field: a
// spec the loader must refuse, or a written meaning for the accepted empty
// value. It is the structural half of the fix for silently-discarded empty
// collections — the refusals live in the validators, and this is what stops the
// next list or map from shipping without the question being asked.
func TestSpec_EveryCollectionFieldHasAnEmptyPolicy(t *testing.T) {
	t.Parallel()

	seen := map[reflect.Type]bool{}
	used := map[string]bool{}
	var walk func(reflect.Type)
	walk = func(rt reflect.Type) {
		for rt.Kind() == reflect.Pointer || rt.Kind() == reflect.Slice || rt.Kind() == reflect.Map {
			rt = rt.Elem()
		}
		if rt.Kind() != reflect.Struct || seen[rt] {
			return
		}
		seen[rt] = true
		for i := range rt.NumField() {
			f := rt.Field(i)
			tag := f.Tag.Get("yaml")
			// Unexported fields and fields the spec never decodes ("-", e.g. the
			// matrix row binding the loader fills in) are not part of the surface
			// an author can write empty.
			if f.PkgPath != "" || tag == "" || strings.HasPrefix(tag, "-") {
				continue
			}
			if collectionKind(f.Type) {
				key := rt.Name() + "." + f.Name
				used[key] = true
				policy, ok := emptyPolicies[key]
				switch {
				case !ok:
					t.Errorf("%s is collection-valued and has no policy for an explicitly empty value; "+
						"decide whether writing it empty is refused at load or means something, and record it in emptyPolicies", key)
				case policy.refuse == "" && policy.meaning == "":
					t.Errorf("%s is accepted empty without a stated meaning", key)
				}
			}
			walk(f.Type)
		}
	}
	walk(reflect.TypeOf(spec.Spec{}))

	for key := range emptyPolicies {
		if !used[key] {
			t.Errorf("emptyPolicies has an entry for %s, which is no longer a collection-valued spec field", key)
		}
	}
}

// TestSpec_RefusedEmptyCollectionsAreActuallyRefused loads each refusal spec the
// policy table declares, so the table cannot claim a refusal the loader does not
// perform. The message must name the field, since an author reading it has to be
// told which empty value was the problem.
func TestSpec_RefusedEmptyCollectionsAreActuallyRefused(t *testing.T) {
	t.Parallel()
	for key, policy := range emptyPolicies {
		if policy.refuse == "" {
			continue
		}
		t.Run(key, func(t *testing.T) {
			t.Parallel()
			_, err := LoadBytes("t.atago.yaml", []byte(policy.refuse))
			if err == nil {
				t.Fatalf("%s: LoadBytes() error = nil, want the empty collection refused", key)
			}
			field := key[strings.Index(key, ".")+1:]
			// The YAML key, derived from the Go field name: Compare → compare,
			// NotContains → not_contains.
			var yamlKey strings.Builder
			for i, r := range field {
				if r >= 'A' && r <= 'Z' {
					if i > 0 {
						yamlKey.WriteByte('_')
					}
					r += 'a' - 'A'
				}
				yamlKey.WriteRune(r)
			}
			if !strings.Contains(err.Error(), yamlKey.String()) {
				t.Errorf("%s: error = %q, want it to name the %q key", key, err, yamlKey.String())
			}
		})
	}
}
