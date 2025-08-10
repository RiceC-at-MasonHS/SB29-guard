# Deployment: OPNsense (Unbound / Dnsmasq)

Status: Draft

## Overview
OPNsense can run Unbound (default) or Dnsmasq for DNS services. You can deploy SB29-guard redirects via Unbound includes, RPZ, or a hosts file for Dnsmasq.

## Prerequisites
- OPNsense 23.x+
- Access to Services > Unbound DNS (or Dnsmasq)
- Redirect host reachable from LAN (e.g., 10.10.10.50)
- Generated artifacts from `sb29guard generate-dns`

## Unbound: local-data include
1) Generate Unbound snippet:
```
sb29guard generate-dns --policy policy/domains.yaml --mode a-record --format unbound --redirect-ipv4 10.10.10.50 --out dist/opnsense/policy.conf
```
2) Copy to firewall:
```
scp dist/opnsense/policy.conf root@opnsense:/usr/local/etc/unbound.opnsense.d/sb29-guard.conf
```
3) In UI: Services > Unbound DNS > Advanced > Custom options, ensure includes path is recognized (OPNsense includes `unbound.opnsense.d` automatically). Apply.

## Unbound: RPZ
1) Generate RPZ zone as BIND format:
```
sb29guard generate-dns --policy policy/domains.yaml --mode rpz --format bind --redirect-host blocked.guard.local --redirect-ipv4 10.10.10.50 --out dist/opnsense/rpz.zone
```
2) Copy to `/usr/local/etc/unbound.opnsense.d/rpz.zone` and reference via `response-policy` in custom options:
```
server:
  response-policy:
    zone:
      name: "rpz.sb29guard"
      zonefile: "/usr/local/etc/unbound.opnsense.d/rpz.zone"
```
3) Apply.

## Dnsmasq: hosts file
1) Generate hosts list:
```
sb29guard generate-dns --mode a-record --format hosts --redirect-ipv4 10.10.10.50 --out dist/opnsense/hosts.txt
```
2) Place at `/usr/local/etc/dnsmasq.d/sb29-guard.hosts` and ensure `addn-hosts=/usr/local/etc/dnsmasq.d/sb29-guard.hosts` is set. Restart Dnsmasq.

## Verification
- Tools > Diagnostics > DNS Lookup: `exampletool.com` resolves to redirect host.
- Clients browsing to blocked domain see the explanation page.

## Header inference notes
- A/AAAA overrides: consider `SB29_ALLOW_HOST_FALLBACK=true`.
- CNAME/RPZ: inject `X-Original-Host` with a reverse proxy.

## Rollback
Keep previous file versions and restore on misconfig; re-apply service.
