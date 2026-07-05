package loader

import (
	"strconv"
	"time"

	"github.com/nao1215/atago/internal/spec"
)

func validateFixture(add func(string, ...any), where string, f *spec.Fixture) {
	if f.File == "" {
		add("%s.fixture.file is required", where)
	}
	n := 0
	if f.Content != "" {
		n++
	}
	if f.Base64 != "" {
		n++
	}
	if f.From != "" {
		n++
	}
	if f.Symlink != "" {
		n++
	}
	if n > 1 {
		add("%s.fixture: set only one of content, base64, from, or symlink", where)
	}
	if f.Symlink != "" && f.Mode != "" {
		add("%s.fixture: mode cannot be applied to a symlink", where)
	}
	if f.Mode != "" {
		if _, err := strconv.ParseUint(f.Mode, 8, 32); err != nil {
			add("%s.fixture.mode %q is not an octal file mode (e.g. \"0444\")", where, f.Mode)
		}
	}
	if f.Mtime != "" {
		if _, err := time.Parse(time.RFC3339, f.Mtime); err != nil {
			add("%s.fixture.mtime %q is not an RFC3339 timestamp (e.g. \"2026-01-02T15:04:05Z\")", where, f.Mtime)
		}
	}
}
