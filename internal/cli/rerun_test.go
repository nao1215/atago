package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
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
			data, err := os.ReadFile(rerunStatePath())
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
		start := make(chan struct{})
		for _, ledger := range ledgers {
			wg.Go(func() {
				<-start
				for range 40 {
					if err := saveRerunState(ledger); err != nil {
						t.Error(err)
						return
					}
					check()
				}
			})
		}
		close(start)
		wg.Wait()
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
