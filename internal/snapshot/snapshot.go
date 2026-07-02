// Package snapshot stores and compares golden output (spec.md §16.10). Captured
// output is normalized — volatile details such as ANSI codes, temp paths, the
// home directory, UUIDs, timestamps, and local ports are masked — so snapshots
// stay stable across machines and runs.
package snapshot

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ErrMissing reports that a snapshot file does not exist yet; the caller should
// suggest re-running with --update-snapshots.
var ErrMissing = errors.New("snapshot file does not exist")

// Options controls normalization.
type Options struct {
	// Workdir is the scenario's temporary directory; occurrences are masked so a
	// random temp path does not churn the snapshot.
	Workdir string
	// Secrets, when set, masks declared secret values in captured output before it
	// is written to or compared against the golden file, so a real credential
	// never lands in a committed snapshot (issue #11). It is applied before the
	// volatile-detail normalization below.
	Secrets func([]byte) []byte
}

var (
	reANSI      = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)
	reUUID      = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
	reTimestamp = regexp.MustCompile(`\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?`)
	rePort      = regexp.MustCompile(`(127\.0\.0\.1|0\.0\.0\.0|localhost|\[::1\]):\d{2,5}`)
)

// Normalize masks volatile details so snapshots are deterministic. Declared
// secrets are masked first so a real credential never reaches the golden file.
func Normalize(data []byte, opt Options) []byte {
	if opt.Secrets != nil {
		data = opt.Secrets(data)
	}
	s := string(data)
	// Line endings are an OS artifact, not observable behavior: fold CRLF so a
	// snapshot recorded on POSIX matches cmd.exe output on Windows (and the
	// committed golden never carries CRs). Matches the equals matcher's rule.
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = reANSI.ReplaceAllString(s, "")
	if opt.Workdir != "" {
		s = strings.ReplaceAll(s, opt.Workdir, "<workdir>")
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		s = strings.ReplaceAll(s, home, "~")
	}
	s = reUUID.ReplaceAllString(s, "<uuid>")
	s = reTimestamp.ReplaceAllString(s, "<timestamp>")
	s = rePort.ReplaceAllString(s, "$1:<port>")
	return []byte(s)
}

// Compare normalizes actual and checks it against the stored snapshot at path.
// It returns the normalized expected and actual text for diffing. If the
// snapshot does not exist it returns ErrMissing.
func Compare(path string, actual []byte, opt Options) (ok bool, expected, actualNorm string, err error) {
	stored, rerr := os.ReadFile(path) //nolint:gosec // snapshot path is user-declared
	if rerr != nil {
		if os.IsNotExist(rerr) {
			return false, "", string(Normalize(actual, opt)), ErrMissing
		}
		return false, "", "", rerr
	}
	actualNorm = string(Normalize(actual, opt))
	expected = string(stored)
	return expected == actualNorm, expected, actualNorm, nil
}

// Update writes the normalized actual output to path, creating parent dirs.
func Update(path string, actual []byte, opt Options) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	return os.WriteFile(path, Normalize(actual, opt), 0o600)
}
