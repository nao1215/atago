package store

import "testing"

func TestGet(t *testing.T) {
	t.Parallel()
	s := New()
	s.Set("k", "v")
	if got, ok := s.Get("k"); !ok || got != "v" {
		t.Errorf("Get(k) = %q,%v want v,true", got, ok)
	}
	if _, ok := s.Get("missing"); ok {
		t.Error("Get(missing) ok = true, want false")
	}
}

func TestExpand(t *testing.T) {
	t.Parallel()
	s := New()
	s.Set("user_id", "42")
	s.Set("name", "Alice")

	tests := []struct {
		in, want string
	}{
		{"id is ${user_id}", "id is 42"},
		{"${name}-${user_id}", "Alice-42"},
		{"no vars here", "no vars here"},
		{"unknown ${missing} stays", "unknown ${missing} stays"},
		// Issue #37: $${name} is a literal escape and must render as ${name}
		// without expansion, even when the variable exists.
		{"literal $${name} stays", "literal ${name} stays"},
		{"$${user_id} and ${user_id}", "${user_id} and 42"},
		// A bare $$ (e.g. shell PID) is not an escape and is left untouched.
		{"pid is $$", "pid is $$"},
		{"cost is $$5", "cost is $$5"},
	}
	for _, tt := range tests {
		if got := s.Expand(tt.in); got != tt.want {
			t.Errorf("Expand(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestExpandMap(t *testing.T) {
	t.Parallel()
	s := New()
	s.Set("dir", "/tmp/x")
	out := s.ExpandMap(map[string]string{"GOBIN": "${dir}/bin"})
	if out["GOBIN"] != "/tmp/x/bin" {
		t.Errorf("GOBIN = %q, want /tmp/x/bin", out["GOBIN"])
	}
}

func TestUnresolved(t *testing.T) {
	t.Parallel()
	s := New()
	s.Set("known", "v")

	tests := []struct {
		in   string
		want int
		name string
	}{
		{"echo ${known}", 0, ""},
		{"echo ${missing}", 1, "missing"},
		{"echo $${literal}", 0, ""}, // escaped: the author wants literal text
		{"plain text $HOME $$", 0, ""},
		{"${known} then ${typo}", 1, "typo"},
	}
	for _, tt := range tests {
		got := s.Unresolved(tt.in)
		if len(got) != tt.want {
			t.Errorf("Unresolved(%q) = %v, want %d unresolved", tt.in, got, tt.want)
			continue
		}
		if tt.want == 1 && got[0] != tt.name {
			t.Errorf("Unresolved(%q) = %v, want [%s]", tt.in, got, tt.name)
		}
	}
}
