# HTTPS / TLS Setup

Goal: serve the explanation pages over HTTPS without browser warnings, with minimal upkeep.

## Strategy Overview
- Terminate TLS at a reverse proxy (Nginx/Traefik/Caddy/IIS) and proxy to sb29guard on 8080.
- Use public certificates from Let’s Encrypt (ACME) for public DNS names, or internal CA certs for private names.
- Automate renewal (cron/systemd timers or built-in resolvers like Traefik/Caddy).

## Let’s Encrypt (Nginx + Certbot)
```
sudo apt install nginx certbot python3-certbot-nginx
sudo certbot --nginx -d blocked.guard.local --redirect
# This edits nginx config to add SSL and HTTP->HTTPS redirect, and sets up renewal.
```
Ensure the server is reachable on 80/443 and public DNS resolves to it.

## Traefik (Docker)
Traefik can obtain and auto-renew certs via ACME:
```
# traefik.yml snippet
entryPoints:
  websecure: { address: ":443" }
certificatesResolvers:
  letsencrypt:
    acme:
      email: admin@example.org
      storage: /letsencrypt/acme.json
      tlsChallenge: {}
```
Add labels to your sb29 container (see container.md) to enable TLS and route.

## Caddy
Caddy auto-issues and renews certs by default:
```
blocked.guard.local {
  reverse_proxy 127.0.0.1:8080
}
```

## IIS (Windows)
- Import a certificate (public or internal CA) into Local Machine → Personal store.
- Bind the site to HTTPS on 443 with that cert.
- Use URL Rewrite or Application Request Routing (ARR) to reverse proxy to http://localhost:8080.
- Add request headers X-Original-Host and X-Forwarded-Host when proxying.

## Internal CA Option
For private hostnames, use your district’s internal CA and distribute the root certificate to managed devices (GPO/MDM). Generate a cert for the blocked host and configure your proxy to use it.

## Security headers
sb29guard already sets conservative headers on /explain and /law (no-store, no-referrer, frame-ancestors none). Keep proxy defaults simple.
