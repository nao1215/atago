package plural

import "testing"

func TestCount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		n        int
		singular string
		plural   string
		want     string
	}{
		{"zero takes the plural", 0, "page", "pages", "0 pages"},
		{"one takes the singular", 1, "page", "pages", "1 page"},
		{"two takes the plural", 2, "page", "pages", "2 pages"},
		{"an irregular plural is spelled out", 1, "entry", "entries", "1 entry"},
		{"an irregular plural for many", 3, "entry", "entries", "3 entries"},
		{"minus one is still singular", -1, "line", "lines", "-1 line"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Count(tt.n, tt.singular, tt.plural); got != tt.want {
				t.Errorf("Count(%d, %q, %q) = %q, want %q", tt.n, tt.singular, tt.plural, got, tt.want)
			}
		})
	}
}

func TestNoun(t *testing.T) {
	t.Parallel()
	if got := Noun(1, "file", "files"); got != "file" {
		t.Errorf("Noun(1) = %q, want %q", got, "file")
	}
	if got := Noun(0, "file", "files"); got != "files" {
		t.Errorf("Noun(0) = %q, want %q", got, "files")
	}
}
