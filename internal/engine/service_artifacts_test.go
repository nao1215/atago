package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// serviceLog returns the recorded service-log artifact for name, or fails.
func serviceLog(t *testing.T, res *SuiteResult, name string) ServiceLog {
	t.Helper()
	for i := range res.Scenarios {
		for _, sl := range res.Scenarios[i].ServiceLogs {
			if sl.Name == name {
				return sl
			}
		}
	}
	t.Fatalf("no service log recorded for %q", name)
	return ServiceLog{}
}

// TestEngine_ServiceLogPreservedOnReadinessFailure covers a service that never
// becomes ready: its captured output must be preserved as a durable artifact
// (#51).
func TestEngine_ServiceLogPreservedOnReadinessFailure(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	res := runSpecWithArtifacts(t, "svc.atago.yaml", `
version: "1"
suite:
  name: s
scenarios:
  - name: never ready but talks
    services:
      - name: chatty
        shell: true
        command: '`+echoThenIdle("booting-up", 5)+`'
        ready:
          file: never.txt
          timeout: 150ms
    steps:
      - run: {shell: true, command: "echo unreached"}
`, root)
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error", res.Status)
	}
	sl := serviceLog(t, res, "chatty")
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(sl.Path)))
	if err != nil {
		t.Fatalf("read service log: %v", err)
	}
	if !strings.Contains(string(data), "booting-up") {
		t.Errorf("service log = %q, want the captured output", data)
	}
}

// TestEngine_ServiceLogPreservedOnLaterStepFailure covers a service that comes up
// fine but a later assertion fails: the still-running service's log must still be
// preserved (#51).
func TestEngine_ServiceLogPreservedOnLaterStepFailure(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	res := runSpecWithArtifacts(t, "svc.atago.yaml", `
version: "1"
suite:
  name: s
scenarios:
  - name: ready then fail
    services:
      - name: peer
        shell: true
        command: '`+echoThenIdle("peer-log", 5)+`'
        ready:
          # A log probe, not a file probe: readiness then guarantees "peer-log"
          # is already in the captured output buffer when the failing step tears
          # the scenario down — a file probe left a window where the buffer was
          # still empty and no artifact was written (flaked under -cover).
          log: peer-log
          timeout: 2s
    steps:
      - run: {shell: true, command: "echo hello"}
      - assert:
          stdout: {contains: goodbye}
`, root)
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed", res.Status)
	}
	sl := serviceLog(t, res, "peer")
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(sl.Path)))
	if err != nil {
		t.Fatalf("read service log: %v", err)
	}
	if !strings.Contains(string(data), "peer-log") {
		t.Errorf("service log = %q, want captured output", data)
	}
}

// TestEngine_SilentServiceReadinessFailureWritesNoArtifact is the regression from
// #51: a readiness failure where the service produced no output must not create
// an empty, noisy artifact.
func TestEngine_SilentServiceReadinessFailureWritesNoArtifact(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	res := runSpecWithArtifacts(t, "svc.atago.yaml", `
version: "1"
suite:
  name: s
scenarios:
  - name: silent and never ready
    services:
      - name: mute
        shell: true
        command: '`+sleepCmd(5)+`'
        ready:
          file: never.txt
          timeout: 150ms
    steps:
      - run: {shell: true, command: "echo unreached"}
`, root)
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error", res.Status)
	}
	if logs := res.Scenarios[0].ServiceLogs; len(logs) != 0 {
		t.Errorf("silent service should write no artifact, got %+v", logs)
	}
	// And no stray files under the artifacts root.
	var files int
	_ = filepath.WalkDir(root, func(_ string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			files++
		}
		return nil
	})
	if files != 0 {
		t.Errorf("expected no artifact files, found %d", files)
	}
}

// TestEngine_ServiceLogMasksSecrets verifies saved service logs mask declared
// secrets, consistent with the rest of the report (#51).
func TestEngine_ServiceLogMasksSecrets(t *testing.T) {
	const secret = "svc-s3cr3t"
	t.Setenv("ATAGO_SVC_SECRET", secret)
	root := t.TempDir()
	res := runSpecWithArtifacts(t, "svc.atago.yaml", `
version: "1"
suite:
  name: s
secrets:
  - ATAGO_SVC_SECRET
scenarios:
  - name: service leaks a secret then fails
    services:
      - name: leaky
        shell: true
        command: '`+echoThenIdle("token="+envRef("ATAGO_SVC_SECRET"), 5)+`'
        ready:
          file: never.txt
          timeout: 150ms
    steps:
      - run: {shell: true, command: "echo unreached"}
`, root)
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error", res.Status)
	}
	sl := serviceLog(t, res, "leaky")
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(sl.Path)))
	if err != nil {
		t.Fatalf("read service log: %v", err)
	}
	if strings.Contains(string(data), secret) {
		t.Errorf("service log leaked the secret: %q", data)
	}
	if !strings.Contains(string(data), "***") {
		t.Errorf("service log not masked: %q", data)
	}
}

// TestEngine_GreenServiceRunWritesNoLog verifies logs stay opt-in: a passing
// scenario with a healthy service writes no service-log artifact (#51).
func TestEngine_GreenServiceRunWritesNoLog(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	res := runSpecWithArtifacts(t, "svc.atago.yaml", `
version: "1"
suite:
  name: s
scenarios:
  - name: healthy service passes
    services:
      - name: ok
        shell: true
        command: '`+publishEchoIdle("up", "ready.txt", "serving", 5)+`'
        ready:
          file: ready.txt
          timeout: 2s
    steps:
      - run: {shell: true, command: "echo hi"}
      - assert:
          stdout: {contains: hi}
`, root)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed", res.Status)
	}
	if logs := res.Scenarios[0].ServiceLogs; len(logs) != 0 {
		t.Errorf("green run should write no service log, got %+v", logs)
	}
}

// TestEngine_MockRequestLogPreservedOnFailure closes the gap RequestLog's doc
// comment promised away: it billed itself as "the durable artifact written next
// to service logs when a scenario fails", but nothing ever called it — a failed
// scenario preserved the service's stdout while discarding the requests the
// mock observed, which is exactly the evidence (a typo'd path 404ing) the
// artifact dir exists to keep. On failure each mock server with at least one
// recorded request now writes its request log next to the service logs.
func TestEngine_MockRequestLogPreservedOnFailure(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	res := runSpecWithArtifacts(t, "mock.atago.yaml", `
version: "1"
suite:
  name: s
runners:
  api:
    type: http
    base_url: ${api.url}
scenarios:
  - name: the client hits a typo'd path
    mock_servers:
      - name: api
        routes:
          - method: GET
            path: /v1/ping
            status: 200
            body: pong
    steps:
      - http:
          runner: api
          method: GET
          path: /v1/pingg
      - assert:
          status: 200
`, root)
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed: %+v", res.Status, res.Scenarios)
	}
	sl := serviceLog(t, res, "api (mock requests)")
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(sl.Path)))
	if err != nil {
		t.Fatalf("read mock request log: %v", err)
	}
	if !strings.Contains(string(data), "GET /v1/pingg -> 404") {
		t.Errorf("mock request log = %q, want the 404'd typo'd request", data)
	}
}

// TestEngine_MockRequestLogSkippedWhenGreenOrSilent pins the two no-artifact
// cases: a passing scenario writes nothing (failure-gated like service logs),
// and a failing scenario whose mock recorded no request writes no empty file.
func TestEngine_MockRequestLogSkippedWhenGreenOrSilent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	res := runSpecWithArtifacts(t, "mock.atago.yaml", `
version: "1"
suite:
  name: s
runners:
  api:
    type: http
    base_url: ${api.url}
scenarios:
  - name: green round trip
    mock_servers:
      - name: api
        routes:
          - method: GET
            path: /v1/ping
            status: 200
            body: pong
    steps:
      - http:
          runner: api
          method: GET
          path: /v1/ping
      - assert:
          status: 200
  - name: failing but the mock saw nothing
    mock_servers:
      - name: quiet
        routes:
          - method: GET
            path: /v1/ping
            status: 200
            body: pong
    steps:
      - run: {shell: true, command: "exit 7"}
      - assert:
          exit_code: 0
`, root)
	if res.Scenarios[0].Status != StatusPassed || res.Scenarios[1].Status != StatusFailed {
		t.Fatalf("statuses = %s/%s, want passed/failed", res.Scenarios[0].Status, res.Scenarios[1].Status)
	}
	for i := range res.Scenarios {
		for _, sl := range res.Scenarios[i].ServiceLogs {
			if strings.Contains(sl.Name, "mock requests") {
				t.Errorf("scenario %d unexpectedly recorded a mock request log: %+v", i, sl)
			}
		}
	}
}
