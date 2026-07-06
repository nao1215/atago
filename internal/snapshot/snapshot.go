// Package snapshot stores and compares golden output. Captured
// output is normalized — volatile details such as ANSI codes, temp paths, the
// home directory, UUIDs, timestamps, and local ports are masked — so snapshots
// stay stable across machines and runs. A spec may also declare its own
// regex→placeholder rewrites via Options.Scrub (#137) for volatile patterns the
// built-ins do not cover.
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
	// Scrub, when set, applies the spec's user-declared regex→placeholder rewrites
	// (#137) to captured output: the open set of volatile patterns the built-in
	// normalizers below do not know about (auto-increment IDs, request
	// identifiers, custom timestamps). It runs AFTER secret masking and BEFORE
	// the built-in ANSI/UUID/timestamp/port/path normalization, so a rule sees the
	// already-secret-masked text and the built-ins see the already-scrubbed text.
	Scrub func([]byte) []byte
}

var (
	// reCSI matches a full ECMA-48 CSI escape: parameter bytes 0x30–0x3F
	// (digits, `;`, and the `?`/`:` used by private-mode and colon-subparam
	// sequences), intermediate bytes 0x20–0x2F, and a final byte 0x40–0x7E. The
	// old `[0-9;]*[A-Za-z]` missed `\x1b[?25l`/`\x1b[?25h` (cursor hide/show),
	// `\x1b[?1049h` (alt-screen), mouse-tracking, and colon SGR like
	// `\x1b[38:2:255:0:0m` — every one of which a spinner/TUI emits — leaking raw
	// escape bytes into golden snapshots.
	reCSI = regexp.MustCompile(`\x1b\[[0-9;?:]*[ -/]*[@-~]`)
	// reOSC matches an OSC sequence (window title, hyperlink) terminated by BEL
	// or ST, another volatile control sequence that must not reach the golden.
	reOSC       = regexp.MustCompile(`\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)`)
	reUUID      = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
	reTimestamp = regexp.MustCompile(`\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?`)
	// rePort masks the whole port after a loopback host. `\d+` (not `\d{2,5}`)
	// consumes every digit, so a 1-digit ephemeral port masks too and a >5-digit
	// value is not left with an orphan trailing digit.
	rePort = regexp.MustCompile(`(127\.0\.0\.1|0\.0\.0\.0|localhost|\[::1\]):\d+`)
)

// Normalize masks volatile details so snapshots are deterministic. Declared
// secrets are masked first so a real credential never reaches the golden file.
func Normalize(data []byte, opt Options) []byte {
	if opt.Secrets != nil {
		data = opt.Secrets(data)
	}
	if opt.Scrub != nil {
		data = opt.Scrub(data)
	}
	s := string(data)
	// Line endings are an OS artifact, not observable behavior: fold CRLF so a
	// snapshot recorded on POSIX matches cmd.exe output on Windows (and the
	// committed golden never carries CRs). Matches the equals matcher's rule.
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = reCSI.ReplaceAllString(s, "")
	s = reOSC.ReplaceAllString(s, "")
	s = reUUID.ReplaceAllString(s, "<uuid>")
	s = reTimestamp.ReplaceAllString(s, "<timestamp>")
	s = rePort.ReplaceAllString(s, "$1:<port>")
	// Path masking is component-boundary aware (it must not turn /home/naoki into
	// ~ki), so it depends on the byte that follows the prefix. Run it AFTER the
	// uuid/timestamp/port maskers so that neighbor byte is already in its final
	// form — otherwise a workdir abutting a timestamp (".../tmp/x2026-..." → mask
	// timestamp → ".../tmp/x<timestamp>") would flip the boundary between passes
	// and break idempotence.
	if opt.Workdir != "" {
		s = maskPathPrefix(s, opt.Workdir, "<workdir>")
	}
	// Skip a root home ("/"): masking it would rewrite every absolute path. It is
	// the home of some container/CI users, where os.UserHomeDir returns "/".
	if home, err := os.UserHomeDir(); err == nil && home != "" && home != "/" {
		s = maskPathPrefix(s, home, "~")
	}
	return []byte(s)
}

// maskPathPrefix replaces every occurrence of prefix in s with replacement, but
// only where prefix ends a path component — the following byte is a separator or
// other boundary character, or the match is at end of string. Masking a bare
// substring instead would corrupt a longer sibling path: the home dir /home/nao
// must not turn /home/naoki into ~ki, and the workdir /tmp/run1 must not turn
// /tmp/run10 into <workdir>0.
func maskPathPrefix(s, prefix, replacement string) string {
	if prefix == "" {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		if strings.HasPrefix(s[i:], prefix) {
			end := i + len(prefix)
			if end >= len(s) || isComponentBoundary(s[end]) {
				b.WriteString(replacement)
				i = end
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// isComponentBoundary reports whether byte c ends a path component so the prefix
// before it can be masked: a path separator or a whitespace/control byte, where a
// path token in captured output ends. Filename-legal bytes — letters, digits, and
// punctuation like '.', '-', '_', '+', '@' — are NOT boundaries, so a longer
// sibling path (/home/naoki for home /home/nao, /tmp/run1+cache for workdir
// /tmp/run1) is left intact rather than corrupted.
func isComponentBoundary(c byte) bool {
	return c == '/' || c == '\\' || c <= ' '
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
	// Fold CRLF in the stored golden too: Normalize already folds the actual, so
	// a golden checked out under git autocrlf=true (or saved by a CRLF editor)
	// would otherwise carry \r\n and never match the LF-folded actual — a
	// spurious failure on every snapshot assertion. Line endings are an OS
	// artifact, not observable behavior (same rule as the equals matcher).
	expected = strings.ReplaceAll(string(stored), "\r\n", "\n")
	return expected == actualNorm, expected, actualNorm, nil
}

// Update writes the normalized actual output to path, creating parent dirs.
func Update(path string, actual []byte, opt Options) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	return os.WriteFile(path, Normalize(actual, opt), 0o600)
}
