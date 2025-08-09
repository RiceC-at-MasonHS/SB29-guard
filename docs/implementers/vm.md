# VM / Bare-metal Deployment

This guide covers Windows and Linux server installs without containers.

## Prereqs
- CPU: 1 vCPU+, RAM: 512MB+ (lightweight)
- Go binary (from Releases) or build from source
- DNS configured to redirect restricted domains to this server’s IP(s)
- Optional: reverse proxy (Nginx/Traefik/Caddy/IIS) for HTTPS

## Linux (systemd)
1) Create a user and directories:
```
sudo useradd --system --no-create-home sb29
sudo mkdir -p /opt/sb29/{bin,policy,logs}
sudo chown -R sb29:sb29 /opt/sb29
```
2) Copy binary and policy:
```
sudo cp ./sb29guard /opt/sb29/bin/
sudo cp policy/domains.yaml /opt/sb29/policy/
```
3) Service unit `/etc/systemd/system/sb29guard.service`:
```
[Unit]
Description=SB29-guard web service
After=network.target

[Service]
User=sb29
WorkingDirectory=/opt/sb29
Environment=SB29_LAW_URL=https://search-prod.lis.state.oh.us/api/v2/general_assembly_135/legislation/sb29/05_EN/pdf/
ExecStart=/opt/sb29/bin/sb29guard serve --policy /opt/sb29/policy/domains.yaml
Restart=on-failure

[Install]
WantedBy=multi-user.target
```
4) Start:
```
sudo systemctl daemon-reload
sudo systemctl enable --now sb29guard
```

## Windows (Service)
Use NSSM or Windows Service Wrapper to run sb29guard as a service.
- Path: `C:\sb29\sb29guard.exe`
- Arguments: `serve --policy C:\sb29\policy\domains.yaml`
- Set environment vars (SB29_LAW_URL, optional SB29_ALLOW_HOST_FALLBACK).

## Reverse Proxy (HTTPS)
See `https.md` for TLS setup. Example Nginx snippet:
```
server {
  listen 80;
  server_name blocked.guard.local;
  return 301 https://$host$request_uri;
}
server {
  listen 443 ssl;
  server_name blocked.guard.local;
  ssl_certificate     /etc/letsencrypt/live/blocked.guard.local/fullchain.pem;
  ssl_certificate_key /etc/letsencrypt/live/blocked.guard.local/privkey.pem;
  location / {
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header X-Original-Host $host; # preserve original host if needed
    proxy_set_header X-Forwarded-Host $host;
  }
}
```
