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

	// The BSDs are separate hosts, not one family: a gate naming FreeBSD
	// does not hold on OpenBSD, and "bsd" is not a host at all.
	currentOS = "freebsd"
	if OS() != "freebsd" {
		t.Errorf("OS() = %q, want freebsd", OS())
	}
	if !Matches("freebsd") {
		t.Error("Matches(freebsd) = false, want true")
	}
	for _, other := range []string{"openbsd", "netbsd", "darwin", "bsd"} {
		if Matches(other) {
			t.Errorf("Matches(%q) = true on freebsd, want false", other)
		}
	}
}
