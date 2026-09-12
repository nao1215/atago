package store

import "slices"

// The built-in variables the engine seeds into every scenario store. They are
// named here, as constants, because two packages depend on the same list and
// had written it out separately: the engine seeds them, and the loader rejects a
// `store:` name or a `matrix:` key that would shadow one. The loader's copy
// listed three of the five, so `store: {name: specdir}` was accepted and
// silently redefined ${specdir} for the rest of the scenario — a later step
// reading a committed file through it read somewhere else — while the same
// mistake on ${workdir} was refused with advice to pick another name.
const (
	// BuiltinAtago is the absolute path of the running atago binary, which lets
	// a self-hosted spec invoke atago from inside its isolated temp workdir.
	BuiltinAtago = "atago"
	// BuiltinWorkdir is the scenario's isolated temp directory.
	BuiltinWorkdir = "workdir"
	// BuiltinSpecdir is the absolute directory the spec file lives in.
	BuiltinSpecdir = "specdir"
	// BuiltinFixtures is the committed fixture tree a directory manifest points
	// at; unset when no manifest applied.
	BuiltinFixtures = "fixtures"
	// BuiltinSuitedir is the suite-wide scratch directory shared by every
	// scenario in one spec.
	BuiltinSuitedir = "suitedir"
)

// Builtins is every name above, sorted, so a caller can both test membership and
// name them all in one message.
var Builtins = []string{
	BuiltinAtago,
	BuiltinFixtures,
	BuiltinSpecdir,
	BuiltinSuitedir,
	BuiltinWorkdir,
}

// IsBuiltin reports whether name is one the engine seeds, and so one a spec must
// not bind itself.
func IsBuiltin(name string) bool { return slices.Contains(Builtins, name) }
