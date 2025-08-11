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
GET /explain
- Renders HTML page.

Query Parameters (display-only; strict validation; headers remain authoritative):
- `d` original domain (hostname only)
- `c` classification key (optional)
- `v` policy version (optional)
- `h` policy hash short (optional)

Header precedence for resolving the original domain (first match wins):
1. X-Original-Host
2. X-Forwarded-Host (first value)
3. Referer (host portion)
4. Host (only if SB29_ALLOW_HOST_FALLBACK=true)

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

## 4. Law Redirect
GET /law
- 302 redirect to configured law URL (default: LIS PDF for SB29). Target can be overridden via `SB29_LAW_URL` environment variable.

## 5. Aggregated Usage (Admin) [Optional]
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
Explanation page is text/html.

## 9. Rate Limiting (Future)
Optional basic IP rate limiting for admin endpoints only.

## 10. Versioning
- Policy version supplied in all responses.
- API structural changes increase a service `api_version` header (e.g., `X-SB29Guard-API: 1`).

## 11. Localization
Optional `Accept-Language` influences localized text blocks if available; fallback to default locale.

## 12. OpenAPI (Planned)
An OpenAPI 3.1 spec will be generated at build time and exposed at `/api/openapi.json`.
