package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestSaveRerunState_Concurrent(t *testing.T) {
	withWorkdir(t, t.TempDir(), func() {
		const writers = 16
		ledgers := make([][]failedEntry, writers)
		valid := make(map[string]bool, writers)
		for i := range ledgers {
			ledgers[i] = []failedEntry{{SpecPath: "suite.atago.yaml", Scenario: fmt.Sprintf("%d-%s", i, strings.Repeat("x", i*4096))}}
			data, err := json.MarshalIndent(rerunState{SchemaVersion: RerunStateSchemaVersion, Failed: ledgers[i]}, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			valid[string(data)+"\n"] = true
		}
		if err := saveRerunState(ledgers[0]); err != nil {
			t.Fatal(err)
		}
		check := func() {
			data, err := readRerunStateFile()
			if err != nil {
				t.Error(err)
				return
			}
			var state rerunState
			// Unmarshal consumes the whole document and rejects trailing bytes.
			if err := json.Unmarshal(data, &state); err != nil {
				t.Errorf("invalid ledger (%d bytes): %v", len(data), err)
				return
			}
			if !valid[string(data)] {
				t.Errorf("ledger (%d bytes) does not match any complete submitted state", len(data))
			}
		}
		var wg sync.WaitGroup
		var succeeded, tolerated atomic.Int64
		start := make(chan struct{})
		for _, ledger := range ledgers {
			wg.Go(func() {
				<-start
				for range 40 {
					if err := saveRerunState(ledger); err != nil {
						if !isTransientRerunIOError(err) {
							t.Error(err)
							return
						}
						// Best-effort saves may exhaust retries under Windows contention.
						tolerated.Add(1)
					} else {
						succeeded.Add(1)
					}
					check()
				}
			})
		}
		close(start)
		wg.Wait()
		t.Logf("successful saves: %d; tolerated transient failures: %d", succeeded.Load(), tolerated.Load())
		if succeeded.Load() == 0 {
			t.Error("no concurrent save succeeded")
		}
		check()
	})
}

func TestSaveRerunState_TempCleanup(t *testing.T) {
	for _, renameFailure := range []bool{false, true} {
		t.Run(fmt.Sprintf("renameFailure=%t", renameFailure), func(t *testing.T) {
			withWorkdir(t, t.TempDir(), func() {
				if renameFailure {
					if err := os.MkdirAll(rerunStatePath(), 0o750); err != nil {
						t.Fatal(err)
					}
				}
				err := saveRerunState([]failedEntry{{SpecPath: "a", Scenario: "b"}})
				if (err != nil) != renameFailure {
					t.Fatalf("save error = %v, want failure=%t", err, renameFailure)
				}
				entries, err := os.ReadDir(rerunStateDir)
				if err != nil {
					t.Fatal(err)
				}
				if len(entries) != 1 || entries[0].Name() != rerunStateFile {
					t.Fatalf("state directory contains leftover temporary files: %v", entries)
				}
				if !renameFailure && runtime.GOOS != "windows" {
					info, err := os.Stat(rerunStatePath())
					if err != nil {
						t.Fatal(err)
					}
					if info.Mode().Perm() != 0o600 {
						t.Errorf("mode = %o, want 600", info.Mode().Perm())
					}
				}
			})
		})
	}
}

func TestIsTransientRerunIOError(t *testing.T) {
	t.Parallel()
	// Win32 codes: ERROR_ACCESS_DENIED=5, ERROR_SHARING_VIOLATION=32.
	// The same numeric errors must never trigger retries on other platforms.
	for _, code := range []syscall.Errno{5, 32, 2, 3, 6, 33, 112} {
		want := runtime.GOOS == "windows" && (code == 5 || code == 32)
		for _, err := range []error{
			code,
			&os.PathError{Op: "open", Path: "ledger", Err: code},
			&os.LinkError{Op: "rename", Old: "temp", New: "ledger", Err: code},
			fmt.Errorf("wrapped: %w", code),
		} {
			if got := isTransientRerunIOError(err); got != want {
				t.Errorf("isTransientRerunIOError(%v) = %t, want %t", err, got, want)
			}
		}
	}
	for _, err := range []error{nil, errors.New("Access is denied."), os.ErrPermission} {
		if isTransientRerunIOError(err) {
			t.Errorf("unexpected retry for %v", err)
		}
	}
}

func TestRetryRerunIO(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("permanent failure")
	for _, result := range []error{nil, sentinel, syscall.Errno(32)} {
		calls := 0
		err := retryRerunIO(func() error {
			calls++
			if calls == 1 {
				return result
			}
			return nil
		})
		wantCalls, wantErr := 1, result
		if runtime.GOOS == "windows" && errors.Is(result, syscall.Errno(32)) {
			wantCalls, wantErr = 2, nil
		}
		if calls != wantCalls || !errors.Is(err, wantErr) {
			t.Errorf("result %v: calls=%d, err=%v; want %d, %v", result, calls, err, wantCalls, wantErr)
		}
	}
}

func TestRetryRerunIO_PersistentConflict(t *testing.T) {
	t.Parallel()
	conflict := &os.LinkError{Op: "rename", Old: "temp", New: "ledger", Err: syscall.Errno(5)}
	calls := 0
	start := time.Now()
	err := retryRerunIO(func() error {
		calls++
		return conflict
	})
	if !errors.Is(err, conflict) {
		t.Fatalf("error = %v, want original conflict", err)
	}
	if runtime.GOOS != "windows" {
		if calls != 1 {
			t.Fatalf("retried %d times on %s", calls, runtime.GOOS)
		}
		return
	}
	if elapsed := time.Since(start); elapsed < time.Second || elapsed > 5*time.Second {
		t.Errorf("retry duration = %s, want about one second", elapsed)
	}
	if calls < 2 || calls > 100 {
		t.Errorf("operation called %d times, want bounded retries with backoff", calls)
	}
}
