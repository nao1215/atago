# Security policy

## Supported versions
Only the latest release of atago gets fixes, including security fixes. If you
hit an issue on an older version, please reproduce it on the latest release
first.

## Reporting a vulnerability
Report security issues privately, not through public issues or pull requests.

- Email: [n.chika156@gmail.com](mailto:n.chika156@gmail.com)
- Or use the "Report a vulnerability" button on the repository's Security tab.

atago is a CLI-oriented Go project, so reports about input parsing, path
handling, command execution, file handling, or resource exhaustion are
especially useful. Please include enough detail to reproduce:

- atago version (`atago -version`)
- OS and architecture
- The command you ran and what happened
- A minimal reproduction, if you have one

## What to expect
atago is maintained by one developer in spare time, so there is no guaranteed
response time. I will acknowledge the report, confirm the issue, and fix it in a
new release. You will be credited in the release notes unless you prefer to stay
anonymous.

## Verifying releases
Release artifacts are signed with cosign and ship with an SBOM and build
provenance. See [Verifying release integrity](./README.md#verifying-release-integrity)
for how to check what you download.
