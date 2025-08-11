# Caddy quickstart (School Mode)

Goal: deploy SB29-guard with Caddy so blocked requests show a friendly explain page.

Preview
![Explain page screenshot](../../screenshot-2025-08-09-204319.png)

Models
- Header-injection reverse proxy (preferred): forward to sb29-guard with X-Original-Host
- Redirect to static explain: 302 to an explain site that reads d,c,v,h

Prereqs
- SB29-guard reachable (e.g., http://127.0.0.1:8080)
- Vhost for blocked traffic, e.g., blocked.school.local

1) Generate a bundle or snippet

Try this first
- Minimal Caddyfile to validate header-injection:
  sb29guard generate-proxy --format caddy --mode header-injection --site-host blocked.school.local --backend-url http://127.0.0.1:8080 --dry-run > Caddyfile
  caddy run --config Caddyfile

Snippet (header-injection):
  sb29guard generate-proxy --format caddy --mode header-injection --site-host blocked.school.local --backend-url http://127.0.0.1:8080 --dry-run

Bundle:
  sb29guard generate-proxy --format caddy --mode header-injection --site-host blocked.school.local --backend-url http://127.0.0.1:8080 --bundle-dir dist/caddy

Redirect version:
  sb29guard generate-proxy --format caddy --mode redirect --site-host blocked.school.local --explain-url https://explain.school.example/explain --bundle-dir dist/caddy-redirect

2) Run Caddy
- Place Caddyfile and run: caddy run --config Caddyfile
- For HTTPS automation, ensure DNS/public reachability or provide certificates if running internally.

3) Verify
- curl -H "X-Original-Host: exampletool.com" http://blocked.school.local/explain
- Expect HTML with the explain page; for non-listed domains, guard returns 404 Not Classified (proxy should pass-through).

Notes
- Header precedence: X-Original-Host > X-Forwarded-Host > Referer.
- Use generate-explain-static to host a static page if you prefer the redirect model.

See example bundle
- dist/caddy/README.md
