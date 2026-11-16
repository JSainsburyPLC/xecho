[![CI](https://github.com/JSainsburyPLC/xecho/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/JSainsburyPLC/xecho/actions/workflows/ci.yml)

# xecho

A thin wrapper around https://echo.labstack.com. Includes
    
* New Relic integration
* Logrus for structured json logging
* Custom error type
* Custom request context with tracing (correlation ID)

## Requirements

* Go 1.27 or later
* Echo v4 (`github.com/labstack/echo/v4`). Versions before v0.2.0 used Echo v3; when upgrading,
  change your `github.com/labstack/echo` imports (including `echo/middleware`) to `echo/v4`.

## Development

Tool versions are pinned in `tools/go.mod` and run with `go tool`, so nothing needs installing globally.

* `make test` - gofmt, go vet and race-enabled tests
* `make tools` - pre-build the pinned tools
* `make fmt` - format with gofmt and goimports
* `make lint` - golangci-lint
* `make vulncheck` - govulncheck
* `make apidiff` - gorelease against the latest tag (needs a clean worktree)
