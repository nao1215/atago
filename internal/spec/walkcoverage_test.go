package spec

import (
	"strings"
	"testing"
)

// stepSpecimens is one fully-populated Step per kind, each carrying a live
// `${probe}` reference and a `${env:PROBE_ENV}` reference in a field the
// engine actually ${name}-expands, plus a declared output where the kind can
// declare one. The walker coverage tests below iterate AllStepKinds against
// this registry, so the chain is closed: adding a step kind fails
// TestStepActions_CoverEveryStepField until the table knows it, then fails
// TestStepSpecimens_CoverEveryKind until a specimen exists, then fails each
// walker's Decides test until the author either wires the walker or records a
// reasoned exception — the decision the walk-forgetting bugs all skipped.
var stepSpecimens = map[StepKind]*Step{
	StepFixture: {Fixture: &Fixture{File: "seed-${probe}.txt", Content: "token=${env:PROBE_ENV}"}},
	StepRun:     {Run: &Run{Command: "tool --who ${probe} --token ${env:PROBE_ENV}", StdoutTo: "run-out.txt"}},
	StepHTTP:    {HTTP: &HTTP{Runner: "api", Method: "GET", Path: "/v1/${probe}?t=${env:PROBE_ENV}", BodyTo: "resp.bin"}},
	StepQuery:   {Query: &Query{Runner: "db", SQL: "SELECT ${probe}, '${env:PROBE_ENV}'"}},
	StepGRPC:    {GRPC: &GRPC{Runner: "rpc", Method: "pkg.S/${probe}", Header: map[string]string{"authorization": "Bearer ${env:PROBE_ENV}"}}},
	StepCDP: {CDP: &CDP{Runner: "web", Actions: []CDPAction{
		{Navigate: "https://host/${probe}?t=${env:PROBE_ENV}"},
		{Screenshot: &CDPScreenshot{Path: "shot.png"}},
	}}},
	StepAssert: {Assert: &Assert{
		Stdout: &StreamAssert{Contains: StringList{"${probe}", "${env:PROBE_ENV}"}},
		File:   &FileAssert{Path: "made.txt", Exists: boolp(true)},
	}},
	StepStore:   {Store: &Store{Name: "v", From: &StoreFrom{File: &FileAssert{Path: "${probe}-${env:PROBE_ENV}.json", Text: boolp(true)}}}},
	StepService: {Service: &Service{Name: "svc", Command: "serve ${probe} ${env:PROBE_ENV}"}},
	StepPTY:     {PTY: &PTY{Command: "repl ${probe} ${env:PROBE_ENV}"}},
	StepSignal:  {Signal: &Signal{Service: "${probe}-${env:PROBE_ENV}", Signal: "TERM"}},
	// A mock server is started verbatim: the engine ${name}-expands none of its
	// fields, so its specimen carries no live reference on purpose — the Decides
	// tests document that as the reason it is exempt everywhere.
	StepMockServer: {MockServer: &MockServer{Name: "stub"}},
}

// specimenRunners backs the runner-carrying specimens, with a remote db so the
// query kind exercises the network note.
var specimenRunners = map[string]Runner{
	"api": {Type: "http", BaseURL: "http://127.0.0.1:1"},
	"db":  {Type: "db", Driver: "postgres", DSN: "postgres://127.0.0.1:1/x"},
	"rpc": {Type: "grpc", Target: "127.0.0.1:1"},
	"web": {Type: "browser"},
}

// TestStepSpecimens_CoverEveryKind pins the registry to the kind list from both
// sides, so the Decides tests below genuinely iterate every kind that exists.
func TestStepSpecimens_CoverEveryKind(t *testing.T) {
	t.Parallel()
	for _, kind := range AllStepKinds() {
		st, ok := stepSpecimens[kind]
		if !ok {
			t.Errorf("step kind %q has no specimen; add one so the walker coverage tests exercise it", kind)
			continue
		}
		if got := st.Kind(); got != kind {
			t.Errorf("specimen for %q reports Kind() = %q", kind, got)
		}
	}
	known := map[StepKind]bool{}
	for _, kind := range AllStepKinds() {
		known[kind] = true
	}
	for kind := range stepSpecimens {
		if !known[kind] {
			t.Errorf("specimen exists for %q, which AllStepKinds does not list", kind)
		}
	}
}

// TestCollectStepVars_DecidesEveryStepKind forces a decision per kind: either
// CollectStepVars sees the specimen's ${probe} reference, or the kind is
// exempted here with the reason. A new kind with a specimen but no entry fails,
// which is the forcing function — the historical bugs were kinds nobody decided
// about (pty's paste and exec went uncounted, the whole assert kind too).
func TestCollectStepVars_DecidesEveryStepKind(t *testing.T) {
	t.Parallel()
	// Exempt kinds carry the reason collection would be wrong, not just absent.
	exempt := map[StepKind]string{
		StepMockServer: "the engine starts the stub verbatim; no field is ${name}-expanded",
	}
	for _, kind := range AllStepKinds() {
		set := map[string]bool{}
		CollectStepVars(set, stepSpecimens[kind], specimenRunners)
		if reason, ok := exempt[kind]; ok {
			if len(set) != 0 {
				t.Errorf("kind %q is exempt (%s) but collected %v; drop the exemption", kind, reason, set)
			}
			continue
		}
		if !set["probe"] {
			t.Errorf("kind %q: CollectStepVars missed the specimen's ${probe} reference (got %v); wire the kind or exempt it with a reason", kind, set)
		}
	}
}

// TestSecurityNotes_DecidesEveryStepKind forces the same decision for the
// security summary. Every specimen reads ${env:PROBE_ENV} through a field the
// engine expands, so every non-exempt kind must yield at least the
// host-environment-read note — fixture, assert, store, and signal were exactly
// the kinds nobody had decided about, and their env reads went unreported.
func TestSecurityNotes_DecidesEveryStepKind(t *testing.T) {
	t.Parallel()
	exempt := map[StepKind]string{
		StepMockServer: "a local stub: nothing is expanded, no process runs a user command, no egress",
	}
	for _, kind := range AllStepKinds() {
		sc := &Scenario{Name: "probe", Steps: []Step{*stepSpecimens[kind]}}
		notes := SecurityNotes(sc, specimenRunners)
		if reason, ok := exempt[kind]; ok {
			if len(notes) != 0 {
				t.Errorf("kind %q is exempt (%s) but produced %v; drop the exemption", kind, reason, notes)
			}
			continue
		}
		found := false
		for _, note := range notes {
			if strings.Contains(note, "${env:PROBE_ENV}") {
				found = true
			}
		}
		if !found {
			t.Errorf("kind %q: SecurityNotes missed the specimen's ${env:PROBE_ENV} read (got %v); wire the kind or exempt it with a reason", kind, notes)
		}
	}
}

// TestGeneratedArtifacts_DecidesEveryStepKind forces the decision for the
// declared-outputs list: each kind either contributes the specimen's declared
// path or is recorded here as declaring none — the pdf assertion was the
// output-declaring form this list never learned (#73-era drift).
func TestGeneratedArtifacts_DecidesEveryStepKind(t *testing.T) {
	t.Parallel()
	wantPath := map[StepKind]string{
		StepRun:    "run-out.txt",
		StepHTTP:   "resp.bin",
		StepAssert: "made.txt",
		StepCDP:    "shot.png",
		// Every other kind declares no output file: a fixture writes an INPUT,
		// and query/grpc/store/pty/signal/service/mock_server carry no
		// file-producing field.
		StepFixture:    "",
		StepQuery:      "",
		StepGRPC:       "",
		StepStore:      "",
		StepService:    "",
		StepPTY:        "",
		StepSignal:     "",
		StepMockServer: "",
	}
	for _, kind := range AllStepKinds() {
		want, ok := wantPath[kind]
		if !ok {
			t.Errorf("step kind %q: decide whether GeneratedArtifacts must report it, and record the decision here", kind)
			continue
		}
		sc := &Scenario{Name: "probe", Steps: []Step{*stepSpecimens[kind]}}
		got := GeneratedArtifacts(sc)
		switch {
		case want == "" && len(got) != 0:
			t.Errorf("kind %q: expected no declared outputs, got %v", kind, got)
		case want != "" && !contains(got, want):
			t.Errorf("kind %q: GeneratedArtifacts = %v, want it to include %q", kind, got, want)
		}
	}
}

// TestDescribers_SeeEveryPhase pins that the shared describers read steps
// through the phase walkers: the same specimen yields its note and its artifact
// from every phase steps can live in, scenario and suite alike.
func TestDescribers_SeeEveryPhase(t *testing.T) {
	t.Parallel()
	run := *stepSpecimens[StepRun]

	sc := &Scenario{Name: "probe", Teardown: []Step{run}}
	if notes := SecurityNotes(sc, specimenRunners); len(notes) == 0 {
		t.Error("SecurityNotes ignored a scenario teardown step")
	}
	if got := GeneratedArtifacts(sc); !contains(got, "run-out.txt") {
		t.Errorf("GeneratedArtifacts ignored a scenario teardown step: %v", got)
	}

	s := &Spec{Suite: Suite{Setup: []Step{run}}}
	if notes := SuiteSecurityNotes(s); len(notes) == 0 {
		t.Error("SuiteSecurityNotes ignored a suite.setup step")
	}
	s = &Spec{Suite: Suite{Teardown: []Step{run}}}
	if notes := SuiteSecurityNotes(s); len(notes) == 0 {
		t.Error("SuiteSecurityNotes ignored a suite.teardown step")
	}
	if got := SuiteGeneratedArtifacts(&Suite{Teardown: []Step{run}}); !contains(got, "run-out.txt") {
		t.Errorf("SuiteGeneratedArtifacts ignored a suite.teardown step: %v", got)
	}
}
