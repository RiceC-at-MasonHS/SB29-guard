# Apache httpd quickstart (School Mode)

Goal: run SB29-guard behind Apache so blocked requests show a friendly explain page.

Preview
![Explain page screenshot](../../screenshot-2025-08-09-204319.png)

Note
- Bundles aren’t committed to git; they’re generated into dist/ and may be overwritten.

Prereqs
- SB29-guard reachable (e.g., http://127.0.0.1:8080)
- Vhost for blocked traffic, e.g., blocked.school.local
- Enable modules: proxy, proxy_http, headers, ssl (if HTTPS)

Generate example bundle (one-liner)
- sb29guard generate-proxy --format apache --mode header-injection --site-host blocked.school.local --backend-url http://127.0.0.1:8080 --bundle-dir dist/apache

Try this first (minimal guard.conf)
- sb29guard generate-proxy --format apache --mode header-injection --site-host blocked.school.local --backend-url http://127.0.0.1:8080 --dry-run > guard.conf
- Include and reload Apache.

Install and verify
- Place guard.conf in sites-available (or conf.d) and enable the site; reload Apache.
- curl -H "X-Original-Host: exampletool.com" http://blocked.school.local/explain
- Expect explain HTML; non-listed domains return 404 Not Classified (pass-through).

Notes
- For HTTPS, configure SSLCertificateFile/SSLCertificateKeyFile on the vhost.
- Static explain (redirect model): sb29guard generate-explain-static --out-dir dist/explain and host it.

See also
- Example bundle: dist/apache/README.md
- Proxy overview: docs/implementers/proxy.md
- GUI/list integrations: docs/implementers/gui-proxy.md
