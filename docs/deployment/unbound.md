# Deployment: Unbound Local Zones / RPZ

Status: Draft

## Overview
Unbound can block / redirect domains using local-data or Response Policy Zones (RPZ, Unbound 1.16+). We map restricted domains to a redirect host.

## Prerequisites
- Unbound 1.13+ (RPZ needs newer for full features)
- Access to unbound.conf
- Redirect host IP(s) (IPv4/IPv6)
- Generated config snippets from `sb29guard generate-dns --mode a-record|rpz --format unbound`

## Option 1: local-zone + local-data
```
server:
  include: "/etc/unbound/sb29-guard/*.conf"
```
Generated snippet example (`/etc/unbound/sb29-guard/policy.conf`):
```
# policy_version=0.1.0 sha256=<HASH>
local-zone: "exampletool.com" redirect
local-data: "exampletool.com A 10.10.10.50"
local-zone: "trackingwidgets.io" redirect
local-data: "trackingwidgets.io A 10.10.10.50"
```
Note: `redirect` sends all names below domain to specified A/AAAA unless more specific.

Wildcard Handling: Generator may choose `local-zone: trackingwidgets.io redirect` to cover `*.trackingwidgets.io`.

## Option 2: RPZ
Add to unbound.conf:
```
server:
  response-policy:
    zone:
      name: "rpz.sb29guard"
      zonefile: "/etc/unbound/sb29-guard/rpz.zone"
```
RPZ zone example:
```
$TTL 300
@ SOA guard.school.local. hostmaster.school.local. 2025080801 3600 900 604800 300
@ NS ns.guard.school.local.
exampletool.com CNAME blocked.guard.local.
trackingwidgets.io CNAME blocked.guard.local.
blocked.guard.local A 10.10.10.50
```

## Reloading
```
unbound-control reload
```

## Verification
```
unbound-control lookup exampletool.com
```
Expect redirect IP or CNAME chain.

Header inference notes:
- If you deploy as A/AAAA overrides, you can set `SB29_ALLOW_HOST_FALLBACK=true` so the app uses the `Host` header for domain detection.
- If you deploy via CNAME to a consolidated host, keep Host fallback disabled and add a reverse proxy that injects `X-Original-Host` or `X-Forwarded-Host`.

### Verification checklist
- DNS returns redirect record:
  - `unbound-control lookup exampletool.com` shows `A 10.10.10.50` or `CNAME blocked.guard.local.`
- Web server health:
  - `GET http://<redirect-host>:8080/health` returns `{ "status": "ok" }`
  - `GET http://<redirect-host>:8080/metrics` shows `policy_version`, `record_count`, and refresh stats
- Policy version and record count match expectations (compare with CLI `sb29guard hash`)
- TTLs reasonable (e.g., 300s) and config reload applied (`unbound-control reload`)
- Rollback plan confirmed (previous conf/zone retained)

## Performance Considerations
- Keep total restricted domains list moderate; local-data is in-memory and efficient.
- Use TTL <=300 for responsive policy changes.

## Privacy
Disable extended query logging. Use only aggregated web redirect logs.

## Automation
```
sb29guard generate-dns --policy policy/domains.yaml --mode a-record --format unbound --out dist/dns/unbound/policy.conf --redirect-ipv4 10.10.10.50
rsync dist/dns/unbound/ server:/etc/unbound/sb29-guard/
ssh server 'unbound-control reload'
```

## Rollback
Retain previous `policy.conf` and restore if needed; reload.
