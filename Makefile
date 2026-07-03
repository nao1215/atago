.PHONY: build test clean vet fmt lint tools release-smoke e2e thirdparty dogfood dogfood-iso8583tool dogfood-jose dogfood-career dogfood-gup dogfood-mimixbox dogfood-mobilepkg demo docs site help

APP         = atago
VERSION     = $(shell git describe --tags --always --dirty 2>/dev/null)
# Scenario concurrency for the spawn/IO-bound suites (atago-on-atago, sqly).
# Override with `make e2e PARALLEL=1`. NOT used for dogfood-gup: its scenarios
# each run a CPU-bound `go install` that already parallelizes compilation, so
# concurrency oversubscribes the CPU and is slower than serial.
PARALLEL    = 8
GO          = go
GO_BUILD    = $(GO) build
GO_FORMAT   = $(GO) fmt
GO_INSTALL  = $(GO) install
GO_LIST     = $(GO) list
GO_TEST     = $(GO) test -v
GO_TOOL     = $(GO) tool
GO_VET      = $(GO) vet
GOOS        = ""
GOARCH      = ""
GO_PKGROOT  = ./...
GO_PACKAGES = $(shell $(GO_LIST) $(GO_PKGROOT))
GO_LDFLAGS  = -ldflags '-X github.com/nao1215/atago/internal/buildinfo.Version=$(VERSION)'

build: ## Build binary
	env GO111MODULE=on CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO_BUILD) $(GO_LDFLAGS) -o $(APP) .

clean: ## Clean project artifacts
	-rm -rf $(APP) cover.out cover.html dist

test: ## Run tests with coverage output
	env GOOS=$(GOOS) $(GO_TEST) -cover -coverpkg=./... -coverprofile=cover.out $(GO_PKGROOT)
	$(GO_TOOL) cover -html=cover.out -o cover.html

vet: ## Run go vet
	$(GO_VET) $(GO_PACKAGES)

fmt: ## Format Go source code
	$(GO_FORMAT) $(GO_PKGROOT)

lint: ## Run golangci-lint
	golangci-lint run --config .golangci.yml

tools: ## Install developer tools used by this repository
	$(GO_INSTALL) github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2

release-smoke: ## Build release artifacts locally and smoke-test them (requires goreleaser; syft adds the SBOM check)
	@if command -v syft >/dev/null; then \
		goreleaser release --snapshot --clean --skip=publish,sign && \
		./scripts/smoke_artifacts.sh dist; \
	else \
		echo "syft not found; skipping SBOM generation (CI still enforces it)"; \
		goreleaser release --snapshot --clean --skip=publish,sign,sbom && \
		SMOKE_SKIP_SBOM=1 ./scripts/smoke_artifacts.sh dist; \
	fi

e2e: ## Build the binary and run the hermetic self-hosted E2E specs (atago tested by atago)
	env CGO_ENABLED=0 $(GO_BUILD) $(GO_LDFLAGS) -o ./dist/$(APP) .
	./dist/$(APP) run --parallel $(PARALLEL) ./test/e2e/atago ./test/e2e/thirdparty/git

thirdparty: ## Run atago against third-party programs (needs git, caddy, pushgateway, webhook, restic, rclone, minio+mc, prometheus+promtool, gitea, coredns+dig, nats-server+nats, mailpit, ntfy on PATH)
	env CGO_ENABLED=0 $(GO_BUILD) $(GO_LDFLAGS) -o ./dist/$(APP) .
	./dist/$(APP) run ./test/e2e/thirdparty

dogfood: ## Run atago against real nao1215 CLIs that just need a binary on PATH (gup, sqly, truss)
	env CGO_ENABLED=0 $(GO_BUILD) $(GO_LDFLAGS) -o ./dist/$(APP) .
	./dist/$(APP) run --parallel $(PARALLEL) ./test/e2e/tools/gup ./test/e2e/tools/sqly ./test/e2e/tools/truss

dogfood-iso8583tool: ## Full iso8583tool e2e (builds latest iso8583tool + its TCP mock; set ISO_REPO)
	bash ./test/e2e/tools/iso8583tool/run.sh --parallel $(PARALLEL)

dogfood-jose: ## Full jose e2e (builds latest jose with GOEXPERIMENT=jsonv2; set JOSE_REPO)
	bash ./test/e2e/tools/jose/run.sh --parallel $(PARALLEL)

dogfood-career: ## Full career e2e (builds latest career; set CAREER_REPO)
	bash ./test/e2e/tools/career/run.sh --parallel $(PARALLEL)

dogfood-gup: ## Full offline gup e2e (builds gup's in-repo module proxy; set GUP_REPO)
	# Runs serially on purpose: each scenario's `go install` is CPU-bound and
	# already parallelizes, so --parallel would oversubscribe and slow it down.
	bash ./test/e2e/tools/gup-offline/run.sh --parallel 1

dogfood-mimixbox: ## Full mimixbox applet e2e (builds + --full-installs applets; set MIMIXBOX_REPO)
	# Builds the mimixbox multi-call binary, installs every applet into an
	# isolated bin dir put FIRST on PATH, then runs the atago specs against the
	# real applets (the atago replacement for mimixbox's ShellSpec suite).
	# --parallel 1 on purpose: the kill-family scenarios (killall/pkill/pgrep)
	# signal processes BY NAME, so running them alongside other scenarios'
	# `sleep` processes is a cross-scenario race (seen as a flaky
	# "sleeps then returns" failure under --parallel 8).
	bash ./test/e2e/tools/mimixbox/run.sh --parallel 1

dogfood-mobilepkg: ## Full mobilepkg e2e (builds latest mobilepkg; set MOBILEPKG_REPO)
	bash ./test/e2e/tools/mobilepkg/run.sh --parallel $(PARALLEL)

docs: ## Regenerate the committed behavior docs under doc/e2e/ from the specs
	env CGO_ENABLED=0 $(GO_BUILD) $(GO_LDFLAGS) -o ./dist/$(APP) .
	./dist/$(APP) doc --out doc/e2e/atago.md      ./test/e2e/atago
	./dist/$(APP) doc --out doc/e2e/git.md         ./test/e2e/thirdparty/git
	./dist/$(APP) doc --out doc/e2e/caddy.md       ./test/e2e/thirdparty/caddy
	./dist/$(APP) doc --out doc/e2e/pushgateway.md ./test/e2e/thirdparty/pushgateway
	./dist/$(APP) doc --out doc/e2e/webhook.md     ./test/e2e/thirdparty/webhook
	./dist/$(APP) doc --out doc/e2e/gitea.md       ./test/e2e/thirdparty/gitea
	./dist/$(APP) doc --out doc/e2e/jq.md          ./test/e2e/thirdparty/jq
	./dist/$(APP) doc --out doc/e2e/fzf.md         ./test/e2e/thirdparty/fzf
	./dist/$(APP) doc --out doc/e2e/redis.md       ./test/e2e/thirdparty/redis
	./dist/$(APP) doc --out doc/e2e/hugo.md        ./test/e2e/thirdparty/hugo
	./dist/$(APP) doc --out doc/e2e/openssl.md     ./test/e2e/thirdparty/openssl
	./dist/$(APP) doc --out doc/e2e/sqlite3.md     ./test/e2e/thirdparty/sqlite3
	./dist/$(APP) doc --out doc/e2e/minio.md       ./test/e2e/thirdparty/minio
	./dist/$(APP) doc --out doc/e2e/prometheus.md  ./test/e2e/thirdparty/prometheus
	./dist/$(APP) doc --out doc/e2e/rclone.md      ./test/e2e/thirdparty/rclone
	./dist/$(APP) doc --out doc/e2e/restic.md      ./test/e2e/thirdparty/restic
	./dist/$(APP) doc --out doc/e2e/coredns.md     ./test/e2e/thirdparty/coredns
	./dist/$(APP) doc --out doc/e2e/nats.md        ./test/e2e/thirdparty/nats
	./dist/$(APP) doc --out doc/e2e/mailpit.md     ./test/e2e/thirdparty/mailpit
	./dist/$(APP) doc --out doc/e2e/ntfy.md        ./test/e2e/thirdparty/ntfy
	./dist/$(APP) doc --out doc/e2e/transfersh.md  ./test/e2e/thirdparty/transfersh
	./dist/$(APP) doc --out doc/e2e/gotify.md      ./test/e2e/thirdparty/gotify
	./dist/$(APP) doc --out doc/e2e/grafana.md     ./test/e2e/thirdparty/grafana
	./dist/$(APP) doc --out doc/e2e/gup.md         ./test/e2e/tools/gup
	./dist/$(APP) doc --out doc/e2e/sqly.md        ./test/e2e/tools/sqly
	./dist/$(APP) doc --out doc/e2e/truss.md       ./test/e2e/tools/truss
	./dist/$(APP) doc --out doc/e2e/iso8583tool.md ./test/e2e/tools/iso8583tool
	./dist/$(APP) doc --out doc/e2e/jose.md        ./test/e2e/tools/jose
	./dist/$(APP) doc --out doc/e2e/career.md      ./test/e2e/tools/career
	./dist/$(APP) doc --out doc/e2e/mimixbox.md    ./test/e2e/tools/mimixbox
	./dist/$(APP) doc --out doc/e2e/mobilepkg.md   ./test/e2e/tools/mobilepkg

site: ## Regenerate the browsable docs site under site/ (drift-guarded by TestSite_InSync)
	env UPDATE_SITE=1 $(GO) test -run TestSite_InSync .

demo: ## Regenerate the README GIFs with vhs (requires vhs on PATH)
	env CGO_ENABLED=0 $(GO_BUILD) $(GO_LDFLAGS) -o ./dist/$(APP) .
	rm -f doc/demo/snapshots/help.txt
	vhs doc/vhs/demo.tape
	vhs doc/vhs/snapshot.tape
	vhs doc/vhs/review.tape

.DEFAULT_GOAL := help
help:
	@grep -E '^[0-9a-zA-Z_-]+[[:blank:]]*:.*?## .*$$' $(MAKEFILE_LIST) | sort \
	| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[1;32m%-15s\033[0m %s\n", $$1, $$2}'
