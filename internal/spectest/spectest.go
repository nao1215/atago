// Package spectest builds minimal spec values for tests in other packages. Five
// layers dispatch on spec.AssertTarget — the loader's validation, the runtime
// check, doc, explain, and the JSON schema — and each needs a coverage test that
// walks spec.AllAssertTargets and exercises its own switch. That needs an Assert
// carrying exactly one target, built without a per-target constructor in every
// test file.
package spectest

import (
	"reflect"
	"strings"

	"github.com/nao1215/atago/internal/spec"
)

// AssertForTarget returns an Assert with exactly the given target's field
// allocated to a zero value. The field is located by its yaml key, which is the
// target's name, so a new target needs no change here.
func AssertForTarget(target spec.AssertTarget) *spec.Assert {
	a := &spec.Assert{}
	rv := reflect.ValueOf(a).Elem()
	rt := rv.Type()
	for i := range rt.NumField() {
		if strings.Split(rt.Field(i).Tag.Get("yaml"), ",")[0] != string(target) {
			continue
		}
		if f := rv.Field(i); f.Kind() == reflect.Pointer {
			f.Set(reflect.New(f.Type().Elem()))
		}
		return a
	}
	return a
}
