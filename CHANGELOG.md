# Changelog

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
- Easy-mode integration tests moved under internal/easymodeint (opt-in)
- Guards against committing .exe binaries in pre-commit and CI
- Publish workflow fix: stable lowercased owner for GHCR tags

Notes:
- Dockerfile uses distroless base (debian12:nonroot). You can pin by digest for scanner quieting; see Dockerfile comment.
- Integration tests for easy-mode require build tags `easymode integration` and SB29_EASYMODE_TEST=1.

