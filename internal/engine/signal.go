package engine

import (
	"fmt"
	"strings"
	"time"

	servicerunner "github.com/nao1215/atago/internal/runner/service"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

// defaultSignalWait bounds `signal.wait` when no timeout is authored (#23).
const defaultSignalWait = 5 * time.Second

// runSignal executes a signal step (#23): resolve the target service by name
// (scenario services first, then suite services), deliver the named POSIX
// signal to its process group, and — when wait is set — block until the
// process exits or the timeout elapses. The service name is ${name}-expanded
// so matrix instances can target parameterized services; the signal name is
// normalized (an optional SIG prefix, any case) to match the loader's rule.
func runSignal(sg *spec.Signal, st *store.Store, scenarioServices, suiteServices []*servicerunner.Proc) error {
	name := st.Expand(sg.Service)
	proc := findServiceProc(name, scenarioServices, suiteServices)
	if proc == nil {
		return fmt.Errorf("signal step targets unknown service %q (no scenario or suite service with that name is running)", name)
	}
	sigName := spec.NormalizeSignalName(sg.Signal)
	if err := proc.Signal(sigName); err != nil {
		return fmt.Errorf("signal step: %w", err)
	}
	if sg.Wait == nil {
		return nil
	}
	timeout := defaultSignalWait
	if sg.Wait.Timeout != "" {
		d, err := time.ParseDuration(sg.Wait.Timeout) // validated at load time
		if err == nil {
			timeout = d
		}
	}
	if !proc.WaitExit(timeout) {
		return fmt.Errorf("service %q did not exit within %s after SIG%s", name, timeout, sigName)
	}
	return nil
}

// findServiceProc resolves a service name against the scenario's own services
// first, then the suite-wide ones, mirroring the store's scenario-over-suite
// precedence.
func findServiceProc(name string, scenarioServices, suiteServices []*servicerunner.Proc) *servicerunner.Proc {
	for _, p := range scenarioServices {
		if p.Name() == name {
			return p
		}
	}
	for _, p := range suiteServices {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

// signalNames is the accepted signal set (#23), shared with the loader's
// validation message.
var signalNames = []string{"TERM", "INT", "HUP", "USR1", "USR2", "KILL"}

// SignalNameList renders the accepted signal names for error messages.
func SignalNameList() string { return strings.Join(signalNames, ", ") }
