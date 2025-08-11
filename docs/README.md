# SB29-guard Documentation Index

Start here if you’re building or integrating SB29-guard.

For humans
- Implementers overview: docs/implementers-overview.md
- Proxy/Gateway guide (School Mode): docs/implementers/proxy.md
 - Quickstarts: NGINX, Caddy, HAProxy, Apache — see docs/implementers/ (nginx-quickstart.md, caddy-quickstart.md, haproxy-quickstart.md, apache-quickstart.md)
 - GUI-driven proxies: docs/implementers/gui-proxy.md
- Deployment targets: docs/deployment/ (Bind/Unbound/pfSense/OPNsense/Infoblox/Pi-hole/Windows DNS)
- HTTPS/TLS specifics: docs/implementers/https.md
- Container/Kubernetes: docs/implementers/container.md
- Roadmap: docs/roadmap.md

For AI coders
- Requirements (human): docs/requirements.md
- Requirements (machine): docs/requirements.proxy.json
- ADR-0001 (proxy-first decision): docs/adr/0001-proxy-first.md
- CLI design: docs/cli-design.md
- API contract: docs/api-contract.md
- Feature specs: docs/feature-specs/
- Test plan: docs/testing/test-plan.md
- Threat model: docs/security/threat-model.md

Contributing
- See docs/CONTRIBUTING.md (top-level CONTRIBUTING.md) and ensure pre-commit hooks are enabled.
