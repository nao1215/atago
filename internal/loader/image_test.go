package loader

import (
	"strings"
	"testing"
)

const imageSpecHead = "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - assert:\n          image:\n"

func TestLoadBytes_ImageValid(t *testing.T) {
	t.Parallel()
	srcs := []string{
		imageSpecHead + "            path: out.png\n            format: png\n",
		imageSpecHead + "            path: out.png\n            width: 800\n            height: 600\n",
		imageSpecHead + "            path: out.png\n            min_width: 10\n            max_width: 100\n",
		imageSpecHead + "            path: out.png\n            alpha: true\n",
		imageSpecHead + "            path: out.png\n            similar_to: base.png\n            max_diff: 0.02\n",
	}
	for i, src := range srcs {
		if _, err := LoadBytes("t.atago.yaml", []byte(src)); err != nil {
			t.Errorf("case %d: unexpected error: %v", i, err)
		}
	}
}

func TestLoadBytes_ImageErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		body    string
		wantMsg string
	}{
		{
			name:    "no constraint",
			body:    "            path: out.png\n",
			wantMsg: "must set at least one",
		},
		{
			name:    "invalid format",
			body:    "            path: out.png\n            format: heic\n",
			wantMsg: "format \"heic\" is invalid",
		},
		{
			name:    "max_diff without similar_to",
			body:    "            path: out.png\n            max_diff: 0.1\n",
			wantMsg: "max_diff requires similar_to",
		},
		{
			name:    "max_diff out of range",
			body:    "            path: out.png\n            similar_to: b.png\n            max_diff: 2\n",
			wantMsg: "max_diff must be between 0 and 1",
		},
		{
			name:    "min exceeds max width",
			body:    "            path: out.png\n            min_width: 200\n            max_width: 100\n",
			wantMsg: "min_width 200 exceeds max_width 100",
		},
		{
			name:    "min exceeds max height",
			body:    "            path: out.png\n            min_height: 200\n            max_height: 100\n",
			wantMsg: "min_height 200 exceeds max_height 100",
		},
		{
			name:    "negative dimension",
			body:    "            path: out.png\n            width: -5\n",
			wantMsg: "dimensions must be >= 0",
		},
		{
			name:    "avif cannot be measured",
			body:    "            path: out.avif\n            format: avif\n            width: 100\n",
			wantMsg: "cannot be measured",
		},
		{
			name:    "svg cannot be compared",
			body:    "            path: out.svg\n            format: svg\n            similar_to: base.svg\n",
			wantMsg: "cannot be measured",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadBytes("t.atago.yaml", []byte(imageSpecHead+tt.body))
			if err == nil {
				t.Fatalf("expected error for %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.wantMsg)
			}
		})
	}
}
