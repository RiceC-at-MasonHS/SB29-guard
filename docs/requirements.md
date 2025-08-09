# SB29-guard Requirements Specification

Version: 0.1.0
Status: Active (living document; sections marked FUTURE deferred)
Owner: Project Maintainers (Educator-led)
Last Updated: 2025-08-08

---
## 1. Purpose / Goal
Provide a lightweight, auditable, and privacy‑respecting mechanism for K‑12 districts in Ohio to assist educators with compliance to Ohio Senate Bill 29 by transparently intercepting access to disallowed digital tools (domains) whose vendors have not executed an approved Digital Privacy Agreement (DPA). When a user (teacher / student) attempts to reach a flagged domain, they are redirected to a locally hosted explainer page describing:
1. The original requested domain.
2. The policy reason / classification for restriction.
3. Guidance for alternatives and escalation (request review process).

The system MUST NOT collect or expose personally identifiable information (PII) of students and MUST minimize data retention.

---
## 2. Definitions
- Disallowed Domain: A FQDN (or wildcard pattern) associated with a product lacking an approved DPA under Ohio SB29.
- Redirect Service (Explainer Page): HTTP(S) endpoint providing compliant messaging & rationale.
- Policy Dataset: Canonical, versioned list of domains, metadata, and policy classification.
- Classification Code: Short code (e.g., NO_DPA, PENDING_REVIEW, EXPIRED_DPA) describing restriction reason.
- Policy Version: Semantic version of the domain dataset used to generate DNS artifacts.

---
## 3. High-Level Overview
1. Maintain a structured domain policy file (YAML) under version control.
2. Generate DNS override artifacts (hosts, bind zone, unbound local-zone, RPZ) pointing to redirect IP or host.
3. Redirect host serves an explainer page using embedded templates; direct parameter model currently `/explain?domain=` (internal lookup planned for FUTURE host header inference).
4. Page displays domain, classification, rationale (optional), reference (optional), policy version, timestamp.
5. (FUTURE) Aggregate logging / metrics; current implementation does not persist request events.

---
## 4. Functional Requirements
Implemented (Current):
FR-1: Maintain a machine-readable policy file (default: `policy/domains.yaml`).
FR-2: Policy record fields: domain, classification, rationale, last_review, status (+ optional source_ref, notes, expires, tags).
FR-3: Support wildcard domains (leftmost label `*.`).
FR-4: CLI command `validate` performs schema + logical validation (strict mode toggle).
FR-5: CLI command `generate-dns` exports: hosts, bind zone, unbound local-zone, RPZ.
FR-6: DNS export allows configurable redirect IPv4 (`--redirect-ipv4`) or redirect host (`--redirect-host`).
FR-6a: Published Google Sheets CSV ingestion via `--sheet-csv` with on-disk caching (ETag / Last-Modified).
FR-6b: In serve mode with `--sheet-csv`, auto-refresh policy on a schedule with graceful error handling/logging.
  - Flags: `--refresh-at HH:MM` (daily) or `--refresh-every <duration>`.
FR-14: Integrity hash (SHA-256 canonical over active records) via `hash` command.
FR-16: `generate-dns --dry-run` prints to stdout.
FR-17: Unit tests cover schema validation, DNS generation (positive + negative), server handlers, hash, CLI.
FR-42: Embed static HTML/CSS templates (layout.html, root.html, explain.html, style.css) via Go embed; allow runtime override with `--templates <dir>`.
FR-46: Dry-run output supported for CI validation.

Partial / Planned (FUTURE):
FR-7: HTTP redirect mode toggle (currently direct page only).
FR-8: Host header inference when params missing.
FR-9: Additional last review date & contact/escalation dynamic text (basic contact line present).
FR-10: Localization readiness (strings presently inline English).
FR-11: Formal accessibility audit & documentation (structure is semantic; needs axe validation) .
FR-12: JSON API endpoint `/api/domain-info`.
FR-13: Aggregated counts & reporting CLI.
FR-18: Container/Dockerfile publishing.
FR-19: Central config file loader.
FR-21..27: Additional DNS formats (pfSense, OPNsense, Infoblox, Route53, Azure, GCloud, plain list).
FR-28: Classification filter flag.
FR-29: Suspended inclusion toggle (currently suspended excluded from hash & generation logic implicitly; explicit flag pending).
FR-30: Manifest JSON generation.
FR-31..38: Sheets integration (advanced: periodic sync, metrics, fallback, error write-back) – PARTIAL (published CSV fetch only implemented as FR-6a).
FR-39: Additional subcommands (classify, export-schema, etc.).
FR-40: Hash check command for artifacts.
FR-41: Read-only policy flag.
FR-43: Integrity fail action.
FR-44: Systemd unit generation.
FR-45: Distroless/Alpine images.
FR-47: Explicit O(log n) matcher structure (current approach acceptable for present scale; optimization later).
FR-48: Rate limiting.
FR-49: CSP nonce injection (not needed sans inline scripts).
FR-50: Public IP safety check.

---
## 5. Non-Functional Requirements
NFR-1: Privacy first: No storage of client IP, user agent beyond short-lived (<=24h) rolling logs for troubleshooting; storage can be disabled.
NFR-2: Performance: Redirect page response time <150 ms server-side (p95) under 50 RPS.
NFR-3: Reliability: Redirect service target uptime goal 99.5% (school hours); graceful degradation (serve static fallback page if dynamic context fails).
NFR-4: Portability: Runs on Linux or Windows Server (DNS) environments; container image multi-arch (amd64, arm64).
NFR-5: Simplicity: Core system deployable with: policy file + generated DNS records + static web server variant (fallback mode).
NFR-6: Observability: Provide structured logs (JSON) with fields: ts, event, domain, classification, policy_version, count (aggregated).
NFR-7: Security: No inbound auth required for user page; admin/reporting features gated behind environment variable-based basic auth or separate network segment.
NFR-8: Code Quality: CI enforces lint, schema validation, tests.
NFR-9: Documentation: README + docs for deployment topologies.
NFR-10: Minimal Dependencies: Keep runtime dependencies small to ease audits.
NFR-11: Implementation language: Go (>=1.22) chosen for static binary, low memory footprint, strong stdlib, portability.
NFR-12: Dependency minimization: third-party libraries limited to YAML parsing, optional JSON Schema validation, and Google Sheets API client; audit list documented.
NFR-13: Startup time < 150ms cold on modest hardware (2 vCPU) with 5k records.
NFR-14: Memory footprint target < 60MB RSS with 10k records loaded.
NFR-15: Supply SBOM (Software Bill of Materials) generation in CI (e.g., Syft) for transparency.
NFR-16: Provide reproducible builds instructions (deterministic Go build flags) and publish SHA-256 sums.
NFR-17: Container hardening: read-only root filesystem, non-root user, dropped Linux capabilities, healthcheck endpoint.
NFR-18: Systemd hardening guidance: `NoNewPrivileges=yes`, `ProtectSystem=strict`, `ProtectHome=yes`, `ReadWritePaths=/var/lib/sb29guard /var/log/sb29guard`, `PrivateTmp=yes`.
NFR-19: No dynamic plugin loading or runtime code execution permitted.
NFR-20: All network listeners bind only to configured interface (default 0.0.0.0) and optionally `--bind` parameter.
NFR-21: All error logs redact secrets (env var names only) and truncate long inputs >2k characters (security + log hygiene).
NFR-22: Time synchronization assumption: system clock accurate within ±60s; drift detection logged if responses show >300s difference from `ts` parameter.
NFR-23: Provide clear exit codes documented for each subcommand (machine parsing reliability).

---
## 6. Policy File Schema (YAML Example)
```yaml
version: 0.1.0
updated: 2025-08-08
records:
  - domain: "exampletool.com"
    classification: NO_DPA
    rationale: "Vendor has not signed district-approved Digital Privacy Agreement."
    last_review: 2025-08-01
    status: active
    source_ref: "District Legal Review #123"
  - domain: "*.trackingwidgets.io"
    classification: EXPIRED_DPA
    rationale: "Agreement expired; renewal pending."
    last_review: 2025-07-15
    status: active
```
Validation Rules:
- Domains normalized to lowercase.
- Wildcards only as leftmost label (`*.domain.tld`).
- Dates ISO8601 (YYYY-MM-DD).
- classification in {NO_DPA, PENDING_REVIEW, EXPIRED_DPA, LEGAL_HOLD, OTHER}.
- status in {active, suspended}.

---
## 7. Redirect Parameter Contract
When a blocked domain is requested, DNS points to redirect host (e.g., 10.10.10.50). A web server vhost/default site issues (option A) an HTTP 302 to `/explain` with parameters OR (option B) serves dynamic content directly.

Current Parameters:
- domain (required) – original requested domain.
Future Parameter Model (deferred): original_domain, classification, policy_version, ts, ref, locale (server currently derives classification from in-memory policy at render time instead of trusting client parameters).

---
## 8. DNS Generation Strategies
Option A (A Record Override): For each domain/wildcard generate zone override mapping to redirect IP.
Option B (CNAME Consolidation): Point each restricted FQDN to `blocked.guard.local.` which resolves to redirect IP (reduces A record duplication).
Option C (RPZ - Response Policy Zone): Provide RPZ zone file for integration with supported DNS resolvers (Bind/Unbound/PowerDNS). Action = CNAME rewrite to redirect host.

Selectable via CLI flags: `--format hosts|bind|unbound|rpz` plus `--mode a-record|cname` (cname for bind/unbound); RPZ is a distinct format.

---
## 9. Logging & Metrics
- Inbound request log (ephemeral): domain, classification, policy_version, minute bucket.
- Aggregator produces rolling counts and flushes to daily JSON file (planned); currently metrics JSON is exposed at `/metrics` (policy version, record count, refresh stats).
- No raw IPs or user agents persisted beyond in-memory counters.

---
## 10. Security & Privacy Controls
- No cookies, no tracking beacons, no third-party scripts.
- CSP header: `default-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self';`.
- Referrer-Policy: `no-referrer`.
- Cache-Control: `no-store` to prevent stale reasons if classification changes.
- Optional HSTS (if TLS).
- Regular dependency vulnerability scan in CI.

---
## 11. Configuration File (app.yaml example)
```yaml
redirect_host: guard.school.local
listen_port: 8080
redirect_mode: direct_page   # direct_page | http_redirect
redirect_status_code: 302
static_fallback_page: ./public/fallback.html
policy_file: ./policy/domains.yaml
log_dir: ./logs
aggregate_flush_interval_sec: 60
locale_default: en-US
feature:
  api_domain_info: true
  metrics: false
  admin_reporting: false
security:
  admin_basic_auth:
    enabled: true
    username_env: SB29_ADMIN_USER
    password_env: SB29_ADMIN_PASS
```

---
## 12. CLI Commands (Planned)
- `sb29guard validate --policy ./policy/domains.yaml`
- `sb29guard generate-dns --policy ./policy/domains.yaml --out ./dist/dns --mode rpz --redirect-host guard.school.local --redirect-ipv4 10.10.10.50`
- `sb29guard serve --config ./config/app.yaml`
- `sb29guard classify --lookup some.domain.com` (prints current classification)

---
## 13. Testing / Acceptance Criteria
TST-1: Invalid policy file (schema violation) returns non-zero exit code.
TST-2: Given a domain in policy, generated DNS file contains corresponding override.
TST-3: Wildcard entry expands / matches subdomains during runtime lookup.
TST-4: Redirect page renders correct rationale & classification for sample domain.
TST-5: No network calls to external hosts during normal page render.
TST-6: Accessibility scan (axe) reports no critical violations.
TST-7: Unit test coverage >= 80% for policy parsing & DNS generation modules.
TST-8: Container image passes vulnerability scan (no critical CVEs) at build time.
TST-9: Logging aggregator produces daily summary file with expected JSON schema.
TST-10: Removing a domain & regenerating DNS removes its record (idempotent build).

---
## 14. Deployment Topologies
- Simple: Internal resolver override (Bind) + single container for explainer page.
- Pi-hole: Import hosts file mapping disallowed domains to redirect IP.
- Windows DNS: Zone file script import.
- RPZ: Distribute RPZ zone to recursive resolvers (central management).

Example deployment docs added under `docs/deployment/` (bind, unbound, pihole, windows-dns). Additional formats FUTURE.

---
## 15. Future Enhancements (Non-Commitment)
- Web-based admin to request review / track approval workflows.
- Automated vendor DPA status ingestion (district SIS / legal system integration).
- Email notifications when DPA statuses change.
- TLS certificate automation (ACME internal CA) for blocked domains (if serving HTTPS directly instead of redirect).
- Multi-district federation (shared baseline policy + local overrides).

---
## 16. Out of Scope (Initial Phase)
- Storage of individual user access attempts.
- Real-time user authentication or personalization.
- Automated legal document management.

---
## 17. Contributing Guidance (Summary)
- Open a PR referencing policy change rationale.
- Run `validate` + tests prior to PR.
- Increment policy version only through CI if content hash changed.
- Provide plain-language rationale for every new domain.

### 17.1 Local Quality Gates (Pre-commit Policy)
To keep history clean and commits focused, the repository enforces a strict local pre-commit policy that mirrors CI. Enable project hooks once per clone:

```
git config core.hooksPath .githooks
```

The pre-commit hook will fail the commit if any of the following checks fail:
- Formatting: gofmt check (no diffs allowed)
- Lint: golangci-lint with a 3-minute timeout
- Static analysis: go vet ./...
- Build: go build ./...
- Tests: go test -race -coverprofile=coverage.out -covermode=atomic ./...
- Coverage gates (per-package minimums):
  - internal/policy >= 70%
  - internal/dnsgen >= 70%
  - internal/server >= 85%
  - internal/sheets >= 80%

Pre-push hook simply re-runs pre-commit for defense-in-depth. To skip hooks in emergencies, use `--no-verify` and open a follow-up issue to restore quality gates.

---
## 18. License & Compliance Notes
License: AGPL-3.0 (GNU Affero General Public License v3). Districts can self-host and modify; if run as a network service with modifications, provide corresponding source to users.
DISCLAIMER: Tool aids but does not guarantee legal compliance; districts must conduct their own legal review.

---
## 19. Acceptance to Move Forward
This document must be reviewed by: (a) Lead Teacher Sponsor, (b) District IT Security, (c) Legal / Compliance Liaison. Upon sign-off, implementation can begin.

---
## 20. Implementation Stack Decision
Language: Go (>=1.22). Rationale: static single binary, cross-compilation (Windows/Linux), minimal runtime deps, strong stdlib for networking & crypto.

### Module Layout (Current / Planned)
- `internal/policy`: load, normalize, validate, hash (implemented)
- `internal/dnsgen`: generators (hosts, bind, unbound, rpz implemented)
- `internal/server`: HTTP server + embedded templates (implemented)
- `internal/hash`: hashing utility (implemented)
- `internal/match`: FUTURE dedicated matcher optimization
- `internal/logging`: FUTURE structured logging package
- `internal/sheets`: FUTURE sheets ingestion
- `internal/integrity`: FUTURE integrity verification helpers

### Security Hardening Points
- Embed templates & schema (immutability)
- Distroless container variant
- Integrity check at startup (policy + templates)
- Strict HTTP headers (CSP, etc.)
- Optional systemd unit generation with hardening directives
- Read-only runtime except cache/log paths

### Build & Release
- Direct Go tooling (no Makefile) via documented commands in README / TECHNICAL: `go test ./...`, `go build ./cmd/sb29guard`. CI mirrors these steps.
- Deterministic build flags: `-trimpath -ldflags "-s -w -buildid="` and embed version/hash via `-ldflags "-X main.version=... -X main.commit=..."`.

### Integrity Strategy
- Hash normalization (sorted active records) reused across DNS generation & CLI hash command.
- (FUTURE) `manifest.json` will enumerate artifact hashes.

END IMPLEMENTATION STACK DECISION
---
END OF DOCUMENT
