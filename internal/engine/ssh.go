package engine

import (
	"time"

	sshrunner "github.com/nao1215/atago/internal/runner/ssh"
	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

// sshConn returns the scenario's connection for a named ssh runner, dialing it on
// first use and caching it for the rest of the scenario. The connection fields
// are ${name}-expanded so host/user/password/key can reference stored values or
// built-ins (e.g. a key written under ${workdir}).
func sshConn(name string, st *store.Store, rc runConfig, conns map[string]*sshrunner.Runner) (*sshrunner.Runner, error) {
	return resolveConn(name, "run step", "ssh", rc, conns, func(rdef spec.Runner, timeout time.Duration) (*sshrunner.Runner, error) {
		cfg := sshrunner.Config{
			Addr:            st.Expand(rdef.Host),
			User:            st.Expand(rdef.User),
			Password:        st.Expand(rdef.Password),
			KeyFile:         st.Expand(rdef.KeyFile),
			KnownHosts:      st.Expand(rdef.KnownHosts),
			InsecureHostKey: rdef.InsecureHostKey,
			Timeout:         timeout,
		}
		// Enforce the network allowlist before dialing (issue #17): ssh egress is
		// confined to permissions.network.allow just like HTTP.
		if err := security.CheckHost(rc.allow, cfg.Addr); err != nil {
			return nil, err
		}
		return sshrunner.Open(cfg)
	})
}
