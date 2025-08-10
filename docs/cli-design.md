# CLI Design Specification

Status: Draft

Binary Name: `sb29guard`

## Command Overview
```
sb29guard
  validate       Validate policy (YAML or published CSV)
  generate-dns   Produce DNS artifacts (hosts/bind/unbound/rpz)
  serve          Start redirect web service
  hash           Output normalized policy hash & version metadata
```

## Global Flags
- `--policy <path>`: Override default policy file path.
- `--config <path>`: Config file path (for serve).
- `--log-level <level>`: trace|debug|info|warn|error (default info).
- `--format json|text`: Output format for machine readability.
- `--no-color`: Disable ANSI color.

## validate
Validate the policy file.
Flags:
- `--strict` (default true) enforce JSON Schema (set false for transitional validation)
Exit Codes:
- 0 success
- 1 schema invalid
- 2 IO error
Output (json format):
```
{
  "status": "ok",
  "records": 128,
  "hash": "<sha256>",
  "version": "0.1.0"
}
```

## generate-dns
Flags:
- `--out <file|dir>` (required)
- `--mode a-record|cname` (default a-record)
- `--format hosts|bind|unbound|rpz|dnsmasq|domain-list|winps` (subset depends on mode)
- `--redirect-ipv4 <ip>` (required for a-record/hosts)
- `--redirect-ipv6 <ip>` (optional)
- `--redirect-host <fqdn>` (required for cname/rpz)
- `--ttl <seconds>` (default 300)
- `--dry-run` (prints to stdout)
- `--serial-strategy date|epoch|hash` (default date: YYYYMMDDNN)

Additional Flags (new):
- `--classification-filter CLASS[,CLASS...]` Limit output to specific classifications.
- `--include-inactive` Include suspended records (override default exclusion).
- `--manifest-out <path>` Override default manifest path (`dist/dns/manifest.json`).

Supported Formats (expanded):
Note: The following list includes exploratory targets; for the authoritative, implemented set in v1.x, see `sb29guard generate-dns --help`.
- `pfSense-unbound`
- `opnsense-unbound`
- `infoblox-rpz`
- `route53-json`
- `azure-cli`
- `gcloud-dns`
- `domain-list` (cloud security products)

Output Header Comment Example:
```
# sb29guard policy_version=0.1.0 hash=<SHA256> generated=2025-08-08T12:00:00Z tool_version=0.1.0
```

## serve
Flags (override config):
- `--policy <path>` or `--sheet-csv <url>` (data source)
- `--listen <addr>` (e.g., `:8080`)
- `--templates <dir>` (override embedded templates/CSS)
- Refresh scheduling (when using `--sheet-csv`):
  - `--refresh-at HH:MM` (daily, local time)
  - `--refresh-every <duration>` (e.g., `30m`, `2h`)

Endpoints:
- `GET /` human-friendly landing.
- `GET /explain` explanation page (HTML).
- `GET /law` 302 to configured law URL (default LIS PDF; override via SB29_LAW_URL).
- `GET /health` liveness probe (200 + minimal JSON).
- `GET /metrics` JSON metrics (policy_version, record_count, refresh stats).

Auto-refresh behavior (current):
- When started with `--sheet-csv`, the server schedules a daily refresh at 23:59 local time.
- Successful refresh hot-swaps in-memory policy; failures log JSON error events and retain the last known-good policy.

## Notes
- The service looks up classification from in-memory policy; no public classify command is currently exposed.

## hash
Computes canonical hash of sorted active records (domain + classification + rationale + last_review + status + optional fields normalized).
Flags:
- `--strict` (default true) enforce JSON Schema before hashing
Output JSON includes: hash, record_count, version, updated.

## export-schema
Prints embedded policy JSON Schema to stdout (machine retrieval), enabling external validators.

## demo-data
Writes a minimal `domains.yaml` if one does not exist (safe create; refuses overwrite unless `--force`).

## sheet CSV notes
Published CSV mode uses ETag/Last-Modified caching and retries; metrics record refresh_count and last_refresh_source (csv|csv-cache).

## JSON Logging Structure
```
{"ts":"2025-08-08T12:00:00Z","level":"info","event":"server.start","port":8080,"policy_version":"0.1.0"}
```

## Error Handling Patterns
- Validation errors: structured list with path and message.
- CLI returns non-zero codes; never swallow errors silently.

## Concurrency
DNS generation may process records concurrently but final output order must be deterministic (sorted by domain) to keep stable hashes.

## Extensibility Interfaces (Pseudo)
```
interface PolicyLoader { Load(path) -> Policy }
interface DnsWriter { Write(policy, options) -> string }
interface Matcher { Match(host) -> Record? }
```

## Performance Targets
- Validate 5k records < 0.5s on modest hardware.
- Generate DNS artifacts for 5k records < 1s.

## Security
- No dynamic code execution.
- Input sanitization for domains & query parameters.

## Telemetry
- None by default; explicit flag needed for any anonymous usage stats (not planned initial).

## Manifest File Schema
`dist/dns/manifest.json` example:
```
{
  "generated":"2025-08-08T12:00:00Z",
  "policy_version":"0.1.0",
  "hash":"<sha256>",
  "artifacts":[
    {"path":"dist/dns/hosts.txt","mode":"a-record","format":"hosts","sha256":"...","bytes":1234},
    {"path":"dist/dns/rpz.zone","mode":"rpz","format":"rpz","sha256":"...","bytes":4567}
  ],
  "tool_version":"0.1.0"
}
```

## Environment Variables (Reference Extract)
```
SB29_MODE=serve|once
SB29_POLICY_SOURCE=file|sheet
SB29_SHEET_ID=
SB29_SHEET_RANGE=Policy!A:Z
SB29_SHEET_API_KEY=
SB29_GOOGLE_CREDENTIALS_JSON=./secrets/google.json
SB29_SHEET_FETCH_INTERVAL_SEC=300
SB29_CACHE_DIR=./cache
SB29_FALLBACK_POLICY=policy/domains.yaml
```
