package loader

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/spec"
)

const viewportSpec = `version: "1"
suite: {name: viewport}
runners:
  web: %s
scenarios:
  - name: viewport
    steps:
      - cdp:
          runner: web
          actions: [%s]
`

func TestLoadBytes_Viewport(t *testing.T) {
	t.Parallel()
	scale := 2.0
	for _, tc := range []struct {
		viewport, want string
		expected       spec.Viewport
	}{
		{viewport: "{width: 390, height: 844}", expected: spec.Viewport{Width: 390, Height: 844}},
		{viewport: "{width: 420, height: 900, device_scale_factor: 2, mobile: true}", expected: spec.Viewport{Width: 420, Height: 900, DeviceScaleFactor: &scale, Mobile: true}},
		{viewport: "{}", want: "width and height"},
		{viewport: "{width: 390}", want: "width and height"},
		{viewport: "{width: 0, height: 844}", want: "width and height"},
		{viewport: "{width: 390, height: -1}", want: "width and height"},
		{viewport: "{width: 10000001, height: 844}", want: "width and height"},
		{viewport: "{width: 390.5, height: 844}", want: "expected a whole number"},
		{viewport: "{width: 390, height: 844, device_scale_factor: 0}", want: "finite and positive"},
		{viewport: "{width: 390, height: 844, device_scale_factor: -1}", want: "finite and positive"},
		{viewport: "{width: 390, height: 844, device_scale_factor: .inf}", want: "finite and positive"},
		{viewport: "{width: 390, height: 844, device_scale_factor: .nan}", want: "finite and positive"},
	} {
		for _, atRunner := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/runner=%v", tc.viewport, atRunner), func(t *testing.T) {
				t.Parallel()
				runner := "{type: browser}"
				action := "{eval: '1'}"
				if atRunner {
					runner = "{type: browser, viewport: " + tc.viewport + "}"
				} else {
					action = "{viewport: " + tc.viewport + "}"
				}
				src := fmt.Sprintf(viewportSpec, runner, action)
				s, err := LoadBytes("viewport.atago.yaml", []byte(src))
				if tc.want != "" {
					if err == nil || !strings.Contains(err.Error(), tc.want) {
						t.Fatalf("error = %v, want %q", err, tc.want)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				v := s.Runners["web"].Viewport
				if !atRunner {
					v = s.Scenarios[0].Steps[0].CDP.Actions[0].Viewport
				}
				if v == nil || !reflect.DeepEqual(*v, tc.expected) {
					t.Fatalf("viewport = %+v, want %+v", v, tc.expected)
				}
			})
		}
	}
}

func TestLoadBytes_ViewportPlacement(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ runner, action, want string }{
		{"{type: http, viewport: {width: 390, height: 844}}", "{eval: '1'}", "cannot be set on a http runner"},
		{"{type: browser}", "{viewport: {width: 390, height: 844}, eval: '1'}", "sets multiple actions"},
	} {
		src := fmt.Sprintf(viewportSpec, tc.runner, tc.action)
		_, err := LoadBytes("viewport.atago.yaml", []byte(src))
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("error = %v, want %q", err, tc.want)
		}
	}
}
