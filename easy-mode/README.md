# Easy-mode: One-command HTTPS deployment

This is the recommended, simplest path for schools to deploy SB29-guard with automatic HTTPS.

What you get:
- Reverse proxy with HTTPS (Caddy auto-issues/renews certificates via ACME)
- sb29guard app behind HTTPS, with original Host preserved for header-based inference
- Minimal config: provide your domain and contact email, plus your policy file

## Prerequisites
- A DNS name (e.g., blocked.guard.school.org) pointing to this server’s public IP
- Docker + Docker Compose installed
- Policy file at `easy-mode/policy/domains.yaml`

## Quick start
1) Copy example policy and edit it:
```
mkdir -p easy-mode/policy
cp policy/domains.example.yaml easy-mode/policy/domains.yaml
```
2) Create a `.env` file in `easy-mode/`:
```
SB29_DOMAIN=blocked.guard.school.org
ACME_EMAIL=it-admin@school.org
# Optional: override default law URL
# SB29_LAW_URL=https://search-prod.lis.state.oh.us/api/v2/general_assembly_135/legislation/sb29/05_EN/pdf/
```
3) Launch:
```
docker compose -f easy-mode/docker-compose.yml up -d
```
4) Test in a browser: https://blocked.guard.school.org/explain?domain=exampletool.com

Optional: Quick header-based CLI tests (local port published for this)
```
curl -s -H "X-Original-Host: exampletool.com" -o /dev/null -w "HTTP %{http_code}\n" http://localhost:8080/explain
curl -s -H "X-Original-Host: not-in-policy.example" -o /dev/null -w "HTTP %{http_code}\n" http://localhost:8080/explain
```

## Notes
- Caddy handles HTTPS automatically. Ensure ports 80/443 are reachable from the Internet and DNS resolves to this host.
- Header inference works out of the box because Caddy forwards the original Host to the app.
- Host fallback stays disabled by default for safety.

## Updating policy
Replace the file at `easy-mode/policy/domains.yaml` and the app will read it on next start. To hot-reload, stop/start the container.
Note: For local CLI tests, sb29guard is also bound on `http://127.0.0.1:8080`.

## Generate DNS artifacts (no docker exec)
Use the helper scripts which handle docker compose run and output paths for you:

- Windows PowerShell:
```
cd easy-mode
./gen-dns.ps1 hosts a-record 10.10.10.50         # hosts file to out/hosts.txt
./gen-dns.ps1 bind cname blocked.guard.local     # BIND zone to out/zone.db
./gen-dns.ps1 domain-list                        # one domain per line to out/domains.txt
```

- Bash (macOS/Linux):
```
cd easy-mode
./gen-dns.sh hosts a-record 10.10.10.50
./gen-dns.sh bind cname blocked.guard.local
./gen-dns.sh domain-list
```

Outputs land in `easy-mode/out/` on your host.

## Removing
```
docker compose -f easy-mode/docker-compose.yml down
```
