package loader

import (
	"testing"

	"github.com/nao1215/atago/internal/spec"
)

// mustLoadSpec loads src under the given spec path and returns the parsed spec,
// failing the test on a load error. It collapses the success-path load
// boilerplate shared by the loader tests; error-path tests keep their own
// LoadBytes call so they can assert on the specific message.
func mustLoadSpec(t *testing.T, path, src string) *spec.Spec {
	t.Helper()
	s, err := LoadBytes(path, []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	return s
}
