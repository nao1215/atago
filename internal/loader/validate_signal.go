package loader

import (
	"maps"
	"slices"
	"strings"
	"time"

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
	if sg.Wait != nil && sg.Wait.Timeout != "" {
		d, err := time.ParseDuration(sg.Wait.Timeout)
		switch {
		case err != nil:
			add("%s.signal.wait.timeout %q is not a valid duration (e.g. \"5s\")", where, sg.Wait.Timeout)
		case d <= 0:
			add("%s.signal.wait.timeout must be positive (got %q); omit it for the 5s default", where, sg.Wait.Timeout)
		}
	}
}
