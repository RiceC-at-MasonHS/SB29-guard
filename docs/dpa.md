# Data Privacy Agreement (DPA) – SB29-guard

Version: 1.0 • Last updated: 2025-08-08

This document is provided to help institutions document privacy expectations when deploying SB29-guard. It is not legal advice. Districts should review with counsel and adapt as needed.

## 1) Parties and Roles
- Controller/Operator: The deploying institution (e.g., school district). Operates the software on its own infrastructure.
- Publisher/Maintainers: The SB29-guard open-source project. Publishes source code only; does not operate services for districts and does not process data on their behalf.

Result: The project maintainers are not a processor/service provider to the district. All processing is performed locally by the district’s deployment.

## 2) Purpose and Processing Description
SB29-guard displays a local “explanation” page when users attempt to visit domains that are restricted by district policy. It generates DNS lists or answers that steer restricted domains to the local explanation page.

Core behavior:
- On each request to the explanation page, the software renders static HTML with the domain name and the policy classification/reason from the district’s policy.
- By default, the software does not store any request data and does not write access logs. No cookies, analytics, or tracking beacons are used.
- Optional feature: If configured with `--sheet-csv`, the server fetches a published CSV from Google Sheets to update the domain policy. Responses may be cached on disk with ETag/Last-Modified metadata to avoid unnecessary network traffic.

## 3) Categories of Data
- Domain names attempted by users (e.g., example.com) – shown transiently in the page when a user hits the explanation endpoint; not stored by the application.
- Policy data (domains, classification, rationale, reference IDs) – provided and managed by the district in YAML or CSV. This data should not contain student personal information.
- Operational metadata (only if `--sheet-csv` is used): ETag/Last-Modified headers and the latest CSV file contents are cached on local disk for efficiency.

Not collected:
- No names, emails, device identifiers, cookies, or telemetry are collected by SB29-guard.
- The application does not transmit usage data to the project or third parties.

Note: Standard web server infrastructure may record connection metadata (e.g., IP address, user agent) if you enable access logging at reverse proxies or gateways. That is outside SB29-guard’s code and under district control.

## 4) Retention, Access, and Deletion
- Request data is not stored by SB29-guard.
- Policy files and (optional) CSV cache are stored locally under district control. Districts should apply their standard retention and deletion policies.
- The CSV cache directory is `./cmd/sb29guard/cache/sheets/` when run from source or `./cache/sheets/` in typical deployments; it can be safely purged at any time (the next refresh will re-download the CSV if needed).

## 5) Security Measures
- SB29-guard renders static pages with strict headers and no active scripts by default (no cookies, no third-party scripts).
- Recommended: Run behind HTTPS (TLS) and a standard reverse proxy; restrict external network egress except for the optional Google Sheets CSV endpoint if used.
- Keep binaries up to date; verify checksums of releases; restrict filesystem access to the service account; limit write access to the cache directory only.

## 6) Subprocessors and International Transfers
- The project maintainers do not operate as a processor and use no subprocessors.
- If the district enables the optional Google Sheets CSV feature, Google LLC is the publisher of the CSV content the district chose to host with Google. The district remains controller for that policy data and should review Google’s terms and data location policies. Disable the feature if such transfer is not desired.

## 7) Data Subject Rights (Access/Correction/Deletion)
- SB29-guard does not store personal data. If any policy data maintained by the district is deemed personal data, the district as controller is responsible for fulfilling requests. SB29-guard has no backend database to export or purge user-specific records.

## 8) Incident Response and Breach Notification
- As SB29-guard does not persist user data, the primary risk surface is limited to policy files and optional CSV cache.
- Districts should apply their standard security incident procedures (log review, cache cleanup, policy file integrity checks, application update). There is no project-operated service to notify.

## 9) Compliance Notes (FERPA/COPPA/GDPR)
- Intended use avoids student PII; domains and policy reasons should be non-personal.
- Districts should avoid including student identifiers in policy rationale or reference fields.
- If your legal framework classifies connection metadata as personal data, ensure reverse proxy logs are configured and retained per policy.

## 10) Contact
- District Contact: [Insert district privacy office contact]
- Project Contact: Use the project’s issue tracker for code-related questions; maintainers do not receive or process deployment data.

## 11) Term and Termination
- This DPA applies while the district runs SB29-guard. Upon decommission, remove the application and delete local policy and cache files per district policy.

## 12) Execution
This document describes the privacy posture of SB29-guard when deployed on-premises by a district. Districts may incorporate this text into their institutional DPA templates.

---
Not legal advice. Adapt to your jurisdiction and institutional policies.
