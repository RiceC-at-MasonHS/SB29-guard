# HAProxy quickstart (School Mode)

Goal: route blocked traffic to an explanation flow using HAProxy.

Models
- Header-injection reverse proxy (preferred): forward to sb29-guard with X-Original-Host
- Redirect to static explain: 302 to an explain site that reads d,c,v,h

Prereqs
- SB29-guard reachable (e.g., 127.0.0.1:8080)
- Vhost for blocked traffic, e.g., blocked.school.local

1) Generate config
Snippet (header-injection):
  sb29guard generate-proxy --format haproxy --mode header-injection --site-host blocked.school.local --backend-url http://127.0.0.1:8080 --dry-run
Bundle:
  sb29guard generate-proxy --format haproxy --mode header-injection --site-host blocked.school.local --backend-url http://127.0.0.1:8080 --bundle-dir dist/haproxy
Redirect:
  sb29guard generate-proxy --format haproxy --mode redirect --site-host blocked.school.local --explain-url https://explain.school.example/explain --bundle-dir dist/haproxy-redirect

2) Install and reload
- Copy haproxy.cfg to the system config location or include it from your main config.
- Reload/Restart HAProxy.

3) Verify
- curl -H "Host: blocked.school.local" -H "X-Original-Host: exampletool.com" http://127.0.0.1:80/explain
- Expect 200 and an explain page; non-listed domains should pass-through.

Selective routing map (optional)
- If you supply --policy or --sheet-csv, a blocked.map file is generated.
- Use -m dom or ACLs to match entries (includes base and .base for wildcards).

Notes
- Ensure required modules/ACLs are enabled; check logs if 404 occurs.
