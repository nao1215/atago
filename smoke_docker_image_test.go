package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSmokeDockerImage_AllowsVersionOnlyLocalTag(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	binDir := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(tmp, "docker-run-image.txt")
	dockerPath := filepath.Join(binDir, "docker")
	dockerStub := `#!/bin/sh
set -eu

case "${1:-}" in
image)
	if [ "${2:-}" = "ls" ]; then
		printf '%s\n' "${DOCKER_STUB_IMAGES:-}"
		exit 0
	fi
	;;
run)
	if [ "${2:-}" = "--rm" ]; then
		printf '%s\n' "${3:-}" >"${DOCKER_STUB_LOG:?}"
		printf '%s\n' "${DOCKER_STUB_VERSION_OUTPUT:-atago test}"
		exit 0
	fi
	;;
buildx)
	if [ "${2:-}" = "imagetools" ] && [ "${3:-}" = "inspect" ]; then
		exit 1
	fi
	;;
esac

echo "unexpected docker invocation: $*" >&2
exit 1
`
	if err := os.WriteFile(dockerPath, []byte(dockerStub), 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.CommandContext(context.Background(), "sh", "./scripts/smoke_docker_image.sh", "ghcr.io/nao1215/atago")
	cmd.Env = append(os.Environ(),
		"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"DOCKER_STUB_IMAGES=ghcr.io/nao1215/atago:v0.10.0-next",
		"DOCKER_STUB_VERSION_OUTPUT=atago v0.10.0-next",
		"DOCKER_STUB_LOG="+logPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("smoke_docker_image.sh failed: %v\n%s", err, out)
	}
	if got := strings.TrimSpace(string(out)); !strings.Contains(got, "docker-smoke: ghcr.io/nao1215/atago:v0.10.0-next runs atago v0.10.0-next") {
		t.Fatalf("stdout = %q, want success output for the version tag", got)
	}
	image, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(image)); got != "ghcr.io/nao1215/atago:v0.10.0-next" {
		t.Fatalf("docker run image = %q, want the only local tag", got)
	}
}
