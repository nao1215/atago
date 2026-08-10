package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/loader"
	runnercmd "github.com/nao1215/atago/internal/runner/cmd"
	"github.com/nao1215/atago/internal/spec"
)

// defaultBuildTimeout bounds a subject build when the manifest sets none. A
// build that hangs has to fail the run rather than the CI job's wall clock.
const defaultBuildTimeout = 10 * time.Minute

// builtSubject is what a completed build hands to the run: where the binary
// landed, and the environment the profile asked for.
type builtSubject struct {
	Name     string
	Artifact string
	Env      map[string]string
}

// buildSubjects builds the binary under test declared by each distinct manifest
// covering the given spec paths, ONCE per `atago run` invocation (#393).
//
// Once per invocation is the whole point. The cookbook's older recipe — a
// suite.setup step building into ${suitedir} — rebuilds per spec FILE, which for
// sqly's 89 specs means 89 builds, and it never puts the binary on PATH, so
// every spec has to spell an absolute path instead of the name a user types.
// That is why all nine downstream repos build in bash instead.
//
// A build failure is a run-level error: no scenario executes, because every one
// of them would be testing either a stale binary or nothing at all.
func buildSubjects(ctx context.Context, paths []string, profile string, scratch string, stderr io.Writer) ([]builtSubject, error) {
	seen := map[string]bool{}
	var built []builtSubject
	for _, p := range paths {
		proj, err := loader.FindProject(filepath.Dir(p))
		if err != nil {
			return nil, err
		}
		if proj == nil || proj.Subject == nil {
			continue
		}
		// One manifest can cover hundreds of specs; build its subject once.
		if seen[proj.Path] {
			continue
		}
		seen[proj.Path] = true
		b, err := buildOne(ctx, proj, profile, scratch, stderr)
		if err != nil {
			return nil, err
		}
		built = append(built, b)
	}
	if profile != "" && len(built) == 0 {
		return nil, fmt.Errorf("--profile %q was given, but no manifest above the given specs declares a subject to build", profile)
	}
	return built, nil
}

func buildOne(ctx context.Context, proj *loader.Project, profile, scratch string, stderr io.Writer) (builtSubject, error) {
	sub := proj.Subject
	build := sub.Build
	env := map[string]string{}
	if profile != "" {
		prof, ok := proj.Profiles[profile]
		if !ok {
			return builtSubject{}, fmt.Errorf("%s declares no profile %q (it has: %s)",
				proj.Path, profile, profileNames(proj.Profiles))
		}
		if prof.Build != nil {
			// Whole-command replacement, not a merge: a half-merged build
			// command is unreadable, and "which flags survived" is not a
			// question a spec author should have to answer.
			build = prof.Build
		}
		for k, v := range prof.Env {
			env[k] = v
		}
	}

	artifact := filepath.Join(scratch, filepath.FromSlash(sub.Artifact))
	// The per-OS suffix jose hand-rolls in bash: without it a Windows build
	// writes mytool.exe while atago looks for mytool and reports a missing
	// artifact for a build that succeeded.
	if runtime.GOOS == "windows" && filepath.Ext(artifact) == "" {
		artifact += ".exe"
	}
	if err := os.MkdirAll(filepath.Dir(artifact), 0o750); err != nil {
		return builtSubject{}, fmt.Errorf("preparing the build output directory: %w", err)
	}

	dir := filepath.Dir(proj.Path)
	if build.Cwd != "" {
		dir = runnercmd.ResolveDir(dir, build.Cwd)
	}
	command := strings.ReplaceAll(build.Command, "${artifact}", artifact)

	timeout := defaultBuildTimeout
	if build.Timeout != "" {
		if d, err := time.ParseDuration(build.Timeout); err == nil && d > 0 {
			timeout = d
		}
	}
	bctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	fmt.Fprintf(stderr, "atago: building %s (%s)\n", sub.Name, command)
	res, err := runnercmd.New().Run(bctx, &spec.Run{
		Command: command,
		Shell:   build.Shell,
		Env:     env,
	}, dir)
	switch {
	case err != nil:
		return builtSubject{}, fmt.Errorf("building %s: %w", sub.Name, err)
	case res.TimedOut:
		return builtSubject{}, fmt.Errorf("building %s timed out after %s", sub.Name, timeout)
	case res.ExitCode != 0:
		// The compiler's own output IS the report: anything else would make the
		// author re-run the build by hand to see what went wrong.
		return builtSubject{}, fmt.Errorf("building %s failed (exit %d)\n%s%s",
			sub.Name, res.ExitCode, res.Stderr, res.Stdout)
	}
	if st, serr := os.Stat(artifact); serr != nil || st.IsDir() {
		return builtSubject{}, fmt.Errorf(
			"building %s succeeded but produced no file at %s: check that the build command writes to ${artifact}",
			sub.Name, artifact)
	}
	return builtSubject{Name: sub.Name, Artifact: artifact, Env: env}, nil
}

func profileNames(profiles map[string]loader.Profile) string {
	if len(profiles) == 0 {
		return "none"
	}
	names := make([]string, 0, len(profiles))
	for k := range profiles {
		names = append(names, k)
	}
	return strings.Join(spec.SortedKeys(toSet(names)), ", ")
}

func toSet(names []string) map[string]bool {
	out := make(map[string]bool, len(names))
	for _, n := range names {
		out[n] = true
	}
	return out
}

// exposeSubjects prepends each artifact's directory to PATH so a bare binary
// name in a spec resolves to the freshly built one, and returns the environment
// the active profile asked for.
//
// Prepending to the process PATH is exactly what every wrapper script does
// (`export PATH="$TMP/bin:$PATH"`), and it is the only way a spec can go on
// saying `mytool convert ...` rather than an absolute path — which is the point
// of testing the binary the way a user invokes it.
func exposeSubjects(built []builtSubject) error {
	for _, b := range built {
		dir := filepath.Dir(b.Artifact)
		if err := os.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH")); err != nil {
			return err
		}
		for k, v := range b.Env {
			if err := os.Setenv(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}
