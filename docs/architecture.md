# Architecture Overview

Status: Draft (see ADR-0001 for proxy-first decision)

## Components
1. Policy Dataset (YAML) + Schema
2. CLI Tool (Policy validation, DNS artifact generation, classification lookup, serving redirect service)
3. Redirect Web Service (Static+Dynamic explanation page, metrics JSON)
4. DNS Integration Artifacts (BIND, Unbound, Pi-hole, Windows DNS, RPZ)
5. Metrics & (future) Aggregation (in-memory counters, optional summaries)
6. Deployment Assets (Dockerfile, compose, example systemd unit)

## Flow (Request to Blocked Domain)
1. User device queries disallowed domain.
2. Internal DNS returns redirect host IP (A/AAAA) or CNAME chain to redirect host.
3. Browser requests `http(s)://original-domain/...` but traffic is handled by a proxy/gateway.
4. Web server (sb29-guard) resolves the original domain using header inference with this precedence:
  a. `X-Original-Host`
  b. first `X-Forwarded-Host`
  c. `Referer` (host portion)
  d. Optional query param `d` (display-only) if no informative header is present
  e. `Host` header only if explicitly enabled via SB29_ALLOW_HOST_FALLBACK=true
  Notes:
  - Query params d/c/v/h are validated and affect display only; classification still comes from the in-memory policy.
5. Response: Explanation page (HTML) or JSON (if API call) with rationale.
6. Metrics counters update (refresh events visible at /metrics).

## Data Model (Policy Record)
```
Record {
  domain: string
  classification: enum
  rationale: string
  last_review: date
  status: enum
  notes?: string
  source_ref?: string
  expires?: date
  tags?: string[]
}
```

## DNS Generation Modes
- a-record: Direct A/AAAA mapping per domain.
- cname: CNAME to central blocked host name.
- rpz: CNAME rewrite using Response Policy Zone.

## Redirect Strategies
- Model A: Header-injection reverse proxy (preferred). Proxy terminates TLS and forwards to sb29-guard while setting `X-Original-Host: <blocked-domain>`.
- Model B: 302 redirect to a static explain site. Proxy redirects to `/explain?d=<blocked-domain>&c=<classification>&v=<version>&h=<hash>` on a host you control with a valid cert.
  - In both models, sb29-guard treats d/c/v/h as display-only and resolves classification from policy.

See ADR: `docs/adr/0001-proxy-first.md`.

## Reliability Safeguards
- Static fallback page if dynamic data load fails.
- Graceful policy reload (SIGHUP or CLI trigger).
- Atomic write of regenerated artifacts (write temp + move).

## Security Controls
- Strict CSP, no third-party resources.
- Optional basic auth for admin endpoints only.
- Hash + version embedding in generated files for tamper detection.

## Extensibility
- Additional DNS formats by implementing interface `DnsWriter`.
- Additional classification enums allowed with minor schema update.
- Localization via message catalogs (JSON/YAML key-value).

## Sequence (DNS Generation)
1. Load & validate policy.
2. Compute content hash (SHA-256) of normalized records.
3. Determine policy_version (Git tag or derived from version + hash prefix).
4. For each record (active) produce lines based on mode.
5. Insert header comment with metadata (policy_version, hash, timestamp, tool version) into output.
6. Write file atomically to target path.

## Performance Considerations
- Keep policy file small (<=10k domains) for quick startup.
- Use radix/prefix tree or wildcard aware matcher for Host header lookups.

## Logging Format (Aggregated)
```
{
  "date": "2025-08-08",
  "policy_version": "0.1.0",
  "entries": [
    { "domain": "exampletool.com", "classification": "NO_DPA", "count": 42 },
    { "domain": "trackingwidgets.io", "classification": "EXPIRED_DPA", "count": 8 }
  ]
}
```

## Open Questions
- Should policy_version auto-bump via Git commit hook?
- Provide signed JSON (JWS) for aggregated logs?
- Add optional caching layer for high RPS scenarios?
