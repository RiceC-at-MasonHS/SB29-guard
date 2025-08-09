# Contributing  
[README](./README.md) • [Technical Reference](./TECHNICAL.md) • [Customizing](./CUSTOMIZING.md)

Thanks for your interest!

- Open issues with clear context and repro steps.
- Run tests locally: `go test ./...`
- Follow Go formatting and linting (CI will verify).
- Keep PRs focused and reference requirement IDs from `docs/requirements.md` when applicable.
- For templates/branding changes, see `CUSTOMIZING.md`.

## Development Quickstart
- Go 1.22+
- Build: `go build ./cmd/sb29guard`
- Run server: `sb29guard serve --policy policy/domains.yaml` or `--sheet-csv <csv_url>`

## Tests & Coverage
We maintain per-package coverage gates in CI for critical packages. Please add tests alongside changes.
See badges and coverage summaries in CI artifacts.

## License
By contributing, you agree that your contributions will be licensed under the AGPL-3.0 (see `LICENSE`).
