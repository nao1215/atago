package browser

import "testing"

// TestSplitFlag covers the launch-flag parsing used to plumb browser_args into
// chromedp: bare switches become (name, true), key=value flags keep their string
// value, a leading "--" is tolerated, and empty tokens are dropped.
func TestSplitFlag(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in        string
		wantName  string
		wantValue any
	}{
		{"disable-gpu", "disable-gpu", true},
		{"--disable-gpu", "disable-gpu", true},
		{"window-size=1280,720", "window-size", "1280,720"},
		{"--window-size=1280,720", "window-size", "1280,720"},
		{"proxy-server=http://127.0.0.1:8080", "proxy-server", "http://127.0.0.1:8080"},
		{"  lang=en-US  ", "lang", "en-US"},
		{"", "", nil},
		{"--", "", nil},
		{"=novalue", "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			name, value := splitFlag(tt.in)
			if name != tt.wantName {
				t.Errorf("splitFlag(%q) name = %q, want %q", tt.in, name, tt.wantName)
			}
			if name != "" && value != tt.wantValue {
				t.Errorf("splitFlag(%q) value = %v, want %v", tt.in, value, tt.wantValue)
			}
		})
	}
}
