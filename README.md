groupme.com	LEGAL_HOLD	Pending legal/privacy assessment due to chat features	2025-08-08	suspended	TCK-1202	Temporarily suspended pending risk evaluation	2025-12-31	COMMUNICATION,RISK
# SB29-guard

District-friendly tool to show a clear, plain-language “Why was I redirected?” page when staff or students try to use an online tool without an approved Data Privacy Agreement (SB29 context).

If you just need to get it running, follow the Quick Start below. For deeper technical details, see `TECHNICAL.md`.

## What It Does (Plain Language)
When a blocked site is requested, your DNS redirects the user to this service. The service shows:
- The original site name
- The reason it’s not currently approved (e.g., NO_DPA, EXPIRED_DPA)
- Optional rationale and reference (ticket / review ID)

## Basic Concepts
- Policy File: A simple list of domains and reasons (you can edit a YAML file). Wildcards like `*.example.com` are supported.
- Redirect Page: A minimal web page explaining the block.
- DNS Lists: Generated files you can load into DNS platforms (hosts file, BIND zone, Unbound local-data, RPZ). All point those domains at your redirect host/IP.

## Quick Start (File Mode)
1. Copy the example: `cp policy/domains.example.yaml policy/domains.yaml`
2. Edit `policy/domains.yaml` – change or add domains & reasons.
3. Validate:
	```
	sb29guard validate --policy policy/domains.yaml
	```
4. Generate a simple hosts file that redirects to an internal IP (replace 10.10.10.50):
	```
	sb29guard generate-dns --policy policy/domains.yaml --format hosts --mode a-record --redirect-ipv4 10.10.10.50 --out dist/dns/hosts.txt
	```
5. Or create a BIND zone with a redirect host:
	```
	sb29guard generate-dns --policy policy/domains.yaml --format bind --mode cname --redirect-host blocked.guard.local --out dist/dns/zone.db
	```
6. Run the explanation server (default port 8080):
	```
	sb29guard serve --policy policy/domains.yaml
	```
7. Test in a browser:
	`http://localhost:8080/explain?domain=exampletool.com`

## Adding / Updating a Domain
Open `policy/domains.yaml`, duplicate an entry, change the domain and classification, keep dates in `YYYY-MM-DD`.

Common classifications (choose one):
`NO_DPA`, `PENDING_REVIEW`, `EXPIRED_DPA`, `LEGAL_HOLD`, `OTHER`

Set `status: active` to enforce. Use `suspended` to temporarily disable an entry (it will not appear in new DNS outputs).

## Wildcards
Use `*.trackingwidgets.io` to cover any subdomain like `api.trackingwidgets.io`. The explanation page will match both the base domain and subdomains.

## Optional: Spreadsheet (Planned)
Future versions will allow maintaining the list in a secure Google Sheet instead of editing the YAML manually. (See `TECHNICAL.md`.)

## Updating DNS
Regenerate the DNS file and deploy to your DNS resolver whenever you change the policy. Keep the redirect IP/host pointing at the server that runs `sb29guard serve`.

## Integrity / Audit
You can get a stable hash of the active policy:
```
sb29guard hash --policy policy/domains.yaml
```
Record that hash if you need an audit trail.

## Releases
Pre-built binaries are published on the Releases page. Download the one for your OS, place it on a server (or in a container), and follow Quick Start.

### Building From Source (Developers)
Ensure Go 1.22+ is installed, then:
```
go test ./...
go build -trimpath -ldflags "-s -w" ./cmd/sb29guard
./sb29guard --help
```
No Makefile required; CI uses the same commands.

## Need More Detail?
See `TECHNICAL.md` for schema, advanced environment variables, roadmap, and contributor guidance. For branding or page layout changes, read `CUSTOMIZING.md`.

## Disclaimer
This tool assists with transparency and process—it does not replace formal legal review. Always confirm status with your district’s data privacy / legal team.
