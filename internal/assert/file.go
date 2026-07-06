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
func checkFile(f *spec.FileAssert, env Env) *CheckResult {
	path, err := security.ResolveWorkdirPath("assert.file.path", env.Workdir, f.Path)
	if err != nil {
		return &CheckResult{Desc: fmt.Sprintf("assert file %q", f.Path), Hint: err.Error()}
	}

	switch {
	case f.Exists != nil:
		desc := fmt.Sprintf("assert file %q exists: %t", f.Path, *f.Exists)
		_, err := os.Stat(path)
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

	case f.Contains != nil:
		data, cr := readFile(f.Path, path)
		if cr != nil {
			return cr
		}
		desc := fileContainsDesc(f.Path, f.Contains, true)
		if sub, idx, missing := firstMissing(string(data), f.Contains); missing {
			return &CheckResult{
				Desc:           desc,
				Expected:       fmt.Sprintf("file %q contains %q", f.Path, sub),
				Actual:         excerpt(string(data)),
				Hint:           fmt.Sprintf("the substring %q%s was not present in %q", sub, elementLabel(idx, len(f.Contains)), f.Path),
				ArtifactKind:   "file",
				ArtifactActual: data,
			}
		}
		return pass(desc)

	case f.NotContains != nil:
		data, cr := readFile(f.Path, path)
		if cr != nil {
			return cr
		}
		desc := fileContainsDesc(f.Path, f.NotContains, false)
		if sub, idx, present := firstPresent(string(data), f.NotContains); present {
			return &CheckResult{
				Desc:           desc,
				Expected:       fmt.Sprintf("file %q without %q", f.Path, sub),
				Actual:         excerpt(string(data)),
				Hint:           fmt.Sprintf("the substring %q%s was unexpectedly present in %q", sub, elementLabel(idx, len(f.NotContains)), f.Path),
				ArtifactKind:   "file",
				ArtifactActual: data,
			}
		}
		return pass(desc)

	case f.Executable != nil:
		info, statErr := os.Stat(path)
		if statErr != nil {
			return &CheckResult{
				Desc:     fmt.Sprintf("assert file %q executable: %t", f.Path, *f.Executable),
				Expected: fmt.Sprintf("readable file %q", f.Path),
				Actual:   statErr.Error(),
				Hint:     fmt.Sprintf("could not stat file %q", f.Path),
			}
		}
		isExec := info.Mode().Perm()&0o111 != 0
		desc := fmt.Sprintf("assert file %q executable: %t", f.Path, *f.Executable)
		if isExec == *f.Executable {
			return pass(desc)
		}
		return &CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("file %q executable=%t", f.Path, *f.Executable),
			Actual:   fmt.Sprintf("executable=%t (mode %s)", isExec, info.Mode().Perm()),
			Hint:     fmt.Sprintf("expected file %q to %s executable", f.Path, executability(*f.Executable)),
		}

	case f.Equals != nil:
		data, cr := readFile(f.Path, path)
		if cr != nil {
			return cr
		}
		// Byte-exact: no CRLF or trailing-newline normalization, unlike the stdout
		// equals matcher. A round-trip test needs to prove the bytes are identical.
		desc := fmt.Sprintf("assert file %q equals exact bytes", f.Path)
		if string(data) == *f.Equals {
			return pass(desc)
		}
		return &CheckResult{
			Desc:             desc,
			Expected:         excerpt(*f.Equals),
			Actual:           excerpt(string(data)),
			Hint:             fmt.Sprintf("file %q did not equal the expected bytes exactly (no CRLF/newline normalization)", f.Path),
			ArtifactKind:     "file",
			ArtifactActual:   data,
			ArtifactExpected: []byte(*f.Equals),
		}

	case f.EqualsFile != nil:
		data, cr := readFile(f.Path, path)
		if cr != nil {
			return cr
		}
		otherPath, err := security.ResolveWorkdirPath("assert.file.equals_file", env.Workdir, *f.EqualsFile)
		if err != nil {
			return &CheckResult{Desc: fmt.Sprintf("assert file %q equals_file %q", f.Path, *f.EqualsFile), Hint: err.Error()}
		}
		other, cr := readFile(*f.EqualsFile, otherPath)
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
			ArtifactKind:     "file",
			ArtifactActual:   data,
			ArtifactExpected: other,
		}

	case f.JSON != nil:
		data, cr := readFile(f.Path, path)
		if cr != nil {
			return cr
		}
		res := checkJSON(fmt.Sprintf("assert file %q json", f.Path), f.Path, data, f.JSON)
		if !res.OK && res.ArtifactKind == "" {
			res.ArtifactKind = "file"
			res.ArtifactActual = data
		}
		return res

	case f.Snapshot != "":
		data, cr := readFile(f.Path, path)
		if cr != nil {
			return cr
		}
		return checkSnapshot(fmt.Sprintf("assert file %q snapshot", f.Path), f.Path, f.Snapshot, data, env)

	default:
		return &CheckResult{Desc: "assert file", Hint: "file assertion must set exists/contains/not_contains/executable/equals/equals_file/json/snapshot"}
	}
}

func readFile(label, path string) ([]byte, *CheckResult) {
	// The program under test may have planted a symlink at the assertion target
	// pointing outside the workdir; reading through it would disclose an arbitrary
	// host file into the report/artifacts, so refuse to follow it (issue #16).
	data, err := security.ReadFileNoFollow(path)
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
