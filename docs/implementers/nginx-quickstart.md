# NGINX quickstart (School Mode)

Goal: run SB29-guard behind NGINX so blocked requests show a friendly explain page.

Preview
![Explain page screenshot](../../screenshot-2025-08-09-204319.png)

Note
- Bundles aren’t committed to git; they’re generated on demand into dist/ and may be overwritten when you re-run the CLI.

Prereqs
- SB29-guard reachable (e.g., http://127.0.0.1:8080)
- Vhost name for blocked traffic, e.g., blocked.school.local
- For HTTPS: valid cert/key on managed devices

Generate example bundle (one-liner)
- sb29guard generate-proxy --format nginx --mode header-injection --site-host blocked.school.local --backend-url http://127.0.0.1:8080 --bundle-dir dist/nginx

Try this first (minimal snippet)
- sb29guard generate-proxy --format nginx --mode header-injection --site-host blocked.school.local --backend-url http://127.0.0.1:8080 --dry-run > nginx.conf
- Include nginx.conf in your NGINX site and reload.

Install
- Copy dist/nginx/site.conf into sites-available or conf.d and enable it.
- If using TLS, set ssl_certificate/ssl_certificate_key (bundle includes 80->443 redirect when --tls-* was used).
- If you generated blocked_map.conf, include it under the global http {} in nginx.conf.

Reload and verify
- Reload: nginx -s reload
- Quick check (PowerShell): Invoke-WebRequest -UseBasicParsing -Uri http://blocked.school.local/explain -Headers @{ 'X-Original-Host'='exampletool.com' } | Select-Object -ExpandProperty StatusCode
- Or run: ./dist/nginx/smoke.ps1 -Guard 'http://127.0.0.1:8080' -HostName 'exampletool.com'

Notes
- Header precedence: X-Original-Host > first X-Forwarded-Host > Referer (Host fallback off by default).
- Denylist model: non-listed domains return 404 Not Classified (proxy should pass through).
- Static explain (redirect model): sb29guard generate-explain-static --out-dir dist/explain and host it.

See also
- Example bundle: dist/nginx/README.md
- Proxy overview: docs/implementers/proxy.md
- GUI/list integrations: docs/implementers/gui-proxy.md
