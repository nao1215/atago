package buildinfo

import (
	"strings"
	"testing"
)

// TestSchemaRef proves the emitted schema URL pins to a clean release tag when
// this build is one, and otherwise falls back to "main" — a ref GitHub raw can
// always resolve — for "dev" and pseudo-version builds (#121).
func TestSchemaRef(t *testing.T) {
	orig := Version
	t.Cleanup(func() { Version = orig })

	tests := []struct {
		version string
		want    string
	}{
		{"v0.4.0", "v0.4.0"},
		{"v1.2.3", "v1.2.3"},
		{"dev", "main"}, // source checkout
		{"v0.3.5-0.20260101120000-abcdef123456", "main"}, // pseudo-version
		{"v0.4.0-rc1", "main"},                           // pre-release tag
		{"garbage", "main"},
	}
	for _, tt := range tests {
		Version = tt.version
		if got := SchemaRef(); got != tt.want {
			t.Errorf("SchemaRef() with Version=%q = %q, want %q", tt.version, got, tt.want)
		}
	}
}

// TestSchemaHeader proves the header is a single yaml-language-server comment
// line carrying an absolute schema URL, terminated with a newline.
func TestSchemaHeader(t *testing.T) {
	orig := Version
	t.Cleanup(func() { Version = orig })
	Version = "v0.4.0"

	h := SchemaHeader()
	if !strings.HasPrefix(h, "# yaml-language-server: $schema=https://") {
		t.Errorf("header %q must start with an absolute yaml-language-server schema comment", h)
	}
	if !strings.HasSuffix(h, "/schema/atago.schema.json\n") {
		t.Errorf("header %q must point at the schema file and end with a newline", h)
	}
	if !strings.Contains(h, "/v0.4.0/") {
		t.Errorf("header %q must pin to the release tag", h)
	}
	if strings.Count(h, "\n") != 1 {
		t.Errorf("header must be exactly one line, got %q", h)
	}
}
