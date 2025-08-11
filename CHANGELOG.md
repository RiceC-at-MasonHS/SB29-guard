# Changelog

## v1.2.1 (2025-08-11)

Patch:
- Fix CI failure in script tests by resolving script paths from repo root.
- Linux automation script now honors env overrides (GUARD_BASE, OUT_FILE, ONLY_WHEN_CHANGED) for easier testing and customization.

## v1.2.0 (2025-08-11)

Highlights:
- Proxy-first quickstarts unified (NGINX, Caddy, HAProxy, Apache) with consistent structure and crosslinks.
- New GUI-proxy guide using /classify and /domain-list endpoints.
- Set-and-forget automation scripts added (Linux bash and Windows PowerShell) with scheduling instructions.
- Makefile convenience targets to generate example proxy bundles and static explain site; clean target added.
- HAProxy runtime socket notes for zero-reload map updates.
- Documentation sweep: updated top-level README, operator checklist, hands-off operations.

Testing & quality:
- New test suite for automation scripts (Linux E2E via mock server; Windows static checks).
- CI/coverage gates remain green across packages.

Notes:
- The legacy “easy-mode DNS” path is retired; proxy-first is the recommended approach.
- dist/* remains ignored in git; generate local examples via Makefile when needed.

## v1.1.2

Maintenance: release automation polish. No functional code changes from v1.1.1.

Fixes:
- Prevent duplicate upload of SHA256SUMS.txt in release step (single glob; strict fail_on_unmatched_files)

## v1.1.1

Maintenance release to fix release automation. No functional code changes from v1.1.0.

Fixes:
- Avoid artifact name collisions in release matrix builds (unique names + merge on download)
- Concurrency guard for tag-triggered releases to prevent overlapping runs
- Consolidate to a single tag-based release workflow; legacy release kept dispatch-only
- CI artifacts set to overwrite on re-runs to prevent 409 conflicts

## v1.1.0

Highlights:
- New DNS formats in generate-dns: dnsmasq, domain-list, winps
- Clarified CLI help and docs; easier to discover formats and modes
- Guards against committing .exe binaries in pre-commit and CI
- Publish workflow fix: stable lowercased owner for GHCR tags

Notes:
- Dockerfile uses distroless base (debian12:nonroot). You can pin by digest for scanner quieting; see Dockerfile comment.

