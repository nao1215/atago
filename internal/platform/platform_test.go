package platform

import "testing"

func TestOSAndMatches(t *testing.T) {
	// OS reflects the current host; Matches compares against it. Override the
	// package var so the assertions are host-independent.
	orig := currentOS
	t.Cleanup(func() { currentOS = orig })

	currentOS = "linux"
	if OS() != "linux" {
		t.Errorf("OS() = %q, want linux", OS())
	}
	if !Matches("linux") {
		t.Error("Matches(linux) = false, want true")
	}
	if Matches("windows") {
		t.Error("Matches(windows) = true, want false")
	}
	// An empty condition OS matches nothing.
	if Matches("") {
		t.Error("Matches(\"\") = true, want false")
	}
}
