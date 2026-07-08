package loader

import (
	"maps"
	"slices"
	"strings"

	"github.com/nao1215/atago/internal/spec"
)

// validateSignal checks a signal step (#23): a declared target service, an
// accepted signal name, and a parseable wait timeout. A ${name}-referencing
// target is resolved at run time and skips the declared-name check.
func validateSignal(add func(string, ...any), where string, sg *spec.Signal, serviceNames map[string]bool) {
	switch {
	case sg.Service == "":
		add("%s.signal.service is required (the scenario or suite service to signal)", where)
	case !strings.Contains(sg.Service, "${") && !serviceNames[sg.Service]:
		declared := "none"
		if len(serviceNames) > 0 {
			declared = strings.Join(slices.Sorted(maps.Keys(serviceNames)), ", ")
		}
		add("%s.signal.service %q is not a declared service (declared: %s)", where, sg.Service, declared)
	}
	switch {
	case sg.Signal == "":
		add("%s.signal.signal is required (TERM, INT, HUP, USR1, USR2, or KILL)", where)
	case !spec.ValidSignalName(sg.Signal):
		add("%s.signal.signal %q is not an accepted signal (TERM, INT, HUP, USR1, USR2, or KILL, with an optional SIG prefix)", where, sg.Signal)
	}
	if sg.Wait != nil {
		positiveDuration(add, where+".signal.wait.timeout", sg.Wait.Timeout, "5s", "5s")
	}
}
