package loader

import (
	"fmt"
	"strings"
	"testing"
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
	for _, tc := range []struct{ viewport, want string }{
		{"{width: 390, height: 844}", ""},
		{"{width: 420, height: 900, device_scale_factor: 2, mobile: true}", ""},
		{"{}", "width and height"},
		{"{width: 390}", "width and height"},
		{"{width: 0, height: 844}", "width and height"},
		{"{width: 390, height: -1}", "width and height"},
		{"{width: 10000001, height: 844}", "width and height"},
		{"{width: 390.5, height: 844}", "expected a whole number"},
		{"{width: 390, height: 844, device_scale_factor: 0}", "finite and positive"},
		{"{width: 390, height: 844, device_scale_factor: -1}", "finite and positive"},
		{"{width: 390, height: 844, device_scale_factor: .inf}", "finite and positive"},
		{"{width: 390, height: 844, device_scale_factor: .nan}", "finite and positive"},
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
				if v == nil || (v.Width != 390 && v.Width != 420) || (v.Height != 844 && v.Height != 900) {
					t.Fatalf("viewport not decoded: %+v", v)
				}
				if strings.Contains(tc.viewport, "mobile") && (!v.Mobile || v.DeviceScaleFactor == nil || *v.DeviceScaleFactor != 2) {
					t.Fatalf("metrics not decoded: %+v", v)
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
