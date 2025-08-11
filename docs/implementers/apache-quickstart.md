# Apache httpd quickstart (School Mode)

Goal: show a friendly explain page for blocked sites using Apache.

Models
- Header-injection reverse proxy (preferred): ProxyPass to sb29-guard and set X-Original-Host/X-Forwarded-Host
- Redirect to static explain: 302 to an explain site that reads d,c,v,h

Prereqs
- SB29-guard reachable (e.g., http://127.0.0.1:8080)
- Vhost for blocked traffic, e.g., blocked.school.local
- Enable modules: proxy, proxy_http, headers, ssl (if HTTPS)

1) Generate config
Snippet:
  sb29guard generate-proxy --format apache --mode header-injection --site-host blocked.school.local --backend-url http://127.0.0.1:8080 --dry-run
Bundle:
  sb29guard generate-proxy --format apache --mode header-injection --site-host blocked.school.local --backend-url http://127.0.0.1:8080 --bundle-dir dist/apache
Redirect variant:
  sb29guard generate-proxy --format apache --mode redirect --site-host blocked.school.local --explain-url https://explain.school.example/explain --bundle-dir dist/apache-redirect

2) Install
- Place guard.conf in sites-available (or conf.d) and enable the site.
- Reload Apache.

3) Verify
- curl -H "X-Original-Host: exampletool.com" http://blocked.school.local/explain
- Expect the explain HTML; for non-listed domains, guard returns 404 Not Classified.

Notes
- For HTTPS, configure SSLCertificateFile/SSLCertificateKeyFile on the vhost.
- To host a static explain page, run generate-explain-static and serve the folder.
