## SB29-guard Technical Reference

This document contains advanced / implementation details for developers and technical operators. Non-technical users should start with the simplified `README.md`.

### Contents
- Policy Data Model & Schema
- Google Sheets Integration Details
- Environment Variables
- Sample Policy Rows
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

### Google Sheets Integration
Intended for districts that prefer editing a spreadsheet over Git.

Expected columns (header row):
```
domain	classification	rationale	last_review	status	source_ref	notes	expires	tags
```
Columns not present are treated as empty/optional. Extra columns are ignored (future versions may warn).

Sync Flow (planned implementation):
1. Fetch sheet (API key or service account).
2. Map rows -> records; normalize lowercase domains.
3. Validate against JSON Schema.
4. Compute canonical hash; write cache file `cache/policy.<hash>.json`.
5. Update pointer `cache/current.json`.
6. DNS generation / server uses cached structure in memory.
7. On validation failure fallback to last good cache or local file.

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

### Sample Policy Rows (for Sheets)
Header:
```
domain	classification	rationale	last_review	status	source_ref	notes	expires	tags
```
Examples:
```
kahoot.com	NO_DPA	Vendor has not signed required DPA	2025-08-08	active	TCK-1201	High classroom usage awaiting legal review		ENGAGEMENT
groupme.com	LEGAL_HOLD	Pending legal/privacy assessment due to chat features	2025-08-08	suspended	TCK-1202	Temporarily suspended pending risk evaluation	2025-12-31	COMMUNICATION,RISK
quizlet.com	EXPIRED_DPA	Prior DPA expired – renewal in progress	2025-08-08	active	TCK-1203	Allowed read-only until renewal finalized	2025-10-15	STUDY,ASSESSMENT
*.trackingwidgets.io	EXPIRED_DPA	Expired DPA; tracking script disabled	2025-08-08	active	TCK-1204	Wildcard for widget subdomains		AD_TECH
exampletool.com	PENDING_REVIEW	Awaiting initial privacy review	2025-08-08	active	TCK-1205	Teacher requested addition last week		PILOT
```

### Coverage Strategy
CI enforces per-package thresholds (currently 70% for `internal/policy` and `internal/dnsgen`). See `.github/workflows/ci.yml` for inline steps. Raise thresholds as test depth improves.

### Hashing & Integrity
- Canonical hash: SHA-256 over normalized active records (domain, classification, rationale, last_review, status)
- Exposed via CLI `hash` command.
- Intended for artifact attestation (future: signed manifests).

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
