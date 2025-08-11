# NGINX quickstart (School Mode)

Goal: deploy SB29-guard behind NGINX so blocked requests show a friendly explain page. Two models are supported; pick one.

Preview
![Explain page screenshot](../../screenshot-2025-08-09-204319.png)

- Model A — header-injection reverse proxy (preferred): NGINX proxies to sb29-guard and sets X-Original-Host.
- Model B — redirect to static explain: NGINX sends an HTTP 302 to a static page that reads d,c,v,h from the URL.

Prereqs
- SB29-guard running (or plan to run) at http://127.0.0.1:8080 or similar
- A vhost name for blocked traffic, e.g., blocked.school.local
- For HTTPS: a valid cert/key for your vhost on managed devices

1) Generate config

Try this first
- Minimal header-injection config to validate the flow quickly:
  sb29guard generate-proxy --format nginx --mode header-injection --site-host blocked.school.local --backend-url http://127.0.0.1:8080 --dry-run > nginx.conf
  # Include nginx.conf in your nginx site and reload.

Header-injection (recommended):
- Use the CLI to emit a snippet or a bundle.

Snippet (stdout):
  sb29guard generate-proxy --format nginx --mode header-injection --site-host blocked.school.local --backend-url http://127.0.0.1:8080 --dry-run

Bundle (ready-to-use folder with site.conf, optional blocked_map.conf, README, smoke.ps1):
  sb29guard generate-proxy --format nginx --mode header-injection --site-host blocked.school.local --backend-url http://127.0.0.1:8080 --bundle-dir dist/nginx
  # Optional extras
  #  --tls-cert /etc/ssl/certs/blocked.crt --tls-key /etc/ssl/private/blocked.key
  #  --policy policy/domains.yaml  (writes blocked_map.conf for selective routing)
  #  --redirect-unknown --explain-url https://explain.school.example/explain  (redirect 404s to static explain)

Redirect model:
  sb29guard generate-proxy --format nginx --mode redirect --site-host blocked.school.local --explain-url https://explain.school.example/explain --dry-run

2) Install
- Copy site.conf to NGINX (sites-available or conf.d) and enable it.
- If you generated blocked_map.conf, include it under the global http {} in nginx.conf.
- If using TLS, update ssl_certificate/ssl_certificate_key and add the 80->443 redirect server block (bundled when --tls-* provided).

3) Reload and smoke test
- Reload: nginx -s reload
- Quick check (PowerShell example):
  Invoke-WebRequest -UseBasicParsing -Uri http://blocked.school.local/explain -Headers @{ 'X-Original-Host'='exampletool.com' } | Select-Object -ExpandProperty StatusCode
- Or run the bundle's smoke.ps1: ./smoke.ps1 -Guard 'http://127.0.0.1:8080' -HostName 'exampletool.com'

Notes
- Header precedence in sb29-guard: X-Original-Host > X-Forwarded-Host(1st) > Referer; Host fallback is off by default.
- Policy lookup is denylist style: if not found, guard returns 404 Not Classified. Your proxy should pass-through in that case (non-interruptive).
- Use blocked_map.conf for selective routing only if your main proxy needs to detect blocked hosts without calling the API.
- For static explain, generate with: sb29guard generate-explain-static --out-dir dist/explain, and host it on your web server.

See example bundle
- dist/nginx/README.md
