# Release process

This describes how a atago release is cut. It is for maintainers.

## Overview
Releases are driven by Git tags. Pushing a tag that matches `v*` triggers the
[release workflow](../.github/workflows/release.yml), which runs
[GoReleaser](https://goreleaser.com/) using [.goreleaser.yml](../.goreleaser.yml).
There is no manual upload step.

## Versioning
- atago follows [Semantic Versioning](https://semver.org/): `vMAJOR.MINOR.PATCH`.
- Release notes are generated from commit messages, so use
  [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`,
  `perf:`, `docs:`, and `!` for breaking changes). `chore:`, `ci:`, `style:`,
  and `test:` commits are excluded from the notes.

## Before tagging
- Make sure `main` is green (build, unit tests, coverage, lint, gitleaks, and release smoke).
- Locally you can dry-run the build with `goreleaser release --snapshot --clean`.

## Cut a release
```shell
git switch main
git pull --ff-only
git tag vX.Y.Z
git push origin vX.Y.Z
```

The release workflow then:

- builds binaries for Linux, macOS, and Windows (amd64 and arm64)
- publishes archives and `checksums.txt`
- signs the checksums with cosign (keyless) and attaches SBOMs
- attests build provenance via GitHub OIDC
- publishes a Homebrew cask to [nao1215/homebrew-tap](https://github.com/nao1215/homebrew-tap)

## Required secrets
- `GITHUB_TOKEN`: provided automatically; used to create the GitHub Release.
- `TAP_GITHUB_TOKEN`: a repo-scoped token for `nao1215/homebrew-tap`; used by
  GoReleaser to push the Homebrew cask on a tagged release. The push-time
  release smoke skips publishing (`--skip=publish`), so only real tag releases
  need it.

## After releasing
- Check the [Releases page](https://github.com/nao1215/atago/releases) for the
  generated notes and artifacts.
- Verify a downloaded artifact as described in
  [Verifying release integrity](https://nao1215.github.io/atago/install/#verifying-release-integrity).

## If a release fails
- Re-run the failed job from the Actions tab once the cause is fixed.
- If the tag itself is wrong, delete it locally and remotely, then tag again:
  ```shell
  git tag -d vX.Y.Z
  git push origin :refs/tags/vX.Y.Z
  ```
