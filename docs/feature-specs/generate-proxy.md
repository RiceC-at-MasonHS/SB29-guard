# Feature Spec: sb29guard generate-proxy

Goal
Generate copy/paste proxy snippets for School Mode.

CLI
- name: generate-proxy
- flags:
  - --format: caddy|nginx|haproxy|apache (required)
  - --mode: header-injection|redirect (required)
  - --site-host: blocked.guard.school.org (required)
  - --backend-url: http://127.0.0.1:8080 (required)
  - --explain-url: https://explain.school.org/explain (optional; redirect mode)
  - --out: file path (optional)
  - --dry-run: print to stdout

Outputs
- caddy: Caddyfile site block; inject X-Original-Host and proxy to backend.
- nginx: server/location directives with proxy_set_header and return 302 for redirect mode.
- haproxy: frontend/backend snippets with http-response set-header or redirect.
- apache: VirtualHost with RequestHeader set / RewriteRule for redirect.

Validation
- Require site-host and backend-url; explain-url required in redirect mode.
- No secrets; outputs are static text.

Acceptance
- PX-1: Valid snippets render for all formats and both modes.
- Examples included in docs/implementers/proxy.md and one-page quickstarts:
  - docs/implementers/nginx-quickstart.md
  - docs/implementers/caddy-quickstart.md
  - docs/implementers/haproxy-quickstart.md
  - docs/implementers/apache-quickstart.md
