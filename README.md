# SB29-guard

![CI](https://github.com/RiceC-at-MasonHS/SB29-guard/actions/workflows/ci.yml/badge.svg)
![Coverage](https://codecov.io/gh/RiceC-at-MasonHS/SB29-guard/branch/main/graph/badge.svg)
![Release](https://img.shields.io/github/v/tag/RiceC-at-MasonHS/SB29-guard?label=release)
![Lint](https://img.shields.io/badge/lint-golangci--lint-blue)

> Coverage: generated each run (see Actions artifacts). You can later wire a badge via a coverage service.

### Coverage Strategy
CI enforces per-package coverage thresholds inline (no external script). Current targets: 70% for `internal/policy` and `internal/dnsgen`. Threshold logic lives in the workflow (`.github/workflows/ci.yml`) so adjustments only require editing that file.
DNS-based page redirect to support teachers with requirements of Ohio SB29.

## Overview
SB29-guard helps districts comply with Ohio Senate Bill 29 by redirecting access to digital tools lacking an approved Digital Privacy Agreement to an explanatory page that:
- Names the original site
- Explains the classification reason (e.g., NO_DPA, EXPIRED_DPA)
- Provides guidance and escalation contact

## Key Features
- Single canonical policy file (YAML) or Google Sheets sync for non-technical admins
- DNS artifact generation (hosts, BIND, Unbound, RPZ; more platforms planned)
- Lightweight redirect service (container deployable)
- Privacy-preserving aggregated usage metrics (no PII)
- Machine-readable requirements & schema enforcement

## Non-Technical Admin Mode (Google Sheets)
For administrators uncomfortable with Git/YAML, SB29-guard can pull the domain policy from a protected Google Sheet on a schedule.

Two operation modes:
1. Local File Mode (default) – uses `policy/domains.yaml` under version control.
2. Sheets Sync Mode – CLI/service loads a cached copy of a Google Sheet, converts rows to records, validates against schema, and (optionally) writes an updated `domains.generated.yaml` (read-only for admins).

### Google Sheet Expected Columns
| Column | Required | Description |
|--------|----------|-------------|
| domain | yes | FQDN or wildcard `*.example.com` |
| classification | yes | Enum: NO_DPA, PENDING_REVIEW, EXPIRED_DPA, LEGAL_HOLD, OTHER |
| rationale | yes | Human-friendly explanation |
| last_review | yes | YYYY-MM-DD |
| status | yes | active or suspended |
| source_ref | no | Reference / ticket ID |
| notes | no | Internal notes (not displayed) |
| expires | no | YYYY-MM-DD (optional sunset) |
| tags | no | Comma-separated tags (A-Z0-9_-) |

### Environment Variables (.env)
Create a `.env` file (use `.env.example` template) with:
```
SB29_MODE=serve
SB29_POLICY_SOURCE=sheet          # sheet | file
SB29_SHEET_ID=<google_sheet_id>
SB29_SHEET_RANGE=Policy!A:Z       # Adjust to your tab name
SB29_SHEET_API_KEY=<api_key_or_use_service_account>
SB29_SHEET_FETCH_INTERVAL_SEC=300
SB29_CACHE_DIR=./cache
SB29_FALLBACK_POLICY=policy/domains.yaml
```
If using a service account JSON key file instead of API key:
```
SB29_GOOGLE_CREDENTIALS_JSON=./secrets/google-service-account.json
```

### Sync Flow
1. Service starts; loads `.env`.
2. If `SB29_POLICY_SOURCE=sheet`, attempt fetch (Sheet API or service account + Sheets API).
3. Convert rows -> internal record objects; normalize domains/lowercase.
4. Validate structure against JSON Schema (same constraints as YAML file).
5. If valid, compute hash & store cached JSON/YAML in `cache/policy.<hash>.json` and update symlink/marker `cache/current.json`.
6. If fetch fails or validation fails, revert to last known good cache or fall back file `policy/domains.yaml`.
7. DNS generation / server runtime uses in-memory policy from cache.

### Public Example Sheet
An example public Google Sheet (read-only) you can view / copy:

Sheet URL:
https://docs.google.com/spreadsheets/d/1UiaBnVMaDgB00H1C50VUEkssWDA7S_11Mk604S2kw4w/copy

Sheet ID (for `SB29_SHEET_ID`):
`1UiaBnVMaDgB00H1C50VUEkssWDA7S_11Mk604S2kw4w`

CSV Export (first sheet / gid 0):
`https://docs.google.com/spreadsheets/d/1UiaBnVMaDgB00H1C50VUEkssWDA7S_11Mk604S2kw4w/export?format=csv&gid=0`

Alternative (gviz) CSV endpoint:
`https://docs.google.com/spreadsheets/d/1UiaBnVMaDgB00H1C50VUEkssWDA7S_11Mk604S2kw4w/gviz/tq?tqx=out:csv&sheet=Sheet1`

> NOTE: For production, restrict access instead of leaving the sheet fully public; use a service account with explicit share.

### Sample Rows to Copy/Paste
Below are sample records you can paste into the sheet (one row per domain). Provide the header row first if your sheet is empty.

Header Row:
```
domain	classification	rationale	last_review	status	source_ref	notes	expires	tags
```

Sample Data Rows:
```
kahoot.com	NO_DPA	Vendor has not signed required DPA	2025-08-08	active	TCK-1201	High classroom usage awaiting legal review		ENGAGEMENT
groupme.com	LEGAL_HOLD	Pending legal/privacy assessment due to chat features	2025-08-08	suspended	TCK-1202	Temporarily suspended pending risk evaluation	2025-12-31	COMMUNICATION,RISK
quizlet.com	EXPIRED_DPA	Prior DPA expired – renewal in progress	2025-08-08	active	TCK-1203	Allowed read-only until renewal finalized	2025-10-15	STUDY,ASSESSMENT
```

Optional Additional Examples:
```
*.trackingwidgets.io	EXPIRED_DPA	Expired DPA; tracking script disabled	2025-08-08	active	TCK-1204	Wildcard for widget subdomains		AD_TECH
exampletool.com	PENDING_REVIEW	Awaiting initial privacy review	2025-08-08	active	TCK-1205	Teacher requested addition last week		PILOT
```

Guidance:
- Use ISO dates (YYYY-MM-DD) in `last_review` and `expires`.
- `status` should be `active` (enforced) or `suspended` (temporarily excluded from DNS outputs unless include-inactive added later).
- Tags are comma-separated; system will split, uppercase expected.
- Avoid freeform formatting (bold, colors) – only cell values are read.

Environment variable snippet (for this example sheet):
```
SB29_POLICY_SOURCE=sheet
SB29_SHEET_ID=1UiaBnVMaDgB00H1C50VUEkssWDA7S_11Mk604S2kw4w
SB29_SHEET_RANGE=Sheet1!A:I
```

Future automation (planned): a `sheet-pull` command will fetch this sheet, validate rows via JSON Schema, compute policy hash, and materialize a cached JSON/YAML version for DNS generation.

### Handling Admin Edits
- Admin edits Google Sheet cells only (no formatting dependence).
- A background log entry notes new hash when change detected.
- Optional email/slack notification (future) when policy hash changes.

### Safety / Validation Feedback
If invalid row(s) encountered, an error summary can be (future) written back to a dedicated “ValidationErrors” tab with line numbers & messages (read-only for service account).

## Quick Start (Local File Mode)
1. Copy `policy/domains.example.yaml` to `policy/domains.yaml` and edit.
2. Run: `sb29guard validate --policy policy/domains.yaml` (add `--strict=false` to bypass JSON Schema temporarily).
3. Generate DNS: `sb29guard generate-dns --policy policy/domains.yaml --mode a-record --format hosts --redirect-ipv4 10.10.10.50 --out dist/dns/hosts.txt`.
4. Deploy hosts/zone file to DNS platform.
5. Run server: `sb29guard serve --config config/app.yaml`.

## .env Example
An `.env.example` file will ship containing common variables for both modes (file & sheet). Copy to `.env` and edit sensitive values.

## Policy Versioning & Hash
The tool computes a SHA-256 hash of normalized active records. Hash + `version` from data source embedded in generated artifacts for audit.

## Aggregated Logs
Daily JSON (no PII) summarizing blocked lookups. Example:
```
{"date":"2025-08-08","policy_version":"0.1.0","entries":[{"domain":"exampletool.com","classification":"NO_DPA","count":42}]}
```

## Roadmap (Abbrev.)
## Releases
Tagged releases (vX.Y.Z) publish multi-platform binaries via GitHub Actions. Create a tag to trigger:
```
git tag v0.1.0
git push origin v0.1.0
```
Artifacts include SHA256SUMS for integrity verification.

- Additional DNS formats (pfSense, OPNsense, Infoblox, Route53, Azure, GCP)
- Google Sheets validation feedback loop
- OpenAPI spec & web UI improvement

## Contributing
See `docs/requirements.md` for enumerated FR/NFR/TST items. Pull Requests should run validation & tests locally.

## Disclaimer
This project aids compliance but does not replace legal review. District remains responsible for verifying DPA status and local policies.
