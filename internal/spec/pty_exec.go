package spec

import (
	"fmt"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// PTYExec is one host command run mid-session (#380). It accepts either a
// scalar string (argv-parsed like run.command) or a mapping carrying the shell
// and timeout knobs.
//
// The command runs in the scenario workdir with the same environment the pty
// child received, so `sandbox_home` / `clear_env` isolation is not quietly
// punctured by the helper. Its output never joins the transcript — the
// transcript is what the TERMINAL showed — and is kept only to explain a
// failure. A non-zero exit, a timeout, or a failure to start is a hard step
// error: the command is scaffolding, not the subject under test, so a broken
// one must stop the run rather than leave the following expect_screen waiting
// for a change that was never made.
//
// It is deliberately not a scripting language: a fixed command at a fixed point
// in the session, no branching, no captured output feeding later steps.
type PTYExec struct {
	// Command is the program to run. Required.
	Command string `yaml:"command"`
	// Shell runs Command through the shell, like run.shell.
	Shell *bool `yaml:"shell,omitempty"`
	// Timeout bounds this command as a Go duration (default 10s). The remaining
	// session budget bounds it too, whichever is shorter.
	Timeout string `yaml:"timeout,omitempty"`

	// mapped records that the author used the mapping form, so an empty mapping
	// ({}) is rejected rather than read as "no command".
	mapped bool
}

// ShellEnabled reports whether the exec command runs through the shell.
func (e *PTYExec) ShellEnabled() bool { return e.Shell != nil && *e.Shell }

// DefaultPTYExecTimeout bounds a mid-session host command when the spec does
// not (#380). Scaffolding a session waits on should be quick; a command that is
// not is one the author should have to bound on purpose.
const DefaultPTYExecTimeout = 10 * time.Second

// UnmarshalYAML decodes exec as a scalar command string or a mapping, rejecting
// unknown mapping keys (a custom unmarshaler bypasses the loader's strict
// decode). It decodes from the AST node so every shape error carries the
// offending value's [line:col].
func (e *PTYExec) UnmarshalYAML(node ast.Node) error {
	fail := func(format string, args ...any) error {
		return &yaml.SyntaxError{Message: fmt.Sprintf(format, args...), Token: node.GetToken()}
	}
	var one string
	if err := yaml.NodeToValue(node, &one); err == nil {
		e.Command = one
		return nil
	}
	var raw map[string]any
	if err := yaml.NodeToValue(node, &raw); err != nil {
		return fail("exec must be a string or {command: ..., shell: bool, timeout: duration}")
	}
	e.mapped = true
	for k, v := range raw {
		switch k {
		case "command":
			str, ok := v.(string)
			if !ok {
				return fail("exec.command must be a string")
			}
			e.Command = str
		case "shell":
			b, ok := v.(bool)
			if !ok {
				return fail("exec.shell must be true or false")
			}
			e.Shell = &b
		case "timeout":
			str, ok := v.(string)
			if !ok {
				return fail("exec.timeout must be a duration string (e.g. \"10s\")")
			}
			e.Timeout = str
		default:
			return fail("exec: unknown key %q (accepted: command, shell, timeout)", k)
		}
	}
	return nil
}

// MarshalYAML emits the shape UnmarshalYAML accepts, so a loaded exec
// round-trips: the scalar form when nothing else was set, the mapping otherwise.
func (e PTYExec) MarshalYAML() (any, error) {
	if !e.mapped && e.Shell == nil && e.Timeout == "" {
		return e.Command, nil
	}
	m := map[string]any{"command": e.Command}
	if e.Shell != nil {
		m["shell"] = *e.Shell
	}
	if e.Timeout != "" {
		m["timeout"] = e.Timeout
	}
	return m, nil
}

// Label renders the exec symbolically for explain/doc (#380).
func (e *PTYExec) Label() string { return fmt.Sprintf("exec %q", e.Command) }
