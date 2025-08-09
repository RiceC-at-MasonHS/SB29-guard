# Contributing
[README](./README.md) • [Technical Reference](./TECHNICAL.md) • [Customizing](./CUSTOMIZING.md)

Thanks for your interest!

- Open issues with clear context and repro steps.
- Run tests locally: `go test ./...`
- Keep PRs focused and reference requirement IDs from `docs/requirements.md` when applicable.
- For templates/branding changes, see `CUSTOMIZING.md`.

## Development quickstart
- Install Go 1.22+
- Optional: install golangci-lint locally
	- `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8`
- Enable repo hooks (once per clone):
	- `git config core.hooksPath .githooks`

## Required local quality gates (pre-commit)
Commits will be blocked unless all of the following pass:
- Auto-format staged Go files (gofmt -s -w)
- Formatting check on staged files
- golangci-lint (3m timeout)
- go vet ./...
- go build ./...
- go test -race -coverprofile=coverage.out -covermode=atomic ./... (race only when CGO enabled)
- Per-package coverage minimums (only for changed packages locally; CI checks all):
	- internal/policy ≥ 70%
	- internal/dnsgen ≥ 70%
	- internal/server ≥ 85%
	- internal/sheets ≥ 80%

Performance tweaks in local hooks:
- Doc-only commits (no .go changes) skip tests/coverage gates.
- Coverage gates only run for packages touched by the commit (CI still checks all).
- Local tests use the Go build cache (no -count=1); CI uses -count=1 for freshness.

Pre-push re-runs pre-commit for defense-in-depth.

Windows notes:
- Hooks are bash scripts run via Git Bash; they work fine from PowerShell as long as Git Bash is installed.
- On Windows without a C toolchain, tests run without -race (CGO disabled). CI will still run -race on Linux.

## Tests & coverage in CI
CI mirrors the local checks and publishes coverage artifacts.

## License
By contributing, you agree that your contributions will be licensed under the AGPL-3.0 (see `LICENSE`).
