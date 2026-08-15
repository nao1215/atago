package assert

import (
	"bytes"
	"fmt"
	"os"

	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
)

// checkFile evaluates a file assertion. Relative paths resolve
// against the scenario workdir and may not escape it.
func checkFile(f *spec.FileAssert, env Env) (out *CheckResult) {
	path, err := security.ResolveWorkdirPath("assert.file.path", env.Workdir, f.Path)
	if err != nil {
		return &CheckResult{Desc: fmt.Sprintf("assert file %q", f.Path), Hint: err.Error()}
	}

	// Attach the compared file bytes to any failing result that did not shape its
	// own artifact, so --artifacts-dir can persist it (#48). This mirrors
	// checkStream's single deferred hook, so a new file matcher inherits the
	// attachment instead of each branch remembering it by hand (#247). A branch
	// that reads the file records its bytes via read(); exists/executable never
	// read and leave fileData nil (they attach nothing); checkSnapshot shapes its
	// own artifact and is left untouched (it reads via readFile, not read).
	var fileData []byte
	defer func() {
		if out != nil && !out.OK && out.ArtifactKind == "" && fileData != nil {
			out.ArtifactKind = "file"
			if out.ArtifactActual == nil {
				out.ArtifactActual = fileData
			}
		}
	}()
	read := func(label, p string) ([]byte, *CheckResult) {
		data, cr := readFile(label, env.Workdir, p)
		if cr == nil {
			fileData = data
		}
		return data, cr
	}

	// Size bounds compose with the content matchers instead of being one of them
	// (#397): they answer a different question, and a spec routinely wants both
	// ("the export parses as JSON AND stays under a megabyte"). They are checked
	// first because a wrong size explains a content mismatch better than the
	// other way round, and when nothing else is set they are the whole assertion.
	if f.HasSize() {
		if cr := checkFileSize(f, env.Workdir, path); !cr.OK || !f.HasContentMatcher() {
			return cr
		}
	}

	// A count bound turns `contains` into an occurrence check (#396), which
	// subsumes presence. The loader has already required a single-element
	// contains next to it.
	if f.HasCount() {
		data, cr := read(f.Path, path)
		if cr != nil {
			return cr
		}
		return checkOccurrences(fmt.Sprintf("file %q", f.Path), string(data),
			occurrenceTarget{Literal: f.Contains[0]},
			countBounds{Count: f.Count, Min: f.MinCount, Max: f.MaxCount})
	}

	switch {
	case f.Exists != nil:
		return checkFileExists(f, path)

	case f.Contains != nil:
		data, cr := read(f.Path, path)
		if cr != nil {
			return cr
		}
		return checkFileContains(f, data)

	case f.NotContains != nil:
		data, cr := read(f.Path, path)
		if cr != nil {
			return cr
		}
		return checkFileNotContains(f, data)

	case f.Executable != nil:
		return checkFileExecutable(f, path)

	case f.Equals != nil:
		data, cr := read(f.Path, path)
		if cr != nil {
			return cr
		}
		return checkFileEquals(f, data)

	case f.EqualsFile != nil:
		data, cr := read(f.Path, path)
		if cr != nil {
			return cr
		}
		return checkFileEqualsFile(f, data, env)

	case len(f.JSON) > 0:
		data, cr := read(f.Path, path)
		if cr != nil {
			return cr
		}
		// A failing json check inherits the "file" artifact from the deferred hook
		// unless checkJSONChecks shaped its own.
		return checkJSONChecks(fmt.Sprintf("assert file %q json", f.Path), f.Path, data, f.JSON, false)

	case f.Snapshot != "":
		// Read plainly so fileData stays nil: checkSnapshot shapes its own
		// (normalized) artifact, and the deferred hook must not overwrite it.
		data, cr := readFile(f.Path, env.Workdir, path)
		if cr != nil {
			return cr
		}
		return checkSnapshot(fmt.Sprintf("assert file %q snapshot", f.Path), f.Path, f.Snapshot, data, env)

	default:
		return &CheckResult{Desc: "assert file", Hint: "file assertion must set exists/contains/not_contains/executable/equals/equals_file/json/snapshot"}
	}
}

// checkFileExists evaluates exists:true/false against a stat of path.
func checkFileExists(f *spec.FileAssert, path string) *CheckResult {
	desc := fmt.Sprintf("assert file %q exists: %t", f.Path, *f.Exists)
	info, err := os.Stat(path)
	// Only a genuine "not exist" result participates in exists:true/false.
	// Permission, I/O, and other stat failures are surfaced as an error so
	// users do not debug a "missing file" that is really unreadable (#39).
	if err != nil && !os.IsNotExist(err) {
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("stat-able file %q", f.Path),
			Actual:   err.Error(),
			Hint:     fmt.Sprintf("could not stat file %q: %v", f.Path, err),
		}
	}
	// A directory is not the file this assertion is about. Counting it as one
	// made `exists: true` pass for a tool that produced a directory where a
	// file was expected — the assertion's whole job is to catch that.
	if err == nil && info.IsDir() {
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("file %q exists=%t", f.Path, *f.Exists),
			Actual:   fmt.Sprintf("%q is a directory", f.Path),
			Hint:     fmt.Sprintf("%q exists but is a directory, not a file; use a dir: assertion for it", f.Path),
		}
	}
	exists := err == nil
	if exists == *f.Exists {
		return pass(desc)
	}
	return &CheckResult{
		Desc:     desc,
		Expected: fmt.Sprintf("file %q exists=%t", f.Path, *f.Exists),
		Actual:   fmt.Sprintf("exists=%t", exists),
		Hint:     fmt.Sprintf("expected file %q to %s", f.Path, existence(*f.Exists)),
	}
}

// checkFileContains requires every listed substring to appear in data.
func checkFileContains(f *spec.FileAssert, data []byte) *CheckResult {
	desc := fileContainsDesc(f.Path, f.Contains, true)
	if sub, idx, missing := firstMissing(string(data), f.Contains); missing {
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("file %q contains %q", f.Path, sub),
			Actual:   excerpt(string(data)),
			Hint:     fmt.Sprintf("the substring %q%s was not present in %q", sub, elementLabel(idx, len(f.Contains)), f.Path),
		}
	}
	return pass(desc)
}

// checkFileNotContains requires every listed substring to be absent from data.
func checkFileNotContains(f *spec.FileAssert, data []byte) *CheckResult {
	desc := fileContainsDesc(f.Path, f.NotContains, false)
	if sub, idx, present := firstPresent(string(data), f.NotContains); present {
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("file %q without %q", f.Path, sub),
			Actual:   excerpt(string(data)),
			Hint:     fmt.Sprintf("the substring %q%s was unexpectedly present in %q", sub, elementLabel(idx, len(f.NotContains)), f.Path),
		}
	}
	return pass(desc)
}

// checkFileExecutable evaluates executable:true/false against path's mode bits.
func checkFileExecutable(f *spec.FileAssert, path string) *CheckResult {
	desc := fmt.Sprintf("assert file %q executable: %t", f.Path, *f.Executable)
	info, statErr := os.Stat(path)
	if statErr != nil {
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("readable file %q", f.Path),
			Actual:   statErr.Error(),
			Hint:     fmt.Sprintf("could not stat file %q", f.Path),
		}
	}
	// Every directory carries the execute bit, so reading it as "this is an
	// executable" turned a tool that produced a directory into a passing
	// executable check.
	if info.IsDir() {
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("file %q executable=%t", f.Path, *f.Executable),
			Actual:   fmt.Sprintf("%q is a directory", f.Path),
			Hint:     fmt.Sprintf("%q is a directory, not an executable file; a directory's execute bit means it can be entered", f.Path),
		}
	}
	isExec := info.Mode().Perm()&0o111 != 0
	if isExec == *f.Executable {
		return pass(desc)
	}
	return &CheckResult{
		Desc:     desc,
		Expected: fmt.Sprintf("file %q executable=%t", f.Path, *f.Executable),
		Actual:   fmt.Sprintf("executable=%t (mode %s)", isExec, info.Mode().Perm()),
		Hint:     fmt.Sprintf("expected file %q to %s executable", f.Path, executability(*f.Executable)),
	}
}

// checkFileEquals compares data to the expected literal byte-exactly: no CRLF
// or trailing-newline normalization, unlike the stdout equals matcher. A
// round-trip test needs to prove the bytes are identical.
func checkFileEquals(f *spec.FileAssert, data []byte) *CheckResult {
	desc := fmt.Sprintf("assert file %q equals exact bytes", f.Path)
	if string(data) == *f.Equals {
		return pass(desc)
	}
	return &CheckResult{
		Desc:             desc,
		Expected:         excerpt(*f.Equals),
		Actual:           excerpt(string(data)),
		Hint:             fmt.Sprintf("file %q did not equal the expected bytes exactly (no CRLF/newline normalization)", f.Path),
		ArtifactExpected: []byte(*f.Equals),
	}
}

// checkFileEqualsFile compares data byte-exactly to another workdir file.
func checkFileEqualsFile(f *spec.FileAssert, data []byte, env Env) *CheckResult {
	otherPath, err := security.ResolveWorkdirPath("assert.file.equals_file", env.Workdir, *f.EqualsFile)
	if err != nil {
		return &CheckResult{Desc: fmt.Sprintf("assert file %q equals_file %q", f.Path, *f.EqualsFile), Hint: err.Error()}
	}
	// The comparison file is read plainly (not via checkFile's read): the failure
	// artifact is the file under test, and the other file is carried as
	// ArtifactExpected.
	other, cr := readFile(*f.EqualsFile, env.Workdir, otherPath)
	if cr != nil {
		return cr
	}
	desc := fmt.Sprintf("assert file %q equals file %q", f.Path, *f.EqualsFile)
	if bytes.Equal(data, other) {
		return pass(desc)
	}
	return &CheckResult{
		Desc:             desc,
		Expected:         fmt.Sprintf("bytes identical to %q", *f.EqualsFile),
		Actual:           excerpt(string(data)),
		Hint:             fmt.Sprintf("file %q is not byte-identical to %q (no CRLF/newline normalization)", f.Path, *f.EqualsFile),
		ArtifactExpected: other,
	}
}

func readFile(label, root, path string) ([]byte, *CheckResult) {
	// The program under test may have planted a symlink at the assertion target
	// pointing outside the workdir; reading through it would disclose an arbitrary
	// host file into the report/artifacts, so refuse to follow it (issue #16), and
	// bind the read to the workdir so an ancestor swapped for a link cannot
	// redirect it either (issue #430).
	data, err := security.ReadFileNoFollow(root, path)
	if err != nil {
		return nil, &CheckResult{
			Desc:     fmt.Sprintf("assert file %q", label),
			Expected: fmt.Sprintf("readable file %q", label),
			Actual:   err.Error(),
			Hint:     fmt.Sprintf("could not read file %q", label),
		}
	}
	return data, nil
}

func existence(want bool) string {
	if want {
		return "exist"
	}
	return "not exist"
}

func executability(want bool) string {
	if want {
		return "be"
	}
	return "not be"
}
