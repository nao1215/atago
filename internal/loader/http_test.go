package loader

import (
	"strings"
	"testing"
)

func TestLoadBytes_HTTPValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		src     string
		wantMsg string // "" means the spec must load cleanly
	}{
		{
			name:    "header assert needs a name",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - assert:\n          header:\n            equals: y",
			wantMsg: "name is required",
		},
		{
			name:    "header assert needs a matcher",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - assert:\n          header:\n            name: X",
			wantMsg: "must set one of contains/equals",
		},
		{
			name:    "store from.header is a valid source",
			src:     "version: \"1\"\nsuite:\n  name: x\nrunners:\n  api:\n    type: http\n    base_url: http://localhost:8080\nscenarios:\n  - name: a\n    steps:\n      - http:\n          runner: api\n          method: GET\n          path: /x\n      - store:\n          name: loc\n          from:\n            header: Location",
			wantMsg: "",
		},
		{
			name:    "store with two from sources is rejected",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - store:\n          name: v\n          from:\n            header: X\n            body: {contains: y}",
			wantMsg: "exactly one source",
		},
		{
			name:    "http step needs a method",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - http:\n          path: /x",
			wantMsg: "http.method is required",
		},
		{
			name:    "http step with a raw body is valid",
			src:     "version: \"1\"\nsuite:\n  name: x\nrunners:\n  api:\n    type: http\n    base_url: http://localhost:8080\nscenarios:\n  - name: a\n    steps:\n      - http:\n          runner: api\n          method: POST\n          path: /x\n          body: \"metric 1\"",
			wantMsg: "",
		},
		{
			name:    "http step with both json and body is rejected",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - http:\n          runner: api\n          method: POST\n          path: /x\n          json: {a: 1}\n          body: \"raw\"",
			wantMsg: "sets json and body; a request has one payload",
		},
		{
			name:    "valid http step + status assert",
			src:     "version: \"1\"\nsuite:\n  name: x\nrunners:\n  api:\n    type: http\n    base_url: http://localhost:8080\nscenarios:\n  - name: a\n    steps:\n      - http:\n          runner: api\n          method: GET\n          path: /x\n      - assert:\n          status: 200",
			wantMsg: "",
		},
		{
			name:    "store stdout matches invalid regexp is rejected at load",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo hi}\n      - store:\n          name: v\n          from:\n            stdout: {matches: \"[unterminated\"}",
			wantMsg: "not a valid regexp",
		},
		{
			name:    "store stdout json valid path loads",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo hi}\n      - store:\n          name: v\n          from:\n            stdout: {json: {path: \"$.token\"}}",
			wantMsg: "",
		},
		{
			name:    "store stdout without json or matches is rejected",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo hi}\n      - store:\n          name: v\n          from:\n            stdout: {contains: hi}",
			wantMsg: "json path, a matches regexp, or trim",
		},
		{
			name:    "store name shadowing a built-in is rejected",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo hi}\n      - store:\n          name: workdir\n          from:\n            stdout: {matches: \"(.*)\"}",
			wantMsg: "shadows a built-in variable",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadBytes("t.atago.yaml", []byte(tt.src))
			if tt.wantMsg == "" {
				if err != nil {
					t.Fatalf("LoadBytes() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("LoadBytes() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.wantMsg)
			}
		})
	}
}
