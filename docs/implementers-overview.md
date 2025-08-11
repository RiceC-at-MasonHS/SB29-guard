# Implementers Overview

This document summarizes deployment options and points to detailed guides.

Recommended path:
- Proxy/Gateway integration (School Mode): docs/implementers/proxy.md
	- One-page quickstarts: docs/implementers/nginx-quickstart.md, docs/implementers/caddy-quickstart.md, docs/implementers/haproxy-quickstart.md, docs/implementers/apache-quickstart.md
	- GUI/list integrations: docs/implementers/gui-proxy.md

Other options:
- VM/Bare-metal: docs/implementers/vm.md
- Containers/Kubernetes: docs/implementers/container.md
- HTTPS/TLS patterns: docs/implementers/https.md
- DNS platform guides: docs/deployment/

DNS artifacts supported by the tool:
- hosts, BIND, Unbound, RPZ, dnsmasq, Windows DNS PowerShell, domain-list

Header inference quick note:
- Prefer passing the original host via proxy headers (X-Original-Host or X-Forwarded-Host).
- Keep SB29_ALLOW_HOST_FALLBACK=false unless using direct A/AAAA redirects without a proxy.
