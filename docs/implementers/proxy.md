# Proxy/Gateway integration (School Mode — recommended)

Status: Stable recommendation for K–12 networks

This guide shows how to integrate SB29-guard with your school’s web filter/forward proxy so teachers and students get a seamless, friendly explanation page with zero warnings or extra clicks.

Operator checklist
- Pick your proxy: NGINX, Caddy, HAProxy, Apache (see quickstarts below)
- Choose a model: header-injection (preferred) or redirect to static explain
- Set a vhost (e.g., blocked.school.local) with a trusted cert
- Forward X-Original-Host (and X-Forwarded-Host) to sb29-guard OR send 302 to your explain host
- Verify: blocked domain shows the explain page; non-listed returns 404 (pass-through)

Hands-off (set-and-forget)
- Policy source: default is a YAML file; alternatively use a Google Sheet published as CSV with `--sheet-csv <url>`.
- Auto-refresh: when started with `--sheet-csv`, the server refreshes nightly at 23:59 by default (configurable via `--refresh-at HH:MM` or `--refresh-every 2h`). It hot-swaps the in-memory policy without restarts.
- No proxy reloads: in header-injection mode the proxy just forwards; updated classifications take effect immediately in SB29‑guard. Keep proxy config static.
- Resilience: conditional requests (ETag/Last-Modified), on-disk cache, last-known-good policy retained on failures. Check `/metrics` for refresh stats and last error.

Quickstarts
- NGINX: docs/implementers/nginx-quickstart.md
- Caddy: docs/implementers/caddy-quickstart.md
- HAProxy: docs/implementers/haproxy-quickstart.md
- Apache httpd: docs/implementers/apache-quickstart.md
- GUI-driven proxies (APIs/lists): docs/implementers/gui-proxy.md

Why proxy-first?
- HTTPS reality: DNS-only CNAME/A overrides for blocked sites won’t display a friendly page on HTTPS due to certificate/SNI mismatch. The browser fails TLS before any redirect.
- Proxies solve this: School gateways terminate TLS (using a trusted internal CA on managed devices) and can inject headers or send controlled 302 redirects.

Core contract
- Provide the original requested host to SB29-guard by either:
  1) Header-injection reverse proxy: Set X-Original-Host: <blocked-domain> and forward to sb29-guard.
  2) 302 redirect to an explanation host: https://explain.school.example/explain?d=<blocked-domain>&c=<classification>&v=<version>&h=<hash> (all display-only)
- SB29-guard’s inference precedence (authoritative):
  X-Original-Host > first X-Forwarded-Host > Referer(host) > Host (only if SB29_ALLOW_HOST_FALLBACK=true). Query params (d/c/v/h) are display-only.
  
See also
- CLI to generate configs: docs/feature-specs/generate-proxy.md
- Static explain page bundle: docs/feature-specs/generate-explain-static.md

Security and privacy
- SB29-guard and static pages strictly validate and escape the domain (hostname only; IDNA/punycode safe; length and charset limits).
- Explanation responses include Cache-Control: no-store.
- No analytics/PII by default; metrics avoid logging d unless explicitly enabled in future.

Model A: Header-injection reverse proxy (preferred when proxy can forward)
- Flow: Proxy terminates TLS → forwards to sb29-guard with X-Original-Host set → sb29-guard renders.
- Keep SB29_ALLOW_HOST_FALLBACK=false (safer) and rely on explicit headers.

Examples (minimal, vendor-agnostic)

Caddy
```
# guard.school.internal serves SB29-guard; inject original host
example.org {
  reverse_proxy 127.0.0.1:8080 {
    header_up X-Original-Host {host}
  }
}
```

NGINX
```
server {
  listen 443 ssl;
  server_name guard.school.internal;
  # ssl_certificate ...; ssl_certificate_key ...;

  location / {
    proxy_set_header X-Original-Host $host;
    proxy_pass http://127.0.0.1:8080;
    add_header Cache-Control "no-store" always;
  }
}
```

HAProxy
```
frontend fe_guard
  bind :443 ssl crt /etc/haproxy/certs
  default_backend be_guard

backend be_guard
  http-request set-header X-Original-Host %[req.hdr(Host)]
  server s1 127.0.0.1:8080 check
```

Apache httpd
```
<VirtualHost *:443>
  ServerName guard.school.internal
  SSLEngine on
  # SSLCertificateFile ... SSLCertificateKeyFile ...
  RequestHeader set X-Original-Host "%{Host}e" env=HTTPS
  ProxyPass / http://127.0.0.1:8080/
  ProxyPassReverse / http://127.0.0.1:8080/
  Header always set Cache-Control "no-store"
</VirtualHost>
```

Model B: 302 redirect to a static explanation page (Pages or similar)
- Flow: Proxy detects a blocked domain → 302 Location to explain host with query params → static page renders.
- Recommended when your proxy supports external redirect URLs easily.

Param schema
- d: required, original domain (hostname only)
- c: optional classification key (e.g., unapproved)
- v: optional policy version
- h: optional short policy hash

Examples

Caddy (dedicated redirect endpoint)
```
block.school.internal {
  @blocked path /blocked
  handle @blocked {
    # {host} is the original host in this simplified example
    redir 302 https://explain.school.example/explain?d={host}
  }
}
```

NGINX (redirect)
```
server {
  listen 443 ssl;
  server_name block.school.internal;
  location /blocked {
    return 302 https://explain.school.example/explain?d=$host;
  }
}
```

HAProxy (redirect)
```
frontend fe_block
  bind :443 ssl crt /etc/haproxy/certs
  http-request redirect code 302 location https://explain.school.example/explain?d=%[req.hdr(Host)]
```

Apache httpd (redirect)
```
<VirtualHost *:443>
  ServerName block.school.internal
  RedirectMatch 302 ^/blocked$ https://explain.school.example/explain?d=%{HTTP_HOST}e
</VirtualHost>
```

Squid (concept)
- Many filters/UIs allow an external redirect URL template for blocked categories. Use:
  https://explain.school.example/explain?d=<original-domain-token>
- Refer to your product docs for the exact token name (often a macro like %un or similar).

TLS/SNI realities (important)
- Do NOT rely on DNS CNAME/A overrides to a different hostname for HTTPS and expect a friendly page; you’ll hit a certificate mismatch before HTTP starts.
- Use a proxy/gateway that can either forward with header injection or send a 302 to a host you control with a valid certificate.

Verification
- Header-injection: curl -H "X-Original-Host: example.com" https://guard.school.internal/explain
  Expect HTML showing example.com.
- Redirect: visit https://block.school.internal/blocked for an HTTPS site; ensure you land on https://explain.school.example/explain?d=example.com with no warnings.

FAQ
- Can we keep DNS RPZ? Yes. Use RPZ to coarsely classify, but rely on your proxy to deliver the user experience. DNS-only won’t produce a seamless page on HTTPS.
- What about metrics/privacy? Keep them lean; avoid PII. SB29-guard avoids logging d by default.
