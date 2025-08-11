# GUI proxy integrations (lists and API)

Some school web filters and GUI-driven proxies prefer an API call or a plain domain list instead of custom config. SB29-guard exposes both to keep setup simple.

Endpoints (served by sb29-guard)
- /classify (GET): JSON lookup
  Request: /classify?d=<domain>
  Response: { "found": bool, "classification": string, "policy_version": string }
- /domain-list (GET): plaintext list
  - Each line is a domain; wildcards appear as base and .base for easy matching.

Behavior
- Denylist model: if a domain is not present, it’s treated as allowed. The server returns 404 Not Classified for /explain, and {found:false} for /classify.
- Normalization: inputs are sanitized (hostnames only, lowercased). Wildcards (*.example.com) are represented by example.com and .example.com in /domain-list.

Usage examples
- Block page via redirect template (most GUI filters):
  https://explain.school.example/explain?d=%ORIGINAL_DOMAIN%
- Pre-check via JSON (some proxies allow a decision hook):
  GET https://guard.school.internal/classify?d=exampletool.com
  - If found=true, redirect the user to your explain page.
- Ingest list (scheduled):
  GET https://guard.school.internal/domain-list > blocked.txt
  - Load or import blocked.txt into your proxy’s custom denylist.

Tips
- Keep TLS trusted for your internal host so clients never see warnings.
- Cache results conservatively and refresh on policy updates. The JSON response includes policy_version to help invalidations.
- For a static explain page, generate it with: sb29guard generate-explain-static --out-dir dist/explain
