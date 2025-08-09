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
GET /explain?domain=...
- Renders HTML page.

## 3. Metrics (JSON)
GET /metrics
```
{
  "policy_version": "0.1.0",
  "record_count": 245,
  "last_refresh_time": "2025-08-08T12:00:00Z",
  "last_refresh_source": "csv|csv-cache",
  "refresh_count": 3,
  "refresh_error_count": 0,
  "last_refresh_error": ""
}
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

## 4. Error Responses (General)
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
