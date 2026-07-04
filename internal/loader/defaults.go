package loader

import (
	"maps"

	"github.com/nao1215/atago/internal/spec"
)

// applyDefaults expands the top-level `defaults:` block into the concrete
// scenario model. It runs after matrix expansion and before
// validation, so the merged result is validated and the engine only ever sees
// fully-resolved scenarios — `defaults` is pure authoring sugar with no runtime
// model of its own.
//
// Merge rules: an explicitly-authored value always wins; maps shallow-merge (the
// authored key wins per key); a nil pointer / empty string counts as "unset" and
// takes the default — so an authored `shell: false` beats a defaulted
// `shell: true` (Shell is a *bool precisely to keep unset and false distinct).
func applyDefaults(s *spec.Spec) {
	d := s.Defaults
	if d == nil {
		return
	}
	for i := range s.Scenarios {
		sc := &s.Scenarios[i]
		if d.Scenario != nil {
			sc.Env = mergeStringMap(d.Scenario.Env, sc.Env)
		}
		if d.Service != nil {
			for j := range sc.Services {
				mergeServiceDefaults(d.Service, &sc.Services[j])
			}
		}
		if d.Run != nil {
			for j := range sc.Steps {
				if sc.Steps[j].Run != nil {
					mergeRunDefaults(d.Run, sc.Steps[j].Run)
				}
			}
			// Teardown steps are steps too: shared run defaults (shell, env, ...)
			// apply so cleanup does not need to re-declare them.
			for j := range sc.Teardown {
				if sc.Teardown[j].Run != nil {
					mergeRunDefaults(d.Run, sc.Teardown[j].Run)
				}
			}
		}
	}
}

// mergeStringMap returns a shallow merge of def under own: every default key is
// present unless own overrides it. It returns own unchanged when there is nothing
// to add, and never mutates either input.
func mergeStringMap(def, own map[string]string) map[string]string {
	if len(def) == 0 {
		return own
	}
	out := make(map[string]string, len(def)+len(own))
	maps.Copy(out, def)
	maps.Copy(out, own)
	return out
}

// mergeRunDefaults layers def beneath an authored run step. Command and Retry are
// intentionally not merged (they are per-step; the validator rejects them on
// defaults.run so a stray value is never silently ignored).
func mergeRunDefaults(def, r *spec.Run) {
	if r.Runner == "" {
		r.Runner = def.Runner
	}
	if r.Shell == nil {
		r.Shell = def.Shell
	}
	if r.Cwd == "" {
		r.Cwd = def.Cwd
	}
	// Timeout is deliberately NOT merged here: the engine resolves the full
	// precedence chain (step > runner > defaults.run > suite > built-in, #17)
	// itself, and merging at load time would erase which level supplied the
	// value — the timeout-kill hint names that level, and a runner-common
	// timeout must beat defaults.run.timeout, which a string-fill here would
	// invert.
	// Stdin is deliberately NOT merged: like command it is per-step input
	// data, and the validator rejects it on defaults.run (#18).
	r.Env = mergeStringMap(def.Env, r.Env)
	if r.ClearEnv == nil {
		r.ClearEnv = def.ClearEnv
	}
	// pass_env is meaningless without clear_env, so a step that authored
	// `clear_env: false` does not inherit the default allowlist (#16).
	if r.PassEnv == nil && r.ClearEnvEnabled() {
		r.PassEnv = def.PassEnv
	}
}

// mergeServiceDefaults layers def beneath an authored service. Name and Command
// identify a service and are never defaulted (the validator rejects them on
// defaults.service). A whole Ready probe is copied in when the service declares
// none, so shared readiness (e.g. `ready.store`/`ready.timeout`) need not repeat.
func mergeServiceDefaults(def, svc *spec.Service) {
	if svc.Shell == nil {
		svc.Shell = def.Shell
	}
	if svc.Cwd == "" {
		svc.Cwd = def.Cwd
	}
	svc.Env = mergeStringMap(def.Env, svc.Env)
	if svc.ClearEnv == nil {
		svc.ClearEnv = def.ClearEnv
	}
	// Mirrors mergeRunDefaults: no allowlist inheritance without clear_env (#16).
	if svc.PassEnv == nil && svc.ClearEnvEnabled() {
		svc.PassEnv = def.PassEnv
	}
	if svc.Ready == nil && def.Ready != nil {
		ready := *def.Ready
		svc.Ready = &ready
	}
}
