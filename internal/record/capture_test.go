package record

import (
	"io"
	"os"
	"strings"
	"testing"
)

// TestCapturePTY_RecordsOutputAndExit drives the whole capture path with a
// self-exiting command and no interactive input: start the child in a real pty
// (POSIX) / ConPTY (Windows), drain its output into the recording, and reap its
// exit code. It runs on every OS, so it is the automated backstop for the
// otherwise interactive, human-driven `record --pty` — including the Windows
// ConPTY capture, which has no other test.
func TestCapturePTY_RecordsOutputAndExit(t *testing.T) {
	inR, inW, err := os.Pipe() // an empty input pipe: the command needs no keystrokes
	if err != nil {
		t.Fatalf("input pipe: %v", err)
	}
	defer func() { _ = inR.Close(); _ = inW.Close() }()
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("output pipe: %v", err)
	}
	defer func() { _ = outR.Close(); _ = outW.Close() }()
	// Drain the output pipe so the terminal's writes never block on a full buffer.
	go func() { _, _ = io.Copy(io.Discard, outR) }()

	// shell:true → `sh -c` / `cmd /S /C`, so echo is a builtin on both shells.
	rec, err := CapturePTY("echo capture-marker", true, inR, outW)
	if err != nil {
		t.Fatalf("CapturePTY() error = %v", err)
	}
	if rec.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", rec.ExitCode)
	}
	var out strings.Builder
	for _, seg := range rec.Segments {
		out.Write(seg.Output)
	}
	if !strings.Contains(out.String(), "capture-marker") {
		t.Errorf("recording missing the child's output: %q", out.String())
	}
}
