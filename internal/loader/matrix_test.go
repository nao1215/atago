package loader

import (
	"strings"
	"testing"
)

func TestLoadBytes_MatrixExpands(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: matrix
scenarios:
  - name: "greets ${who}"
    matrix:
      - { who: Alice, lang: en }
      - { who: Bob, lang: fr }
    steps:
      - run:
          command: echo ${who}
      - assert:
          stdout:
            contains: ${who}
`
	s, err := LoadBytes("m.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	if len(s.Scenarios) != 2 {
		t.Fatalf("scenarios = %d, want 2 (matrix expansion)", len(s.Scenarios))
	}
	if s.Scenarios[0].Name != "greets Alice" || s.Scenarios[1].Name != "greets Bob" {
		t.Errorf("names = %q, %q; want templated", s.Scenarios[0].Name, s.Scenarios[1].Name)
	}
	if s.Scenarios[0].Vars["lang"] != "en" || s.Scenarios[1].Vars["lang"] != "fr" {
		t.Errorf("row vars not bound: %+v / %+v", s.Scenarios[0].Vars, s.Scenarios[1].Vars)
	}
	if s.Scenarios[0].Matrix != nil {
		t.Errorf("matrix should be cleared on expanded instance")
	}
}

func TestLoadBytes_MatrixSuffixWhenNameNotTemplated(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: matrix
scenarios:
  - name: same
    matrix:
      - { n: "1" }
      - { n: "2" }
    steps:
      - run:
          command: echo ${n}
`
	s, err := LoadBytes("m.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	if s.Scenarios[0].Name != "same [n=1]" || s.Scenarios[1].Name != "same [n=2]" {
		t.Errorf("names = %q, %q; want deterministic suffixes", s.Scenarios[0].Name, s.Scenarios[1].Name)
	}
}

func TestLoadBytes_MatrixErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		src     string
		wantMsg string
	}{
		{
			name:    "empty matrix",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    matrix: []\n    steps:\n      - run: {command: echo}",
			wantMsg: "at least one row",
		},
		{
			name:    "empty row",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    matrix:\n      - {}\n    steps:\n      - run: {command: echo}",
			wantMsg: "at least one variable",
		},
		{
			name:    "expanded names collide",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: \"fixed ${x}\"\n    matrix:\n      - { x: dup }\n      - { x: dup }\n    steps:\n      - run: {command: echo}",
			wantMsg: "duplicate scenario name",
		},
		{
			name:    "matrix key shadows a built-in",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: \"m ${workdir}\"\n    matrix:\n      - { workdir: /tmp/x }\n    steps:\n      - run: {command: echo}",
			wantMsg: "shadows a built-in variable",
		},
		{
			name:    "template omits a distinguishing variable",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: \"check ${who}\"\n    matrix:\n      - { who: alice, lang: en }\n      - { who: alice, lang: fr }\n    steps:\n      - run: {command: echo}",
			wantMsg: "omits row variable \"lang\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadBytes("t.atago.yaml", []byte(tt.src))
			if err == nil {
				t.Fatalf("LoadBytes() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestLoadBytes_RetryValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		src     string
		wantMsg string
	}{
		{
			name:    "times below 1",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run:\n          command: echo\n          retry:\n            times: 0\n            until: {exit_code: 0}",
			wantMsg: "times must be >= 1",
		},
		{
			name:    "bad interval",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run:\n          command: echo\n          retry:\n            times: 3\n            interval: nope\n            until: {exit_code: 0}",
			wantMsg: "not a valid duration",
		},
		{
			name:    "missing until",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run:\n          command: echo\n          retry:\n            times: 3",
			wantMsg: "until is required",
		},
		{
			name:    "valid retry",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run:\n          command: echo\n          retry:\n            times: 3\n            interval: 10ms\n            until: {stdout: {contains: hi}}",
			wantMsg: "",
		},
		{
			name:    "until changes cannot be satisfied",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run:\n          command: echo\n          retry:\n            times: 3\n            until: {changes: {created: [out.txt]}}",
			wantMsg: "retry.until.changes cannot be satisfied",
		},
		{
			name:    "until screen cannot be satisfied",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run:\n          command: echo\n          retry:\n            times: 3\n            until: {screen: {contains: hi}}",
			wantMsg: "retry.until.screen cannot be satisfied",
		},
		{
			name:    "until duration is accepted",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run:\n          command: echo\n          retry:\n            times: 3\n            until: {duration: {lt: 2s}}",
			wantMsg: "",
		},
		{
			name:    "until exit_code is accepted",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run:\n          command: echo\n          retry:\n            times: 3\n            until: {exit_code: 0}",
			wantMsg: "",
		},
		{
			name:    "http retry times below 1",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - http:\n          method: GET\n          path: /job\n          retry:\n            times: 0\n            until: {status: 200}",
			wantMsg: "http.retry.times must be >= 1",
		},
		{
			name:    "http retry missing until",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - http:\n          method: GET\n          path: /job\n          retry:\n            times: 3",
			wantMsg: "http.retry.until is required",
		},
		{
			name:    "valid http retry",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - http:\n          method: GET\n          path: /job\n          retry:\n            times: 3\n            interval: 10ms\n            until: {status: 200}",
			wantMsg: "",
		},
		{
			name:    "http json and form conflict",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - http:\n          method: POST\n          path: /x\n          json: {a: 1}\n          form: {b: two}",
			wantMsg: "a request has one payload",
		},
		{
			name:    "http body_file and body conflict",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - http:\n          method: PUT\n          path: /x\n          body: raw\n          body_file: blob.bin",
			wantMsg: "a request has one payload",
		},
		{
			name:    "http file part missing field",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - http:\n          method: POST\n          path: /x\n          files:\n            - path: a.png",
			wantMsg: "files[0].field is required",
		},
		{
			name:    "valid multipart upload",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - http:\n          method: POST\n          path: /x\n          form: {title: t}\n          files:\n            - field: upload\n              path: a.png\n              content_type: image/png",
			wantMsg: "",
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
