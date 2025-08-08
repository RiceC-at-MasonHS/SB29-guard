# SB29-guard Requirements Specification

Version: 0.1.0 (INITIAL DRAFT)
Status: Draft
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
1. Maintain a structured domain policy file (e.g., YAML/JSON/CSV) under version control.
2. Generate DNS override records for each disallowed domain pointing to a local redirect host IP (IPv4 + optional IPv6) OR use a wildcard CNAME inside an internal override zone.
3. Redirect host serves an explainer page; it inspects query parameters to render dynamic rationale.
4. Query parameters include: original_domain, classification, policy_version, timestamp (UTC ISO8601), locale (optional), ref=sb29guard.
5. System logs ONLY aggregate counts (domain, classification, day) – no user identifiers, no IP retention beyond ephemeral operational logs (<24h, rotation or anonymization required).

---
## 4. Functional Requirements
FR-1: Maintain a machine-readable policy file (default: `policy/domains.yaml`).
FR-2: Each policy record MUST contain: domain (string), classification (enum), rationale (string, teacher-friendly), last_review (date), status (active|suspended), notes (optional), source_ref (optional citation).
FR-3: Support wildcard domains (e.g., `*.example.com`).
FR-4: Provide a CLI command to validate the policy file schema.
FR-5: Provide a CLI command to export DNS zone snippets for: (a) BIND, (b) Windows DNS (zone file), (c) Pi-hole / dnsmasq hosts format, (d) Unbound local-zone directives.
FR-6: DNS export MUST allow configurable redirect IPv4/IPv6 targets.
FR-7: Provide an HTTP redirect OR direct resolution to the explainer host (configurable 301/302 or direct A record to web server).
FR-8: Explainer page MUST accept query parameters; if absent, it should attempt internal mapping by Host header.
FR-9: Explainer page MUST render: original domain, classification description, human rationale, last review date, policy version, contact/escalation instructions.
FR-10: Provide localization readiness (string keys; default en-US).
FR-11: Provide accessibility: WCAG 2.1 AA (semantic HTML, contrast, keyboard nav, ARIA labels as needed).
FR-12: Provide an optional JSON API endpoint: `GET /api/domain-info?domain=...` returning domain metadata.
FR-13: Provide an optional lightweight admin/report CLI to aggregate counts by domain & classification from access logs (with privacy constraints).
FR-14: Provide integrity check: hash (SHA-256) of policy file embedded into generated artifacts.
FR-15: Provide policy version (semantic) computed from Git tag or content hash.
FR-16: Provide dry-run mode for DNS export.
FR-17: Provide unit tests covering schema validation, DNS generation, query parameter parsing.
FR-18: Provide containerized deployment (Dockerfile + sample compose) for the redirect service.
FR-19: Support configuration file (YAML) for service settings (`config/app.yaml`).
FR-20: Provide signed release artifacts (optional future) to ensure tamper detection.
FR-21: `generate-dns` supports `--format pfSense-unbound` emitting Unbound include fragment with header metadata.
FR-22: `generate-dns` supports `--format opnsense-unbound` (variant of pfSense format; naming differences only).
FR-23: `generate-dns` supports `--format infoblox-rpz` (Infoblox‑compatible RPZ zone + optional CSV import variant).
FR-24: `generate-dns` supports `--format route53-json` producing AWS Route53 change batch JSON for private hosted zone creation/update (idempotent output ordering).
FR-25: `generate-dns` supports `--format azure-cli` producing Azure CLI script (idempotent) for Private DNS zone record sets.
FR-26: `generate-dns` supports `--format gcloud-dns` producing Google Cloud DNS transaction script template.
FR-27: `generate-dns` supports domain-only plain list variant for cloud security products (Umbrella, Cloudflare Gateway) with optional classification suffix as comment.
FR-28: Add export flag `--classification-filter <CLASS[,CLASS...]>` to restrict output to given classifications (phased rollout).
FR-29: Add export flag `--inactive-exclude/--no-inactive-exclude` (default exclude) controlling inclusion of suspended records.
FR-30: Generator writes `dist/dns/manifest.json` enumerating produced artifacts: path, mode, format, sha256, bytes, generated timestamp, policy_version, tool_version.
FR-31: Support Google Sheets as a policy source (`SB29_POLICY_SOURCE=sheet`) with periodic fetch interval configurable via env var.
FR-32: Sheets sync performs schema validation and falls back to last known good cache or local file on error.
FR-33: Provide `.env.example` documenting all environment variables with comments.
FR-34: Cache each successfully fetched sheet as normalized JSON/YAML with content hash in filename and maintain a `current` symlink/marker file.
FR-35: Provide CLI command `sb29guard sheet-pull` to force an immediate fetch & validation cycle (dry-run mode optional).
FR-36: Provide optional write-back of validation errors to a specified Google Sheet tab (future flag, initially placeholder / no-op).
FR-37: Redact secrets (API keys, credentials) from logs; never print full `.env` contents.
FR-38: Maintain metrics counter for successful vs failed sheet sync attempts (exposed in logs/metrics if enabled).
FR-39: Implementation delivered as a single statically linked Go binary (`sb29guard`) providing all subcommands (validate, generate-dns, serve, classify, hash, export-schema, demo-data, sheet-pull) to minimize operational complexity.
FR-40: Provide integrity command `sb29guard hash --check <artifact>` to verify embedded policy hash comment matches actual file contents (tamper detection).
FR-41: `serve` mode must support `--read-only-policy` flag rejecting runtime modifications to policy cache (except sheet sync updates) for hardened environments.
FR-42: Embed static HTML/CSS templates using Go `embed` (no runtime template directory dependency) and allow optional override via env `SB29_TEMPLATE_DIR`.
FR-43: Provide `--integrity-fail-action exit|warn` controlling behavior when hash mismatch detected at startup.
FR-44: Provide systemd hardening recommendations output via `sb29guard serve --print-systemd-unit` generating a sample unit with security directives.
FR-45: Provide container image build (multi-stage Dockerfile) producing distroless and Alpine variants.
FR-46: Provide `sb29guard generate-dns --dry-run` output to stdout without file writes for CI validation.
FR-47: Provide wildcard / exact domain matching structure with O(log n) lookup performance for 10k domains.
FR-48: Provide optional basic rate limiting for admin endpoints (`--admin-rps-limit`), default disabled.
FR-49: Provide CSP nonce injection for inline critical script/style blocks (if any) or ensure no inline scripts required.
FR-50: Provide configuration sanity check rejecting redirect IP if it matches a public (non-RFC1918/ULA) space unless `--allow-public-redirect-ip` set (prevent accidental external redirection).

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

Query Parameters (lowercase, kebab or snake consistent - choose snake):
- original_domain (required)
- classification (required)
- policy_version (required)
- ts (UTC ISO8601)
- ref = sb29guard (constant marker)
- locale (optional)

Example URL:
```
https://guard.school.local/explain?original_domain=exampletool.com&classification=NO_DPA&policy_version=0.1.0&ts=2025-08-08T12:00:00Z&ref=sb29guard
```
If params missing, backend attempts lookup by Host header or SNI (future TLS termination logic).

---
## 8. DNS Generation Strategies
Option A (A Record Override): For each domain/wildcard generate zone override mapping to redirect IP.
Option B (CNAME Consolidation): Point each restricted FQDN to `blocked.guard.local.` which resolves to redirect IP (reduces A record duplication).
Option C (RPZ - Response Policy Zone): Provide RPZ zone file for integration with supported DNS resolvers (Bind/Unbound/PowerDNS). Action = CNAME rewrite to redirect host.

Selectable via CLI flag: `--mode a-record|cname|rpz`.

---
## 9. Logging & Metrics
- Inbound request log (ephemeral): domain, classification, policy_version, minute bucket.
- Aggregator produces rolling counts and flushes to daily JSON file (e.g., `logs/aggregates/2025-08-08.json`).
- No raw IPs or user agents persisted beyond in-memory counters.
- Provide optional Prometheus metrics endpoint (future) guarded by network ACLs.

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

Provide example docs for each under `docs/deployment/` (future).

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

---
## 18. License & Compliance Notes
License: (TBD - recommend permissive OSS, e.g., MIT or Apache-2.0) ensuring districts can self-host.
Include DISCLAIMER: Tool aids but does not guarantee legal compliance; districts must conduct their own legal review.

---
## 19. Acceptance to Move Forward
This document must be reviewed by: (a) Lead Teacher Sponsor, (b) District IT Security, (c) Legal / Compliance Liaison. Upon sign-off, implementation can begin.

---
## 20. Implementation Stack Decision
Language: Go (>=1.22). Rationale: static single binary, cross-compilation (Windows/Linux), minimal runtime deps, strong stdlib for networking & crypto.

### Module Layout (Planned)
- `internal/policy`: load, normalize, validate, hash
- `internal/sheets`: fetch & convert Google Sheet rows
- `internal/dnsgen`: generators (bind, unbound, rpz, hosts, route53-json, etc.)
- `internal/server`: HTTP server, templates, headers, rate limiting
- `internal/match`: wildcard + exact matcher
- `internal/logging`: structured JSON logger
- `internal/integrity`: hash verification helpers

### Security Hardening Points
- Embed templates & schema (immutability)
- Distroless container variant
- Integrity check at startup (policy + templates)
- Strict HTTP headers (CSP, etc.)
- Optional systemd unit generation with hardening directives
- Read-only runtime except cache/log paths

### Build & Release
- Makefile / CI pipeline: lint -> test -> build (GOOS/GOARCH matrix) -> generate SBOM -> sign checksums (optional cosign) -> publish artifacts.
- Deterministic build flags: `-trimpath -ldflags "-s -w -buildid="` and embed version/hash via `-ldflags "-X main.version=... -X main.commit=..."`.

### Integrity Strategy
- Hash normalization (sorted active records) reused across DNS artifact stamping & integrity command.
- `manifest.json` enumerates artifact hashes for quick verification.

END IMPLEMENTATION STACK DECISION
---
END OF DOCUMENT
