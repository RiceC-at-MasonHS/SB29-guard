# Deployment: Infoblox (BIND-based)

Status: Draft

## Overview
Infoblox NIOS appliances are BIND-based. You can deploy SB29-guard with an override zone (A/AAAA or CNAME) or RPZ. This guide focuses on importing generated artifacts and scheduling updates.

## Prerequisites
- Access to Infoblox Grid Manager
- A redirect host (blocked.guard.local) and IP(s)
- Artifacts generated via `sb29guard generate-dns`

## Options
- Override Zone (master): A-record overrides or CNAMEs
- RPZ (Response Policy Zone)

## Generating Artifacts
- CNAME zone (recommended for manageability):
```
sb29guard generate-dns --policy policy/domains.yaml --mode cname --redirect-host blocked.guard.local --format bind --out dist/infoblox/zone.db
```
- A-record override zone:
```
sb29guard generate-dns --policy policy/domains.yaml --mode a-record --redirect-ipv4 10.10.10.50 --format bind --out dist/infoblox/zone.db
```
- RPZ zone:
```
sb29guard generate-dns --policy policy/domains.yaml --mode rpz --redirect-host blocked.guard.local --redirect-ipv4 10.10.10.50 --format bind --out dist/infoblox/rpz.zone
```

## Importing into Infoblox
1) Create a new Zone (Authoritative for override/CNAME or RPZ for policy) in Grid Manager.
2) Use Data Management > DNS > Zone > Import to upload the BIND-format file.
3) Ensure SOA/NS records in the file align with your naming (generator sets placeholders you may edit).
4) Save and apply changes; Infoblox will distribute across the grid.

## Scheduling Updates
- Use a CI pipeline to regenerate files and push via Infoblox APIs (WAPI) or manual upload.
- Ensure serial increases with each update (the generator handles serial bumping).

## Verification
- Test resolution using Tools > NS Lookup or from a client.
- Expect A to redirect IP (a-record) or CNAME to `blocked.guard.local` (cname/rpz).

## Header inference notes
- A/AAAA overrides: may enable `SB29_ALLOW_HOST_FALLBACK=true` on the web app if no proxy injects headers.
- CNAME/RPZ: place a reverse proxy ahead of the app and inject `X-Original-Host`.

## Rollback
Keep prior imported file versions; re-import to revert.
