<div align="center">

# SB29-guard 🚦📘

Making blocked ed‑tech domains less confusing (and more transparent) for teachers & students. Protecting student data is a noble objective, and can be done in ways that lift up our communities.

<em>Built by a teacher, for teachers — to protect educators, reduce friction, and make compliance humane.</em>

<!-- Badges -->
[![Build](https://github.com/RiceC-at-MasonHS/SB29-guard/actions/workflows/ci.yml/badge.svg)](../../actions)
![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-AGPL--3.0-blue)
![Status](https://img.shields.io/badge/status-stable-brightgreen)
![Coverage](https://img.shields.io/badge/coverage-core%2080%25+-brightgreen)

</div>

<div align="center">
  <img src="./screenshot-2025-08-09-204319.png" alt="SB29-guard explanation page (dark mode)" style="max-width: 980px; width: 100%; border-radius: 8px;" />
</div>

District‑friendly tool that shows a clear, plain‑language “Why was I redirected?” page when staff or students try a site without an approved Data Privacy Agreement (SB29 context). One small self‑contained binary. No tracking. No student data stored.

## The gist
- Show a friendly explanation page when a restricted domain is visited.
- Keep a simple, auditable policy (YAML or Google Sheet).
- Point DNS (A/CNAME/RPZ/hosts) to this service; the page infers the original domain safely.

## Get started
- Easiest (recommended): Easy‑mode with auto‑HTTPS using Docker + Caddy — see `easy-mode/README.md`.
- Prefer your own stack? See implementers overview and deployment guides:
  - `docs/implementers/README.md`
  - `docs/deployment/` (BIND, Unbound, Pi-hole, Windows DNS, pfSense, OPNsense, Infoblox)

## Downloads and images
- Releases (binaries + SHA256SUMS): GitHub → Releases
- Container image (GHCR): `ghcr.io/ricec-at-masonhs/sb29-guard:v1.x.x` (pin a version or digest; latest stable is v1.1.x)

## Operators: where to read more
- Technical Reference (headers, caching, CSV refresh, metrics, verification): `TECHNICAL.md`
- Deployment guides (platform specifics): `docs/deployment/`
- Customizing the UI/templates: `CUSTOMIZING.md`

## Developers
- Contributing: `CONTRIBUTING.md`
- CLI design and behavior: `docs/cli-design.md`
- Print version: `sb29guard version` (release builds embed version/commit/date)

## License and notice
- License: AGPL‑3.0 (see `LICENSE`)
- Disclaimer: Coordinate with your legal/data privacy teams; this aids transparency and workflow but doesn’t replace legal review.

Questions or ideas? Open an issue. Contributions welcome.
