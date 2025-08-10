# Deployment: pfSense (Unbound / DNS Resolver)

Status: Draft

## Overview
pfSense uses Unbound as its DNS Resolver. You can deploy SB29-guard redirects via local-data or RPZ. This guide covers both approaches and where to place generated files.

## Prerequisites
- pfSense 2.5+
- Access to pfSense WebGUI (Services > DNS Resolver)
- Redirect host reachable from LAN (e.g., 10.10.10.50)
- Generated artifacts from `sb29guard generate-dns`

## Option A: Local Overrides (host records)
If you have a small list, you can use pfSense Host Overrides (GUI). For large lists, use Unbound includes.

### Using Unbound includes
1) Generate an Unbound snippet:
```
sb29guard generate-dns \
  --policy policy/domains.yaml \
  --mode a-record \
  --redirect-ipv4 10.10.10.50 \
  --format unbound \
  --out sb29-unbound/policy.conf
```
2) Copy `policy.conf` to pfSense:
- scp to `/var/unbound/sb29-guard/policy.conf`
- Ensure directory exists and permissions are 644 (file) and 755 (dir)

3) In pfSense: Services > DNS Resolver > Custom options:
```
server:
  include: "/var/unbound/sb29-guard/policy.conf"
```
4) Save & Apply. Check `Status > System Logs > System > General` for Unbound reload messages.

## Option B: RPZ
1) Generate RPZ zone:
```
sb29guard generate-dns --policy policy/domains.yaml --mode rpz --format bind --out sb29-bind/rpz.zone --redirect-host blocked.guard.local --redirect-ipv4 10.10.10.50
```
2) Copy `rpz.zone` to `/var/unbound/sb29-guard/rpz.zone`.
3) In pfSense Custom options:
```
server:
  response-policy:
    zone:
      name: "rpz.sb29guard"
      zonefile: "/var/unbound/sb29-guard/rpz.zone"
```
4) Save & Apply.

## Verification
- Diagnostics > DNS Lookup: query `exampletool.com` and expect A 10.10.10.50 or a CNAME per your mode.
- From a client: `nslookup exampletool.com <pfsense-ip>`.

## Header inference notes
- A/AAAA overrides: you may enable `SB29_ALLOW_HOST_FALLBACK=true` so the app uses the `Host` header as original domain.
- CNAME/RPZ: keep Host fallback disabled; add a reverse proxy that injects `X-Original-Host`.

## Automation
Use a CI or cron to SCP updated files and issue a config reload by toggling DNS Resolver or using `unbound-control reload` via SSH.

## Rollback
Keep previous files under `/var/unbound/sb29-guard/backup/` and restore as needed before applying changes.
