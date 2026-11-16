[![CI](https://github.com/JSainsburyPLC/xecho/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/JSainsburyPLC/xecho/actions/workflows/ci.yml)

# xecho

A thin wrapper around https://echo.labstack.com. Includes
    
* New Relic integration
* Logrus for structured json logging
* Custom error type
* Custom request context with tracing (correlation ID)

## Requirements

* Go 1.27 or later
* Echo v4 (`github.com/labstack/echo/v4`)

## Breaking changes in v0.2.0

Upgrading from v0.1.x requires changes in consuming services:

* **Echo v3 → v4.** Exported functions (`Echo`, `EchoHandler`, `ContextMiddleware` and the other
  middleware constructors) now take and return `github.com/labstack/echo/v4` types. Change all
  `github.com/labstack/echo` imports, including `echo/middleware`, to `github.com/labstack/echo/v4`,
  then run `go mod tidy`. See the [Echo v4.0.0 release notes](https://github.com/labstack/echo/releases/tag/v4.0.0)
  for changes to Echo's own API.
* **Go 1.27 minimum.** The `go` directive moved from 1.12 to 1.27, so consumers must build with Go 1.27 or later.
* **New Relic transaction end errors are logged.** A failure to end the New Relic transaction is now
  logged at error level on the request logger instead of being silently ignored.

## Client IP behind a load balancer

The `ip` field in request logs and New Relic comes from Echo's `RealIP()`. By default xecho uses
`echo.ExtractIPDirect()`, so the IP is the connection's remote address and `X-Forwarded-For` /
`X-Real-IP` are ignored. Behind a load balancer this is the balancer's IP. Set
`Config.IPExtractor` to trust forwarding headers only from your proxies:

```go
conf := xecho.NewConfig()
conf.IPExtractor = echo.ExtractIPFromXFFHeader(echo.TrustIPRange(vpcCIDR))
```

`ExtractIPFromXFFHeader()` trusts loopback, link-local and private ranges by default. Never trust
forwarding headers from addresses a client can reach directly, as the header can be spoofed.

## Development

Tool versions are pinned in `tools/go.mod` and run with `go tool`, so nothing needs installing globally.

* `make test` - gofmt, go vet and race-enabled tests
* `make tools` - pre-build the pinned tools
* `make fmt` - format with gofmt and goimports
* `make lint` - golangci-lint
* `make vulncheck` - govulncheck; fails on reachable vulnerabilities unless listed in the Makefile as
  `VULNCHECK_ACCEPTED` (long-term) or `VULNCHECK_UNTIL_FIXED` (fails once a fix is released). Also runs
  weekly on `master` via the "Scheduled vulncheck" workflow, which can be triggered manually from the Actions tab
* `make apidiff` - gorelease against the latest tag (needs a clean worktree)
