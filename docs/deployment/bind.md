# Deployment: BIND (Named) with A Record Overrides

Status: Draft

## Overview
This guide shows how to deploy SB29-guard using BIND with direct A record overrides or RPZ.

## Prerequisites
- BIND 9.10+ (RPZ recommended 9.11+)
- Access to internal recursive resolver config
- Redirect host (web server) reachable IP(s): e.g. 10.10.10.50 / 2001:db8:10::50
- Generated artifacts from `sb29guard generate-dns`

## Modes
1. A Record Override Zone
2. CNAME Consolidation Zone
3. RPZ (Response Policy Zone) with CNAME rewrite

## Directory Layout Example
```
/etc/named/
  sb29-guard/
    zone.override            # A/CNAME zone (mode a-record or cname)
    rpz.zone                 # RPZ zone file (mode rpz)
```

## 1. A Record Override
### Sample Generated Zone Snippet
```
$TTL 300
@   IN SOA guard.school.local. hostmaster.school.local. (
        2025080801 ; serial
        3600       ; refresh
        900        ; retry
        604800     ; expire
        300 )      ; minimum
    IN NS  ns.guard.school.local.
exampletool.com.          300 IN A 10.10.10.50
trackingwidgets.io.       300 IN A 10.10.10.50
sub.trackingwidgets.io.   300 IN A 10.10.10.50 ; if wildcard expanded
```

### named.conf Include
```
zone "override.internal" IN {
  type master;
  file "/etc/named/sb29-guard/zone.override";
  allow-update { none; };
};
```
Use search list or explicit zone design so queries for listed FQDNs match.

## 2. CNAME Consolidation
```
$TTL 300
@   IN SOA guard.school.local. hostmaster.school.local. (
        2025080801 3600 900 604800 300 )
    IN NS  ns.guard.school.local.
exampletool.com.     300 IN CNAME blocked.guard.local.
trackingwidgets.io.  300 IN CNAME blocked.guard.local.
blocked.guard.local. 300 IN A 10.10.10.50
```
Pros: Single A record to change redirect IP. Cons: Extra lookup (minor).

## 3. RPZ Deployment
### named.conf
```
response-policy { zone "rpz.sb29guard"; } break-dnssec yes;
zone "rpz.sb29guard" {
  type master;
  file "/etc/named/sb29-guard/rpz.zone";
};
```

### RPZ File Example
```
$TTL 300
@ IN SOA guard.school.local. hostmaster.school.local. (2025080801 3600 900 604800 300)
@ IN NS ns.guard.school.local.
; policy_version=0.1.0 sha256=<HASH>
exampletool.com. CNAME blocked.guard.local.
trackingwidgets.io. CNAME blocked.guard.local.
*.trackingwidgets.io. CNAME blocked.guard.local.
blocked.guard.local. A 10.10.10.50
```

## Reloading BIND
```
rndc reload
# or specific zone
rndc reload rpz.sb29guard
```

## Verifying
```
dig exampletool.com @127.0.0.1
# Expect CNAME or A to redirect host
```

## Logging Considerations
Disable query logging for standard traffic if not required; rely on aggregated web logs.

## Security Notes
- RPZ can interfere with DNSSEC; `break-dnssec yes;` required when rewriting signed domains.
- Restrict file permissions (640) and directory (750) to named user.

## Automation Workflow
1. Update `policy/domains.yaml`.
2. Run validator.
3. Generate zone files.
4. Copy artifacts to BIND server path.
5. Increment serials automatically (generator handles).
6. Reload zones.

## Rollback
Keep previous artifact set; revert file and reload.

## Monitoring
Use `dig +trace` for debugging mismatches. Ensure TTL manageable (<=300) for quick policy changes.
