package store

import (
	"math/rand"
	"testing"
	"testing/quick"
)

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

// TestExpand_EnvRefs proves ${env:NAME} resolves from the host environment:
// set variables expand (including set-but-empty), unset ones stay verbatim so
// they surface as failures, $${env:NAME} stays literal, and env names never
// fall back to same-named store variables.
func TestExpand_EnvRefs(t *testing.T) {
	t.Setenv("ATAGO_STORE_TEST", "from-env")
	t.Setenv("ATAGO_STORE_EMPTY", "")
	s := New()
	s.Set("ATAGO_STORE_UNSET", "from-store") // must NOT satisfy ${env:ATAGO_STORE_UNSET}

	tests := []struct {
		in, want string
	}{
		{"v is ${env:ATAGO_STORE_TEST}", "v is from-env"},
		{"empty [${env:ATAGO_STORE_EMPTY}]", "empty []"},
		{"unset ${env:ATAGO_STORE_UNSET} stays", "unset ${env:ATAGO_STORE_UNSET} stays"},
		{"literal $${env:ATAGO_STORE_TEST}", "literal ${env:ATAGO_STORE_TEST}"},
	}
	for _, tt := range tests {
		if got := s.Expand(tt.in); got != tt.want {
			t.Errorf("Expand(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}

	// Unresolved reports the unset env reference with its env: prefix.
	got := s.Unresolved("run ${env:ATAGO_STORE_UNSET}")
	if len(got) != 1 || got[0] != "env:ATAGO_STORE_UNSET" {
		t.Errorf("Unresolved = %v, want [env:ATAGO_STORE_UNSET]", got)
	}
	if got := s.Unresolved("run ${env:ATAGO_STORE_TEST}"); len(got) != 0 {
		t.Errorf("Unresolved(set env) = %v, want none", got)
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

// TestEscapeExpandRoundTrip proves store.Escape is the exact inverse of Expand:
// Expand(Escape(x)) == x for any observed text. record relies on this to embed a
// raw recorded command/output/pty-send in a spec without it being re-expanded at
// replay. A blind ${→$${ rewrite broke it for a ${ not followed by a valid name
// (e.g. a tool that prints ${1}): the expander only unescapes $${<valid-name>},
// so $${1} never round-tripped back to ${1}.
func TestEscapeExpandRoundTrip(t *testing.T) {
	t.Parallel()
	s := New()
	s.Set("name", "Alice")
	s.Set("user_id", "42")
	cases := []string{
		"plain text with no refs",
		"live ${name} reference",
		"unknown ${missing} reference",
		"${user_id} and ${name} together",
		"already escaped $${name}",
		"double escaped $$${user_id}",
		"digit name ${1}",
		"leading-digit name ${9zzz}",
		"empty name ${}",
		"space in name ${ x}",
		"unclosed ${name",
		"env ref ${env:PATH}",
		"bare pid $$ and price $$5",
		"adjacent ${name}${user_id}",
		"mixed ${name}/${1}/$${user_id}/${}",
		"dollar-brace-digit output val=${1} end",
	}
	for _, in := range cases {
		if got := s.Expand(Escape(in)); got != in {
			t.Errorf("Expand(Escape(%q)) = %q, want %q (round-trip broken)", in, got, in)
		}
	}
}

// TestEscapeExpandRoundTrip_Quick fuzzes the same inverse law over random inputs
// with a fixed seed, so a future change to the reference grammar that breaks the
// escape/expand symmetry is caught.
func TestEscapeExpandRoundTrip_Quick(t *testing.T) {
	t.Parallel()
	s := New()
	s.Set("a", "X")
	s.Set("name", "Alice")
	roundTrips := func(in string) bool { return s.Expand(Escape(in)) == in }
	cfg := &quick.Config{Rand: rand.New(rand.NewSource(1)), MaxCount: 3000}
	if err := quick.Check(roundTrips, cfg); err != nil {
		t.Errorf("escape/expand round-trip law broken: %v", err)
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
