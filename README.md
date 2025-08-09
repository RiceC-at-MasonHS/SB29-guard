<div align="center">

# SB29-guard 🚦📘

Making blocked ed‑tech domains less confusing (and more transparent) for teachers & students.

<!-- Badges -->
[![Build](https://github.com/RiceC-at-MasonHS/SB29-guard/actions/workflows/ci.yml/badge.svg)](../../actions)
![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-AGPL--3.0-blue)
![Status](https://img.shields.io/badge/status-early%20preview-orange)
![Coverage](https://img.shields.io/badge/coverage-core%2080%25+-brightgreen)

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
* DNS outputs – hosts, BIND, Unbound, RPZ to steer blocked domains to this page.

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

## Integrity / Audit 🔐
Stable hash of active (non‑suspended) records:
```
sb29guard hash --policy policy/domains.yaml
```
Record that hash if you need an audit trail.

## Releases 📦
Pre-built binaries (see Releases). Download, place on a server (or container), run commands above.

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

## License
SB29-guard is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0).
- You can use and modify it freely.
- If you modify and run it as a network service, you must provide the corresponding source to users of the service.
- You may charge for installation, support, or hosting; the source must remain available under AGPL-3.0.
- See `LICENSE` for the full terms.

## Disclaimer ⚖️
Helps with transparency & workflow. Does NOT replace district legal review. Always verify with your data privacy / legal team.

---
Questions or ideas? Open an issue. Contributions welcome.
