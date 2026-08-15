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
- pushes the winget manifests to `nao1215/winget-pkgs` and opens the pull request against [microsoft/winget-pkgs](https://github.com/microsoft/winget-pkgs), on a stable tag only

## Required secrets
- `GITHUB_TOKEN`: provided automatically; used to create the GitHub Release.
- `TAP_GITHUB_TOKEN`: a repo-scoped token for `nao1215/homebrew-tap`; used by
  GoReleaser to push the Homebrew cask on a tagged release. The push-time
  release smoke skips publishing (`--skip=publish`), so only real tag releases
  need it.
- `WINGET_GITHUB_TOKEN`: a **classic** personal access token with the `public_repo` scope. GoReleaser v2.16.0 uses this one token for both writes: committing the manifests to a branch on `nao1215/winget-pkgs` and opening the pull request against microsoft/winget-pkgs. A fine-grained token cannot do the second — it can only be scoped to repositories you own, and microsoft/winget-pkgs is not one of them, so the push succeeds and the pull request fails with 403. Same publish-time-only rule as the tap token.

## The winget pull request
A **stable** tagged release opens a pull request on microsoft/winget-pkgs; a moderator merges it, usually within a day, once their validation pipeline passes. A pre-release tag opens nothing: `skip_upload: auto` skips the winget pipe whenever the tag carries a pre-release suffix, because the community repository takes stable versions only.

Nothing in the release depends on that pull request, so a rejected or delayed one never blocks a version — GoReleaser logs the failure and finishes the release. Watch the [pull requests authored by nao1215](https://github.com/microsoft/winget-pkgs/pulls/nao1215) after the first submission of a new package, which gets more review than an update to an existing one.

If a release finishes but no pull request appears, the manifests were still generated and the recovery is manual: `dist/winget/manifests/n/nao1215/atago/<version>/` holds the three files, and they can be committed to a `atago-<version>` branch on the fork and submitted by hand. A failure at the push step points at the token's scope; a failure only at the pull-request step points at a fine-grained token.

## After releasing
- Check the [Releases page](https://github.com/nao1215/atago/releases) for the
  generated notes and artifacts.
- Nothing to do for the website: the release workflow republishes it from the
  tag, deriving the reference's Since column from the schema at every tag, so
  the keys the release introduced stop reading `unreleased`. Refresh the
  committed copy (`website/data/spec_keys.json`, which is what a local
  `hugo server` reads) whenever convenient:
  ```shell
  python3 website/tools/gen-spec-keys.py
  ```
- Verify a downloaded artifact as described in
  [Verifying release integrity](https://nao1215.github.io/atago/install/#verifying-release-integrity).

## If a release fails
- Re-run the failed job from the Actions tab once the cause is fixed.
- If the tag itself is wrong, delete it locally and remotely, then tag again:
  ```shell
  git tag -d vX.Y.Z
  git push origin :refs/tags/vX.Y.Z
  ```
