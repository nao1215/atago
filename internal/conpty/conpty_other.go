//go:build !windows

// Package conpty is Windows-only; this file provides non-Windows stubs so the
// package (and anything that imports it behind a build tag) still compiles
// everywhere. None of these are reached on POSIX — the callers are all
// Windows-tagged — but keeping the surface identical means a cross-platform
// dispatcher could reference it without extra guards.
package conpty

import (
	"context"
	"errors"
)

// errUnsupported is returned by every stub; ConPTY exists only on Windows.
var errUnsupported = errors.New("conpty: pseudo consoles are a Windows-only feature")

// PseudoConsole is the non-Windows stub of the Windows pseudo console.
type PseudoConsole struct{}

// IsAvailable is always false off Windows.
func IsAvailable() bool { return false }

// CommandLine is unsupported off Windows.
func CommandLine(string, bool) (string, error) { return "", errUnsupported }

// Start is unsupported off Windows.
func Start(string, string, []string, int, int) (*PseudoConsole, error) { return nil, errUnsupported }

// Read is unsupported off Windows.
func (c *PseudoConsole) Read([]byte) (int, error) { return 0, errUnsupported }

// Write is unsupported off Windows.
func (c *PseudoConsole) Write([]byte) (int, error) { return 0, errUnsupported }

// Resize is unsupported off Windows.
func (c *PseudoConsole) Resize(int, int) error { return errUnsupported }

// Wait is unsupported off Windows.
func (c *PseudoConsole) Wait(context.Context) int { return -1 }

// Pid is unsupported off Windows.
func (c *PseudoConsole) Pid() int { return 0 }

// Close is unsupported off Windows.
func (c *PseudoConsole) Close() error { return errUnsupported }
