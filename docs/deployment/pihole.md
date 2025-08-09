# Deployment: Pi-hole / dnsmasq Hosts File

Status: Draft

## Overview
Pi-hole can ingest a hosts-format file mapping restricted domains to the redirect IP.

## Prerequisites
- Pi-hole >= v5
- Redirect host reachable from LAN (e.g., 10.10.10.50)
- Generated hosts file from: `sb29guard generate-dns --mode a-record --format hosts`

## Generation (Example)
```
sb29guard generate-dns --policy policy/domains.yaml \
  --out dist/dns/pihole-hosts.txt \
  --mode a-record \
  --redirect-ipv4 10.10.10.50 \
  --format hosts
```

## Sample Hosts Output
```
# sb29guard policy_version=0.1.0 sha256=<HASH>
10.10.10.50 exampletool.com
10.10.10.50 trackingwidgets.io
10.10.10.50 sub.trackingwidgets.io
```

## Import Strategies
1. Local Custom List: Copy file to `/etc/pihole/custom.list` (overwrites on regen).
2. Adlist Subscription (Optional): Host file via internal HTTP and add URL to Pi-hole adlists (noting Pi-hole expects one domain per line; requires transform mode that outputs only domains, no IP).

### Option 1 (Direct Replace)
```
scp dist/dns/pihole-hosts.txt pi-hole:/etc/pihole/custom.list
ssh pi-hole 'pihole restartdns reload'
```

### Option 2 (Adlist Style)
Generate domain-only list:
```
sb29guard generate-dns --mode a-record --format domain-list --out dist/dns/pihole-domains.txt
```
Add resulting internal URL to Pi-hole web admin (Group Management -> Adlists).

## Verification
```
dig exampletool.com @<pihole-ip>
# Expect A 10.10.10.50
```
Navigate to blocked domain in browser; expect redirect explanation page.

### Verification checklist
- DNS returns redirect record:
  - `dig exampletool.com @<pihole-ip>` shows `A 10.10.10.50`
- Web server health:
  - `GET http://<redirect-host>:8080/health` returns `{ "status": "ok" }`
  - `GET http://<redirect-host>:8080/metrics` shows `policy_version`, `record_count`, and refresh stats
- Policy version and record count match expectations (compare with CLI `sb29guard hash`)
- Pi-hole DNS reload applied (`pihole restartdns reload`)
- Rollback plan confirmed (backup `custom.list` retained)

## Updating
Automate via cron or CI pipeline:
```
0 * * * * /usr/local/bin/sb29guard generate-dns --policy /opt/sb29/policy/domains.yaml --out /opt/sb29/dist/dns/pihole-hosts.txt --mode a-record --redirect-ipv4 10.10.10.50 --format hosts && cp /opt/sb29/dist/dns/pihole-hosts.txt /etc/pihole/custom.list && pihole restartdns reload
```

## Privacy Considerations
Pi-hole query logs may contain client IPs; configure retention per district policy (consider 24h or less) and restrict dashboard access.

## Troubleshooting
- If domain still resolves to original IP, flush client DNS cache (ipconfig /flushdns, browser restart).
- Check Pi-hole `pihole -t` for live query to ensure block applied.

## Rollback
Restore previous `custom.list` backup and reload DNS.
