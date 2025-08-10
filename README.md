<div align="center">

# SB29-guard 🚦📘

Making blocked ed‑tech domains less confusing (and more transparent) for teachers & students. Protecting student data is a noble objective, and can be done in ways that lift up our communities. 

<em>Built by a teacher, for teachers — to protect educators, reduce friction, and make compliance humane.</em>

<!-- Badges -->
[![Build](https://github.com/RiceC-at-MasonHS/SB29-guard/actions/workflows/ci.yml/badge.svg)](../../actions)
![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-AGPL--3.0-blue)
![Status](https://img.shields.io/badge/status-stable-brightgreen)
![Coverage](https://img.shields.io/badge/coverage-core%2080%25+-brightgreen)

</div>

<div align="center">
	<img src="./screenshot-2025-08-09-204319.png" alt="SB29-guard explanation page (dark mode)" style="max-width: 980px; width: 100%; border-radius: 8px;" />
</div>

District‑friendly tool that shows a clear, plain‑language “Why was I redirected?” page when staff or students try a site without an approved Data Privacy Agreement (SB29 context). One small self‑contained binary (HTML & CSS embedded). No tracking. No student data stored. ✨

If you just need to get it running, follow the Quick Start below. For deeper technical details, see the Technical Reference: TECHNICAL.md.

## What It Does (Plain Language)
When a blocked site is requested, school DNS points the browser here. The page clearly explains:
* 🔗 The site name
* 🏷️ Why it’s restricted (e.g., NO_DPA, EXPIRED_DPA)
* 📝 Optional rationale / ticket reference
* 📌 Policy version (for audits)

## Core Pieces
* Policy file (YAML) – you edit it; wildcards like `*.example.com` allowed.
* Explanation page – simple, readable, accessible.
* DNS outputs – hosts, BIND, Unbound, RPZ, dnsmasq, Windows DNS PowerShell, and domain-list to steer blocked domains to this page.

## Quick Start (🪄 ~2 minutes)
1. Copy example policy:
	```
	cp policy/domains.example.yaml policy/domains.yaml
	```
2. Edit `policy/domains.yaml` (add or change a domain entry).
3. Validate it:
	```
	sb29guard validate --policy policy/domains.yaml
	```
4. Generate a hosts file (swap in your internal IP):
	```
	sb29guard generate-dns --policy policy/domains.yaml --format hosts --mode a-record --redirect-ipv4 10.10.10.50 --out dist/dns/hosts.txt
	```
5. Or generate a BIND zone using a redirect host:
	```
	sb29guard generate-dns --policy policy/domains.yaml --format bind --mode cname --redirect-host blocked.guard.local --out dist/dns/zone.db
	```
6. Run the server:
	```
	sb29guard serve --policy policy/domains.yaml
	```
7. Open:
	`http://localhost:8080/explain?domain=exampletool.com`

### Law link customization
The footer’s “Ohio SB29” points to an internal `/law` redirect. By default, it redirects to the LIS PDF for SB29. You can change the target via an environment variable:

```
# PowerShell (Windows)
$env:SB29_LAW_URL = "https://search-prod.lis.state.oh.us/api/v2/general_assembly_135/legislation/sb29/05_EN/pdf/"

# Bash (Linux/macOS)
export SB29_LAW_URL="https://search-prod.lis.state.oh.us/api/v2/general_assembly_135/legislation/sb29/05_EN/pdf/"
```
Restart the server after changing the variable.

### How the server knows the original domain
The explanation page needs to know which domain was blocked:
- If you call `/explain?domain=example.com`, that value is used.
- If there’s no query param (typical DNS redirect), the server usually detects the original site automatically from standard headers set by browsers/proxies.

Want the nuts and bolts or special setups (A‑record vs CNAME, reverse proxy headers, optional Host fallback)? See the Technical Reference: [Header‑based domain inference](./TECHNICAL.md#header-based-domain-inference).

## Add / Update a Domain 🧾
Edit `policy/domains.yaml`. Duplicate an existing record and change the domain.

Classifications: `NO_DPA`, `PENDING_REVIEW`, `EXPIRED_DPA`, `LEGAL_HOLD`, `OTHER`.

`status: active` = enforced. `status: suspended` = ignored in new DNS outputs & hash.

## Wildcards
Use `*.trackingwidgets.io` to cover any subdomain like `api.trackingwidgets.io`. The explanation page will match both the base domain and subdomains.

## Use a Google Sheet (Published CSV) 📊
Prefer editing a Google Sheet instead of a YAML file? Publish the sheet as CSV (File → Share → Publish to the web → select the sheet + CSV) and copy the link ending with `output=csv`.

Serving with automatic updates:
- `sb29guard serve --sheet-csv <csv_url>` automatically checks for updates daily at 11:59 PM (local time) and hot‑swaps the in‑memory policy on success. No restart required.
- Errors (HTTP/network/validation) are logged as JSON and the server keeps the last good policy.
- JSON log events: `policy.refresh.scheduled`, `policy.refresh.start`, `policy.refresh.success`, `policy.refresh.error`.

CLI/one‑off usage with the sheet:
```
sb29guard validate --sheet-csv "https://docs.google.com/.../output=csv"
sb29guard hash --sheet-csv "https://docs.google.com/.../output=csv"
sb29guard generate-dns --sheet-csv "https://docs.google.com/.../output=csv" --format hosts --mode a-record --redirect-ipv4 10.10.10.50 --dry-run
```

Caching: CSV responses are cached under `./cache/sheets/` using ETag/Last‑Modified. If unchanged, logs show `"source":"csv-cache"`.

Column details (required/optional) and deep‑dive notes live in the Technical Reference under Google Sheets Integration:
- Technical Reference: [TECHNICAL.md](./TECHNICAL.md#google-sheets-integration-published-csv-%E2%80%93-implemented-v01)

## Updating DNS 🔁
Whenever the policy changes:
```
sb29guard generate-dns --policy policy/domains.yaml --format hosts --mode a-record --redirect-ipv4 10.10.10.50 --out dist/dns/hosts.txt
```
Deploy the refreshed file to your DNS platform.

## Verify deployment ✅
- DNS returns redirect for a sample domain (platform-specific):
	- BIND/Unbound: dig exampletool.com @<resolver-ip>
	- Pi-hole: dig exampletool.com @<pihole-ip>
	- Windows DNS: Resolve-DnsName exampletool.com -Server <dns-ip>
- Web server health: GET http://<redirect-host>:8080/health → {"status":"ok"}
- Metrics: GET http://<redirect-host>:8080/metrics → policy_version, record_count, refresh stats
- Browser: Visit a blocked domain and confirm the explanation page renders

See platform guides with detailed checklists:
- BIND: docs/deployment/bind.md
- Unbound: docs/deployment/unbound.md
- Pi-hole: docs/deployment/pihole.md
- Windows DNS: docs/deployment/windows-dns.md
 - pfSense: docs/deployment/pfSense.md
 - OPNsense: docs/deployment/OPNsense.md
 - Infoblox: docs/deployment/infoblox.md
 - Implementers (VM/Containers/HTTPS): docs/implementers/
 - Easy‑mode (recommended, auto‑HTTPS): easy-mode/ (Caddy + Docker Compose)

## Integrity / Audit 🔐
Stable hash of active (non‑suspended) records:
```
sb29guard hash --policy policy/domains.yaml
```
Record that hash if you need an audit trail.

## Releases 📦
Pre-built binaries (see Releases). Download, place on a server (or container), run commands above.

Container image (GHCR):

- ghcr.io/ricec-at-masonhs/sb29-guard:v1.0.0
- Prefer pinning a version tag (or digest) over latest for stability.
	- Example (Compose): `image: ghcr.io/ricec-at-masonhs/sb29-guard:v1.0.0`
	- Optional strict pin: `image: ghcr.io/ricec-at-masonhs/sb29-guard@sha256:<digest>`

	Binary downloads verification:

	1) Download the appropriate sb29guard-<os>-<arch> file and SHA256SUMS.txt from the Release page.
	2) Verify checksums:
		- Windows (PowerShell): `Get-FileHash .\sb29guard-windows-amd64.exe -Algorithm SHA256`
		- macOS/Linux: `shasum -a 256 sb29guard-darwin-arm64` or `sha256sum sb29guard-linux-amd64`
	3) Confirm the hex matches the corresponding line in SHA256SUMS.txt.

### Building From Source (Developers) 🛠️
Need to hack? Install Go 1.22+ then:
```
go test ./...
go build -trimpath -ldflags "-s -w" ./cmd/sb29guard
./sb29guard --help
```
No Makefile. CI mirrors these steps. Core logic (policy, DNS generation) has coverage gates; CLI & server also tested.

## Need More Detail? 🔍
- Technical Reference (internals, roadmap, caching details): [TECHNICAL.md](./TECHNICAL.md)
- Customizing the UI/templates: [CUSTOMIZING.md](./CUSTOMIZING.md)
- Contributing guide: [CONTRIBUTING.md](./CONTRIBUTING.md)
 - Data Privacy Agreement (template): [docs/dpa.md](./docs/dpa.md)

## License
SB29-guard is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0).

### Try it: Easy‑mode (HTTPS, one command)
If you have Docker Desktop, you can spin up the auto‑HTTPS stack:

```powershell
# 1) Prepare policy
New-Item -ItemType Directory -Force -Path easy-mode\policy | Out-Null
Copy-Item -Force policy\domains.example.yaml easy-mode\policy\domains.yaml

# 2) Create easy-mode/.env with your public domain and email
#    (Domain must resolve publicly to this host for HTTPS issuance.)
@'
SB29_DOMAIN=blocked.guard.school.org
ACME_EMAIL=it-admin@school.org
# Optional: override default law URL
# SB29_LAW_URL=https://search-prod.lis.state.oh.us/api/v2/general_assembly_135/legislation/sb29/05_EN/pdf/
'@ | Set-Content -NoNewline -Path easy-mode\.env

# 3) Launch the stack
docker compose -f easy-mode\docker-compose.yml up -d

# 4) Test in a browser (replace with your domain):
#    https://blocked.guard.school.org/explain?domain=exampletool.com

# Optional: Quick header-based CLI tests (local port published for this)
# Expect 200 for a domain in the example policy (exampletool.com)
curl.exe -s -H "X-Original-Host: exampletool.com" -o NUL -w "HTTP %{http_code}\n" http://localhost:8080/explain

# Expect 404 for a domain not in policy
curl.exe -s -H "X-Original-Host: not-in-policy.example" -o NUL -w "HTTP %{http_code}\n" http://localhost:8080/explain

# 5) Tear down when finished
# docker compose -f easy-mode\docker-compose.yml down
```

## Disclaimer ⚖️
You can use and modify this software freely. If you modify and run it as a network service, you must provide the corresponding source to users of the service. You may charge for installation, support, or hosting; the source must remain available under AGPL‑3.0. See `LICENSE` for the full terms.

Helps with transparency & workflow. Does NOT replace district legal review. It just makes it a little harder for someone to accidentally use a disallowed web-service. Always verify with your data privacy / legal team.

---
Questions or ideas? Open an issue. Contributions welcome.
