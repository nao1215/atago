#!/bin/sh
# Combine unit-test coverage with self-hosted E2E coverage into a single
# cover.out. Unit tests report line coverage, but they never exercise the real
# atago binary the way an end user does; the self-hosted E2E specs do. Go 1.20+
# lets us instrument a built binary (`go build -cover`) and collect its runtime
# coverage via GOCOVERDIR, so we can merge "what the tests cover" with "what a
# real run covers" and get one honest number.
#
# This is intentionally a separate, heavier target: `make test` / `make e2e`
# stay fast. Everything lands under .coverage/ (gitignored) except the final
# cover.out / cover.html, which are the same artifacts `make test` already
# produces so octocov and local tooling need no changes.
#
# Override scenario concurrency with PARALLEL (defaults to 8, matching `make e2e`).
set -eu

cd "$(CDPATH= cd "$(dirname "$0")/.." && pwd)"
root="$(pwd)"
cov="${root}/.coverage"
parallel="${PARALLEL:-8}"

rm -rf "${cov}"
mkdir -p "${cov}/unit" "${cov}/e2e" "${cov}/bin" "${cov}/merged"

# 1. Unit-test coverage as raw covdata (GOCOVERDIR form) so it can be merged
#    with the E2E covdata below. -covermode=atomic must match the binary build.
echo ">> unit coverage -> ${cov}/unit"
go test -count=1 -cover -covermode=atomic -coverpkg=./... ./... \
	-args -test.gocoverdir="${cov}/unit"

# 2. Coverage-instrumented atago binary. E2E must run THIS, not dist/atago.
echo ">> building coverage-instrumented atago -> ${cov}/bin/atago-cover"
go build -cover -covermode=atomic -coverpkg=./... -o "${cov}/bin/atago-cover" .

# 3. Self-hosted E2E via the instrumented binary. ${atago} inside the specs is
#    os.Executable(), i.e. this same cover binary, so every atago-on-atago child
#    is also instrumented. GOCOVERDIR is inherited by those children (the specs
#    that invoke ${atago} do not use clear_env), so each writes its own covdata.
echo ">> e2e coverage -> ${cov}/e2e"
GOCOVERDIR="${cov}/e2e" \
	"${cov}/bin/atago-cover" run --parallel "${parallel}" \
	./test/e2e/atago ./test/e2e/thirdparty/git

# 4. Merge the raw covdata and render the combined text profile + reports.
echo ">> merging unit + e2e covdata -> cover.out"
go tool covdata merge -i="${cov}/unit,${cov}/e2e" -o="${cov}/merged"
go tool covdata textfmt -i="${cov}/merged" -o="${root}/cover.out"

go tool cover -func=cover.out | tail -n 1
go tool cover -html=cover.out -o cover.html
echo ">> wrote cover.out and cover.html (unit + e2e combined)"
