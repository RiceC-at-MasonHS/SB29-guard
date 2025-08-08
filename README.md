<div align="center">

# SB29-guard 🚦📘

Making blocked ed‑tech domains less confusing (and more transparent) for teachers & students.

<!-- Badges -->
[![Build](https://github.com/RiceC-at-MasonHS/SB29-guard/actions/workflows/ci.yml/badge.svg)](../../actions)
![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-blue)
![Status](https://img.shields.io/badge/status-early%20preview-orange)
![Coverage](https://img.shields.io/badge/coverage-core%2080%25+-brightgreen)

</div>

District‑friendly tool that shows a clear, plain‑language “Why was I redirected?” page when staff or students try a site without an approved Data Privacy Agreement (SB29 context). One small self‑contained binary (HTML & CSS embedded). No tracking. No student data stored. ✨

If you just need to get it running, follow the Quick Start below. For deeper technical details, see `TECHNICAL.md`.

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

## (Planned) Spreadsheet Input 📊
Future option to sync from a Google Sheet (see TECHNICAL.md roadmap section).

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
See `TECHNICAL.md` (internals, roadmap) and `CUSTOMIZING.md` (branding/templates). Keep this README teacher‑/admin‑friendly.

## Disclaimer ⚖️
Helps with transparency & workflow. Does NOT replace district legal review. Always verify with your data privacy / legal team.

---
Questions or ideas? Open an issue. Contributions welcome.
