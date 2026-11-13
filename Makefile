GOTOOL := go tool -modfile=tools/go.mod

.PHONY: test tools fmt lint vulncheck apidiff

test:
	bash -c 'diff -u <(echo -n) <(gofmt -s -d .)'
	bash -c 'diff -u <(echo -n) <(go vet ./...)'
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
	$(GOTOOL) govulncheck ./...

# Reports API changes since the latest tag and suggests the next version. Requires a clean worktree.
apidiff:
	$(GOTOOL) gorelease -base=$$(git describe --tags --abbrev=0)
