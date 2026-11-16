SHELL := /bin/bash
GOTOOL := go tool -modfile=tools/go.mod

# Reachable advisories accepted by `make vulncheck`; see scripts/vulncheck.sh. Comment why for each ID.
# Long-term accepted risks (never fail while reported):
VULNCHECK_ACCEPTED :=
# Accepted only until a fix is released; vulncheck fails once govulncheck reports a fixed version.
# GO-2026-6872: rate limiter keyed on spoofable IP; GO-2026-6875: proxy forwards X-Real-IP. xecho uses neither middleware.
VULNCHECK_UNTIL_FIXED := GO-2026-6872 GO-2026-6875

.PHONY: test tools fmt lint vulncheck apidiff

test:
	bash -c 'diff -u <(echo -n) <(gofmt -s -d .)'
	go vet ./...
	go test -v -race ./...

# Builds the tools pinned in tools/go.mod so later targets start quickly.
tools:
	go build -modfile=tools/go.mod tool

fmt:
	gofmt -s -w .
	$(GOTOOL) goimports -w .

lint:
	$(GOTOOL) golangci-lint run ./...

vulncheck:
	@GOTOOL="$(GOTOOL)" VULNCHECK_ACCEPTED="$(VULNCHECK_ACCEPTED)" \
		VULNCHECK_UNTIL_FIXED="$(VULNCHECK_UNTIL_FIXED)" ./scripts/vulncheck.sh

# Reports API changes since the latest tag and suggests the next version. Requires a clean worktree.
apidiff:
	$(GOTOOL) gorelease -base=$$(git describe --tags --abbrev=0)
