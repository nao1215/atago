## Contributing to atago
Thank you for building atago with us.
Every report, patch, test, and review helps turn an early repository into a
reliable project. Let's keep the code, docs, and release process clear and
predictable together.

## Contributing as a Developer
### 1. Start with clear communication
- Bug report: Use the issue template and include reproducible steps, expected behavior, and actual behavior.
- New feature: Open an issue first so we can agree on direction before implementation.
- Bug fix or improvement: Open a PR with a clear problem statement and solution summary.

### 2. Keep the quality bar high
- Add or update unit tests when you add features or fix bugs.
- Avoid regressions on supported OSes (Linux, macOS, Windows).
- Keep CLI behavior, file handling, and error messages clear and consistent.

### 3. Run checks before opening a PR
```shell
make build
make test
make lint
make e2e
```

`make vet` runs `go vet` and `make fmt` formats the Go sources. `make test` also
generates `cover.out` and `cover.html` locally.

`make e2e` builds the binary and runs atago against its own self-hosted specs in
`test/e2e/atago/` plus the git suite in `test/e2e/thirdparty/git/`. atago is also
dogfooded against real CLIs with `make dogfood` and the per-tool
`make dogfood-<tool>` targets; bugs found while dogfooding become regression
tests. The README demos are recorded with
[charmbracelet/vhs](https://github.com/charmbracelet/vhs) from `doc/vhs/*.tape`
(regenerate with `make demo`).

When you change a spec under `test/e2e/`, regenerate the committed behavior
docs with `make docs` (guarded by `TestDocs_E2EDocsInSync`) and the site with
`make site`. The specs under `examples/` are linked from the README and tested
by `examples_test.go`: a new example must be registered there as hermetic
(executed in CI) or validate-only.

### 4. Install developer tools
```shell
make tools
```

### 5. Update documentation when behavior changes
- Keep `README.md` accurate for user-facing behavior.
- Update `CHANGELOG.md` under `## [Unreleased]` for changes worth calling out in releases.
- If the release process changes, update [doc/RELEASE.md](./doc/RELEASE.md).

## Releasing
Maintainers cut releases by pushing a `v*` tag. The workflow and follow-up
checks are documented in [doc/RELEASE.md](./doc/RELEASE.md).

## Need help?
See [SUPPORT.md](./.github/SUPPORT.md) for where to ask questions and report
problems.

## Contributing Outside of Coding
You can still help a lot even if you are not writing code:

- Give atago a GitHub Star
- Open issues with clear reproduction steps
- Review pull requests
- Sponsor the project
