//go:build windows

package ptyrun

import (
	"context"
	"errors"

	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// Run reports that pty steps are POSIX-only for now. The loader accepts the
// step on every platform so specs stay portable to author; at execution time
// Windows gets one clear error (an execution error, exit 4) instead of a
// confusing terminal failure. ConPTY support can lift this later.
func Run(_ context.Context, _ *spec.PTY, _ string, _ []string) (*runner.Result, *ExpectFailure, error) {
	return nil, nil, errors.New("pty steps are not supported on Windows yet (POSIX-only; gate the scenario with `skip: {os: windows}`)")
}
