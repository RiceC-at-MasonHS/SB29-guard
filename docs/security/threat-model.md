# Threat Model (Proxy-first)

Assets
- Policy classification data; explain endpoint integrity; user privacy.

Trust boundaries
- Proxy/gateway to app; external static hosting (if used); user browser.

Threats & mitigations
- Param spoofing (d/c/v/h): display-only, strict validation; headers/policy authoritative.
- Header forgery: trust only from known proxy; recommend network placement and ACLs.
- XSS/Injection: no scripts; template escapes rationale/source; strict CSP.
- Leakage via referer: Referrer-Policy: no-referrer; no third-party calls.
- Caching stale info: Cache-Control: no-store.

Operational guidance
- Run app behind trusted proxy or on same host; block direct Internet access to backend port.
- Rotate static hosting URLs if accidentally indexed; add robots noindex if desired.
