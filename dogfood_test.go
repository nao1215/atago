package main

import (
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/loader"
)

// dogfoodRoots are the directories whose real *.atago.yaml specs the project
// commits and runs.
var dogfoodRoots = []string{"test", "doc"}

// TestDogfood_SpecsLoad loads every committed spec and asserts it passes the
// loader's schema and semantic validation. This is the self-hosted acceptance
// check that the repo's own specs stay valid before a release rather than after
// users hit them.
func TestDogfood_SpecsLoad(t *testing.T) {
	var specs []string
	for _, root := range dogfoodRoots {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(path, ".atago.yaml") {
				specs = append(specs, path)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
	if len(specs) == 0 {
		t.Fatal("found no *.atago.yaml specs to dogfood")
	}

	for _, p := range specs {
		if _, err := loader.Load(p); err != nil {
			t.Errorf("%s: failed to load: %v", p, err)
		}
	}
}
