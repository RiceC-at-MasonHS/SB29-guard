# Implementers Guide

This folder contains practical deployment guides for school IT teams and administrators.

Contents:
- vm.md – Bare‑metal/VM deployment (Windows/Linux)
- container.md – Container deployment (Docker/Podman/Kubernetes)
- https.md – HTTPS/TLS: reverse proxy, certs, automated renewal
 - nginx-quickstart.md – One-page NGINX setup for School Mode
 - caddy-quickstart.md – One-page Caddy setup for School Mode
 - haproxy-quickstart.md – One-page HAProxy setup for School Mode
 - apache-quickstart.md – One-page Apache httpd setup for School Mode
 - gui-proxy.md – GUI proxy/list integrations using /classify and /domain-list
 - scripts/ – Set-and-forget automation scripts
	 - linux-fetch-and-reload.sh – Pull /domain-list and reload/update proxy maps (cron-ready)
	 - windows-fetch-and-import.ps1 – Pull /domain-list and import into GUI systems (Task Scheduler-ready)
 - See also detailed DNS platform guides under docs/deployment/ (BIND, Unbound, Pi-hole, pfSense, OPNsense, Infoblox)

These guides complement the DNS platform docs under docs/deployment/.
