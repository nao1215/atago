package spec

import (
	"reflect"
	"strings"
	"testing"
)

// assertTargetFields maps every yaml key on Assert to the field that carries it.
// It is how the coverage tests build a minimal Assert for one target and how they
// prove the target list matches the struct.
func assertTargetFields(t *testing.T) map[string]reflect.StructField {
	t.Helper()
	out := map[string]reflect.StructField{}
	rt := reflect.TypeOf(Assert{})
	for i := range rt.NumField() {
		f := rt.Field(i)
		tag := strings.Split(f.Tag.Get("yaml"), ",")[0]
		if tag == "" || tag == "-" {
			continue
		}
		out[tag] = f
	}
	return out
}

// assertForTarget returns an Assert with exactly the given target's field
// allocated to a zero value. internal/spectest carries the same helper for the
// packages that dispatch on a target; this copy keeps the spec package's own
// tests free of an import cycle.
func assertForTarget(target AssertTarget) *Assert {
	a := &Assert{}
	rv := reflect.ValueOf(a).Elem()
	rt := rv.Type()
	for i := range rt.NumField() {
		if strings.Split(rt.Field(i).Tag.Get("yaml"), ",")[0] != string(target) {
			continue
		}
		f := rv.Field(i)
		if f.Kind() == reflect.Pointer {
			f.Set(reflect.New(f.Type().Elem()))
		}
		return a
	}
	return a
}

// TestAssertTargets_CoverEveryAssertField is the drift guard for the whole
// assert-target structure. Five layers switch on a target — loader validation,
// the runtime check, doc, explain, and the JSON schema — and each of those has a
// coverage test that walks AllAssertTargets. That only protects anything if the
// list itself is complete, so this test derives the truth from the Assert struct:
// a new target field with no entry in assertTargets fails here first.
func TestAssertTargets_CoverEveryAssertField(t *testing.T) {
	t.Parallel()
	fields := assertTargetFields(t)
	listed := map[AssertTarget]bool{}
	for _, target := range AllAssertTargets() {
		if listed[target] {
			t.Errorf("target %q is listed twice", target)
		}
		listed[target] = true
		if _, ok := fields[string(target)]; !ok {
			t.Errorf("target %q has no field on Assert carrying that yaml key", target)
		}
	}
	for key := range fields {
		if !listed[AssertTarget(key)] {
			t.Errorf("Assert field %q is a target family with no entry in assertTargets; SetTargets will never report it", key)
		}
	}
}

// TestSetTargets_ReportsEachTargetAlone pins the pairing between a target and
// its field: an Assert built for one target reports exactly that target, so the
// table's presence checks cannot be wired to the wrong field.
func TestSetTargets_ReportsEachTargetAlone(t *testing.T) {
	t.Parallel()
	for _, target := range AllAssertTargets() {
		got := assertForTarget(target).SetTargets()
		if len(got) != 1 || got[0] != target {
			t.Errorf("SetTargets for a %q assert = %v, want exactly [%s]", target, got, target)
		}
	}
	if got := (&Assert{}).SetTargets(); len(got) != 0 {
		t.Errorf("SetTargets on an empty assert = %v, want none", got)
	}
}
