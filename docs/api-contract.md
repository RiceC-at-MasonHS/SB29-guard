# API Contract (Redirect Service)

Status: Draft

Base URL: `https://guard.school.local` (example)

## 1. Health
GET /health
Response 200:
```
{"status":"ok","policy_version":"0.1.0"}
```

## 2. Explanation Page (HTML)
GET /explain?original_domain=...&classification=...&policy_version=...&ts=...&ref=sb29guard
- Renders HTML page.
- If params missing, server attempts host header lookup.

## 3. Domain Info API
GET /api/domain-info?domain=<fqdn>
Responses:
200 OK
```
{
  "domain": "exampletool.com",
  "matched_record_domain": "exampletool.com",
  "classification": "NO_DPA",
  "rationale": "Vendor has not signed district-approved Digital Privacy Agreement.",
  "last_review": "2025-08-01",
  "status": "active",
  "policy_version": "0.1.0"
}
```
404 Not Found
```
{"error":"not_found","domain":"unknown.com"}
```

## 4. Aggregated Usage (Admin) [Optional]
GET /admin/summary?date=YYYY-MM-DD (auth required if enabled)
```
{
  "date": "2025-08-08",
  "policy_version": "0.1.0",
  "entries": [
    {"domain":"exampletool.com","classification":"NO_DPA","count":42}
  ]
}
```

## 5. Static Assets
GET /static/* served from public directory (immutable hash filenames recommended).

## 6. Error Responses (General)
```
{"error":"invalid_parameter","detail":"classification missing"}
```

## 7. Security Headers
Applied to all HTML/JSON:
- Content-Security-Policy
- Referrer-Policy: no-referrer
- X-Content-Type-Options: nosniff
- X-Frame-Options: DENY
- Cache-Control: no-store

## 8. Content Negotiation
- `Accept: application/json` forces JSON for domain-info.
- Explanation page always text/html.

## 9. Rate Limiting (Future)
Optional basic IP rate limiting for admin endpoints only.

## 10. Versioning
- Policy version supplied in all responses.
- API structural changes increase a service `api_version` header (e.g., `X-SB29Guard-API: 1`).

## 11. Localization
Optional `Accept-Language` influences localized text blocks if available; fallback to default locale.

## 12. OpenAPI (Planned)
An OpenAPI 3.1 spec will be generated at build time and exposed at `/api/openapi.json`.
