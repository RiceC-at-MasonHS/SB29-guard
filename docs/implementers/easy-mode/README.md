# Easy-mode: One-command HTTPS deployment (Recommended)

A minimal, production-friendly deployment using Docker Compose and Caddy for automatic HTTPS.

## Prerequisites
- A DNS name (e.g., blocked.guard.school.org) pointing to this server’s public IP
- Docker + Docker Compose installed
- Policy file at `docs/implementers/easy-mode/policy/domains.yaml`

## Quick start
1) Copy example policy and edit it:
```
mkdir -p docs/implementers/easy-mode/policy
cp policy/domains.example.yaml docs/implementers/easy-mode/policy/domains.yaml
```
2) Create a `.env` file in `docs/implementers/easy-mode/`:
```
SB29_DOMAIN=blocked.guard.school.org
ACME_EMAIL=it-admin@school.org
# Optional: SB29_LAW_URL override
# SB29_LAW_URL=https://search-prod.lis.state.oh.us/api/v2/general_assembly_135/legislation/sb29/05_EN/pdf/
```
3) Launch:
```
docker compose -f docs/implementers/easy-mode/docker-compose.yml up -d
```
4) Test: https://$env:SB29_DOMAIN/explain?domain=exampletool.com

## Notes
- Caddy handles HTTPS automatically. Ensure ports 80/443 are reachable and public DNS resolves to SB29_DOMAIN.
- Header inference works out of the box; Host fallback stays disabled by default.

## Remove
```
docker compose -f docs/implementers/easy-mode/docker-compose.yml down
```
