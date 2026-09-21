package loader

import (
	"fmt"
	"sort"
	"strings"

	"github.com/nao1215/atago/internal/diag"
	"github.com/nao1215/atago/internal/generator"
	"github.com/nao1215/atago/internal/spec"
)

// defaultForallRuns is how many instances a `forall:` block expands to when it
// states no `runs:`. Every instance is a scenario that starts a process, so the
// default is sized for "try a handful of shapes of this input" rather than for
// a fuzzing campaign; the boundary values come first, so the first few
// instances are the ones most likely to find something.
const defaultForallRuns = 10

// maxForallRuns and maxForallLength bound what one block may ask for. Both are
// authoring guards rather than technical limits: a spec that expands to
// thousands of scenarios, or that builds a megabyte-long argument, has stopped
// being a spec someone reads.
const (
	maxForallRuns   = 1000
	maxForallLength = 4096
)

// validateForall checks every scenario's forall block before it is turned into
// matrix rows. It runs on the raw decoded spec, so the block is still there to
// report against: once expandForall has run, a scenario carries generated rows
// and nothing downstream could name the generator that produced them.
func validateForall(s *spec.Spec) []string {
	var errs errorList
	add := errs.add
	for i := range s.Scenarios {
		sc := &s.Scenarios[i]
		if sc.Forall == nil {
			continue
		}
		where := fmt.Sprintf("scenarios[%d]", i)
		if sc.Name != "" {
			where = fmt.Sprintf("scenario %q", sc.Name)
		}
		if sc.Matrix != nil {
			add(diag.ExclusiveKeys, "%s sets both matrix and forall; a scenario states its inputs or generates them, not both", where)
		}
		validateForallRuns(add, where, sc.Forall)
		if len(sc.Forall.Vars) == 0 {
			add(diag.EmptyValue, "%s.forall.vars must declare at least one variable", where)
			continue
		}
		for _, name := range sortedVarNames(sc.Forall.Vars) {
			g := sc.Forall.Vars[name]
			validateGenerator(add, fmt.Sprintf("%s.forall.vars.%s", where, name), name, &g)
			validateExamples(add, fmt.Sprintf("%s.forall.vars.%s", where, name), &g, forallRuns(sc.Forall))
		}
		validateForallName(add, where, sc)
	}
	return errs.msgs
}

// validateForallRuns bounds the instance count. A negative or zero `runs:` is
// not "generate nothing" — a scenario that expands to no instances asserts
// nothing while still reading like a test.
func validateForallRuns(add addFunc, where string, f *spec.Forall) {
	switch {
	case f.Runs < 0:
		add(diag.NegativeValue, "%s.forall.runs must not be negative", where)
	case f.Runs > maxForallRuns:
		add(diag.OutOfRange, "%s.forall.runs is %d; at most %d (each run is a scenario that starts a process)", where, f.Runs, maxForallRuns)
	}
}

// validateGenerator checks one variable's generator.
func validateGenerator(add addFunc, where, name string, g *spec.Generator) {
	if reservedVarName(name) {
		add(diag.ReservedName, "%s shadows a built-in variable (%s); choose another name", where, builtinList())
	}
	if g.OneOf != nil {
		validateOneOf(add, where, g)
		return
	}
	if g.Type == "" {
		add(diag.RequiredKey, "%s must name a generator type (%s) or list values with one_of", where, strings.Join(generator.Kinds(), ", "))
		return
	}
	if !generator.Known(g.Type) {
		add(diag.NotAllowedValue, "%s type %q is not a generator; use one of %s, or list values with one_of", where, g.Type, strings.Join(generator.Kinds(), ", "))
		return
	}
	validateGeneratorRange(add, where, g)
}

// validateExamples checks the must-try values a generator names (#661). They
// are instances of their own, so more of them than `runs` would either drop the
// last ones or quietly run more scenarios than the block says — both of which
// are worse than saying so.
func validateExamples(add addFunc, where string, g *spec.Generator, runs int) {
	if g.Examples == nil {
		return
	}
	switch {
	case len(g.Examples) == 0:
		add(diag.EmptyList, "%s examples must list at least one value", where)
		return
	case len(g.OneOf) > 0:
		add(diag.ExclusiveKeys, "%s sets examples together with one_of; a choice generator already tries every value it lists", where)
		return
	case len(g.Examples) > runs:
		add(diag.OutOfRange, "%s names %d examples but forall.runs is %d, so the last of them would never run; raise runs to at least %d",
			where, len(g.Examples), runs, len(g.Examples))
	}
	seen := map[string]bool{}
	for _, v := range g.Examples {
		if seen[v] {
			add(diag.DuplicateEntry, "%s lists the example %q twice; rows are deduplicated, so the second one can only be a typo", where, v)
		}
		seen[v] = true
	}
}

// validateOneOf checks a choice generator: it carries values and nothing else,
// because a range over a list of choices has no meaning to narrow.
func validateOneOf(add addFunc, where string, g *spec.Generator) {
	if len(g.OneOf) == 0 {
		add(diag.EmptyList, "%s one_of must list at least one value", where)
		return
	}
	if g.Type != "" || g.Min != nil || g.Max != nil {
		add(diag.ExclusiveKeys, "%s sets one_of together with type/min/max; a choice generator is the values it lists", where)
	}
	seen := map[string]bool{}
	for _, v := range g.OneOf {
		if seen[v] {
			add(diag.DuplicateEntry, "%s lists %q twice; the duplicate cannot produce a distinct instance", where, v)
		}
		seen[v] = true
	}
}

// validateGeneratorRange checks the bounds of a kind that reads them, and
// refuses them on the kinds that do not: `{type: bool, max: 3}` is a mistake
// that would otherwise be accepted and ignored.
func validateGeneratorRange(add addFunc, where string, g *spec.Generator) {
	kind := generator.Kind(g.Type)
	if !generator.Bounded(kind) {
		if g.Min != nil || g.Max != nil {
			add(diag.KeyNotHere, "%s sets min/max on a %s generator, which has no range to narrow", where, g.Type)
		}
		return
	}
	low, high := generator.Defaults(kind)
	if g.Min != nil {
		low = *g.Min
	}
	if g.Max != nil {
		high = *g.Max
	}
	if kind != generator.Int {
		if low < 0 {
			add(diag.NegativeValue, "%s min is %d; a %s generator's bounds are a string LENGTH", where, low, g.Type)
			return
		}
		if high > maxForallLength {
			add(diag.OutOfRange, "%s max is %d; at most %d characters", where, high, maxForallLength)
			return
		}
	}
	if low > high {
		add(diag.EmptyInterval, "%s has min %d greater than max %d, so it can produce no value", where, low, high)
	}
}

// validateForallName keeps generated instances distinguishable. A scenario name
// that references a generated variable is substituted per instance, exactly as
// a matrix name is; referencing SOME of the variables is what collapses two
// instances that differ in the ones it left out onto one name, and the duplicate
// that reports is a puzzle when nobody wrote the values.
func validateForallName(add addFunc, where string, sc *spec.Scenario) {
	refs := spec.VarRefs(sc.Name)
	if len(refs) == 0 {
		return // no reference: every instance gets the [k=v ...] suffix.
	}
	var missing []string
	for _, ref := range refs {
		if _, declared := sc.Forall.Vars[ref]; !declared {
			add(diag.NameNotDeclared,
				"%s references ${%s} in its name, which forall does not generate, so every instance carries the literal text ${%s}; declare it under forall.vars or drop it from the name",
				where, ref, ref)
		}
	}
	referenced := map[string]bool{}
	for _, ref := range refs {
		referenced[ref] = true
	}
	for _, name := range sortedVarNames(sc.Forall.Vars) {
		if !referenced[name] {
			missing = append(missing, "${"+name+"}")
		}
	}
	if len(missing) > 0 {
		add(diag.DuplicateName,
			"%s references some generated variables in its name but not %s, so two instances that differ only there collapse to the same name; reference every forall variable in the name, or none of them (the values are appended automatically)",
			where, strings.Join(missing, ", "))
	}
}

// expandForall turns each forall block into the matrix rows it describes, then
// clears it. Everything after this point — matrix validation and expansion, the
// engine, the reports, explain, doc, list, the manifest — sees a scenario whose
// inputs were authored, because by then they were: generation happens once, at
// load time, from a seed the spec carries.
func expandForall(s *spec.Spec) {
	for i := range s.Scenarios {
		sc := &s.Scenarios[i]
		if sc.Forall == nil {
			continue
		}
		runs := forallRuns(sc.Forall)
		// The suite and scenario names are part of the seed so that two
		// scenarios in one file do not test the same generated inputs twice.
		seed := generator.Seed(sc.Forall.Seed, s.Suite.Name, sc.Name)
		rows := generator.Rows(seed, runs, forallVars(sc.Forall))
		sc.Forall = nil
		if len(rows) == 0 {
			continue
		}
		sc.Matrix = rows
	}
}

// forallRuns is how many instances a block expands to: what it asked for, or
// the default. Validation and expansion both read it, so the number an error
// message quotes is the number the loader would have used.
func forallRuns(f *spec.Forall) int {
	if f.Runs == 0 {
		return defaultForallRuns
	}
	return f.Runs
}

// forallVars renders a forall block's declarations as generator inputs, sorted
// by variable name. The order is what the values depend on, so it has to come
// from the names rather than from Go's map iteration: the same spec must
// generate the same inputs on every run.
func forallVars(f *spec.Forall) []generator.Var {
	names := sortedVarNames(f.Vars)
	vars := make([]generator.Var, 0, len(names))
	for _, name := range names {
		g := f.Vars[name]
		v := generator.Var{Name: name, Kind: generator.Kind(g.Type), Values: g.OneOf, Examples: g.Examples}
		if len(g.OneOf) > 0 {
			v.Kind = generator.OneOf
		}
		v.Min, v.Max = generator.Defaults(v.Kind)
		if g.Min != nil {
			v.Min = *g.Min
		}
		if g.Max != nil {
			v.Max = *g.Max
		}
		vars = append(vars, v)
	}
	return vars
}

// sortedVarNames returns a forall block's variable names in a fixed order, so
// both validation messages and generation are reproducible.
func sortedVarNames(vars map[string]spec.Generator) []string {
	names := make([]string, 0, len(vars))
	for name := range vars {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
