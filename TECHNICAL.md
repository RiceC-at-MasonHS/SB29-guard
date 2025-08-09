## SB29-guard Technical Reference  
[README](./README.md) • [Customizing](./CUSTOMIZING.md) • [Contributing](./CONTRIBUTING.md)

This document contains advanced / implementation details for developers and technical operators. Non-technical users should start with the simplified `README.md`.

### Contents
- Policy Data Model & Schema
- Google Sheets Integration Details
- Environment Variables
- Sample Policy Rows
- Templating & Embedding
- Coverage Strategy & Quality Gates
- Hashing / Integrity
- Aggregated Logs Format
- Roadmap (Abbrev.)
- Contributing

### Policy Data Model
See `internal/policy/policy.schema.json` for authoritative JSON Schema. Key fields per record:

| Field | Type | Notes |
|-------|------|-------|
| domain | string | FQDN or wildcard `*.example.com` |
| classification | enum | `NO_DPA`, `PENDING_REVIEW`, `EXPIRED_DPA`, `LEGAL_HOLD`, `OTHER` |
| rationale | string | Plain-language reason users see |
| last_review | date | `YYYY-MM-DD` |
| status | enum | `active` or `suspended` |
| source_ref | string? | Ticket / reference ID |
| notes | string? | Internal only (not exposed) |
| expires | date? | Optional sunset date |
| tags | array | Optional metadata |

### Google Sheets Integration (Published CSV – Implemented v0.1)
Current mechanism uses a published CSV URL (no API keys) provided by Google Sheets “Publish to the web” feature.

Usage flags (mutually exclusive with --policy):
```
sb29guard validate --sheet-csv <csv_url>
sb29guard hash --sheet-csv <csv_url>
sb29guard generate-dns --sheet-csv <csv_url> --format hosts --mode a-record --redirect-ipv4 10.10.10.50 --dry-run
sb29guard serve --sheet-csv <csv_url>
```
Caching:
- Files stored under `cache/sheets/<fnvhash>.csv` + metadata JSON (ETag, Last-Modified).
- Conditional requests (If-None-Match / If-Modified-Since) reduce bandwidth; when unchanged, server prints `source":"csv-cache"`.
- Safety limits: 5MB max read.
- Retry/backoff: up to 3 attempts (500ms, 1s, 2s delays) on transient errors.

Server auto-refresh:
- In `serve` mode with `--sheet-csv`, a background scheduler refreshes the CSV daily at 23:59 local time.
- On success, server swaps the active policy in-memory using a RWMutex-protected `UpdatePolicy` to avoid races.
- Failures (HTTP errors, parse/validation) emit structured JSON logs and leave the current policy active.
- Log events: `policy.refresh.scheduled` (next run time), `policy.refresh.start`, `policy.refresh.success` (records, source csv|csv-cache, version), `policy.refresh.error` (message).

Columns:
Required: `domain, classification, rationale, last_review, status` (case-insensitive).  
Optional: `source_ref, notes, expires, tags` (tags comma-separated inside one cell; internally split & sorted).

Validation & Normalization:
- Domains lowercased, wildcards allowed only as leading label (`*.example.com`).
- Suspended records excluded from canonical hash and DNS outputs.

Planned Enhancements:
- Configurable refresh interval/clock time via flag or config file; immediate `--refresh-now` trigger.
- Error write-back to separate sheet/tab.
- Classification filtering and explicit suspended inclusion flag.
- Manifest of CSV provenance (timestamp, hash) for audit.

### Environment Variables (Planned / Future)
```
SB29_POLICY_SOURCE=sheet | file
SB29_SHEET_ID=<google_sheet_id>
SB29_SHEET_RANGE=Policy!A:Z
SB29_SHEET_API_KEY=<api_key>
SB29_GOOGLE_CREDENTIALS_JSON=./secrets/service-account.json
SB29_SHEET_FETCH_INTERVAL_SEC=300
SB29_CACHE_DIR=./cache
SB29_FALLBACK_POLICY=policy/domains.yaml
```
File mode typically only needs the fallback policy path.

### Sample Policy Rows (Sheets CSV)
Header:
```
domain,classification,rationale,last_review,status,source_ref,notes,expires,tags
```
Examples:
```
kahoot.com	NO_DPA	Vendor has not signed required DPA	2025-08-08	active	TCK-1201	High classroom usage awaiting legal review		ENGAGEMENT
groupme.com	LEGAL_HOLD	Pending legal/privacy assessment due to chat features	2025-08-08	suspended	TCK-1202	Temporarily suspended pending risk evaluation	2025-12-31	COMMUNICATION,RISK
quizlet.com	EXPIRED_DPA	Prior DPA expired – renewal in progress	2025-08-08	active	TCK-1203	Allowed read-only until renewal finalized	2025-10-15	STUDY,ASSESSMENT
*.trackingwidgets.io	EXPIRED_DPA	Expired DPA; tracking script disabled	2025-08-08	active	TCK-1204	Wildcard for widget subdomains		AD_TECH
exampletool.com	PENDING_REVIEW	Awaiting initial privacy review	2025-08-08	active	TCK-1205	Teacher requested addition last week		PILOT
```

### Templating & Embedding
Runtime UI (root + explanation pages) uses Go `html/template` with three files: `layout.html`, `root.html`, `explain.html` plus `style.css`. All are embedded using `//go:embed` so no external assets are required. Snapshot copies for documentation: `docs/templates/`. Future enhancement: optional `--templates` directory override.

### Coverage Strategy
CI enforces per-package thresholds (currently 70%+ for `internal/policy`; `internal/dnsgen` tracked similarly). Other packages (server, hash, CLI) have growing coverage but are not yet gate-enforced. Policy, DNS generation, and server negative paths are explicitly tested; hash utility 100%.

### Hashing & Integrity
Canonical hash: SHA-256 over normalized ACTIVE records only (suspended excluded). Fields: domain, classification, rationale, last_review, status (normalized + newline joined). Exposed via CLI `hash` command for audit trails. Planned: signed manifest for attestation.

### Aggregated Logs (Planned)
Daily JSON structure (no PII):
```json
{
  "date":"2025-08-08",
  "policy_version":"0.1.0",
  "entries":[{"domain":"exampletool.com","classification":"NO_DPA","count":42}]
}
```

### Roadmap (Abbrev.)
- Additional DNS formats (pfSense, OPNsense, Infoblox, Route53, Azure, GCP)
- Sheet validation feedback loop (write errors to separate tab)
- Signed artifacts & manifest
- OpenAPI + richer web UI
- Metrics endpoint / Prometheus

### Contributing
1. Fork & branch
2. `go test ./...` (ensure per-package coverage passes)
3. Run linter (CI will enforce) – supply comments for exported symbols
4. Submit PR referencing requirement IDs (see `docs/requirements.md`).

### License
See `LICENSE`.
