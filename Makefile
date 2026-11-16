SHELL := /bin/bash
GOTOOL := go tool -modfile=tools/go.mod

# Known advisories with no upstream fix yet. Remove each ID once echo ships a fix.
# GO-2026-6872: rate limiter keyed on spoofable IP; GO-2026-6875: proxy forwards X-Real-IP. xecho uses neither middleware.
VULNCHECK_IGNORE := GO-2026-6872 GO-2026-6875

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

# Fails on any reachable vulnerability not listed in VULNCHECK_IGNORE.
vulncheck:
	@set -euo pipefail; \
	found=$$($(GOTOOL) govulncheck -format json ./... \
		| jq -r 'select(.finding.trace[0].function != null) | .finding.osv' | sort -u); \
	ignore="$(VULNCHECK_IGNORE)"; status=0; \
	for id in $$found; do \
		if [[ " $$ignore " == *" $$id "* ]]; then echo "ignored (known): $$id"; \
		else echo "VULNERABLE: $$id - https://pkg.go.dev/vuln/$$id"; status=1; fi; \
	done; \
	for id in $$ignore; do \
		[[ " $$(echo $$found) " == *" $$id "* ]] || echo "stale ignore (no longer reported, remove it): $$id"; \
	done; \
	if [[ $$status -ne 0 ]]; then echo "Run '$(GOTOOL) govulncheck ./...' for traces."; fi; \
	exit $$status

# Reports API changes since the latest tag and suggests the next version. Requires a clean worktree.
apidiff:
	$(GOTOOL) gorelease -base=$$(git describe --tags --abbrev=0)
