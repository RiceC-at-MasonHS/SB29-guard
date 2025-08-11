# ADR 0001: Proxy-first (School Mode) as the default

Status: Accepted
Date: 2025-08-10

Context
- HTTPS + DNS-only redirects can’t present the friendly page without cert/SNI mismatch.
- Schools already operate proxies/gateways that can add headers or issue 302s.

Decision
- Make proxy/gateway integration the default (“School Mode”).
- Two primary models:
  1) Header-injection reverse proxy to dynamic /explain (preferred where possible).
  2) 302 redirect to a static explain site that reads d/c/v/h params (display-only).
- Deprecate legacy easy-mode; remove code/docs and steer operators to proxy guide.

Consequences
- Seamless HTTPS UX; fewer warnings and support issues.
- CLI adds generators for proxy configs and static bundle.
- /explain strictly validates any display-only params; headers/policy remain authoritative.
- Docs and examples focus on proxy setups (Caddy/NGINX/HAProxy/Apache).

Out of scope
- Authenticating end-users or personalized messages.
- Logging identifiable information.
