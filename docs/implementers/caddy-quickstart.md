# Caddy quickstart (School Mode)

Goal: run SB29-guard with Caddy so blocked requests show a friendly explain page.

Preview
![Explain page screenshot](../../screenshot-2025-08-09-204319.png)

Note
- Bundles aren’t committed to git; they’re generated into dist/ and may be overwritten.

Prereqs
- SB29-guard reachable (e.g., http://127.0.0.1:8080)
- Vhost for blocked traffic, e.g., blocked.school.local

Generate example bundle (one-liner)
- sb29guard generate-proxy --format caddy --mode header-injection --site-host blocked.school.local --backend-url http://127.0.0.1:8080 --bundle-dir dist/caddy

Try this first (minimal Caddyfile)
- sb29guard generate-proxy --format caddy --mode header-injection --site-host blocked.school.local --backend-url http://127.0.0.1:8080 --dry-run > Caddyfile
- caddy run --config Caddyfile

Run and verify
- Use the generated Caddyfile or bundle and start Caddy.
- curl -H "X-Original-Host: exampletool.com" http://blocked.school.local/explain
- Expect explain HTML; non-listed domains return 404 Not Classified (pass-through).

Notes
- Header precedence: X-Original-Host > X-Forwarded-Host > Referer.
- Static explain (redirect model): sb29guard generate-explain-static --out-dir dist/explain and host it.

See also
- Example bundle: dist/caddy/README.md
- Proxy overview: docs/implementers/proxy.md
- GUI/list integrations: docs/implementers/gui-proxy.md
